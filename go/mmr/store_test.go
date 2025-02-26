// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package mmr_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	mmr "github.com/datatrails/go-datatrails-merklelog/mmr"
	"github.com/jackc/pgx/v5"
	"github.com/peterldowns/testy/assert"

	massmmr "github.com/masslbs/network-schema/go/mmr"
)

// mostly a sanity check
// mmr.FirstMMRSize is the smallest size that can contain a leaf at index i
func TestSmallestSize(t *testing.T) {
	type testCase struct {
		index        uint64
		smallestSize uint64
	}

	testCases := []testCase{
		{index: 0, smallestSize: 1},
		{index: 2, smallestSize: 3},
		{index: 4, smallestSize: 7},
		{index: 6, smallestSize: 7},
		{index: 8, smallestSize: 10},
		{index: 10, smallestSize: 11},
		{index: 12, smallestSize: 15},
		{index: 14, smallestSize: 15},
		{index: 16, smallestSize: 18},
		{index: 18, smallestSize: 19},
		{index: 20, smallestSize: 22},
		{index: 22, smallestSize: 23},
		{index: 24, smallestSize: 25},
		{index: 26, smallestSize: 31},
		{index: 28, smallestSize: 31},
		{index: 30, smallestSize: 31},
		{index: 31, smallestSize: 32},
		{index: 32, smallestSize: 34},
		{index: 34, smallestSize: 35},
		{index: 36, smallestSize: 38},
		{index: 38, smallestSize: 39},
		{index: 40, smallestSize: 41},
		{index: 42, smallestSize: 46},
		{index: 44, smallestSize: 46},
		{index: 46, smallestSize: 47},
		{index: 48, smallestSize: 49},
		{index: 50, smallestSize: 53},
		{index: 52, smallestSize: 53},
		{index: 54, smallestSize: 56},
		{index: 56, smallestSize: 57},
		{index: 58, smallestSize: 63},
		{index: 60, smallestSize: 63},
		{index: 62, smallestSize: 63},
	}
	for _, tc := range testCases {
		assert.Equal(t, tc.smallestSize, mmr.FirstMMRSize(tc.index))
	}
}

const mmrTestSize = 7

func TestVerifyInMemoryStore(t *testing.T) {
	db := &massmmr.InMemoryNodeStore{}
	const mmrSize uint64 = mmrTestSize
	testStore(t, db, mmrSize)
}

func TestVerifyPostgresStore(t *testing.T) {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		t.Skip("DATABASE_URL not set")
	}

	connPool, err := pgx.Connect(context.Background(), dbUrl)
	assert.Nil(t, err)

	const testTreeId uint64 = 23 * 42

	_, err = connPool.Exec(context.Background(), "DELETE FROM pgmmr_nodes where tree_id = $1", testTreeId)
	assert.Nil(t, err)

	store, err := massmmr.NewPostgresNodeStore(connPool, testTreeId)
	assert.Nil(t, err)

	testStore(t, store, mmrTestSize)
}

func testStore(t *testing.T, db mmr.NodeAppender, mmrMaxSize uint64) {
	hasher := sha256.New()

	numLeafs := mmr.LeafCount(mmrMaxSize)
	assert.Equal(t, mmrMaxSize/2+1, numLeafs)

	// fill the tree with some data
	var lastIdx uint64
	for i := 0; i < int(numLeafs); i++ {
		input := []byte(fmt.Sprintf("hello %02d", i))
		hasher.Reset()
		hasher.Write(input)
		data := hasher.Sum(nil)
		// data := input
		idx, err := mmr.AddHashedLeaf(db, hasher, data)
		assert.Nil(t, err)
		t.Logf("idx: %02d: %s", idx, input)
		lastIdx = idx
	}
	assert.Equal(t, mmrMaxSize, lastIdx)

	root, err := mmr.GetRoot(mmrMaxSize, db, hasher)
	assert.Nil(t, err)
	t.Logf("root: %x", root)
	assert.NotEqual(t, nil, root)

	for iLeaf := uint64(0); iLeaf < numLeafs; iLeaf++ {
		mmrIndex := mmr.MMRIndex(iLeaf)
		t.Log("================")
		t.Logf("iLeaf: %d", iLeaf)
		t.Log("================")

		// s is the size of the mmr at which the leaf is included
		for s := mmr.FirstMMRSize(mmr.MMRIndex(iLeaf)); s <= mmrMaxSize; s = mmr.FirstMMRSize(s + 1) {
			t.Logf("s: %d", s)

			proof, err := mmr.InclusionProof(db, s-1, mmrIndex)
			assert.Nil(t, err)
			t.Logf("proof len(): %d", len(proof))
			nodeHash, err := db.Get(mmrIndex)
			assert.Nil(t, err)

			accumulator, err := mmr.PeakHashes(db, s-1)
			assert.Nil(t, err)
			t.Logf("accumulator len(): %d", len(accumulator))
			iacc := mmr.PeakIndex(mmr.LeafCount(s), len(proof))
			t.Logf("iacc: %d", iacc)
			assert.LessThan(t, iacc, len(accumulator))

			peak := accumulator[iacc]
			root := mmr.IncludedRoot(hasher, mmrIndex, nodeHash, proof)

			ok := bytes.Equal(root, peak)
			if !ok {
				t.Logf("%d %d VerifyInclusion() failed\n", mmrIndex, iLeaf)
			}
			assert.True(t, ok)
		}
	}
}
