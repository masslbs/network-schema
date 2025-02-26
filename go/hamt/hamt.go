// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

// This file implements a Hash Array Mapped Trie (HAMT), a persistent
// data structure for efficiently storing key-value pairs.
// This implementation supports CBOR serialization and uses SHA-256 for hashing.

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"math/bits"
	"slices"

	masscbor "github.com/masslbs/network-schema/go/cbor"
)

const (
	// bitsPerStep determines how many bits from the hash are used at each level of the trie
	bitsPerStep = 6
	// maxDepth is the maximum depth of the trie based on the hash size and bits per step
	maxDepth = 256 / bitsPerStep
)

var (
	// newHAMTHashFn provides the hash function used by the HAMT (SHA-256 by default)
	newHAMTHashFn = sha256.New
	// nilHAMTHash is the hash of an empty HAMT
	nilHAMTHash []byte
)

func init() {
	nilHAMTHash = newHAMTHashFn().Sum(nil)
}

// hashState tracks the consumption of bits from a key's hash to navigate the trie.
// It stores the full 256-bit (32-byte) hash and tracks how many bits have been consumed.
type hashState struct {
	originalKey []byte
	hashBuf     [32]byte // The full SHA-256 hash of the key
	consumed    uint16   // Number of bits consumed from hashBuf
}

// newHashState creates a new hashState from a key by computing its SHA-256 hash.
func newHashState(key []byte) *hashState {
	var hs hashState
	hs.originalKey = key

	h := newHAMTHashFn()
	h.Write(key)
	copy(hs.hashBuf[:], h.Sum(nil))

	return &hs
}

// next returns the next chunk of bits from the hash to be used for trie navigation.
// Each call consumes bitsPerStep bits from the hash.
func (hs *hashState) next() uint32 {
	// Calculate our current offset in bits
	bitOffset := hs.consumed
	hs.consumed += bitsPerStep

	// Identify which byte(s) the bits cross
	byteIndex := bitOffset / 8
	bitInByte := bitOffset % 8

	// Gather up to 16 bits from two adjacent bytes
	var raw uint16
	raw = uint16(hs.hashBuf[byteIndex]) << 8
	if byteIndex+1 < 32 {
		raw |= uint16(hs.hashBuf[byteIndex+1])
	}

	// Extract the desired bits
	shift := 16 - bitsPerStep - bitInByte
	mask := uint16((1 << bitsPerStep) - 1)
	chunk := (raw >> shift) & mask

	return uint32(chunk)
}

// Trie represents a persistent Hash Array Mapped Trie (HAMT) data structure.
// It provides efficient key-value storage with immutable operations that
// share structure between versions.
type Trie[V any] struct {
	root *Node[V]
	size int
}

// Node is an internal node in the HAMT. It contains a bitmap indicating which
// positions are occupied and a slice of entries. The bitmap enables efficient
// sparse representation of child nodes.
type Node[V any] struct {
	_       struct{} `cbor:",toarray"`
	Bitmap  uint64   // Bitmap indicating which positions contain entries
	Entries []Entry[V]

	// hash caches the computed hash of this node (not serialized)
	hash []byte
}

// Entry represents either a leaf node (with Key-Value and Node=nil) or
// a branch node (with Node!=nil and Key/Value empty).
type Entry[V any] struct {
	_     struct{} `cbor:",toarray"`
	Key   []byte
	Value V
	Node  *Node[V]
}

// NewTrie creates a new empty HAMT.
func NewTrie[V any]() *Trie[V] {
	return &Trie[V]{
		root: &Node[V]{},
		size: 0,
	}
}

// MarshalCBOR marshals the HAMT into a CBOR encoded byte slice using the
// canonical encoding options defined in the schema package.
func (t *Trie[V]) MarshalCBOR() ([]byte, error) {
	return masscbor.Marshal(t.root)
}

// UnmarshalCBOR unmarshals a CBOR encoded byte slice into a HAMT and
// recalculates the size by counting all leaf entries.
func (t *Trie[V]) UnmarshalCBOR(data []byte) error {
	if err := masscbor.Unmarshal(data, &t.root); err != nil {
		return err
	}

	// Recalculate size by counting all leaf entries
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

// Insert adds or updates a key-value pair in the HAMT.
// Returns an error if the operation fails.
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

// insert adds a key-value pair to a node, returning true if a new entry was created
// or false if an existing entry was updated.
func (n *Node[V]) insert(key []byte, value V, hs *hashState) (bool, error) {
	var nilValue V
	if hs.consumed >= maxDepth*bitsPerStep {
		return n.insertFallback(key, value)
	}

	// Determine which slot this key should occupy based on the next hash bits
	idx := hs.next()
	if idx >= 64 {
		return false, fmt.Errorf("idx out of range: %d", idx)
	}

	// Find position in the sparse entries array
	pos := bits.OnesCount64(n.Bitmap & ((1 << idx) - 1))

	// If the position is not occupied, insert a new entry
	if n.Bitmap&(1<<idx) == 0 {
		n.Bitmap |= (1 << idx) // Set bit for this position
		n.Entries = append(n.Entries, Entry[V]{})
		// Make space for the new entry at the correct position
		if pos < len(n.Entries)-1 {
			copy(n.Entries[pos+1:], n.Entries[pos:])
		}
		n.Entries[pos] = Entry[V]{Key: key, Value: value}
		// Invalidate cached hash
		n.hash = nil
		return true, nil
	}

	// Position is already occupied
	entry := &n.Entries[pos]
	if entry.Node == nil {
		// Direct entry - check if it's the same key
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
			// Update existing value
			entry.Value = value
			// invalidate the cached hash since we are mutating the node
			n.hash = nil
			return false, nil
		}

		// Hash collision - convert leaf to branch
		branch := &Node[V]{}

		// Insert existing entry into the new branch
		oldHS := newHashState(entry.Key)
		oldHS.consumed = hs.consumed
		_, err := branch.insert(entry.Key, entry.Value, oldHS)
		if err != nil {
			return false, err
		}

		// Insert new entry into the new branch
		newHS := newHashState(key)
		newHS.consumed = hs.consumed
		_, err = branch.insert(key, value, newHS)
		if err != nil {
			return false, err
		}

		// Replace leaf with branch node
		entry.Node = branch
		entry.Key = nil
		entry.Value = nilValue

		// invalidate the cached hash since we are mutating the node
		n.hash = nil
		return true, nil
	}

	// Recursively insert into branch node
	inserted, err := entry.Node.insert(key, value, hs)
	if inserted {
		// invalidate the cached hash since we are mutating the node
		n.hash = nil
	}
	return inserted, err
}

// insertFallback handles insertion when the maximum depth is reached by
// performing a linear search through entries.
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

// Get retrieves the value associated with a key from the HAMT.
// Returns the value and a boolean indicating whether the key was found.
func (t *Trie[V]) Get(key []byte) (V, bool) {
	if t.root == nil {
		var nilValue V
		return nilValue, false
	}
	return t.root.Find(key)
}

// Find searches for a key in the trie and returns its value if found.
// Uses an iterative approach to avoid stack overflow with deep tries.
func (n *Node[V]) Find(key []byte) (V, bool) {
	var nilValue V
	hs := newHashState(key)
	currentNode := n

	for {
		if currentNode == nil {
			return nilValue, false
		}

		// Check if we've reached maximum depth
		if hs.consumed >= maxDepth*bitsPerStep {
			return currentNode.findFallback(key)
		}

		// Determine which slot this key should occupy
		idx := hs.next()
		if currentNode.Bitmap&(1<<idx) == 0 {
			return nilValue, false
		}

		// Find position in the sparse entries array
		pos := bits.OnesCount64(currentNode.Bitmap & ((1 << idx) - 1))
		if pos >= len(currentNode.Entries) {
			return nilValue, false
		}
		entry := &currentNode.Entries[pos]

		// Check if this is a leaf node with our key
		if entry.Node == nil {
			if bytes.Equal(entry.Key, key) {
				return entry.Value, true
			}
			return nilValue, false
		}

		// Continue traversal with the branch node
		currentNode = entry.Node
	}
}

// findFallback handles key lookup when the maximum depth is reached
// by performing a linear search through entries.
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

// Delete removes a key from the HAMT.
// Returns an error if the operation fails.
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

// delete removes a key from a node, returning true if a key was removed
// and false if the key was not found.
func (n *Node[V]) delete(key []byte, hs *hashState) (bool, error) {
	// Check if we've reached maximum depth
	if hs.consumed >= maxDepth*bitsPerStep {
		return n.deleteFallback(key)
	}

	// Determine which slot this key should occupy
	idx := hs.next()
	if n.Bitmap&(1<<idx) == 0 {
		return false, nil
	}

	pos := bits.OnesCount64(n.Bitmap & ((1 << idx) - 1))
	if pos >= len(n.Entries) {
		return false, fmt.Errorf("pos for idx %d out of range: %d", idx, pos)
	}

	entry := &n.Entries[pos]

	// If leaf node, check if key matches
	if entry.Node == nil {
		if !bytes.Equal(entry.Key, key) {
			return false, nil
		}

		// invalidate hash since we are mutating the node
		n.hash = nil
		n.Bitmap &= ^(1 << idx)
		n.Entries = slices.Delete(n.Entries, pos, pos+1)
		return true, nil
	}

	// Handle branch node deletion
	deleted, err := entry.Node.delete(key, hs)
	if err != nil {
		return false, err
	}
	if !deleted {
		return false, nil
	}

	// Clean up branch nodes
	if len(entry.Node.Entries) == 0 {
		// Remove empty branch
		n.Bitmap &= ^(1 << idx)
		n.Entries = slices.Delete(n.Entries, pos, pos+1)
	} else if len(entry.Node.Entries) == 1 && entry.Node.Entries[0].Node == nil {
		// Collapse branch with single entry
		childEntry := entry.Node.Entries[0]
		entry.Key = childEntry.Key
		entry.Value = childEntry.Value
		entry.Node = nil
	}

	// invalidate hash since we are mutating the node
	n.hash = nil
	return true, nil
}

// deleteFallback handles deletion when the maximum depth is reached
// by performing a linear search through entries.
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

// Size returns the number of key-value pairs in the HAMT.
func (t *Trie[V]) Size() int {
	return t.size
}

// Hash computes the cryptographic hash of the entire HAMT.
// This is useful for comparing tries or verifying integrity.
func (t *Trie[V]) Hash() ([]byte, error) {
	if t.root == nil {
		return nilHAMTHash, nil
	}
	return t.root.Hash()
}

// Hash computes the cryptographic hash of a node and its children.
// Uses caching to avoid recomputing unchanged subtrees.
func (n *Node[V]) Hash() ([]byte, error) {
	h := newHAMTHashFn()
	if n == nil {
		return h.Sum(nil), nil
	}

	// Return cached hash if available
	if n.hash != nil {
		return n.hash, nil
	}

	// Hash entries in order of presence in the node
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

// encodeValue serializes a value to a writer using the default CBOR encoder.
func (n *Node[V]) encodeValue(v V, w io.Writer) error {
	enc := masscbor.DefaultEncoder(w)
	return enc.Encode(v)
}

// All iterates over all key-value pairs in the trie and calls the provided
// function for each pair. The iteration stops if the function returns false.
func (t *Trie[V]) All(fn func([]byte, V) bool) {
	if t.root == nil {
		return
	}
	t.root.all(fn)
}

// all recursively visits all entries in a node and its children,
// calling the provided function for each key-value pair.
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
