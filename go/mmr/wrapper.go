// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package mmr

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash"

	"github.com/datatrails/go-datatrails-merklelog/mmr"
	pgx "github.com/jackc/pgx/v5"
)

// VerifierTree is a tree that can be used to verify the inclusion of values in the tree.
type VerifierTree interface {
	// Add adds a value to the tree and returns the index of the value.
	Add(value []byte) (uint64, error)
	// GetValue returns the value at the given index.
	GetValue(i uint64) ([]byte, error)
	// GetNode returns the (hashed) node at the given index.
	GetNode(i uint64) ([]byte, error)
	// LeafCount returns the number of leaves in the tree.
	LeafCount() (uint64, error)
	// Root returns the root of the tree.
	Root() ([]byte, error)
	// MakeProof returns a proof that the value at the given index is in the tree.
	MakeProof(i uint64) (*Proof, error)
	// VerifyProof verifies a proof that the value at the given index is in the tree.
	VerifyProof(proof Proof) error
}

// Proof is a proof that a value is in the tree.
type Proof struct {
	_         struct{} `cbor:",toarray"`
	NodeIndex uint64
	TreeSize  uint64
	Path      [][]byte
}

// PostgresVerifierTree is a tree that can be used to verify the inclusion of values in the tree.
type PostgresVerifierTree struct {
	db     *pgx.Conn
	hasher hash.Hash
	treeID uint64
	nodes  *PostgresNodeStore
}

var _ VerifierTree = (*PostgresVerifierTree)(nil)

// NewPostgresVerifierTree creates a new PostgresVerifierTree.
func NewPostgresVerifierTree(db *pgx.Conn, hasher hash.Hash, id uint64) (*PostgresVerifierTree, error) {
	nodes, err := NewPostgresNodeStore(db, id)
	if err != nil {
		return nil, err
	}
	return &PostgresVerifierTree{
		db:     db,
		hasher: hasher,
		treeID: id,
		nodes:  nodes,
	}, nil
}

// Add adds a value to the tree and returns the index of the value in the tree.
func (t *PostgresVerifierTree) Add(value []byte) (uint64, error) {
	hasher := t.hasher
	hasher.Reset()
	hasher.Write(value)
	data := hasher.Sum(nil)

	newSize, err := mmr.AddHashedLeaf(t.nodes, t.hasher, data)
	if err != nil {
		return 0, err
	}
	// AddHashedLeaf returns the new size of the tree
	// which is equal to the last node _position_ in the tree
	leafIdx := mmr.LeafIndex(newSize - 1)

	const insertValueQry = "INSERT INTO pgmmr_values (tree_id, leaf_idx, data) VALUES ($1, $2, $3)"
	_, err = t.db.Exec(context.Background(), insertValueQry, t.treeID, leafIdx, value)
	if err != nil {
		return 0, err
	}

	return leafIdx, nil
}

// GetNode returns the node at the given index.
func (t *PostgresVerifierTree) GetNode(i uint64) ([]byte, error) {
	// turn leaf index into node index
	nodeIdx := mmr.MMRIndex(i)
	return t.nodes.Get(nodeIdx)
}

// GetValue returns the value of a node at the given index.
func (t *PostgresVerifierTree) GetValue(i uint64) ([]byte, error) {
	var value []byte
	err := t.db.QueryRow(context.Background(), "SELECT data FROM pgmmr_values WHERE tree_id = $1 AND leaf_idx = $2", t.treeID, i).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("failed to get value %d: %w", i, mmr.ErrNotFound)
		}
		return nil, err
	}
	return value, nil
}

// LeafCount returns the number of leaves in the tree.
func (t *PostgresVerifierTree) LeafCount() (uint64, error) {
	count, err := t.nodes.Size()
	if err != nil {
		return 0, err
	}
	return mmr.LeafCount(count), nil
}

// Root returns the root of the tree.
func (t *PostgresVerifierTree) Root() ([]byte, error) {
	count, err := t.nodes.Size()
	if err != nil {
		return nil, err
	}
	return mmr.GetRoot(count, t.nodes, t.hasher)
}

// MakeProof returns a proof that the value at the given index is in the tree.
func (t *PostgresVerifierTree) MakeProof(i uint64) (*Proof, error) {
	return makeProof(t.nodes, i)
}

// VerifyProof verifies a proof that the value at the given index is in the tree.
func (t *PostgresVerifierTree) VerifyProof(proof Proof) error {
	return verifyPath(t.nodes, t.hasher, proof)
}

// InMemoryVerifierTree is a tree that can be used to verify the inclusion of values in the tree.
type InMemoryVerifierTree struct {
	hasher hash.Hash
	nodes  *InMemoryNodeStore
	values map[uint64][]byte
}

var _ VerifierTree = (*InMemoryVerifierTree)(nil)

// NewInMemoryVerifierTree creates a new InMemoryVerifierTree.
func NewInMemoryVerifierTree(hasher hash.Hash, initalSize uint64) *InMemoryVerifierTree {
	return &InMemoryVerifierTree{
		hasher: hasher,
		nodes:  &InMemoryNodeStore{nodes: make([][]byte, initalSize)},
		values: make(map[uint64][]byte),
	}
}

// Add adds a value to the tree and returns the index of the value in the tree.
func (t *InMemoryVerifierTree) Add(value []byte) (uint64, error) {
	h := t.hasher
	h.Reset()
	h.Write(value)
	data := h.Sum(nil)
	newSize, err := mmr.AddHashedLeaf(t.nodes, t.hasher, data)
	if err != nil {
		return 0, err
	}
	// AddHashedLeaf returns the new size of the tree
	// which is equal to the last node _position_ in the tree
	leafIdx := mmr.LeafIndex(newSize - 1)
	if _, ok := t.values[leafIdx]; ok {
		return 0, fmt.Errorf("value already exists at index %d", leafIdx)
	}
	t.values[leafIdx] = value
	return leafIdx, nil
}

// GetNode returns the node at the given index.
func (t *InMemoryVerifierTree) GetNode(i uint64) ([]byte, error) {
	mmrIndex := mmr.MMRIndex(i)
	return t.nodes.Get(mmrIndex)
}

// GetValue returns the value of a node at the given index.
func (t *InMemoryVerifierTree) GetValue(i uint64) ([]byte, error) {
	value, ok := t.values[i]
	if !ok {
		return nil, fmt.Errorf("value not found at index %d", i)
	}
	return value, nil
}

// LeafCount returns the number of leaves in the tree.
func (t *InMemoryVerifierTree) LeafCount() (uint64, error) {
	count, err := t.nodes.Size()
	if err != nil {
		return 0, err
	}
	return mmr.LeafCount(count), nil
}

// Root returns the root of the tree.
func (t *InMemoryVerifierTree) Root() ([]byte, error) {
	count, err := t.nodes.Size()
	if err != nil {
		return nil, err
	}
	return mmr.GetRoot(count, t.nodes, t.hasher)
}

// MakeProof returns a proof that the value at the given index is in the tree.
func (t *InMemoryVerifierTree) MakeProof(i uint64) (*Proof, error) {
	return makeProof(t.nodes, i)
}

// VerifyProof verifies a proof that the value at the given index is in the tree.
func (t *InMemoryVerifierTree) VerifyProof(proof Proof) error {
	return verifyPath(t.nodes, t.hasher, proof)
}

// NodeAppenderWithSize is a type that can append nodes to a tree and return the size of the tree.
type NodeAppenderWithSize interface {
	mmr.NodeAppender
	Size() (uint64, error)
}

// makeProof creates a proof that the value at the given index is in the tree.
func makeProof(nodes NodeAppenderWithSize, i uint64) (*Proof, error) {
	count, err := nodes.Size()
	if err != nil {
		return nil, err
	}
	mmrIndex := mmr.MMRIndex(i)
	proof, err := mmr.InclusionProof(nodes, count-1, mmrIndex)
	if err != nil {
		return nil, err
	}
	return &Proof{
		TreeSize:  count,
		NodeIndex: mmrIndex,
		Path:      proof,
	}, nil
}

// verifyPath verifies a proof that the value at the given index is in the tree.
func verifyPath(tree NodeAppenderWithSize, hasher hash.Hash, proof Proof) error {
	count, err := tree.Size()
	if err != nil {
		return err
	}

	if proof.TreeSize > count {
		return fmt.Errorf("proof tree size %d is greater than current tree size %d", proof.TreeSize, count)
	}

	if proof.NodeIndex >= count {
		return fmt.Errorf("proof node index %d is greater than current tree size %d", proof.NodeIndex, count)
	}

	node, err := tree.Get(proof.NodeIndex)
	if err != nil {
		return err
	}

	accumulator, err := mmr.PeakHashes(tree, proof.TreeSize-1)
	if err != nil {
		return err
	}

	iacc := mmr.PeakIndex(mmr.LeafCount(proof.TreeSize), len(proof.Path))
	if iacc >= len(accumulator) {
		return fmt.Errorf("proof peak index %d is greater than accumulator length %d", iacc, len(accumulator))
	}

	peak := accumulator[iacc]
	root := mmr.IncludedRoot(hasher, proof.NodeIndex, node, proof.Path)

	ok := bytes.Equal(root, peak)
	if !ok {
		return fmt.Errorf("proof verification for %d failed: %w", proof.NodeIndex, mmr.ErrVerifyInclusionFailed)
	}
	return nil
}
