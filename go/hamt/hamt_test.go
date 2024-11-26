// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package hamt

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/peterldowns/testy/assert"
)

func TestHAMT(t *testing.T) {
	trie := NewTrie[string]()

	// Insert some values
	err := trie.Insert([]byte("name"), "Alice")
	assert.Nil(t, err)

	trie1, err := copyTrie(trie)
	assert.Nil(t, err)

	err = trie1.Insert([]byte("age"), "Bob")
	assert.Nil(t, err)
	assert.Equal(t, 2, trie1.Size())

	// Verify insertions
	val, ok := trie1.Get([]byte("name"))
	assert.True(t, ok)
	assert.Equal(t, "Alice", val)

	val, ok = trie1.Get([]byte("age"))
	assert.True(t, ok)
	assert.Equal(t, "Bob", val)

	// Original trie should be unchanged
	val, ok = trie.Get([]byte("age"))
	assert.False(t, ok)
	assert.Equal(t, "", val)

	// should also work with literal types, like uint64
	trie2 := NewTrie[uint64]()
	err = trie2.Insert([]byte("age"), 1)
	assert.Nil(t, err)
	valInt, ok := trie2.Get([]byte("age"))
	assert.True(t, ok)
	assert.Equal(t, 1, valInt)
}

func TestHAMTComplexOperations(t *testing.T) {
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
	assert.Nil(t, err)
	err = trie.Insert([]byte("b"), values[1])
	assert.Nil(t, err)
	err = trie.Insert([]byte("c"), values[2])
	assert.Nil(t, err)
	err = trie.Insert([]byte("d"), values[3])
	assert.Nil(t, err)
	assert.Equal(t, 4, trie.Size())

	// Test replacing existing values
	trie2, err := copyTrie(trie)
	assert.Nil(t, err)
	err = trie2.Insert([]byte("b"), newValues[1])
	assert.Nil(t, err)
	assert.Equal(t, 4, trie2.Size())

	// Verify original is unchanged
	val, ok := trie.Get([]byte("b"))
	assert.True(t, ok)
	assert.Equal(t, values[1], val)

	// Verify new value in new trie
	val, ok = trie2.Get([]byte("b"))
	assert.True(t, ok)
	assert.Equal(t, newValues[1], val)

	// Test deleting values
	trie3, err := copyTrie(trie2)
	assert.Nil(t, err)
	err = trie3.Delete([]byte("a"))
	assert.Nil(t, err)
	assert.Equal(t, 3, trie3.Size())

	// Verify deletion
	_, ok = trie3.Get([]byte("a"))
	assert.False(t, ok)

	// Other values should remain
	val, ok = trie3.Get([]byte("b"))
	assert.True(t, ok)
	assert.Equal(t, newValues[1], val)

	val, ok = trie3.Get([]byte("c"))
	assert.True(t, ok)
	assert.Equal(t, values[2], val)

	// Test deleting non-existent key
	trie4, err := copyTrie(trie3)
	assert.Nil(t, err)
	err = trie4.Delete([]byte("not-exists"))
	assert.Nil(t, err)
	assert.Equal(t, 3, trie4.Size())
	t3hash, err := trie3.Hash()
	assert.Nil(t, err)
	t4hash, err := trie4.Hash()
	assert.Nil(t, err)
	assert.Equal(t, t3hash, t4hash)

	// Test multiple operations
	trie5, err := copyTrie(trie4)
	assert.Nil(t, err)
	err = trie5.Delete([]byte("b"))
	assert.Nil(t, err)
	err = trie5.Delete([]byte("c"))
	assert.Nil(t, err)
	err = trie5.Insert([]byte("x"), mustEncode(t, 10))
	assert.Nil(t, err)
	assert.Equal(t, 2, trie5.Size())

	val, ok = trie5.Get([]byte("d"))
	assert.True(t, ok)
	assert.Equal(t, values[3], val)

	val, ok = trie5.Get([]byte("x"))
	assert.True(t, ok)
	assert.Equal(t, mustEncode(t, 10), val)

	// Test that older versions of the trie are not affected by new operations
	trie6, err := copyTrie(trie5)
	assert.Nil(t, err)
	err = trie6.Delete([]byte("d"))
	assert.Nil(t, err)
	err = trie6.Insert([]byte("y"), []byte("11"))
	assert.Nil(t, err)
	assert.Equal(t, 2, trie6.Size())

	val, ok = trie6.Get([]byte("d"))
	assert.False(t, ok)
	assert.Equal(t, nil, val)
	assert.Equal(t, 4, trie.Size())
	assert.Equal(t, 3, trie3.Size())
	assert.Equal(t, 2, trie6.Size())
}

func TestHAMTHash(t *testing.T) {

	// Empty trie should have consistent hash
	trie1 := NewTrie[string]()
	hash1, err := trie1.Hash()
	assert.Nil(t, err)

	// Same empty trie should have same hash
	trie2 := NewTrie[string]()
	hash2, err := trie2.Hash()
	assert.Nil(t, err)
	assert.Equal(t, hash1, hash2)

	// Adding elements should change hash
	err = trie1.Insert([]byte("a"), "1")
	assert.Nil(t, err)
	hash3, err := trie1.Hash()
	assert.Nil(t, err)
	assert.NotEqual(t, hash1, hash3)

	// Same elements added in same order should have same hash
	err = trie1.Insert([]byte("a"), "1")
	assert.Nil(t, err)
	hash4, err := trie1.Hash()
	assert.Nil(t, err)
	assert.Equal(t, hash3, hash4)

	// Different elements should have different hashes
	trie5 := NewTrie[string]()
	err = trie5.Insert([]byte("a"), "1")
	assert.Nil(t, err)
	err = trie5.Insert([]byte("b"), "2")
	assert.Nil(t, err)
	hash5, err := trie5.Hash()
	assert.Nil(t, err)
	assert.NotEqual(t, hash3, hash5)

	// Order of insertion shouldn't matter
	trie6 := NewTrie[string]()
	err = trie6.Insert([]byte("a"), "1")
	assert.Nil(t, err)
	err = trie6.Insert([]byte("b"), "2")
	assert.Nil(t, err)

	trie7 := NewTrie[string]()
	err = trie7.Insert([]byte("b"), "2")
	assert.Nil(t, err)
	err = trie7.Insert([]byte("a"), "1")
	assert.Nil(t, err)

	hash6, err := trie6.Hash()
	assert.Nil(t, err)
	hash7, err := trie7.Hash()
	assert.Nil(t, err)
	assert.Equal(t, hash6, hash7)

	// Deleting should change hash
	trie8 := NewTrie[string]()
	err = trie8.Insert([]byte("a"), "1")
	assert.Nil(t, err)
	err = trie8.Insert([]byte("b"), "2")
	assert.Nil(t, err)

	err = trie8.Delete([]byte("a"))
	assert.Nil(t, err)
	hash8, err := trie8.Hash()
	assert.Nil(t, err)
	assert.NotEqual(t, hash6, hash8)

	// Deleting and re-adding same element should restore hash
	err = trie8.Insert([]byte("a"), "1")
	assert.Nil(t, err)
	hash9, err := trie8.Hash()
	assert.Nil(t, err)
	assert.Equal(t, hash6, hash9)
}

func TestHAMTCBORSerialization(t *testing.T) {
	// Create a new trie
	trie := NewTrie[string]()

	// empty trie should (de-)serialize to
	data, err := trie.MarshalCBOR()
	assert.Nil(t, err)
	assert.Equal(t, "8200f6", hex.EncodeToString(data))

	var decoded = NewTrie[string]()
	err = decoded.UnmarshalCBOR(data)
	assert.Nil(t, err)
	assert.Equal(t, 0, decoded.Size())

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
		assert.Nil(t, err)
	}

	// Serialize the trie
	newTrie, err := copyTrie(trie)
	assert.Nil(t, err)

	// Verify the contents
	for _, tc := range testCases {
		value, found := newTrie.Get(tc.key)
		assert.True(t, found)
		assert.Equal(t, tc.expected, value)
	}

	// Verify size
	assert.Equal(t, trie.Size(), newTrie.Size())
}

func TestHAMTSizeTracking(t *testing.T) {

	trie := NewTrie[string]()
	assert.Equal(t, 0, trie.Size())

	// Insert new keys
	err := trie.Insert([]byte("a"), "1")
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.Size())

	err = trie.Insert([]byte("b"), "2")
	assert.Nil(t, err)
	assert.Equal(t, 2, trie.Size())

	// Update existing key
	err = trie.Insert([]byte("a"), "updated-1")
	assert.Nil(t, err)
	assert.Equal(t, 2, trie.Size()) // Size should not change

	// Delete existing key
	err = trie.Delete([]byte("a"))
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.Size())

	// Delete non-existent key
	err = trie.Delete([]byte("non-existent"))
	assert.Nil(t, err)
	assert.Equal(t, 1, trie.Size()) // Size should not change
}

func TestHAMTInsertDepth(t *testing.T) {
	trie := NewTrie[string]()
	numElements := 100000

	// Insert a large number of elements
	for i := range numElements {
		key := []byte(fmt.Sprintf("key-%d", i))
		value := fmt.Sprintf("value-%d", i)
		err := trie.Insert(key, value)
		assert.Nil(t, err)
		_, ok := trie.Get(key)
		assert.True(t, ok)
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
	assert.True(t, averageDepth <= expectedDepth*1.2) // "Average depth (%f) is higher than acceptable depth (%f)", averageDepth, expectedDepth*1.2)
	assert.True(t, maxDepth <= int(expectedDepth*3))  // "Max depth (%d) is higher than expected (%f)", maxDepth, expectedDepth*3)
}

func TestHAMTHashOrderIndependence(t *testing.T) {
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
	for i := range numTries {
		trie := NewTrie[string]()
		// Range over map which will be in random order each time
		for k, v := range items {
			key := bytes.Repeat([]byte(k), 32)
			err := trie.Insert(key, v)
			assert.Nil(t, err)
			_, ok := trie.Get(key)
			assert.True(t, ok)
		}
		if i == 0 {
			hash, err = trie.Hash()
			assert.Nil(t, err)
		} else {
			var hash2 []byte
			hash2, err = trie.Hash()
			assert.Nil(t, err)
			assert.Equal(t, hash, hash2)
		}
	}
}

func TestHAMTLargeScaleInsertGetDelete(t *testing.T) {
	trie := NewTrie[string]()
	numElements := 100000
	keys := make([][]byte, numElements)
	values := make([]string, numElements)

	// Insert a large number of elements
	for i := range numElements {
		key := []byte(fmt.Sprintf("key-%d", i))
		value := fmt.Sprintf("value-%d", i)
		keys[i] = key
		values[i] = value
		err := trie.Insert(key, value)
		assert.Nil(t, err)
	}

	assert.Equal(t, numElements, trie.Size())

	// Verify that all elements can be retrieved
	for i, key := range keys {
		val, ok := trie.Get(key)
		assert.True(t, ok)
		assert.Equal(t, values[i], val)
	}

	// Delete every other element
	for i := range numElements {
		if i%2 == 0 {
			err := trie.Delete(keys[i])
			assert.Nil(t, err)
		}
	}

	// Verify that the correct elements have been deleted
	for i, key := range keys {
		val, ok := trie.Get(key)
		if i%2 == 0 {
			assert.False(t, ok)
			assert.Equal(t, "", val)
		} else {
			assert.True(t, ok)
			assert.Equal(t, values[i], val)
		}
	}

	assert.Equal(t, numElements/2, trie.Size())
}

func TestHAMTIterator(t *testing.T) {
	trie := NewTrie[string]()

	// Empty trie should not call function
	called := false
	trie.All(func(_ []byte, _ string) bool {
		called = true
		return true
	})
	assert.False(t, called)

	// Insert some test data
	testData := map[string]string{
		"a": "value-a",
		"b": "value-b",
		"c": "value-c",
		"d": "value-d",
	}

	for k, v := range testData {
		err := trie.Insert([]byte(k), v)
		assert.Nil(t, err)
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
	assert.Equal(t, len(testData), len(gotValues))

	// Values should match
	for _, kv := range gotValues {
		assert.Equal(t, testData[kv.Key], kv.Value)
	}

	// Early termination
	count := 0
	trie.All(func(_ []byte, _ string) bool {
		count++
		return count < 2 // Stop after first entry
	})
	assert.Equal(t, 2, count)
}

func BenchmarkHAMTOperations(b *testing.B) {
	type KeyGenerator func(i int) []byte

	myRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	// a difference between these would show a problem with the hash function
	keyDistributions := map[string]KeyGenerator{
		"sequential": func(i int) []byte {
			return []byte(fmt.Sprintf("key-%d", i))
		},
		"sparse": func(i int) []byte {
			return []byte(fmt.Sprintf("key-%d", i*1000))
		},
		"random": func(_ int) []byte {
			return []byte(fmt.Sprintf("key-%d", myRand.Int()))
		},
	}

	initSize := []int{1000, 10_000, 100_000, 1_000_000}
	benchSize := 5000 // how many operations to do

	for distName, genFn := range keyDistributions {
		for _, size := range initSize {
			b.Run(fmt.Sprintf("%s_size_%d", distName, size), func(b *testing.B) {
				b.StopTimer()
				fillKeys := make([][]byte, size)
				for i := 0; i < size; i++ {
					fillKeys[i] = genFn(i)
				}

				var trie = NewTrie[string]()
				for _, key := range fillKeys {
					_ = trie.Insert(key, "value")
				}

				b.Log("init done")
				opKeys := make([][]byte, benchSize)
				for i := 0; i < benchSize; i++ {
					opKeys[i] = genFn(i)
				}
				b.ResetTimer()
				b.StartTimer()

				b.Run("insert", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						for _, key := range opKeys {
							_ = trie.Insert(key, "value")
						}
					}
				})

				b.Run("lookup", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						for _, key := range opKeys {
							_, _ = trie.Get(key)
						}
					}
				})

				b.Run("delete", func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						for _, key := range opKeys {
							_ = trie.Delete(key)
						}
					}
				})
			})
		}
	}
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

type TestOperation struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TestVector struct {
	Operations []TestOperation `json:"operations"`
	Hashes     []string        `json:"hashes"`
}

func TestHAMTVectors(t *testing.T) {
	// Read test vectors
	data, err := os.ReadFile("../../vectors/hamt_test.cbor")
	assert.Nil(t, err)

	var vectors []TestVector
	err = cbor.Unmarshal(data, &vectors)
	assert.Nil(t, err)

	for i, vector := range vectors {
		trie := NewTrie[string]()

		t.Run(fmt.Sprintf("vector-%d", i), func(t *testing.T) {
			for j, op := range vector.Operations {
				key, err := hex.DecodeString(op.Key)
				assert.Nil(t, err)

				switch op.Type {
				case "insert":
					if err := trie.Insert(key, op.Value); err != nil {
						t.Fatalf("Vector %d, operation %d: insert failed: %v", i, j, err)
					}
				case "delete":
					if err := trie.Delete(key); err != nil {
						t.Fatalf("Vector %d, operation %d: delete failed: %v", i, j, err)
					}
				default:
					t.Fatalf("Vector %d, operation %d: invalid operation: %s", i, j, op.Type)
				}

				hash, err := trie.Hash()
				assert.Nil(t, err)

				actualHash := hex.EncodeToString(hash)
				assert.Equal(t, vector.Hashes[j], actualHash)
			}
		})
	}
}

// helpers

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

func mustEncode(t *testing.T, v any) cbor.RawMessage {
	data, err := masscbor.Marshal(v)
	if err != nil {
		t.Fatalf("unable to encode value: %v", err)
	}
	return data
}
