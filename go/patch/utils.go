// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"crypto/sha256"
	"fmt"

	"github.com/datatrails/go-datatrails-merklelog/mmr"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	massmmr "github.com/masslbs/network-schema/go/mmr"
	"github.com/masslbs/network-schema/go/objects"
)

// RootHash computes the root hash of a list of patches.
// It returns the root hash, the MMR tree, and an error if one occurs.
func RootHash(patches []Patch) (objects.Hash, massmmr.VerifierTree, error) {
	sz := mmr.FirstMMRSize(uint64(len(patches)))

	tree := massmmr.NewInMemoryVerifierTree(sha256.New(), sz)
	for _, patch := range patches {
		data, err := masscbor.Marshal(patch)
		if err != nil {
			return objects.Hash{}, nil, fmt.Errorf("failed to marshal patch: %w", err)
		}
		_, err = tree.Add(data)
		if err != nil {
			return objects.Hash{}, nil, fmt.Errorf("failed to add patch to tree: %w", err)
		}
	}

	// fill up the tree to the next power of 2
	cnt, err := tree.LeafCount()
	if err != nil {
		return objects.Hash{}, nil, fmt.Errorf("failed to get leaf count: %w", err)
	}
	nextSquare := NextPowerOf2(cnt)
	for i := cnt; i < nextSquare; i++ {
		_, err = tree.Add([]byte{})
		if err != nil {
			return objects.Hash{}, nil, fmt.Errorf("failed to add empty leaf to tree: %w", err)
		}
	}
	root, err := tree.Root()
	if err != nil {
		return objects.Hash{}, nil, fmt.Errorf("failed to get root: %w", err)
	}
	return objects.Hash(root), tree, nil
}

// NextPowerOf2 calculates the smallest power of 2 that is greater than or equal to n.
// It works by:
//   - n--: First decrements n by 1. This is done to handle the case where n is already a power of 2.
//   - The series of bit-shifting operations (|= with right shifts):
//     This sequence "fills" all the bits to the right of the highest set bit with 1s. For example:
//     If n = 00100000, after these operations it becomes 00111111
//   - n++: Finally increments n by 1, which gives us the next power of 2.
//
// Here's a concrete example:
// Start with n = 33 (00100001 in binary)
// After n--, n = 32 (00100000)
// After bit-shifting operations, n = 00111111
// After n++, n = 01000000 (64 in decimal)
func NextPowerOf2(n uint64) uint64 {
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	n++
	return n
}
