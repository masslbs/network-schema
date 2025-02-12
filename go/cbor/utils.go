// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/datatrails/go-datatrails-merklelog/mmr"
	"github.com/fxamacker/cbor/v2"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	"github.com/masslbs/go-pgmmr"
	"golang.org/x/crypto/sha3"
)

func MassMarketTags() cbor.TagSet {
	tags := cbor.NewTagSet()

	// Register tag for Status enum type (using tag 1000)
	// tags.Add(
	// 	cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
	// 	reflect.TypeOf(ListingViewState(0)),
	// 	1000,
	// )
	return tags
}

func DefaultDecoder(rd io.Reader) cbor.Decoder {
	opts := cbor.DecOptions{
		BinaryUnmarshaler: cbor.BinaryUnmarshalerByteString,
	}
	mode, err := opts.DecModeWithTags(MassMarketTags())
	check(err)
	return *mode.NewDecoder(rd)
}

func Unmarshal(data []byte, v interface{}) error {
	dec := DefaultDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}

func DefaultEncoder(w io.Writer) *cbor.Encoder {
	opts := cbor.CanonicalEncOptions()
	opts.BigIntConvert = cbor.BigIntConvertShortest
	opts.Time = cbor.TimeRFC3339
	mode, err := opts.EncModeWithTags(MassMarketTags())
	check(err)
	return mode.NewEncoder(w)
}

func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)
	err := enc.Encode(v)
	return buf.Bytes(), err
}

func DefaultValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("notblank", validators.NotBlank)
	validate.RegisterAlias("nonEmptyMapKeys", "dive,keys,required,notblank,endkeys,required")
	validate.RegisterStructValidation(OrderValidation, Order{})
	// we cant "nonEmptyMapKeys" via struct tags, since the library cant iterate through the HAMT
	validate.RegisterStructValidation(HAMTValidation, Tags{}, Accounts{}, Listings{}, Orders{})
	return validate
}

func bytesToId(buf []byte) ObjectId {
	if len(buf) != 8 {
		panic(fmt.Sprintf("expected 8 bytes, got %d", len(buf)))
	}
	return ObjectId(binary.BigEndian.Uint64(buf))
}

func combinedIDtoBytes(id ObjectId, variations []string) []byte {
	buf := idToBytes(id)
	sort.Strings(variations)
	buf = append(buf, []byte(strings.Join(variations, "|"))...)
	return buf
}

func bytesToCombinedID(buf []byte) (ObjectId, []string) {
	id := bytesToId(buf[:8])
	variations := strings.Split(string(buf[8:]), "|")
	if len(variations) == 1 && variations[0] == "" {
		variations = []string{}
	}
	return id, variations
}

func RootHash(patches []Patch) (Hash, pgmmr.VerifierTree, error) {
	sz := mmr.FirstMMRSize(uint64(len(patches)))

	tree := pgmmr.NewInMemoryVerifierTree(sha3.NewLegacyKeccak256(), sz)
	for _, patch := range patches {
		data, err := Marshal(patch)
		if err != nil {
			return Hash{}, nil, fmt.Errorf("failed to marshal patch: %w", err)
		}
		_, err = tree.Add(data)
		if err != nil {
			return Hash{}, nil, fmt.Errorf("failed to add patch to tree: %w", err)
		}
	}

	// fill up the tree to the next power of 2
	cnt, err := tree.LeafCount()
	if err != nil {
		return Hash{}, nil, fmt.Errorf("failed to get leaf count: %w", err)
	}
	nextSquare := NextPowerOf2(cnt)
	for cnt < nextSquare {
		_, err = tree.Add([]byte{})
		if err != nil {
			return Hash{}, nil, fmt.Errorf("failed to add empty leaf to tree: %w", err)
		}
		cnt, err = tree.LeafCount()
		if err != nil {
			return Hash{}, nil, fmt.Errorf("failed to get leaf count: %w", err)
		}
	}

	root, err := tree.Root()
	if err != nil {
		return Hash{}, nil, fmt.Errorf("failed to get root: %w", err)
	}
	return Hash(root), tree, nil
}

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

func hash(value []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(value)
	return hash.Sum(nil)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func dump(val any) []byte {
	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)

	err := enc.Encode(val)
	check(err)

	fmt.Printf("CBOR of: %+v\n", val)
	data := buf.Bytes()
	fmt.Println(hex.EncodeToString(data))
	return data
}

func diag(val any) {
	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)

	err := enc.Encode(val)
	check(err)

	diagStr, err := cbor.Diagnose(buf.Bytes())
	check(err)

	fmt.Println(diagStr)
}

func pretty(data []byte) string {
	if os.Getenv("PRETTY") == "" {
		return hex.EncodeToString(data)
	}
	shell := exec.Command("cbor2pretty.rb")
	shell.Stdin = bytes.NewReader(data)

	out, err := shell.CombinedOutput()
	check(err)
	return string(out)
}
