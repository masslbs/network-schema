// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/masslbs/go-pgmmr"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
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

func initVectors(t *testing.T, vectors *vectorFileOkay, shopID Uint256) ethKeyPair {
	kp, seed := newTestKeyPair(t)
	vectors.Signer.Secret = seed
	vectors.Signer.Address = kp.Wallet()

	vectors.PatchSet.Header.ShopID = shopID
	vectors.PatchSet.Header.KeyCardNonce = kcNonce
	kcNonce++
	vectors.PatchSet.Header.Timestamp = time.Unix(0, 0).UTC()
	return kp
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
	copy(pk[:], hash(b))
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

// signing the vectors, also returns the seed
func newTestKeyPair(t *testing.T) (ethKeyPair, []byte) {
	priv, err := crypto.GenerateKey()
	require.NoError(t, err)
	return ethKeyPair{secret: priv}, crypto.FromECDSA(priv)
}

// copied from relay
type ethKeyPair struct {
	secret *ecdsa.PrivateKey
}

func (kp ethKeyPair) Wallet() EthereumAddress {
	publicKey := kp.secret.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	return EthereumAddress(crypto.PubkeyToAddress(*publicKeyECDSA))
}

func (kp ethKeyPair) PublicKey() PublicKey {
	publicKey := kp.secret.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	return PublicKey(crypto.CompressPubkey(publicKeyECDSA))
}

func (kp ethKeyPair) CompressedPubKey() []byte {
	return crypto.CompressPubkey(&kp.secret.PublicKey)
}

func (kp ethKeyPair) Sign(data []byte) ([]byte, error) {
	sighash := accounts.TextHash(data)
	signature, err := crypto.Sign(sighash, kp.secret)
	if err != nil {
		return nil, fmt.Errorf("crypto.Sign failed: %w", err)
	}
	return signature, nil
}

func (kp ethKeyPair) TestSign(t testing.TB, data []byte) Signature {
	signature, err := kp.Sign(data)
	require.NoError(t, err)
	if n := len(signature); n != 65 {
		panic(fmt.Sprintf("signature length is not 65: %d", n))
	}
	signature[64] += 27
	var sig Signature
	copy(sig[:], signature)
	return sig
}

func (kp ethKeyPair) TestSignPatchSet(t testing.TB, patchSet *SignedPatchSet) {
	r := require.New(t)
	r.Greater(len(patchSet.Patches), 0)

	var err error
	patchSet.Header.RootHash, _, err = RootHash(patchSet.Patches)
	r.NoError(err)

	// sign the header
	headerEncoded, err := Marshal(patchSet.Header)
	r.NoError(err)
	patchSet.Signature = kp.TestSign(t, headerEncoded)
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
		objID, vars := bytesToCombinedID(id)
		mapKey := strconv.FormatUint(uint64(objID), 10)
		if len(vars) > 0 {
			mapKey += ":" + strings.Join(vars, "-")
		}
		stringID[mapKey] = inv
		return true
	})
	return json.Marshal(stringID)
}

// use default json encoding for ethereum addresses
func (addr EthereumAddress) MarshalJSON() ([]byte, error) {
	common := common.Address(addr)
	return json.Marshal(common)
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

func TestGenerateVectorsMerkleProofs(t *testing.T) {
	type testCase struct {
		Name     string
		Patches  []Patch
		RootHash Hash
		Proofs   []pgmmr.Proof
	}

	_, listing := newTestListing()
	encodedListing := mustEncode(t, listing)

	vectors := []testCase{
		{
			Name: "SinglePatch",
			Patches: []Patch{
				{
					Op: AddOp,
					Path: PatchPath{
						Type:     ObjectTypeListing,
						ObjectID: uint64ptr(1),
					},
					Value: encodedListing,
				},
			},
		},
		{
			Name: "TwoPatches",
			Patches: []Patch{
				{
					Op: AddOp,
					Path: PatchPath{
						Type:     ObjectTypeListing,
						ObjectID: uint64ptr(1),
					},
					Value: encodedListing,
				},
				{
					Op: AddOp,
					Path: PatchPath{
						Type:     ObjectTypeListing,
						ObjectID: uint64ptr(2),
					},
					Value: encodedListing,
				},
			},
		},
		{
			Name: "ThreePatches",
			Patches: []Patch{
				{
					Op: AddOp,
					Path: PatchPath{
						Type:     ObjectTypeListing,
						ObjectID: uint64ptr(1),
					},
					Value: encodedListing,
				},
				{
					Op: AddOp,
					Path: PatchPath{
						Type:     ObjectTypeListing,
						ObjectID: uint64ptr(2),
					},
					Value: encodedListing,
				},
				{
					Op: ReplaceOp,
					Path: PatchPath{
						Type:     ObjectTypeListing,
						ObjectID: uint64ptr(1),
					},
					Value: encodedListing,
				},
			},
		},

		{
			Name: "FourPatches",
			Patches: slices.Repeat([]Patch{
				{
					Op: AddOp,
					Path: PatchPath{
						Type:     ObjectTypeListing,
						ObjectID: uint64ptr(4),
					},
					Value: encodedListing,
				},
			}, 4),
		},

		{
			Name: "FivePatches",
			Patches: slices.Repeat([]Patch{
				{
					Op: AddOp,
					Path: PatchPath{
						Type:     ObjectTypeListing,
						ObjectID: uint64ptr(5),
					},
					Value: encodedListing,
				},
			}, 5),
		},

		{
			Name: "SixteenPatches",
			Patches: slices.Repeat([]Patch{
				{
					Op: AddOp,
					Path: PatchPath{
						Type:     ObjectTypeListing,
						ObjectID: uint64ptr(16),
					},
					Value: encodedListing,
				},
			}, 16),
		},
	}

	// Process each test case to generate merkle roots and proofs
	for idx := range vectors {
		t.Run(vectors[idx].Name, func(t *testing.T) {
			tc := &vectors[idx]

			// Store root hash
			var err error
			var tree pgmmr.VerifierTree
			tc.RootHash, tree, err = RootHash(tc.Patches)
			require.NoError(t, err)

			// Generate and store proofs for each patch
			tc.Proofs = make([]pgmmr.Proof, len(tc.Patches))
			for i := range tc.Patches {
				proof, err := tree.MakeProof(uint64(i))
				require.NoError(t, err)
				require.NotNil(t, proof)
				tc.Proofs[i] = *proof
				err = tree.VerifyProof(*proof)
				require.NoError(t, err)
			}
		})
	}

	// Write test vectors to file
	writeVectors(t, vectors)
}

func TestNextPowerOf2(t *testing.T) {
	require.EqualValues(t, NextPowerOf2(1), 1)
	require.EqualValues(t, NextPowerOf2(2), 2)
	require.EqualValues(t, NextPowerOf2(3), 4)
	require.EqualValues(t, NextPowerOf2(4), 4)
	require.EqualValues(t, NextPowerOf2(5), 8)
	require.EqualValues(t, NextPowerOf2(6), 8)
	require.EqualValues(t, NextPowerOf2(7), 8)
	require.EqualValues(t, NextPowerOf2(8), 8)
	require.EqualValues(t, NextPowerOf2(9), 16)
	require.EqualValues(t, NextPowerOf2(16), 16)
	require.EqualValues(t, NextPowerOf2(17), 32)
	require.EqualValues(t, NextPowerOf2(32), 32)
	require.EqualValues(t, NextPowerOf2(33), 64)
	require.EqualValues(t, NextPowerOf2(64), 64)
	require.EqualValues(t, NextPowerOf2(65), 128)
	require.EqualValues(t, NextPowerOf2(128), 128)
	require.EqualValues(t, NextPowerOf2(256), 256)
	require.EqualValues(t, NextPowerOf2(257), 512)
	require.EqualValues(t, NextPowerOf2(512), 512)
	require.EqualValues(t, NextPowerOf2(513), 1024)
	require.EqualValues(t, NextPowerOf2(1024), 1024)
	require.EqualValues(t, NextPowerOf2(1025), 2048)
	require.EqualValues(t, NextPowerOf2(2048), 2048)
	require.EqualValues(t, NextPowerOf2(2049), 4096)
	require.EqualValues(t, NextPowerOf2(4096), 4096)
	require.EqualValues(t, NextPowerOf2(4097), 8192)
	require.EqualValues(t, NextPowerOf2(8192), 8192)
	require.EqualValues(t, NextPowerOf2(8193), 16384)
	require.EqualValues(t, NextPowerOf2(16384), 16384)
	require.EqualValues(t, NextPowerOf2(16385), 32768)
	require.EqualValues(t, NextPowerOf2(1<<60), 1<<60)
	require.EqualValues(t, NextPowerOf2(1<<60+1), 1<<61)
	require.EqualValues(t, NextPowerOf2((1<<61)-1), 1<<61)
	require.EqualValues(t, NextPowerOf2(1<<61), 1<<61)
	require.EqualValues(t, NextPowerOf2(1<<61+1), 1<<62)
	require.EqualValues(t, NextPowerOf2(1<<62), 1<<62)
}
