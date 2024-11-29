// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"
)

func TestPatchPath(t *testing.T) {

	r := require.New(t)
	patch := Patch{
		Op:   OpString(t.Name()),
		Path: PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "color", "variations", "pink"}},
	}

	data, err := Marshal(patch)
	r.NoError(err)
	t.Log("CBOR encoded:")
	t.Log("\n" + pretty(data))

	var rxPatch Patch
	err = Unmarshal(data, &rxPatch)
	r.NoError(err)
	r.Equal("listing", rxPatch.Path.Type)
	r.Equal(ObjectId(1), rxPatch.Path.ID)
	r.Equal([]string{"options", "color", "variations", "pink"}, rxPatch.Path.Fields)

	var testPath struct {
		Path []any
	}
	err = Unmarshal(data, &testPath)
	r.NoError(err)
	r.Equal([]any{"listing", uint64(1), "options", "color", "variations", "pink"}, testPath.Path)
}

func TestPatchAdd(t *testing.T) {
	var (
		err error
		buf bytes.Buffer
		r   = require.New(t)
		enc = DefaultEncoder(&buf)
	)

	var createListing Patch
	createListing.Op = AddOp
	createListing.Path = PatchPath{Type: "listing", ID: 1}

	lis := testListing()
	createListing.Value, err = Marshal(lis)
	r.NoError(err)
	err = enc.Encode(createListing)
	r.NoError(err)

	opData := buf.Bytes()
	t.Log("OP encoded:")
	t.Log("\n" + pretty(opData))

	dec := DefaultDecoder(bytes.NewReader(opData))
	var rxOp Patch
	err = dec.Decode(&rxOp)
	r.NoError(err)
	r.Equal("listing", rxOp.Path.Type)
	r.Equal(ObjectId(1), rxOp.Path.ID)
	r.NoError(validate.Struct(rxOp))

	dec = DefaultDecoder(bytes.NewReader(rxOp.Value))
	var rxLis Listing
	err = dec.Decode(&rxLis)
	r.NoError(err)

	t.Logf("listing received: %+v", rxLis)
	r.EqualValues(lis, rxLis)
}

func testListing() Listing {
	var lis Listing
	lis.ID = 1
	lis.ViewState = ListingViewStatePublished
	lis.Metadata.Title = "test Listing"
	lis.Metadata.Description = "short desc"
	lis.Metadata.Images = []string{"https://http.cat/images/100.jpg"}
	price := big.NewInt(12345)
	lis.Price = *price
	lis.Options = ListingOptions{
		"color": {
			Title: "Color",
			Variations: map[string]ListingVariation{
				"r": {
					ID: 1,
					VariationInfo: ListingMetadata{
						Title:       "Red",
						Description: "Red color",
					},
				},
				"b": {
					ID: 2,
					VariationInfo: ListingMetadata{
						Title:       "Blue",
						Description: "Blue color",
					},
				},
			},
		},
	}
	lis.StockStatuses = []ListingStockStatus{
		{
			VariationIDs: []ObjectId{1},
			InStock:      boolptr(true),
		},
	}
	return lis
}

func TestPatchListing(t *testing.T) {
	testColorOption := ListingOption{
		Title: "Color",
		Variations: map[string]ListingVariation{
			"pink": {
				ID: 333,
				VariationInfo: ListingMetadata{
					Title:       "Pink",
					Description: "Pink color",
				},
			},
			"orange": {
				ID: 444,
				VariationInfo: ListingMetadata{
					Title:       "Orange",
					Description: "Orange color",
				},
			},
		},
	}

	testSizeOption := ListingOption{
		Title: "Size",
		Variations: map[string]ListingVariation{
			"s": {
				ID: 33,
				VariationInfo: ListingMetadata{
					Title:       "Small",
					Description: "Small size",
				},
			},
			"m": {
				ID: 44,
				VariationInfo: ListingMetadata{
					Title:       "Medium",
					Description: "Medium size",
				},
			},
		},
	}
	testTimeFuture := time.Unix(10000000000, 0)

	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    interface{}
		expected func(*assert.Assertions, Listing)
	}{
		{
			name:  "replace price",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"price"}},
			value: *big.NewInt(66666),
			expected: func(r *assert.Assertions, l Listing) {
				r.Equal(*big.NewInt(66666), l.Price)
			},
		},
		{
			name:  "replace description",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"metadata", "description"}},
			value: "new description",
			expected: func(r *assert.Assertions, l Listing) {
				r.Equal("new description", l.Metadata.Description)
			},
		},
		{
			name:  "replace whole metadata",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"metadata"}},
			value: testListing().Metadata,
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(testListing().Metadata, l.Metadata)
			},
		},
		{
			name:  "append an image",
			op:    AddOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"metadata", "images", "-"}},
			value: "https://http.cat/images/200.jpg",
			expected: func(a *assert.Assertions, l Listing) {
				if !a.Len(l.Metadata.Images, 2) {
					return
				}
				a.Equal("https://http.cat/images/100.jpg", l.Metadata.Images[0])
				a.Equal("https://http.cat/images/200.jpg", l.Metadata.Images[1])
			},
		},
		{
			name:  "prepend an image",
			op:    AddOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"metadata", "images", "0"}},
			value: "https://http.cat/images/200.jpg",
			expected: func(a *assert.Assertions, l Listing) {
				if !a.Len(l.Metadata.Images, 2) {
					return
				}
				a.Equal("https://http.cat/images/200.jpg", l.Metadata.Images[0])
				a.Equal("https://http.cat/images/100.jpg", l.Metadata.Images[1])
			},
		},
		{
			name: "remove an image",
			op:   RemoveOp,
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"metadata", "images", "0"}},
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal([]string{}, l.Metadata.Images)
			},
		},
		{
			name:  "replace view state",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"viewState"}},
			value: ListingViewStatePublished,
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(ListingViewStatePublished, l.ViewState)
			},
		},

		{
			name: "append a stock status",
			op:   AddOp,
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"stockStatuses", "-"}},
			value: ListingStockStatus{
				VariationIDs: []ObjectId{2},
				InStock:      boolptr(true),
			},
			expected: func(a *assert.Assertions, l Listing) {
				if !a.Len(l.StockStatuses, 2) {
					return
				}
				stockStatus := l.StockStatuses[1]
				a.Equal([]ObjectId{2}, stockStatus.VariationIDs)
				a.True(*stockStatus.InStock)
			},
		},
		{
			name: "prepend a stock status",
			op:   AddOp,
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"stockStatuses", "0"}},
			value: ListingStockStatus{
				VariationIDs: []ObjectId{23},
				InStock:      boolptr(true),
			},
			expected: func(a *assert.Assertions, l Listing) {
				if !a.Len(l.StockStatuses, 2) {
					return
				}
				stockStatus := l.StockStatuses[0]
				a.Equal([]ObjectId{23}, stockStatus.VariationIDs)
				a.True(*stockStatus.InStock)
			},
		},
		{
			name: "replace stock status",
			op:   ReplaceOp,
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"stockStatuses", "0"}},
			value: ListingStockStatus{
				VariationIDs: []ObjectId{1},
				InStock:      boolptr(false),
			},
			expected: func(a *assert.Assertions, l Listing) {
				if !a.Len(l.StockStatuses, 1) {
					return
				}
				stockStatus := l.StockStatuses[0]
				a.Equal([]ObjectId{1}, stockStatus.VariationIDs)
				a.False(*stockStatus.InStock)
			},
		},
		{
			name:  "replace expectedInStockBy",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"stockStatuses", "0", "expectedInStockBy"}},
			value: testTimeFuture,
			expected: func(a *assert.Assertions, l Listing) {
				a.Nil(l.StockStatuses[0].InStock)
				a.Equal(testTimeFuture, *l.StockStatuses[0].ExpectedInStockBy)
			},
		},
		{
			name:  "replace inStock",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"stockStatuses", "0", "inStock"}},
			value: true,
			expected: func(a *assert.Assertions, l Listing) {
				a.True(*l.StockStatuses[0].InStock)
				a.Nil(l.StockStatuses[0].ExpectedInStockBy)
			},
		},
		{
			name: "remove stock status",
			op:   RemoveOp,
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"stockStatuses", "0"}},
			expected: func(a *assert.Assertions, l Listing) {
				a.Len(l.StockStatuses, 0)
			},
		},

		// map manipulation of Options
		{
			name:  "add an option",
			op:    AddOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "size"}},
			value: testSizeOption,
			expected: func(a *assert.Assertions, l Listing) {
				sizeOption, ok := l.Options["size"]
				a.True(ok)
				a.Equal(testSizeOption, sizeOption)
			},
		},
		{
			name:  "replace one option",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "color"}},
			value: testColorOption,
			expected: func(a *assert.Assertions, l Listing) {
				colorOption, ok := l.Options["color"]
				a.True(ok)
				a.Equal(testColorOption, colorOption)
			},
		},
		{
			name: "replace whole options",
			op:   ReplaceOp,
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"options"}},
			value: ListingOptions{
				"color": testColorOption,
			},
			expected: func(a *assert.Assertions, l Listing) {
				colorOption, ok := l.Options["color"]
				a.True(ok)
				a.Equal(testColorOption, colorOption)
			},
		},
		{
			name: "remove an option",
			op:   RemoveOp,
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "color"}},
			expected: func(a *assert.Assertions, l Listing) {
				_, ok := l.Options["color"]
				a.False(ok)
			},
		},
		{
			name:  "add a variation to an option",
			op:    AddOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "color", "variations", "pink"}},
			value: testColorOption.Variations["pink"],
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(testColorOption.Variations["pink"], l.Options["color"].Variations["pink"])
			},
		},
		{
			name:  "replace title of an option",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "color", "title"}},
			value: "FARBE",
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal("FARBE", l.Options["color"].Title)
			},
		},
		{
			name:  "replace variations of an option",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "color", "variations", "b"}},
			value: testColorOption.Variations["pink"],
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(testColorOption.Variations["pink"], l.Options["color"].Variations["b"])
			},
		},
		{
			name:  "replace one variation's info",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "color", "variations", "b", "variationInfo"}},
			value: testColorOption.Variations["pink"].VariationInfo,
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(testColorOption.Variations["pink"].VariationInfo, l.Options["color"].Variations["b"].VariationInfo)
			},
		},
		{
			name: "remove a variation from an option",
			op:   RemoveOp,
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "color", "variations", "b"}},
			expected: func(a *assert.Assertions, l Listing) {
				_, ok := l.Options["color"].Variations["b"]
				a.False(ok)
			},
		},
	}

	type vectorEntry struct {
		Patch   Patch
		Value   Listing
		Encoded []byte
		Hash    []byte
	}
	type vectorFile struct {
		Encoded []byte
		Value   Listing
		Hash    []byte
		Patches []vectorEntry
	}
	var vectors vectorFile
	var err error
	vectors.Value = testListing()
	vectors.Encoded, err = Marshal(vectors.Value)
	require.NoError(t, err)
	vectors.Hash = hash(vectors.Encoded)

	var patcher Patcher
	patcher.validator = validate
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			a := assert.New(t)

			lis := testListing()

			// round trip to make sure we can encode/decode the patch
			patch := createPatch(tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			r.Equal(tc.op, decodedPatch.Op)

			err := patcher.Listing(&lis, decodedPatch)
			r.NoError(err)
			tc.expected(a, lis)

			var entry vectorEntry
			entry.Patch = patch
			entry.Value = lis
			entry.Encoded, err = Marshal(lis)
			require.NoError(t, err)
			entry.Hash = hash(entry.Encoded)
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	if !t.Failed() {
		tempFile, err := os.Create("vectors_patch_listing.json")
		require.NoError(t, err)
		// enc := DefaultEncoder(tempFile)
		err = json.NewEncoder(tempFile).Encode(vectors)
		require.NoError(t, err)
		tempFile.Close()
		tempFile, err = os.Create("vectors_patch_listing.cbor")
		require.NoError(t, err)
		enc := DefaultEncoder(tempFile)
		err = enc.Encode(vectors)
		require.NoError(t, err)
		tempFile.Close()
	}
}

func testManifest() Manifest {
	return Manifest{
		ShopId: *big.NewInt(1),
		Payees: map[string]Payee{
			"default": {
				CallAsContract: false,
				Address: ChainAddress{
					ChainID: 1337,
					Address: EthereumAddress([20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}),
				},
			},
		},
		AcceptedCurrencies: []ChainAddress{
			{
				ChainID: 1337,
				Address: EthereumAddress{},
			},
			{
				ChainID: 1337,
				Address: EthereumAddress([20]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
		PricingCurrency: ChainAddress{
			ChainID: 1337,
			Address: EthereumAddress{},
		},
		ShippingRegions: map[string]ShippingRegion{
			"default": {
				Country: "DE",
			},
		},
	}
}

func TestPatchManifest(t *testing.T) {

	testPayee := Payee{
		CallAsContract: true,
		Address: ChainAddress{
			ChainID: 1337,
			Address: EthereumAddress([20]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
		},
	}
	testCurrency := ChainAddress{
		ChainID: 1337,
		Address: EthereumAddress([20]byte{0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00}),
	}

	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    interface{}
		expected func(*assert.Assertions, Manifest)
	}{
		// simple field mutations
		{
			name:  "replace pricing currency",
			op:    ReplaceOp,
			path:  PatchPath{Type: "manifest", ID: 0, Fields: []string{"pricingCurrency"}},
			value: testCurrency,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Equal(testCurrency, m.PricingCurrency)
			},
		},

		// array mutations
		{
			name:  "append accepted currency",
			op:    AddOp,
			path:  PatchPath{Type: "manifest", ID: 0, Fields: []string{"acceptedCurrencies", "-"}},
			value: testCurrency,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.AcceptedCurrencies, 3)
				a.Equal(testCurrency, m.AcceptedCurrencies[2])
			},
		},
		{
			name:  "insert accepted currency at index",
			op:    AddOp,
			path:  PatchPath{Type: "manifest", ID: 0, Fields: []string{"acceptedCurrencies", "0"}},
			value: testCurrency,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.AcceptedCurrencies, 3)
				a.Equal(testCurrency, m.AcceptedCurrencies[0])
			},
		},
		{
			name:  "replace accepted currency",
			op:    ReplaceOp,
			path:  PatchPath{Type: "manifest", ID: 0, Fields: []string{"acceptedCurrencies", "0"}},
			value: testCurrency,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Equal(testCurrency, m.AcceptedCurrencies[0])
			},
		},
		{
			name: "remove accepted currency",
			op:   RemoveOp,
			path: PatchPath{Type: "manifest", ID: 0, Fields: []string{"acceptedCurrencies", "0"}},
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.AcceptedCurrencies, 1)
			},
		},

		// map mutations
		{
			name:  "replace payee",
			op:    ReplaceOp,
			path:  PatchPath{Type: "manifest", ID: 0, Fields: []string{"payees", "default"}},
			value: testPayee,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Equal(testPayee, m.Payees["default"])
			},
		},
		{
			name:  "add a payee",
			op:    AddOp,
			path:  PatchPath{Type: "manifest", ID: 0, Fields: []string{"payees", "test"}},
			value: testPayee,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.Payees, 2)
				a.Equal(testPayee, m.Payees["test"])
			},
		},
		{
			name: "remove a payee",
			op:   RemoveOp,
			path: PatchPath{Type: "manifest", ID: 0, Fields: []string{"payees", "default"}},
			expected: func(a *assert.Assertions, m Manifest) {
				_, ok := m.Payees["default"]
				a.False(ok)
			},
		},
		{
			name:  "add a shipping region",
			op:    AddOp,
			path:  PatchPath{Type: "manifest", ID: 0, Fields: []string{"shippingRegions", "germany"}},
			value: ShippingRegion{Country: "DE"},
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.ShippingRegions, 2)
				a.Equal("DE", m.ShippingRegions["germany"].Country)
			},
		},
		{
			name:  "replace a shipping region",
			op:    ReplaceOp,
			path:  PatchPath{Type: "manifest", ID: 0, Fields: []string{"shippingRegions", "default"}},
			value: ShippingRegion{Country: "DE"},
			expected: func(a *assert.Assertions, m Manifest) {
				a.Equal("DE", m.ShippingRegions["default"].Country)
			},
		},
		{
			name: "remove a shipping region",
			op:   RemoveOp,
			path: PatchPath{Type: "manifest", ID: 0, Fields: []string{"shippingRegions", "default"}},
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.ShippingRegions, 0)
				_, ok := m.ShippingRegions["default"]
				a.False(ok)
			},
		},
	}

	type vectorEntry struct {
		Patch   Patch
		Value   Manifest
		Hash    []byte
		Encoded []byte
	}
	type vector struct {
		Value   Manifest
		Encoded []byte
		Hash    []byte
		Patches []vectorEntry
	}
	var err error
	var vectors vector
	vectors.Value = testManifest()
	vectors.Encoded, err = Marshal(vectors.Value)
	require.NoError(t, err)
	vectors.Hash = hash(vectors.Encoded)

	var patcher Patcher
	patcher.validator = validate
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := testManifest()
			r := require.New(t)
			a := assert.New(t)

			patch := createPatch(tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			r.Equal(tc.op, decodedPatch.Op)

			err = patcher.Manifest(&manifest, decodedPatch)
			r.NoError(err)
			tc.expected(a, manifest)

			var entry vectorEntry
			entry.Patch = patch
			entry.Value = manifest
			entry.Encoded, err = Marshal(manifest)
			require.NoError(t, err)
			entry.Hash = hash(entry.Encoded)
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	if !t.Failed() {
		tempFile, err := os.Create("vectors_patch_manifest.json")
		require.NoError(t, err)
		// enc := DefaultEncoder(tempFile)
		err = json.NewEncoder(tempFile).Encode(vectors)
		require.NoError(t, err)
		tempFile.Close()
		tempFile, err = os.Create("vectors_patch_manifest.cbor")
		require.NoError(t, err)
		enc := DefaultEncoder(tempFile)
		err = enc.Encode(vectors)
		require.NoError(t, err)
		tempFile.Close()
	}
}

// utility functions

func createPatch(op OpString, path PatchPath, value interface{}) Patch {
	encodedValue, err := Marshal(value)
	check(err)
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

// fix formatting for test vectors
// go defaults to encode []byte as base64 encoded string

func (sig Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(sig[:]))
}

func (adde EthereumAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(adde[:]))
}
