// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"fmt"
	"io"
	"math/bits"
	"slices"

	"github.com/cespare/xxhash/v2"
)

const (
	bitsPerStep = 5
	maxDepth    = (64 + bitsPerStep - 1) / bitsPerStep
)

var (
	newHAMTHashFn = xxhash.New
	nilHAMTHash   []byte
)

func init() {
	nilHAMTHash = newHAMTHashFn().Sum(nil)
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
	h := newHAMTHashFn()
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
type Trie[V any] struct {
	root *Node[V]
	size int
}

// Node is a node in the HAMT
type Node[V any] struct {
	_       struct{} `cbor:",toarray"`
	Bitmap  uint32
	Entries []Entry[V]

	// not serialized
	hash []byte
}

// Entry is either a leaf node (Node == nil) or a branch node (Node != nil)
type Entry[V any] struct {
	_     struct{} `cbor:",toarray"`
	Key   []byte
	Value V
	Node  *Node[V]
}

// NewTrie creates a new empty HAMT
func NewTrie[V any]() *Trie[V] {
	return &Trie[V]{
		root: &Node[V]{},
		size: 0,
	}
}

// MarshalCBOR marshals the HAMT into a CBOR encoded byte slice
func (t *Trie[V]) MarshalCBOR() ([]byte, error) {
	// use the schema package to inherit the canonical encoding options
	return Marshal(t.root)
}

// UnmarshalCBOR unmarshals a CBOR encoded byte slice into a HAMT
func (t *Trie[V]) UnmarshalCBOR(data []byte) error {
	if err := Unmarshal(data, &t.root); err != nil {
		return err
	}

	// Recalculate size
	size := 0
	var countEntries func(*Node[V])
	countEntries = func(n *Node[V]) {
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
func (t *Trie[V]) Insert(key []byte, value V) error {
	if t.root == nil {
		t.root = &Node[V]{}
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

func (n *Node[V]) insert(key []byte, value V, hs *hashState) (bool, error) {
	var nilValue V
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
		n.Entries = append(n.Entries, Entry[V]{})
		// shift entries to the right to make space for the new entry
		if pos < len(n.Entries)-1 {
			copy(n.Entries[pos+1:], n.Entries[pos:])
		}
		n.Entries[pos] = Entry[V]{Key: key, Value: value}
		// invalidate the cached hash since we are mutating the node
		n.hash = nil
		return true, nil
	}

	// if the position is occupied, check if it's a branch node
	entry := &n.Entries[pos]
	if entry.Node == nil {
		// if it's not a branch node, check if the key is already in the entry
		if bytes.Equal(entry.Key, key) {
			var currentHash []byte
			if n.hash != nil {
				currentHash = n.hash
			} else {
				var h1 = newHAMTHashFn()
				err := n.encodeValue(entry.Value, h1)
				if err != nil {
					return false, err
				}
				currentHash = h1.Sum(nil)
			}
			var h2 = newHAMTHashFn()
			err := n.encodeValue(value, h2)
			if err != nil {
				return false, err
			}
			newHash := h2.Sum(nil)
			if bytes.Equal(currentHash, newHash) {
				return false, nil
			}
			// update the value directly
			entry.Value = value
			// invalidate the cached hash since we are mutating the node
			n.hash = nil
			return false, nil
		}

		// not a branch node yet, so create a new branch node
		branch := &Node[V]{}

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
		entry.Value = nilValue

		// invalidate the cached hash since we are mutating the node
		n.hash = nil
		return true, nil
	}

	// it is a branch node, so recursively insert the new entry
	inserted, err := entry.Node.insert(key, value, hs)
	if inserted {
		// invalidate the cached hash since we are mutating the node
		n.hash = nil
	}
	return inserted, err
}

// insertFallback is used when the hash state has consumed all bits
func (n *Node[V]) insertFallback(key []byte, value V) (bool, error) {
	for i, e := range n.Entries {
		if bytes.Equal(e.Key, key) {
			n.Entries[i].Value = value
			return false, nil
		}
	}

	// if the key is not found, append it to the entries
	n.Entries = append(n.Entries, Entry[V]{Key: key, Value: value})
	// invalidate the cached hash since we are mutating the node
	n.hash = nil
	return true, nil
}

// Get retrieves the value associated with a key from the HAMT
func (t *Trie[V]) Get(key []byte) (V, bool) {
	if t.root == nil {
		var nilValue V
		return nilValue, false
	}
	return t.root.Find(key)
}

// Find finds a given entry in the trie
func (n *Node[V]) Find(key []byte) (V, bool) {
	var nilValue V
	hs := newHashState(key)
	currentNode := n

	// iterate instead of recursion calls to avoid stack overflow
	for {
		if currentNode == nil {
			return nilValue, false
		}

		// are we at max depth?
		if hs.consumed >= maxDepth*bitsPerStep {
			return currentNode.findFallback(key)
		}

		// figure out which position in the entries array to look for the key
		idx := hs.next()
		if currentNode.Bitmap&(1<<idx) == 0 {
			return nilValue, false
		}

		// figure out which position in the entries array the key is at
		pos := bits.OnesCount32(currentNode.Bitmap & ((1 << idx) - 1))
		if pos >= len(currentNode.Entries) { // invalid position
			return nilValue, false
		}
		entry := &currentNode.Entries[pos]

		// if the entry is not a branch node, check if the key is in the entry
		if entry.Node == nil {
			if bytes.Equal(entry.Key, key) {
				return entry.Value, true
			}
			return nilValue, false
		}

		// if the entry is a branch node, recursively search the branch
		currentNode = entry.Node
	}
}

// findFallback is used when the hash state has consumed all bits
func (n *Node[V]) findFallback(key []byte) (V, bool) {
	for _, e := range n.Entries {
		if e.Node == nil && bytes.Equal(e.Key, key) {
			return e.Value, true
		} else if e.Node != nil {
			if val, found := e.Node.findFallback(key); found {
				return val, true
			}
		}
	}
	var nilValue V
	return nilValue, false
}

// Delete deletes a key from the HAMT
func (t *Trie[V]) Delete(key []byte) error {
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

func (n *Node[V]) delete(key []byte, hs *hashState) (bool, error) {
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
	if pos >= len(n.Entries) {
		return false, fmt.Errorf("pos for idx %d out of range: %d", idx, pos)
	}

	entry := &n.Entries[pos]

	// if the entry is not a branch node, check if the key is in the entry
	if entry.Node == nil {
		if !bytes.Equal(entry.Key, key) {
			return false, nil
		}

		// invalidate hash since we are mutating the node
		n.hash = nil

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

	// invalidate hash since we are mutating the node
	n.hash = nil

	return true, nil
}

func (n *Node[V]) deleteFallback(key []byte) (bool, error) {
	for i, e := range n.Entries {
		if e.Node == nil && bytes.Equal(e.Key, key) {
			n.Entries = slices.Delete(n.Entries, i, i+1)
			// invalidate the cached hash since we are mutating the node
			n.hash = nil
			return true, nil
		}
	}
	return false, nil
}

// Size returns the number of entries in the HAMT
func (t *Trie[V]) Size() int {
	return t.size
}

// Hash hashes the whole HAMT
func (t *Trie[V]) Hash() ([]byte, error) {
	if t.root == nil {
		return nilHAMTHash, nil
	}
	return t.root.Hash()
}

func (n *Node[V]) Hash() ([]byte, error) {
	h := newHAMTHashFn()
	if n == nil {
		return h.Sum(nil), nil
	}

	if n.hash != nil {
		return n.hash, nil
	}

	// Hash entries in order
	for _, e := range n.Entries {
		if e.Node == nil {
			h.Write(e.Key)
			err := n.encodeValue(e.Value, h)
			if err != nil {
				return nil, err
			}
		} else {
			hash, err := e.Node.Hash()
			if err != nil {
				return nil, err
			}
			h.Write(hash)
		}
	}

	n.hash = h.Sum(nil)
	return n.hash, nil
}

func (n *Node[V]) encodeValue(v V, w io.Writer) error {
	enc := DefaultEncoder(w)
	return enc.Encode(v)
}

// All iterates over all key-value pairs in the trie
// The iteration stops if fn returns false.
func (t *Trie[V]) All(fn func([]byte, V) bool) {
	if t.root == nil {
		return
	}
	t.root.all(fn)
}

func (n *Node[V]) all(fn func([]byte, V) bool) bool {
	if n == nil {
		return true
	}
	for _, e := range n.Entries {
		if e.Node == nil {
			if !fn(e.Key, e.Value) {
				return false
			}
		} else {
			if !e.Node.all(fn) {
				return false
			}
		}
	}
	return true
}
