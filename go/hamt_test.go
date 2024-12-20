// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

func TestHAMT(t *testing.T) {
	r := require.New(t)
	trie := NewTrie[string]()

	// Insert some values
	err := trie.Insert([]byte("name"), "Alice")
	r.NoError(err)

	trie1, err := copyTrie[string](trie)
	r.NoError(err)

	err = trie1.Insert([]byte("age"), "Bob")
	r.NoError(err)
	r.Equal(2, trie1.Size())

	// Verify insertions
	val, ok := trie1.Get([]byte("name"))
	r.True(ok)
	r.Equal("Alice", val)

	val, ok = trie1.Get([]byte("age"))
	r.True(ok)
	r.Equal("Bob", val)

	// Original trie should be unchanged
	val, ok = trie.Get([]byte("age"))
	r.False(ok)
	r.Zero(val)
}

func TestHAMTComplexOperations(t *testing.T) {
	r := require.New(t)

	var values = []cbor.RawMessage{
		mustEncode(t, 0),
		mustEncode(t, 1),
		mustEncode(t, 2),
		mustEncode(t, 3),
		mustEncode(t, 4),
	}

	var newValues = []cbor.RawMessage{
		mustEncode(t, "new-0"),
		mustEncode(t, "new-1"),
		mustEncode(t, "new-2"),
		mustEncode(t, "new-3"),
		mustEncode(t, "new-4"),
	}

	// Create initial trie with multiple values
	trie := NewTrie[cbor.RawMessage]()
	err := trie.Insert([]byte("a"), values[0])
	r.NoError(err)
	err = trie.Insert([]byte("b"), values[1])
	r.NoError(err)
	err = trie.Insert([]byte("c"), values[2])
	r.NoError(err)
	err = trie.Insert([]byte("d"), values[3])
	r.NoError(err)
	r.Equal(4, trie.Size())

	// Test replacing existing values
	trie2, err := copyTrie(trie)
	r.NoError(err)
	err = trie2.Insert([]byte("b"), newValues[1])
	r.NoError(err)
	r.Equal(4, trie2.Size())

	// Verify original is unchanged
	val, ok := trie.Get([]byte("b"))
	r.True(ok)
	r.Equal(values[1], val)

	// Verify new value in new trie
	val, ok = trie2.Get([]byte("b"))
	r.True(ok)
	r.Equal(newValues[1], val)

	// Test deleting values
	trie3, err := copyTrie(trie2)
	r.NoError(err)
	err = trie3.Delete([]byte("a"))
	r.NoError(err)
	r.Equal(3, trie3.Size())

	// Verify deletion
	_, ok = trie3.Get([]byte("a"))
	r.False(ok)

	// Other values should remain
	val, ok = trie3.Get([]byte("b"))
	r.True(ok)
	r.Equal(newValues[1], val)

	val, ok = trie3.Get([]byte("c"))
	r.True(ok)
	r.Equal(values[2], val)

	// Test deleting non-existent key
	trie4, err := copyTrie(trie3)
	r.NoError(err)
	err = trie4.Delete([]byte("not-exists"))
	r.NoError(err)
	r.Equal(3, trie4.Size())
	r.Equal(trie3, trie4)

	// Test multiple operations
	trie5, err := copyTrie(trie4)
	r.NoError(err)
	err = trie5.Delete([]byte("b"))
	r.NoError(err)
	err = trie5.Delete([]byte("c"))
	r.NoError(err)
	err = trie5.Insert([]byte("x"), mustEncode(t, 10))
	r.NoError(err)
	r.Equal(2, trie5.Size())

	val, ok = trie5.Get([]byte("d"))
	r.True(ok)
	r.Equal(values[3], val)

	val, ok = trie5.Get([]byte("x"))
	r.True(ok)
	r.Equal(mustEncode(t, 10), val)

	// Test that older versions of the trie are not affected by new operations
	trie6, err := copyTrie(trie5)
	r.NoError(err)
	err = trie6.Delete([]byte("d"))
	r.NoError(err)
	err = trie6.Insert([]byte("y"), []byte("11"))
	r.NoError(err)
	r.Equal(2, trie6.Size())

	val, ok = trie6.Get([]byte("d"))
	r.False(ok)
	r.Nil(val)
	r.Equal(4, trie.Size())
	r.Equal(3, trie3.Size())
	r.Equal(2, trie6.Size())
}

func TestTrieHash(t *testing.T) {
	r := require.New(t)

	// Empty trie should have consistent hash
	trie1 := NewTrie[string]()
	hash1, err := trie1.Hash()
	r.NoError(err)
	r.NotNil(hash1)

	// Same empty trie should have same hash
	trie2 := NewTrie[string]()
	hash2, err := trie2.Hash()
	r.NoError(err)
	r.Equal(hash1, hash2)

	// Adding elements should change hash
	err = trie1.Insert([]byte("a"), "1")
	r.NoError(err)
	hash3, err := trie1.Hash()
	r.NoError(err)
	r.NotEqual(hash1, hash3)

	// Same elements added in same order should have same hash
	err = trie1.Insert([]byte("a"), "1")
	r.NoError(err)
	hash4, err := trie1.Hash()
	r.NoError(err)
	r.Equal(hash3, hash4)

	// Different elements should have different hashes
	trie5 := NewTrie[string]()
	err = trie5.Insert([]byte("a"), "1")
	r.NoError(err)
	err = trie5.Insert([]byte("b"), "2")
	r.NoError(err)
	hash5, err := trie5.Hash()
	r.NoError(err)
	r.NotEqual(hash3, hash5)

	// Order of insertion shouldn't matter
	trie6 := NewTrie[string]()
	err = trie6.Insert([]byte("a"), "1")
	r.NoError(err)
	err = trie6.Insert([]byte("b"), "2")
	r.NoError(err)

	trie7 := NewTrie[string]()
	err = trie7.Insert([]byte("b"), "2")
	r.NoError(err)
	err = trie7.Insert([]byte("a"), "1")
	r.NoError(err)

	hash6, err := trie6.Hash()
	r.NoError(err)
	hash7, err := trie7.Hash()
	r.NoError(err)
	r.Equal(hash6, hash7)

	// Deleting should change hash
	trie8 := NewTrie[string]()
	err = trie8.Insert([]byte("a"), "1")
	r.NoError(err)
	err = trie8.Insert([]byte("b"), "2")
	r.NoError(err)

	err = trie8.Delete([]byte("a"))
	r.NoError(err)
	hash8, err := trie8.Hash()
	r.NoError(err)
	r.NotEqual(hash6, hash8)

	// Deleting and re-adding same element should restore hash
	err = trie8.Insert([]byte("a"), "1")
	r.NoError(err)
	hash9, err := trie8.Hash()
	r.NoError(err)
	r.Equal(hash6, hash9)
}

func TestCBORSerialization(t *testing.T) {
	r := require.New(t)

	// Create a new trie
	trie := NewTrie[string]()

	// empty trie should (de-)serialize to
	data, err := trie.MarshalCBOR()
	r.NoError(err)
	r.Equal("8200f6", hex.EncodeToString(data))

	var decoded = NewTrie[string]()
	err = decoded.UnmarshalCBOR(data)
	r.NoError(err)
	r.Equal(0, decoded.Size())

	testCases := []struct {
		key      []byte
		expected string
	}{
		{[]byte("key1"), "1"},
		{[]byte("key2"), "2"},
		{[]byte("key3"), "3"},
	}

	// Add some test data
	for _, tc := range testCases {
		err = trie.Insert(tc.key, tc.expected)
		r.NoError(err)
	}

	// Serialize the trie
	newTrie, err := copyTrie(trie)
	r.NoError(err)

	// Verify the contents
	for _, tc := range testCases {
		value, found := newTrie.Get(tc.key)
		r.True(found)
		r.Equal(tc.expected, value)
	}

	// Verify size
	r.Equal(trie.Size(), newTrie.Size())
}

func TestHashCollisions(t *testing.T) {
	r := require.New(t)

	// Override the hash function to force collisions for testing
	originalHashKeyWithSeed := hashKeyWithSeed
	originalHashKey := hashKey
	defer func() {
		hashKeyWithSeed = originalHashKeyWithSeed
		hashKey = originalHashKey
	}()
	hashKeyWithSeed = func(key []byte, seed uint64) uint64 {
		t.Logf("hashing key with seed(%d): %s", seed, key)
		// Force a collision for keys starting with "collide"
		if bytes.HasPrefix(key, []byte("collide")) {
			return 42
		}
		return originalHashKeyWithSeed(key, seed)
	}
	hashKey = func(key []byte) uint64 {
		// t.Logf("hashing key: %s", key)
		if bytes.HasPrefix(key, []byte("collide")) {
			return 42
		}
		return originalHashKey(key)
	}

	// Insert keys that will collide
	keys := [][]byte{[]byte("collide1"), []byte("collide2"), []byte("collide3")}
	values := []string{"value1", "value2", "value3"}

	trie := NewTrie[string]()
	for i, key := range keys {
		err := trie.Insert(key, values[i])
		r.NoError(err)
	}

	// Verify all values are retrievable
	for i, key := range keys {
		val, ok := trie.Get(key)
		r.True(ok)
		r.Equal(values[i], val)
	}

	// Ensure that the trie size reflects the correct number of entries
	r.Equal(len(keys), trie.Size())
}

func TestTrieSizeTracking(t *testing.T) {
	r := require.New(t)

	trie := NewTrie[string]()
	r.Equal(0, trie.Size())

	// Insert new keys
	err := trie.Insert([]byte("a"), "1")
	r.NoError(err)
	r.Equal(1, trie.Size())

	err = trie.Insert([]byte("b"), "2")
	r.NoError(err)
	r.Equal(2, trie.Size())

	// Update existing key
	err = trie.Insert([]byte("a"), "updated-1")
	r.NoError(err)
	r.Equal(2, trie.Size()) // Size should not change

	// Delete existing key
	err = trie.Delete([]byte("a"))
	r.NoError(err)
	r.Equal(1, trie.Size())

	// Delete non-existent key
	err = trie.Delete([]byte("non-existent"))
	r.NoError(err)
	r.Equal(1, trie.Size()) // Size should not change
}

func TestTrieDepth(t *testing.T) {
	r := require.New(t)
	trie := NewTrie[string]()
	numElements := 100000

	// Insert a large number of elements
	for i := 0; i < numElements; i++ {
		key := []byte(fmt.Sprintf("key-%d", i))
		value := fmt.Sprintf("value-%d", i)
		err := trie.Insert(key, value)
		r.NoError(err)
	}

	var depths []int
	trie.root.collectDepths(0, &depths)

	totalDepth := 0
	maxDepth := 0
	for _, depth := range depths {
		totalDepth += depth
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	averageDepth := float64(totalDepth) / float64(len(depths))

	// Corrected expected depth calculation
	expectedDepth := math.Log2(float64(numElements)) / bitsPerStep

	t.Logf("Average depth: %f, Expected depth: %f, Max depth: %d", averageDepth, expectedDepth, maxDepth)

	// Adjust acceptable variance
	r.True(averageDepth <= expectedDepth*1.2, "Average depth (%f) is higher than acceptable depth (%f)", averageDepth, expectedDepth*1.2)
	r.True(maxDepth <= int(expectedDepth*2), "Max depth (%d) is higher than expected (%f)", maxDepth, expectedDepth*2)
}

func TestHashOrderIndependence(t *testing.T) {
	r := require.New(t)

	// Create a map of key-value pairs
	items := map[string]string{
		"1":  "value1",
		"2":  "value2",
		"3":  "value3",
		"4":  "value4",
		"5":  "value5",
		"6":  "value6",
		"7":  "value7",
		"8":  "value8",
		"9":  "value9",
		"10": "value10",
	}

	const numTries = 1000
	var hash []byte
	var err error
	// insert items in different orders
	for i := 0; i < numTries; i++ {
		trie := NewTrie[string]()
		// Range over map which will be in random order each time
		for k, v := range items {
			err := trie.Insert([]byte(k), v)
			r.NoError(err)
		}
		if i == 0 {
			hash, err = trie.Hash()
			r.NoError(err)
		} else {
			var hash2 []byte
			hash2, err = trie.Hash()
			r.NoError(err)
			r.True(bytes.Equal(hash, hash2))
		}
	}
}

func TestLargeScaleInsertGetDelete(t *testing.T) {
	r := require.New(t)
	trie := NewTrie[string]()
	numElements := 100000
	keys := make([][]byte, numElements)
	values := make([]string, numElements)

	// Insert a large number of elements
	for i := 0; i < numElements; i++ {
		key := []byte(fmt.Sprintf("key-%d", i))
		value := fmt.Sprintf("value-%d", i)
		keys[i] = key
		values[i] = value
		err := trie.Insert(key, value)
		r.NoError(err)
	}

	r.Equal(numElements, trie.Size())

	// Verify that all elements can be retrieved
	for i, key := range keys {
		val, ok := trie.Get(key)
		r.True(ok)
		r.Equal(values[i], val)
	}

	// Delete every other element
	for i := 0; i < numElements; i += 2 {
		err := trie.Delete(keys[i])
		r.NoError(err)
	}

	// Verify that the correct elements have been deleted
	for i, key := range keys {
		val, ok := trie.Get(key)
		if i%2 == 0 {
			r.False(ok)
			r.Zero(val)
		} else {
			r.True(ok)
			r.Equal(values[i], val)
		}
	}

	r.Equal(numElements/2, trie.Size())
}

func TestTrieIterator(t *testing.T) {
	r := require.New(t)
	trie := NewTrie[string]()

	// Empty trie should not call function
	called := false
	trie.All(func(k []byte, v string) bool {
		called = true
		return true
	})
	r.False(called)

	// Insert some test data
	testData := map[string]string{
		"a": "value-a",
		"b": "value-b",
		"c": "value-c",
		"d": "value-d",
	}

	for k, v := range testData {
		err := trie.Insert([]byte(k), v)
		r.NoError(err)
	}

	// Collect all entries via iterator
	var gotValues []struct {
		Key   string
		Value string
	}
	trie.All(func(k []byte, v string) bool {
		gotValues = append(gotValues, struct {
			Key   string
			Value string
		}{Key: string(k), Value: v})
		return true
	})

	// Should visit all entries
	r.Equal(len(testData), len(gotValues))

	// Values should match
	for _, kv := range gotValues {
		r.Equal(testData[kv.Key], kv.Value, "key: %s", kv.Key)
	}

	// Early termination
	count := 0
	trie.All(func(k []byte, v string) bool {
		count++
		return count < 2 // Stop after first entry
	})
	r.Equal(2, count)
}

func BenchmarkTrieOperations(b *testing.B) {
	type KeyGenerator func(i int) []byte

	my_rand := rand.New(rand.NewSource(time.Now().UnixNano()))

	// a difference between these would show a problem with the hash function
	keyDistributions := map[string]KeyGenerator{
		"sequential": func(i int) []byte {
			return []byte(fmt.Sprintf("key-%d", i))
		},
		"sparse": func(i int) []byte {
			return []byte(fmt.Sprintf("key-%d", i*1000))
		},
		"random": func(i int) []byte {
			return []byte(fmt.Sprintf("key-%d", my_rand.Int()))
		},
	}

	init_size := []int{1000, 10_000, 100_000, 1_000_000}
	bench_size := 5000 // how many operations to do

	for distName, genFn := range keyDistributions {
		for _, size := range init_size {
			b.Run(fmt.Sprintf("%s_size_%d", distName, size), func(b *testing.B) {
				b.StopTimer()
				fill_keys := make([][]byte, size)
				for i := 0; i < size; i++ {
					fill_keys[i] = genFn(i)
				}

				var trie = NewTrie[string]()
				for _, key := range fill_keys {
					_ = trie.Insert(key, "value")
				}

				b.Log("init done")
				op_keys := make([][]byte, bench_size)
				for i := 0; i < bench_size; i++ {
					op_keys[i] = genFn(i)
				}
				b.ResetTimer()
				b.StartTimer()

				b.Run("insert", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						for _, key := range op_keys {
							_ = trie.Insert(key, "value")
						}
					}
				})

				b.Run("lookup", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						for _, key := range op_keys {
							_, _ = trie.Get(key)
						}
					}
				})

				b.Run("delete", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						for _, key := range op_keys {
							_ = trie.Delete(key)
						}
					}
				})
			})
		}
	}
}

// Helper function to create a copy of a trie through serialization
func copyTrie[V any](t *Trie[V]) (*Trie[V], error) {
	data, err := t.MarshalCBOR()
	if err != nil {
		return nil, err
	}
	newTrie := NewTrie[V]()
	err = newTrie.UnmarshalCBOR(data)
	return newTrie, err
}

// test-only helper function to collect depths of all nodes in the trie
func (n *Node[V]) collectDepths(currentDepth int, depths *[]int) {
	if n == nil {
		return
	}
	for _, e := range n.Entries {
		if e.Node == nil {
			*depths = append(*depths, currentDepth)
		} else {
			e.Node.collectDepths(currentDepth+1, depths)
		}
	}
}
