// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

// Package mmr provides merkle mountain range abstraction over the implementation of the datatrails-merklelog package.
package mmr

import (
	"context"
	"errors"
	"fmt"
	"sync"

	dtmmr "github.com/datatrails/go-datatrails-merklelog/mmr"
	pgx "github.com/jackc/pgx/v5"
)

// PostgresNodeStore is a store of tree nodes for a postgres database.
type PostgresNodeStore struct {
	mu            sync.Mutex
	db            *pgx.Conn
	treeID        uint64
	currentNodeID uint64
}

// NewPostgresNodeStore creates a new PostgresNodeStore.
func NewPostgresNodeStore(db *pgx.Conn, id uint64) (*PostgresNodeStore, error) {
	row := db.QueryRow(context.Background(), "SELECT count(*) FROM pgmmr_nodes WHERE tree_id = $1", id)
	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return nil, err
	}

	return &PostgresNodeStore{db: db, treeID: id, currentNodeID: count}, nil
}

// Append appends a node to the store.
func (s *PostgresNodeStore) Append(data []byte) (uint64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	const qry = "INSERT INTO pgmmr_nodes (tree_id, node_id, data) VALUES ($1, $2, $3)"
	_, err := s.db.Exec(context.Background(), qry, s.treeID, s.currentNodeID, data)
	if err != nil {
		return 0, err
	}
	// fmt.Fprintf(debug, "appending: store[%02d] = %x\n", s.currentNodeId, data)
	s.currentNodeID++
	return s.currentNodeID, nil
}

// Get gets a node from the store.
func (s *PostgresNodeStore) Get(i uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	const qry = "SELECT data FROM pgmmr_nodes WHERE tree_id = $1 AND node_id = $2"
	var data []byte
	err := s.db.QueryRow(context.Background(), qry, s.treeID, i).Scan(&data)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("failed to get node %d: %w", i, dtmmr.ErrNotFound)
		}
		return nil, err
	}
	// fmt.Printf("Get(%d): %x\n", i, data)
	return data, nil
}

// Size returns the number of nodes in the store.
func (s *PostgresNodeStore) Size() (uint64, error) {
	const qry = "SELECT count(*) FROM pgmmr_nodes WHERE tree_id = $1"
	var count uint64
	err := s.db.QueryRow(context.Background(), qry, s.treeID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// InMemoryNodeStore is a store of tree nodes in memory.
type InMemoryNodeStore struct {
	nodes [][]byte
	next  uint64
}

// Append appends a node to the store.
func (t *InMemoryNodeStore) Append(value []byte) (uint64, error) {
	if t.next >= uint64(len(t.nodes)) {
		// return 0, fmt.Errorf("out of bounds. %d >= %d", t.next, len(t.nodes))
		t.nodes = append(t.nodes, value)
	} else {
		t.nodes[t.next] = value
	}
	// fmt.Printf("appending: store[%02d] = %x\n", t.next, value)
	t.next++
	return t.next, nil
}

// Get gets a node from the store.
func (t *InMemoryNodeStore) Get(i uint64) ([]byte, error) {
	if i >= t.next {
		return nil, dtmmr.ErrNotFound
	}
	// fmt.Printf("Get(%d): %x, len(t.nodes): %d\n", i, t.nodes[i], t.next)
	return t.nodes[i], nil
}

// Size returns the number of nodes in the store.
func (t *InMemoryNodeStore) Size() (uint64, error) {
	return t.next, nil
}
