// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"
)

var validate = DefaultValidator()

func openTestFile(t testing.TB, fileName string) *os.File {
	path := filepath.Join(os.Getenv("TEST_DATA_OUT"), fileName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
	}
	return file
}

func mustEncode(t *testing.T, v any) cbor.RawMessage {
	data, err := Marshal(v)
	require.NoError(t, err)
	return data
}

func testPubKey(i uint64) PublicKey {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	var pk PublicKey
	n := copy(pk[:], hash(b))
	if n != PublicKeySize {
		panic(fmt.Sprintf("copy failed: %d != %d", n, PublicKeySize))
	}
	return pk
}

func createPatch(t testing.TB, op OpString, path PatchPath, value interface{}) Patch {
	encodedValue, err := Marshal(value)
	require.NoError(t, err)
	return Patch{
		Op:    op,
		Path:  path,
		Value: encodedValue,
	}
}

func encodePatch(t *testing.T, patch Patch) []byte {
	encoded, err := Marshal(patch)
	require.NoError(t, err)
	return encoded
}

func decodePatch(t *testing.T, encoded []byte) Patch {
	var decoded Patch
	dec := DefaultDecoder(bytes.NewReader(encoded))
	err := dec.Decode(&decoded)
	require.NoError(t, err)
	require.NoError(t, validate.Struct(decoded))
	return decoded
}

func hash(value []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(value)
	return hash.Sum(nil)
}

func testHash(i uint) cid.Cid {
	h, err := mh.Sum([]byte(fmt.Sprintf("TEST-%d", i)), mh.SHA3, 4)
	check(err)
	// TODO: check what the codec number should be
	return cid.NewCidV1(666, h)
}

func strptr(s string) *string {
	return &s
}

func boolptr(b bool) *bool {
	return &b
}

func uint64ptr(i uint64) *uint64 {
	return &i
}

// fix formatting for test vectors
// go's json encoder defaults to encode []byte as base64 encoded string

func (sig Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(sig[:]))
}

func (accs Accounts) MarshalJSON() ([]byte, error) {
	// Convert account/userWallet addresses to hex strings for JSON encoding
	hexAccs := make(map[string]Account, len(accs))
	for addr, acc := range accs {
		hexAccs[hex.EncodeToString(addr[:])] = acc
	}
	return json.Marshal(hexAccs)
}

func (addr EthereumAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(addr[:]))
}

func (pub PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(pub[:]))
}

func (patch PatchPath) MarshalJSON() ([]byte, error) {
	path := []any{patch.Type}
	if patch.ObjectID != nil {
		path = append(path, *patch.ObjectID)
	} else if patch.AccountID != nil {
		path = append(path, *patch.AccountID)
	}
	for _, field := range patch.Fields {
		path = append(path, field)
	}
	return json.Marshal(path)
}
