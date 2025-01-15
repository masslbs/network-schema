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

	"github.com/datatrails/go-datatrails-merklelog/mmr"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	"github.com/masslbs/go-pgmmr"
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
	patchSet.Header.RootHash, _, err = rootHash(t, patchSet.Patches)
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
			tc.RootHash, tree, err = rootHash(t, tc.Patches)
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

func rootHash(t testing.TB, patches []Patch) (Hash, pgmmr.VerifierTree, error) {
	r := require.New(t)
	sz := mmr.FirstMMRSize(uint64(len(patches)))

	tree := pgmmr.NewInMemoryVerifierTree(sha3.NewLegacyKeccak256(), sz)
	for _, patch := range patches {
		data, err := Marshal(patch)
		r.NoError(err)
		_, err = tree.Add(data)
		r.NoError(err)
	}

	// fill up the tree to the next power of 2
	cnt, err := tree.LeafCount()
	require.NoError(t, err)
	nextSquare := nextPowerOf2(cnt)
	t.Logf("sz: %d for %d patches. Tree Size: %d", sz, len(patches), nextSquare)
	for cnt < nextSquare {
		_, err = tree.Add([]byte{})
		r.NoError(err)
		cnt, err = tree.LeafCount()
		r.NoError(err)
	}

	root, err := tree.Root()
	r.NoError(err)
	return Hash(root), tree, nil
}


// * n--: First decrements n by 1. This is done to handle the case where n is already a power of 2.
// * The series of bit-shifting operations (|= with right shifts):
//    This sequence "fills" all the bits to the right of the highest set bit with 1s. For example:
//    If n = 00100000, after these operations it becomes 00111111
// * n++: Finally increments n by 1, which gives us the next power of 2.
//
// Here's a concrete example:
// Start with n = 33 (00100001 in binary)
// After n--, n = 32 (00100000)
// After bit-shifting operations, n = 00111111
// After n++, n = 01000000 (64 in decimal)
func nextPowerOf2(n uint64) uint64 {
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

func TestNextPowerOf2(t *testing.T) {
	require.EqualValues(t, nextPowerOf2(1), 1)
	require.EqualValues(t, nextPowerOf2(2), 2)
	require.EqualValues(t, nextPowerOf2(3), 4)
	require.EqualValues(t, nextPowerOf2(4), 4)
	require.EqualValues(t, nextPowerOf2(5), 8)
	require.EqualValues(t, nextPowerOf2(6), 8)
	require.EqualValues(t, nextPowerOf2(7), 8)
	require.EqualValues(t, nextPowerOf2(8), 8)
	require.EqualValues(t, nextPowerOf2(9), 16)
	require.EqualValues(t, nextPowerOf2(16), 16)
	require.EqualValues(t, nextPowerOf2(17), 32)
	require.EqualValues(t, nextPowerOf2(32), 32)
	require.EqualValues(t, nextPowerOf2(33), 64)
	require.EqualValues(t, nextPowerOf2(64), 64)
	require.EqualValues(t, nextPowerOf2(65), 128)
	require.EqualValues(t, nextPowerOf2(128), 128)
	require.EqualValues(t, nextPowerOf2(256), 256)
	require.EqualValues(t, nextPowerOf2(257), 512)
	require.EqualValues(t, nextPowerOf2(512), 512)
	require.EqualValues(t, nextPowerOf2(513), 1024)
	require.EqualValues(t, nextPowerOf2(1024), 1024)
	require.EqualValues(t, nextPowerOf2(1025), 2048)
	require.EqualValues(t, nextPowerOf2(2048), 2048)

}
