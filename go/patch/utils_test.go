// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fxamacker/cbor/v2"
	"github.com/peterldowns/testy/assert"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/masslbs/network-schema/go/internal/testhelper"
	"github.com/masslbs/network-schema/go/objects"
)

var (
	zeroAddress common.Address

	testAddr  = objects.MustAddrFromHex(1, "0x1234567890123456789012345678901234567890")
	testAddr2 = objects.MustAddrFromHex(1, "0x6789012345678901234567890123456789012345")
	testAddr3 = objects.MustAddrFromHex(1, "0x9999999999999999999999999999999999999999")
	testEth   = objects.MustAddrFromHex(1, "0x0000000000000000000000000000000000000000")
	testUsdc  = objects.MustAddrFromHex(1, "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

	testAddr123 = testhelper.CommonEthAddr([20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
)

func TestNextPowerOf2(t *testing.T) {
	assert.Equal(t, NextPowerOf2(1), 1)
	assert.Equal(t, NextPowerOf2(2), 2)
	assert.Equal(t, NextPowerOf2(3), 4)
	assert.Equal(t, NextPowerOf2(4), 4)
	assert.Equal(t, NextPowerOf2(5), 8)
	assert.Equal(t, NextPowerOf2(6), 8)
	assert.Equal(t, NextPowerOf2(7), 8)
	assert.Equal(t, NextPowerOf2(8), 8)
	assert.Equal(t, NextPowerOf2(9), 16)
	assert.Equal(t, NextPowerOf2(16), 16)
	assert.Equal(t, NextPowerOf2(17), 32)
	assert.Equal(t, NextPowerOf2(32), 32)
	assert.Equal(t, NextPowerOf2(33), 64)
	assert.Equal(t, NextPowerOf2(64), 64)
	assert.Equal(t, NextPowerOf2(65), 128)
	assert.Equal(t, NextPowerOf2(128), 128)
	assert.Equal(t, NextPowerOf2(256), 256)
	assert.Equal(t, NextPowerOf2(257), 512)
	assert.Equal(t, NextPowerOf2(512), 512)
	assert.Equal(t, NextPowerOf2(513), 1024)
	assert.Equal(t, NextPowerOf2(1024), 1024)
	assert.Equal(t, NextPowerOf2(1025), 2048)
	assert.Equal(t, NextPowerOf2(2048), 2048)
	assert.Equal(t, NextPowerOf2(2049), 4096)
	assert.Equal(t, NextPowerOf2(4096), 4096)
	assert.Equal(t, NextPowerOf2(4097), 8192)
	assert.Equal(t, NextPowerOf2(8192), 8192)
	assert.Equal(t, NextPowerOf2(8193), 16384)
	assert.Equal(t, NextPowerOf2(16384), 16384)
	assert.Equal(t, NextPowerOf2(16385), 32768)
	assert.Equal(t, NextPowerOf2(1<<60), 1<<60)
	assert.Equal(t, NextPowerOf2(1<<60+1), 1<<61)
	assert.Equal(t, NextPowerOf2((1<<61)-1), 1<<61)
	assert.Equal(t, NextPowerOf2(1<<61), 1<<61)
	assert.Equal(t, NextPowerOf2(1<<61+1), 1<<62)
	assert.Equal(t, NextPowerOf2(1<<62), 1<<62)
}

var validate = objects.DefaultValidator()

func openTestFile(t testing.TB, fileName string) *os.File {
	path := filepath.Join(os.Getenv("TEST_DATA_OUT"), fileName)
	// Check if the target folder exists, and create it if it doesn't
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		t.Fatalf("Failed to create directory for test file: %v", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	assert.Nil(t, err)
	return file
}

func initVectors(t *testing.T, vectors *vectorFileOkay, shopID objects.Uint256) ethKeyPair {
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
	baseName := strings.TrimPrefix(t.Name(), "TestGenerateVectors")
	tempFile := openTestFile(t, baseName+".json")
	jsonEnc := json.NewEncoder(tempFile)
	jsonEnc.SetIndent("", "  ")
	err := jsonEnc.Encode(vectors)
	assert.Nil(t, err)
	assert.Nil(t, tempFile.Close())
	tempFile = openTestFile(t, baseName+".cbor")
	enc := masscbor.DefaultEncoder(tempFile)
	err = enc.Encode(vectors)
	assert.Nil(t, err)
	assert.Nil(t, tempFile.Close())
}

func mustEncode(t *testing.T, v any) cbor.RawMessage {
	data, err := masscbor.Marshal(v)
	assert.Nil(t, err)
	return data
}

func testPubKey(i uint64) objects.PublicKey {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, i)
	var pk objects.PublicKey
	h := sha256.Sum256(b)
	copy(pk[:], h[:])
	return pk
}

func createPatch(t testing.TB, op OpString, path Path, value interface{}) Patch {
	encodedValue, err := masscbor.Marshal(value)
	assert.Nil(t, err)
	return Patch{
		Op:    op,
		Path:  path,
		Value: encodedValue,
	}
}

func decodePatch(t *testing.T, encoded []byte) Patch {
	var decoded Patch
	dec := masscbor.DefaultDecoder(bytes.NewReader(encoded))
	err := dec.Decode(&decoded)
	assert.Nil(t, err)
	assert.Nil(t, validate.Struct(decoded))
	return decoded
}

// signing the vectors, also returns the seed
func newTestKeyPair(t *testing.T) (ethKeyPair, []byte) {
	priv, err := crypto.GenerateKey()
	assert.Nil(t, err)
	return ethKeyPair{secret: priv}, crypto.FromECDSA(priv)
}

// copied from relay
type ethKeyPair struct {
	secret *ecdsa.PrivateKey
}

func (kp ethKeyPair) Wallet() objects.EthereumAddress {
	publicKey := kp.secret.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	commonAddr := crypto.PubkeyToAddress(*publicKeyECDSA)
	return objects.EthereumAddress{Address: commonAddr}
}

func (kp ethKeyPair) PublicKey() objects.PublicKey {
	publicKey := kp.secret.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	return objects.PublicKey(crypto.CompressPubkey(publicKeyECDSA))
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

func (kp ethKeyPair) TestSign(t testing.TB, data []byte) objects.Signature {
	signature, err := kp.Sign(data)
	assert.Nil(t, err)
	if n := len(signature); n != 65 {
		panic(fmt.Sprintf("signature length is not 65: %d", n))
	}
	signature[64] += 27
	var sig objects.Signature
	copy(sig[:], signature)
	return sig
}

func (kp ethKeyPair) TestSignPatchSet(t testing.TB, patchSet *SignedPatchSet) {
	assert.GreaterThan(t, len(patchSet.Patches), 0)

	var err error
	patchSet.Header.RootHash, _, err = RootHash(patchSet.Patches)
	assert.Nil(t, err)

	// sign the header
	headerEncoded, err := masscbor.Marshal(patchSet.Header)
	assert.Nil(t, err)
	patchSet.Signature = kp.TestSign(t, headerEncoded)
}
