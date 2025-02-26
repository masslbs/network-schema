// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package mmr_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"os"
	"testing"

	mmr "github.com/datatrails/go-datatrails-merklelog/mmr"
	"github.com/jackc/pgx/v5"
	"github.com/masslbs/go-pgmmr"
	"github.com/peterldowns/testy/assert"
)

var hashFn = sha256.New

func TestWrapperPostgres(t *testing.T) {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		t.Skip("DATABASE_URL not set")
	}

	connPool, err := pgx.Connect(context.Background(), dbUrl)
	assert.Nil(t, err)

	hasher := hashFn()

	const testId = 42 * 123

	// delete previous values and tree nodes for clean slate
	_, err = connPool.Exec(context.Background(), "DELETE FROM pgmmr_values WHERE tree_id = $1", testId)
	assert.Nil(t, err)
	_, err = connPool.Exec(context.Background(), "DELETE FROM pgmmr_nodes WHERE tree_id = $1", testId)
	assert.Nil(t, err)

	tree, err := pgmmr.NewPostgresVerifierTree(connPool, hasher, testId)
	assert.Nil(t, err)

	testWrapper(t, tree)
}

func TestWrapperInMemory(t *testing.T) {
	db := pgmmr.NewInMemoryVerifierTree(hashFn(), 8)
	testWrapper(t, db)
}

func testWrapper(t *testing.T, tree pgmmr.VerifierTree) {
	const mmrSize = 16 // take care to use a valid mmr size

	type testValue struct {
		idx uint64
		val []byte
	}
	numLeafs := mmr.LeafCount(mmrSize)
	testValues := make([]testValue, numLeafs)

	// roll some random values and save their indices
	var val []byte
	for i := uint64(0); i < numLeafs; i++ {
		val = make([]byte, 32)
		rand.Read(val)
		idx, err := tree.Add(val)
		assert.Nil(t, err)
		testValues[i] = testValue{idx: idx, val: val}
		leafCount, err := tree.LeafCount()
		assert.Nil(t, err)
		assert.Equal(t, uint64(i+1), leafCount)
	}

	root, err := tree.Root()
	assert.Nil(t, err)
	assert.Equal(t, 32, len(root))
	t.Logf("root: %x", root)

	hasher := hashFn()

	// ensure we can get the values back
	for _, tv := range testValues {
		t.Logf("getting value %d", tv.idx)
		val, err := tree.GetValue(tv.idx)
		assert.Nil(t, err)
		assert.Equal(t, tv.val, val)

		hasher.Reset()
		hasher.Write(tv.val)
		data := hasher.Sum(nil)
		node, err := tree.GetNode(tv.idx)
		assert.Nil(t, err)
		assert.Equal(t, data, node)
	}

	t.Log("values checked. now verifying proofs")

	// verify all the values
	for _, tv := range testValues {
		proof, err := tree.MakeProof(tv.idx)
		assert.Nil(t, err)
		assert.NotEqual(t, nil, proof)
		t.Logf("proof for %d created", tv.idx)
		assert.Nil(t, tree.VerifyProof(*proof))
	}
}
