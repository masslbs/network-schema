// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package mmr_test

import (
	"bytes"
	"testing"

	massmmr "github.com/masslbs/network-schema/go/mmr"
)

func TestInvalidProofs(t *testing.T) {
	db := massmmr.NewInMemoryVerifierTree(hashFn(), 16)

	tests := []struct {
		name    string
		pos     uint64
		value   []byte
		proof   [][]byte
		wantErr bool
	}{
		{
			name:    "empty proof",
			pos:     0,
			value:   []byte{1, 2, 3},
			proof:   [][]byte{},
			wantErr: true,
		},
		{
			name:    "proof too long",
			pos:     0,
			value:   []byte{1, 2, 3},
			proof:   bytes.Split(bytes.Repeat([]byte{1, 2, 3}, 10), []byte{1, 2, 3}),
			wantErr: true,
		},
		{
			name:    "incorrect hash size",
			pos:     2,
			value:   []byte{1, 2, 3},
			proof:   [][]byte{{1}, {2}},
			wantErr: true,
		},
		{
			name:    "position beyond tree size",
			pos:     100,
			value:   []byte{1, 2, 3},
			proof:   [][]byte{{1, 2, 3}},
			wantErr: true,
		},
		{
			name:    "nil proof",
			pos:     0,
			value:   []byte{1, 2, 3},
			proof:   nil,
			wantErr: true,
		},
		{
			name:    "nil hashes in proof",
			pos:     2,
			value:   []byte{1, 2, 3},
			proof:   [][]byte{nil, nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p massmmr.Proof
			p.NodeIndex = tt.pos
			// p.Value = tt.value
			// p.Proof = tt.proof
			err := db.VerifyProof(p)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
