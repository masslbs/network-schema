// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
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
	"strconv"
	"strings"
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

func writeVectors(t *testing.T, vectors any) {
	if t.Failed() {
		t.Logf("skipping vector write due to test failure")
		return
	}
	r := require.New(t)
	baseName := strings.TrimPrefix(t.Name(), "TestGenerateVectors")
	tempFile := openTestFile(t, baseName+".json")
	jsonEnc := json.NewEncoder(tempFile)
	jsonEnc.SetIndent("", "  ")
	err := jsonEnc.Encode(vectors)
	r.NoError(err)
	r.NoError(tempFile.Close())
	tempFile = openTestFile(t, baseName+".cbor")
	enc := DefaultEncoder(tempFile)
	err = enc.Encode(vectors)
	r.NoError(err)
	r.NoError(tempFile.Close())
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

func TestCombinedID(t *testing.T) {
	id := ObjectId(1)
	buf := combinedIDtoBytes(id, nil)
	id, variations := bytesToCombinedID(buf)
	require.Equal(t, id, ObjectId(1))
	require.Equal(t, variations, []string{})

	// test with variations
	id = ObjectId(2)
	variations = []string{"a", "b", "c"}
	buf = combinedIDtoBytes(id, variations)
	id, variations = bytesToCombinedID(buf)
	require.Equal(t, id, ObjectId(2))
	require.Equal(t, variations, []string{"a", "b", "c"})
}

// fix formatting for test vectors
// go's json encoder defaults to encode []byte as base64 encoded string

func (sig Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(sig[:]))
}

func (accs Accounts) MarshalJSON() ([]byte, error) {
	// Convert account/userWallet addresses to hex strings for JSON compatible map keys
	hexAccs := make(map[string]Account, accs.Size())
	accs.All(func(addr []byte, acc Account) bool {
		hexAccs[hex.EncodeToString(addr)] = acc
		return true
	})
	return json.Marshal(hexAccs)
}

func (lis Listings) MarshalJSON() ([]byte, error) {
	hexLis := make(map[ObjectId]Listing, lis.Size())
	lis.All(func(id []byte, lis Listing) bool {
		hexLis[bytesToId(id)] = lis
		return true
	})
	return json.Marshal(hexLis)
}

func (inv Inventory) MarshalJSON() ([]byte, error) {
	stringID := make(map[string]uint64, inv.Size())
	inv.All(func(id []byte, inv uint64) bool {
		objId, vars := bytesToCombinedID(id)
		mapKey := strconv.FormatUint(uint64(objId), 10)
		if len(vars) > 0 {
			mapKey += ":" + strings.Join(vars, "-")
		}
		stringID[mapKey] = inv
		return true
	})
	return json.Marshal(stringID)
}

func (addr EthereumAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(addr[:]))
}

func (pub PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(pub[:]))
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(h[:]))
}

func (patch PatchPath) MarshalJSON() ([]byte, error) {
	path := []any{patch.Type}
	var either bool
	if patch.ObjectID != nil {
		path = append(path, *patch.ObjectID)
		either = true
	} else if patch.AccountID != nil {
		path = append(path, *patch.AccountID)
		either = true
	} else if patch.TagName != nil {
		path = append(path, *patch.TagName)
		either = true
	}
	if !either && patch.Type != ObjectTypeManifest {
		return nil, fmt.Errorf("either ObjectID, TagName or AccountID must be set")
	}
	for _, field := range patch.Fields {
		path = append(path, field)
	}
	return json.Marshal(path)
}
