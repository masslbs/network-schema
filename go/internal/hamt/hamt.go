package hamt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/bits"
	"slices"

	"github.com/cespare/xxhash/v2"
	schema "github.com/masslbs/network-schema/go"
)

const (
	bitsPerStep = 5
	maxDepth    = (64 + bitsPerStep - 1) / bitsPerStep
)

var (
	newHashFn = xxhash.New
	nilHash   []byte
)

func init() {
	nilHash = newHashFn().Sum(nil)
}

type hashState struct {
	originalKey []byte
	hash        uint64
	consumed    uint8
	seed        uint64
}

func newHashState(key []byte) *hashState {
	return &hashState{
		originalKey: key,
		hash:        hashKeyWithSeed(key, 0),
		consumed:    0,
		seed:        0,
	}
}

// next returns the next bitsPerStep bits from the hash state
func (hs *hashState) next() uint32 {
	if hs.consumed+bitsPerStep > maxDepth*bitsPerStep {
		hs.seed++
		hs.hash = hashKeyWithSeed(hs.originalKey, hs.seed)
		hs.consumed = 0
	}

	shift := hs.consumed
	mask := uint64((1 << bitsPerStep) - 1)
	chunk := (hs.hash >> shift) & mask
	hs.consumed += bitsPerStep
	return uint32(chunk)
}

var (
	// variables for test stubbing of hash collisions
	hashKeyWithSeed = hashKeyWithSeedFunc
	hashKey         = hashKeyFunc
)

// hashKey is used to hash keys
func hashKeyFunc(key []byte) uint64 {
	h := newHashFn()
	h.Write(key)
	return h.Sum64()
}

// hashKeyWithSeed is used when hashState consumed all bits
func hashKeyWithSeedFunc(key []byte, seed uint64) uint64 {
	// TODO: this uses the seed variation not newHashFn
	h := xxhash.NewWithSeed(seed)
	h.Write(key)
	return h.Sum64()
}

// Trie represents a persistent HAMT
type Trie struct {
	root *Node
	size int
}

// Node is a node in the HAMT
type Node struct {
	_       struct{} `cbor:",toarray"`
	Bitmap  uint32
	Entries []Entry
}

// Entry is either a leaf node (Node == nil) or a branch node (Node != nil)
type Entry struct {
	_     struct{} `cbor:",toarray"`
	Key   []byte
	Value []byte
	Node  *Node
}

// NewTrie creates a new empty HAMT
func NewTrie() *Trie {
	return &Trie{
		root: &Node{},
		size: 0,
	}
}

// MarshalCBOR marshals the HAMT into a CBOR encoded byte slice
func (t *Trie) MarshalCBOR() ([]byte, error) {
	// use the schema package to inherit the canonical encoding options
	return schema.Marshal(t.root)
}

// UnmarshalCBOR unmarshals a CBOR encoded byte slice into a HAMT
func (t *Trie) UnmarshalCBOR(data []byte) error {
	if err := schema.Unmarshal(data, &t.root); err != nil {
		return err
	}

	// Recalculate size
	size := 0
	var countEntries func(*Node)
	countEntries = func(n *Node) {
		if n == nil {
			return
		}
		for _, e := range n.Entries {
			if e.Node == nil {
				size++ // Count direct entries
				continue
			}
			countEntries(e.Node)
		}
	}
	countEntries(t.root)
	t.size = size

	return nil
}

// Insert inserts a key-value pair into the HAMT
func (t *Trie) Insert(key []byte, value []byte) error {
	if t.root == nil {
		t.root = &Node{}
	}

	inserted, err := t.root.insert(key, value, newHashState(key))
	if err != nil {
		return err
	}

	if inserted {
		t.size++
	}
	return nil
}

func (n *Node) insert(key []byte, value []byte, hs *hashState) (bool, error) {
	if hs.consumed >= maxDepth*bitsPerStep {
		return n.insertFallback(key, value)
	}

	// figure out which position in the entries array to insert the new entry
	idx := hs.next()
	if idx >= 32 {
		return false, fmt.Errorf("idx out of range: %d", idx)
	}
	pos := bits.OnesCount32(n.Bitmap & ((1 << idx) - 1))

	// if the position is not occupied, insert the new entry
	if n.Bitmap&(1<<idx) == 0 {
		n.Bitmap |= (1 << idx) // flip bit to 1
		n.Entries = append(n.Entries, Entry{})
		// shift entries to the right to make space for the new entry
		if pos < len(n.Entries)-1 {
			copy(n.Entries[pos+1:], n.Entries[pos:])
		}
		n.Entries[pos] = Entry{Key: key, Value: value}
		return true, nil
	}

	// if the position is occupied, check if it's a branch node
	entry := &n.Entries[pos]
	if entry.Node == nil {
		// if it's not a branch node, check if the key is already in the entry
		if bytes.Equal(entry.Key, key) {
			if bytes.Equal(entry.Value, value) {
				return false, nil
			}
			// update the value directly
			entry.Value = value
			return false, nil
		}

		// not a branch node yet, so create a new branch node
		branch := &Node{}

		// create a new hash state for the existing key
		oldHS := &hashState{
			originalKey: entry.Key,
			hash:        hashKey(entry.Key),
			consumed:    hs.consumed,
			seed:        hs.seed,
		}
		_, err := branch.insert(entry.Key, entry.Value, oldHS)
		if err != nil {
			return false, err
		}

		// create a new hash state for the new key
		newHS := &hashState{
			originalKey: key,
			hash:        hashKey(key),
			consumed:    hs.consumed,
			seed:        hs.seed,
		}
		_, err = branch.insert(key, value, newHS)
		if err != nil {
			return false, err
		}

		// update the current node
		entry.Node = branch
		entry.Key = nil
		entry.Value = nil
		return true, nil
	}

	// it is a branch node, so recursively insert the new entry
	return entry.Node.insert(key, value, hs)
}

// insertFallback is used when the hash state has consumed all bits
func (n *Node) insertFallback(key []byte, value []byte) (bool, error) {
	for i, e := range n.Entries {
		if bytes.Equal(e.Key, key) {
			n.Entries[i].Value = value
			return false, nil
		}
	}

	// if the key is not found, append it to the entries
	n.Entries = append(n.Entries, Entry{Key: key, Value: value})
	return true, nil
}

// Get retrieves the value associated with a key from the HAMT
func (t *Trie) Get(key []byte) ([]byte, bool) {
	if t.root == nil {
		return nil, false
	}
	return t.root.Find(key)
}

// Find finds a given entry in the trie
func (n *Node) Find(key []byte) ([]byte, bool) {
	hs := newHashState(key)
	currentNode := n

	// iterate instead of recursion calls to avoid stack overflow
	for {
		if currentNode == nil {
			return nil, false
		}

		// are we at max depth?
		if hs.consumed >= maxDepth*bitsPerStep {
			return currentNode.findFallback(key)
		}

		// figure out which position in the entries array to look for the key
		idx := hs.next()
		if currentNode.Bitmap&(1<<idx) == 0 {
			return nil, false
		}

		// figure out which position in the entries array the key is at
		pos := bits.OnesCount32(currentNode.Bitmap & ((1 << idx) - 1))
		if pos >= len(currentNode.Entries) { // invalid position
			return nil, false
		}
		entry := &currentNode.Entries[pos]

		// if the entry is not a branch node, check if the key is in the entry
		if entry.Node == nil {
			if bytes.Equal(entry.Key, key) {
				return entry.Value, true
			}
			return nil, false
		}

		// if the entry is a branch node, recursively search the branch
		currentNode = entry.Node
	}
}

// findFallback is used when the hash state has consumed all bits
func (n *Node) findFallback(key []byte) ([]byte, bool) {
	for _, e := range n.Entries {
		if e.Node == nil && bytes.Equal(e.Key, key) {
			return e.Value, true
		} else if e.Node != nil {
			if val, found := e.Node.findFallback(key); found {
				return val, true
			}
		}
	}
	return nil, false
}

// Size returns the number of entries in the HAMT
func (t *Trie) Size() int {
	return t.size
}

// Hash hashes the whole HAMT
func (t *Trie) Hash() []byte {
	if t.root == nil {
		return nilHash
	}
	return t.root.Hash()
}

func (n *Node) Hash() []byte {
	h := newHashFn()
	if n == nil {
		return h.Sum(nil)
	}

	// Hash bitmap
	var bitmapBytes [4]byte
	binary.BigEndian.PutUint32(bitmapBytes[:], n.Bitmap)
	h.Write(bitmapBytes[:])

	// Hash entries in order
	for _, e := range n.Entries {
		if e.Node == nil {
			h.Write(e.Key)
			h.Write(e.Value)
		} else {
			h.Write(e.Node.Hash())
		}
	}

	return h.Sum(nil)
}

// Delete deletes a key from the HAMT
func (t *Trie) Delete(key []byte) error {
	if t.root == nil {
		return nil
	}
	deleted, err := t.root.delete(key, newHashState(key))
	if err != nil {
		return err
	}
	if deleted {
		t.size--
	}
	return nil
}

func (n *Node) delete(key []byte, hs *hashState) (bool, error) {
	// are we at max depth?
	if hs.consumed >= maxDepth*bitsPerStep {
		return n.deleteFallback(key)
	}

	// figure out which position in the entries array the key is at
	idx := hs.next()
	if n.Bitmap&(1<<idx) == 0 {
		return false, nil
	}

	pos := bits.OnesCount32(n.Bitmap & ((1 << idx) - 1))
	entry := &n.Entries[pos]

	// if the entry is not a branch node, check if the key is in the entry
	if entry.Node == nil {
		if !bytes.Equal(entry.Key, key) {
			return false, nil
		}

		// Remove entry in place
		n.Bitmap &= ^(1 << idx)
		n.Entries = slices.Delete(n.Entries, pos, pos+1)
		return true, nil
	}

	// Handle branch node
	deleted, err := entry.Node.delete(key, hs)
	if err != nil {
		return false, err
	}
	if !deleted {
		return false, nil
	}

	// If child is empty, remove it
	if len(entry.Node.Entries) == 0 {
		n.Bitmap &= ^(1 << idx)
		n.Entries = slices.Delete(n.Entries, pos, pos+1)
	} else if len(entry.Node.Entries) == 1 && entry.Node.Entries[0].Node == nil {
		// If child has only one entry, collapse it
		childEntry := entry.Node.Entries[0]
		entry.Key = childEntry.Key
		entry.Value = childEntry.Value
		entry.Node = nil
	}

	return true, nil
}

func (n *Node) deleteFallback(key []byte) (bool, error) {
	for i, e := range n.Entries {
		if e.Node == nil && bytes.Equal(e.Key, key) {
			n.Entries = slices.Delete(n.Entries, i, i+1)
			return true, nil
		}
	}
	return false, nil
}
