// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"
)

func TestPatchObjectIDs(t *testing.T) {
	testAddr := addrFromHex(1, "0x1234567890123456789012345678901234567890").Address

	type testCase struct {
		Path PatchPath
	}

	goodTestCases := []testCase{
		{Path: PatchPath{Type: ObjectTypeManifest}},
		{Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeAccount, AccountID: &testAddr}},
		{Path: PatchPath{Type: ObjectTypeTag, TagName: strptr("test-tag")}},
	}
	for idx, tc := range goodTestCases {
		t.Run(fmt.Sprintf("good-%d", idx), func(t *testing.T) {
			r := require.New(t)
			data, err := Marshal(tc.Path)
			r.NoError(err)

			var got PatchPath
			err = Unmarshal(data, &got)
			r.NoError(err)
			// unset fields for convenience
			got.Fields = nil
			r.EqualValues(tc.Path, got)
		})
	}

	badTestCases := []testCase{
		// missing required field
		{Path: PatchPath{Type: ObjectTypeListing}},
		{Path: PatchPath{Type: ObjectTypeAccount}},
		{Path: PatchPath{Type: ObjectTypeTag}},

		// wrong contextual field
		{Path: PatchPath{Type: ObjectTypeManifest, ObjectID: uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeManifest, AccountID: &testAddr}},

		{Path: PatchPath{Type: ObjectTypeAccount, ObjectID: uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeAccount, TagName: strptr("test-tag")}},

		{Path: PatchPath{Type: ObjectTypeTag, ObjectID: uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeTag, AccountID: &testAddr}},

		{Path: PatchPath{Type: ObjectTypeOrder, TagName: strptr("test-tag")}},
	}
	for idx, tc := range badTestCases {
		t.Run(fmt.Sprintf("bad-%d", idx), func(t *testing.T) {
			r := require.New(t)
			data, err := Marshal(tc.Path)
			r.NoError(err)

			var got PatchPath
			err = Unmarshal(data, &got)
			r.Error(err)
		})
	}
}

func TestPatchPath(t *testing.T) {
	r := require.New(t)
	patch := Patch{
		Op:   OpString(t.Name()),
		Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color", "variations", "pink"}},
	}

	data, err := Marshal(patch)
	r.NoError(err)
	t.Log("CBOR encoded:")
	t.Log("\n" + pretty(data))

	var rxPatch Patch
	err = Unmarshal(data, &rxPatch)
	r.NoError(err)
	r.Equal(ObjectTypeListing, rxPatch.Path.Type)
	r.EqualValues(ObjectId(1), *rxPatch.Path.ObjectID)
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
	createListing.Path = PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1)}

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
	r.Equal(ObjectTypeListing, rxOp.Path.Type)
	r.EqualValues(ObjectId(1), *rxOp.Path.ObjectID)
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
					VariationInfo: ListingMetadata{
						Title:       "Red",
						Description: "Red color",
					},
				},
				"b": {

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
			VariationIDs: []string{"r"},
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

				VariationInfo: ListingMetadata{
					Title:       "Pink",
					Description: "Pink color",
				},
			},
			"orange": {

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
				VariationInfo: ListingMetadata{
					Title:       "Small",
					Description: "Small size",
				},
			},
			"m": {
				VariationInfo: ListingMetadata{
					Title:       "Medium",
					Description: "Medium size",
				},
			},
		},
	}
	testTimeFuture := time.Unix(10000000000, 0).UTC()

	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    interface{}
		expected func(*assert.Assertions, Listing)
	}{
		{
			name:  "create full listing",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1)},
			value: testListing(),
			expected: func(r *assert.Assertions, l Listing) {
				r.EqualValues(testListing(), l)
			},
		},
		{
			name:  "replace price",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"price"}},
			value: *big.NewInt(66666),
			expected: func(r *assert.Assertions, l Listing) {
				r.Equal(*big.NewInt(66666), l.Price)
			},
		},
		{
			name:  "replace description",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"metadata", "description"}},
			value: "new description",
			expected: func(r *assert.Assertions, l Listing) {
				r.Equal("new description", l.Metadata.Description)
			},
		},
		{
			name:  "replace whole metadata",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"metadata"}},
			value: testListing().Metadata,
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(testListing().Metadata, l.Metadata)
			},
		},
		{
			name:  "append an image",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"metadata", "images", "-"}},
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
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"metadata", "images", "0"}},
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
			name:  "replace all images",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"metadata", "images"}},
			value: []string{"https://http.cat/images/300.jpg", "https://http.cat/images/400.jpg"},
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal([]string{"https://http.cat/images/300.jpg", "https://http.cat/images/400.jpg"}, l.Metadata.Images)
			},
		},
		{
			name: "remove an image",
			op:   RemoveOp,
			path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"metadata", "images", "0"}},
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal([]string{}, l.Metadata.Images)
			},
		},
		{
			name:  "replace view state",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"viewState"}},
			value: ListingViewStatePublished,
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(ListingViewStatePublished, l.ViewState)
			},
		},

		{
			name: "append a stock status",
			op:   AddOp,
			path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"stockStatuses", "-"}},
			value: ListingStockStatus{
				VariationIDs: []string{"m"},
				InStock:      boolptr(true),
			},
			expected: func(a *assert.Assertions, l Listing) {
				if !a.Len(l.StockStatuses, 2) {
					return
				}
				stockStatus := l.StockStatuses[1]
				a.Equal([]string{"m"}, stockStatus.VariationIDs)
				a.True(*stockStatus.InStock)
			},
		},
		{
			name: "prepend a stock status",
			op:   AddOp,
			path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"stockStatuses", "0"}},
			value: ListingStockStatus{
				VariationIDs: []string{"m"},
				InStock:      boolptr(true),
			},
			expected: func(a *assert.Assertions, l Listing) {
				if !a.Len(l.StockStatuses, 2) {
					return
				}
				stockStatus := l.StockStatuses[0]
				a.Equal([]string{"m"}, stockStatus.VariationIDs)
				a.True(*stockStatus.InStock)
			},
		},
		{
			name: "replace stock status",
			op:   ReplaceOp,
			path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"stockStatuses", "0"}},
			value: ListingStockStatus{
				VariationIDs: []string{"m"},
				InStock:      boolptr(false),
			},
			expected: func(a *assert.Assertions, l Listing) {
				if !a.Len(l.StockStatuses, 1) {
					return
				}
				stockStatus := l.StockStatuses[0]
				a.Equal([]string{"m"}, stockStatus.VariationIDs)
				a.False(*stockStatus.InStock)
			},
		},
		{
			name:  "replace expectedInStockBy",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"stockStatuses", "0", "expectedInStockBy"}},
			value: testTimeFuture,
			expected: func(a *assert.Assertions, l Listing) {
				a.Nil(l.StockStatuses[0].InStock)
				a.Equal(testTimeFuture, *l.StockStatuses[0].ExpectedInStockBy)
			},
		},
		{
			name:  "replace inStock",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"stockStatuses", "0", "inStock"}},
			value: true,
			expected: func(a *assert.Assertions, l Listing) {
				a.True(*l.StockStatuses[0].InStock)
				a.Nil(l.StockStatuses[0].ExpectedInStockBy)
			},
		},
		{
			name: "remove stock status",
			op:   RemoveOp,
			path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"stockStatuses", "0"}},
			expected: func(a *assert.Assertions, l Listing) {
				a.Len(l.StockStatuses, 0)
			},
		},

		// map manipulation of Options
		{
			name:  "add an option",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "size"}},
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
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color"}},
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
			path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options"}},
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
			path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color"}},
			expected: func(a *assert.Assertions, l Listing) {
				_, ok := l.Options["color"]
				a.False(ok)
			},
		},
		{
			name:  "add a variation to an option",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color", "variations", "pink"}},
			value: testColorOption.Variations["pink"],
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(testColorOption.Variations["pink"], l.Options["color"].Variations["pink"])
			},
		},
		{
			name:  "replace title of an option",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color", "title"}},
			value: "FARBE",
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal("FARBE", l.Options["color"].Title)
			},
		},
		{
			name:  "replace variations of an option",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color", "variations", "b"}},
			value: testColorOption.Variations["pink"],
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(testColorOption.Variations["pink"], l.Options["color"].Variations["b"])
			},
		},
		{
			name:  "replace one variation's info",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color", "variations", "b", "variationInfo"}},
			value: testColorOption.Variations["pink"].VariationInfo,
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(testColorOption.Variations["pink"].VariationInfo, l.Options["color"].Variations["b"].VariationInfo)
			},
		},
		{
			name: "remove a variation from an option",
			op:   RemoveOp,
			path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color", "variations", "b"}},
			expected: func(a *assert.Assertions, l Listing) {
				_, ok := l.Options["color"].Variations["b"]
				a.False(ok)
			},
		},
	}

	type vectorEntry struct {
		Name    string
		Patch   Patch
		After   Listing
		Encoded []byte
		Hash    []byte
	}
	type vectorFile struct {
		Encoded []byte
		Before  Listing
		Hash    []byte
		Patches []vectorEntry
	}
	var vectors vectorFile
	var err error
	vectors.Before = testListing()
	vectors.Encoded, err = Marshal(vectors.Before)
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
			entry.Name = t.Name()
			entry.Patch = patch
			entry.After = lis
			entry.Encoded, err = Marshal(lis)
			require.NoError(t, err)
			entry.Hash = hash(entry.Encoded)
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	if !t.Failed() {
		tempFile, err := os.Create("vectors_patch_listing.json")
		require.NoError(t, err)
		jsonEnc := json.NewEncoder(tempFile)
		jsonEnc.SetIndent("", "  ")
		err = jsonEnc.Encode(vectors)
		require.NoError(t, err)
		tempFile.Close()
		tempFile, err = os.Create("vectors_patch_listing.cbor")
		require.NoError(t, err)
		cborEnc := DefaultEncoder(tempFile)
		err = cborEnc.Encode(vectors)
		require.NoError(t, err)
		tempFile.Close()
	}
}

func TestPatchListingErrors(t *testing.T) {
	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    interface{}
		errMatch string
	}{
		{
			name:     "invalid field path",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"invalid"}},
			value:    "test",
			errMatch: "unsupported field: invalid",
		},
		{
			name:     "remove non-existent metadata field",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"metadata", "nonexistent"}},
			errMatch: "unsupported field: nonexistent",
		},
		{
			name:     "invalid array index",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"stockStatuses", "999"}},
			value:    ListingStockStatus{},
			errMatch: "index out of bounds: 999",
		},
		{
			name:     "remove non-existent stock status",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"stockStatuses", "999"}},
			errMatch: "index out of bounds: 999",
		},
		{
			name:     "invalid value type for price",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"price"}},
			value:    "not a number",
			errMatch: "failed to unmarshal price:",
		},
		{
			name:     "invalid value type for viewState",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"viewState"}},
			value:    123,
			errMatch: "failed to unmarshal viewState:",
		},
		{
			name:     "remove non-existent option",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "nonexistent"}},
			errMatch: "option not found: nonexistent",
		},
		{
			name:     "replace non-existent variation on an option",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color", "variations", "nonexistent"}},
			errMatch: "variation not found: nonexistent",
		},
		{
			name:     "remove non-existent variation from an option",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color", "variations", "nonexistent"}},
			errMatch: "variation not found: nonexistent",
		},
	}

	var patcher Patcher
	patcher.validator = validate

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lis := testListing()

			patch := createPatch(tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			err := patcher.Listing(&lis, decodedPatch)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMatch)
		})
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
			path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"pricingCurrency"}},
			value: testCurrency,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Equal(testCurrency, m.PricingCurrency)
			},
		},

		// array mutations
		{
			name:  "append accepted currency",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"acceptedCurrencies", "-"}},
			value: testCurrency,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.AcceptedCurrencies, 3)
				a.Equal(testCurrency, m.AcceptedCurrencies[2])
			},
		},
		{
			name:  "insert accepted currency at index",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"acceptedCurrencies", "0"}},
			value: testCurrency,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.AcceptedCurrencies, 3)
				a.Equal(testCurrency, m.AcceptedCurrencies[0])
			},
		},
		{
			name:  "replace accepted currency",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"acceptedCurrencies", "0"}},
			value: testCurrency,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Equal(testCurrency, m.AcceptedCurrencies[0])
			},
		},
		{
			name: "remove accepted currency",
			op:   RemoveOp,
			path: PatchPath{Type: ObjectTypeManifest, Fields: []string{"acceptedCurrencies", "0"}},
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.AcceptedCurrencies, 1)
			},
		},

		// map mutations
		{
			name:  "replace payee",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "default"}},
			value: testPayee,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Equal(testPayee, m.Payees["default"])
			},
		},
		{
			name:  "add a payee",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "test"}},
			value: testPayee,
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.Payees, 2)
				a.Equal(testPayee, m.Payees["test"])
			},
		},
		{
			name: "remove a payee",
			op:   RemoveOp,
			path: PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "default"}},
			expected: func(a *assert.Assertions, m Manifest) {
				_, ok := m.Payees["default"]
				a.False(ok)
			},
		},
		{
			name:  "add a shipping region",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"shippingRegions", "germany"}},
			value: ShippingRegion{Country: "DE"},
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.ShippingRegions, 2)
				a.Equal("DE", m.ShippingRegions["germany"].Country)
			},
		},
		{
			name:  "replace a shipping region",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"shippingRegions", "default"}},
			value: ShippingRegion{Country: "DE"},
			expected: func(a *assert.Assertions, m Manifest) {
				a.Equal("DE", m.ShippingRegions["default"].Country)
			},
		},
		{
			name: "remove a shipping region",
			op:   RemoveOp,
			path: PatchPath{Type: ObjectTypeManifest, Fields: []string{"shippingRegions", "default"}},
			expected: func(a *assert.Assertions, m Manifest) {
				a.Len(m.ShippingRegions, 0)
				_, ok := m.ShippingRegions["default"]
				a.False(ok)
			},
		},
	}

	type vectorEntry struct {
		Name    string
		Patch   Patch
		After   Manifest
		Hash    []byte
		Encoded []byte
	}
	type vector struct {
		Before  Manifest
		Encoded []byte
		Hash    []byte
		Patches []vectorEntry
	}
	var err error
	var vectors vector
	vectors.Before = testManifest()
	vectors.Encoded, err = Marshal(vectors.Before)
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
			entry.Name = t.Name()
			entry.Patch = patch
			entry.After = manifest
			entry.Encoded, err = Marshal(manifest)
			require.NoError(t, err)
			entry.Hash = hash(entry.Encoded)
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	if !t.Failed() {
		tempFile, err := os.Create("vectors_patch_manifest.json")
		require.NoError(t, err)
		jsonEnc := json.NewEncoder(tempFile)
		jsonEnc.SetIndent("", "  ")
		err = jsonEnc.Encode(vectors)
		require.NoError(t, err)
		require.NoError(t, tempFile.Close())
		tempFile, err = os.Create("vectors_patch_manifest.cbor")
		require.NoError(t, err)
		enc := DefaultEncoder(tempFile)
		err = enc.Encode(vectors)
		require.NoError(t, err)
		require.NoError(t, tempFile.Close())
	}
}

func TestPatchManifestErrors(t *testing.T) {

	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    any
		errMatch string
	}{
		{
			name:     "unsupported op",
			op:       IncrementOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees"}},
			errMatch: "unsupported op: increment",
		},
		{
			name:     "unsupported field",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"invalid"}},
			errMatch: "unsupported field: invalid",
		},
		{
			name:     "replace non-existent payee",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "nonexistent"}},
			value:    Payee{},
			errMatch: "payee not found: nonexistent",
		},
		{
			name:     "remove non-existent payee",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "nonexistent"}},
			errMatch: "payee not found: nonexistent",
		},
		{
			name:     "replace non-existent shipping region",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"shippingRegions", "nonexistent"}},
			value:    ShippingRegion{},
			errMatch: "shipping region not found: nonexistent",
		},
		{
			name:     "remove non-existent shipping region",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"shippingRegions", "nonexistent"}},
			errMatch: "shipping region not found: nonexistent",
		},
		{
			name:     "invalid index for acceptedCurrencies",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"acceptedCurrencies", "999"}},
			value:    ChainAddress{},
			errMatch: "index out of bounds: 999",
		},
		{
			name:     "invalid value type for pricingCurrency",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"pricingCurrency"}},
			value:    "not a chain address",
			errMatch: "failed to unmarshal currency:",
		},
	}

	var patcher Patcher
	patcher.validator = validate

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			r := require.New(t)

			manifest := testManifest()

			patch := createPatch(tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			err := patcher.Manifest(&manifest, decodedPatch)
			r.Error(err)
			a.Contains(err.Error(), tc.errMatch)
		})
	}
}

func testOrder() Order {
	return Order{
		ID:    666,
		State: OrderStateOpen,
		Items: []OrderedItem{
			{
				ListingID: 5555,
				Quantity:  23,
			},
		},
		InvoiceAddress: &AddressDetails{
			Name:         "John Doe",
			Address1:     "123 Main St",
			City:         "Anytown",
			Country:      "US",
			EmailAddress: "john.doe@example.com",
		},
	}
}

func TestPatchOrder(t *testing.T) {

	testPaymentDetails := PaymentDetails{
		PaymentID: Hash{0x01, 0x02, 0x03},
		Total:     *big.NewInt(1234567890),
		ListingHashes: []cid.Cid{
			testHash(5),
			testHash(6),
			testHash(7),
		},
		TTL:           100,
		ShopSignature: Signature{0xff},
	}

	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    interface{}
		expected func(*assert.Assertions, Order)
	}{
		// item ops
		// ========

		{
			name:  "replace quantity of an item",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"items", "0", "quantity"}},
			value: uint32(42),
			expected: func(a *assert.Assertions, o Order) {
				a.Equal(uint32(42), o.Items[0].Quantity)
			},
		},
		{
			name: "add an item to an order",
			op:   AddOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"items", "-"}},
			value: OrderedItem{
				ListingID: 5555,
				Quantity:  23,
			},
			expected: func(a *assert.Assertions, o Order) {
				if !a.Len(o.Items, 2, "expected 2 items %+v", o.Items) {
					return
				}
				a.Equal(uint32(23), o.Items[1].Quantity)
				a.Equal(ObjectId(5555), o.Items[1].ListingID)
			},
		},
		{
			name: "remove an item from an order",
			op:   RemoveOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"items", "0"}},
			expected: func(a *assert.Assertions, o Order) {
				a.Len(o.Items, 0)
			},
		},
		{
			name:  "remove all items from an order",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"items"}},
			value: []OrderedItem{},
			expected: func(a *assert.Assertions, o Order) {
				a.Len(o.Items, 0)
			},
		},

		// add ops
		// =======

		{
			name:  "set invoice address name",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"invoiceAddress", "name"}},
			value: "John Doe",
			expected: func(a *assert.Assertions, o Order) {
				a.Nil(o.ShippingAddress)
				if !a.NotNil(o.InvoiceAddress) {
					return
				}
				a.Equal("123 Main St", o.InvoiceAddress.Address1)
				a.Equal("Anytown", o.InvoiceAddress.City)
				a.Equal("John Doe", o.InvoiceAddress.Name)
			},
		},
		{
			name: "set shipping address",
			op:   AddOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"shippingAddress"}},
			value: &AddressDetails{
				Name:         "Jane Doe",
				Address1:     "321 Other St",
				City:         "Othertown",
				Country:      "US",
				EmailAddress: "jane.doe@example.com",
			},
			expected: func(a *assert.Assertions, o Order) {
				if !a.NotNil(o.ShippingAddress) {
					return
				}
				a.Equal("321 Other St", o.ShippingAddress.Address1)
				a.Equal("Othertown", o.ShippingAddress.City)
			},
		},
		{
			name: "remove invoice address",
			op:   RemoveOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"invoiceAddress"}},
			expected: func(a *assert.Assertions, o Order) {
				a.Nil(o.InvoiceAddress)
			},
		},
		{
			name: "choose payee",
			op:   AddOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"chosenPayee"}},
			value: Payee{
				Address: ChainAddress{
					ChainID: 1337,
					Address: [20]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90},
				},
			},
			expected: func(a *assert.Assertions, o Order) {
				a.NotNil(o.ChosenPayee)
				a.Equal(ChainAddress{
					ChainID: 1337,
					Address: [20]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90},
				}, o.ChosenPayee.Address)
			},
		},
		{
			name: "choose currency",
			op:   AddOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"chosenCurrency"}},
			value: ChainAddress{
				ChainID: 1337,
				Address: [20]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90},
			},
			expected: func(a *assert.Assertions, o Order) {
				a.NotNil(o.ChosenCurrency)
				a.EqualValues(&ChainAddress{
					ChainID: 1337,
					Address: [20]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90},
				}, o.ChosenCurrency)
			},
		},
		{
			name:  "add payment details",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"paymentDetails"}},
			value: testPaymentDetails,
			expected: func(a *assert.Assertions, o Order) {
				a.NotNil(o.PaymentDetails)
				a.Equal(testPaymentDetails.PaymentID, o.PaymentDetails.PaymentID)
				a.Equal(testPaymentDetails.Total, o.PaymentDetails.Total)
				a.Equal(testPaymentDetails.ListingHashes, o.PaymentDetails.ListingHashes)
				a.Equal(testPaymentDetails.TTL, o.PaymentDetails.TTL)
				a.Equal(testPaymentDetails.ShopSignature, o.PaymentDetails.ShopSignature)
			},
		},
		{
			name: "add tx details",
			op:   AddOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"txDetails"}},
			value: OrderPaid{
				TxHash: &Hash{0x01, 0x02, 0x03},
			},
			expected: func(a *assert.Assertions, o Order) {
				a.NotNil(o.TxDetails)
				a.Equal(&Hash{0x01, 0x02, 0x03}, o.TxDetails.TxHash)
			},
		},

		// replace ops
		// ===========
		{
			name: "replace payee",
			op:   ReplaceOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"chosenPayee"}},
			value: Payee{
				Address: ChainAddress{
					ChainID: 1338,
					Address: [20]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00},
				},
			},
			expected: func(a *assert.Assertions, o Order) {
				a.NotNil(o.ChosenPayee)
				a.Equal(ChainAddress{
					ChainID: 1338,
					Address: [20]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00},
				}, o.ChosenPayee.Address)
			},
		},
		{
			name: "replace currency",
			op:   ReplaceOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"chosenCurrency"}},
			value: ChainAddress{
				ChainID: 1338,
				Address: [20]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00},
			},
			expected: func(a *assert.Assertions, o Order) {
				a.NotNil(o.ChosenCurrency)
				a.EqualValues(&ChainAddress{
					ChainID: 1338,
					Address: [20]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00},
				}, o.ChosenCurrency)
			},
		},

		{
			name:  "replace payment details",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"paymentDetails"}},
			value: testPaymentDetails,
			expected: func(a *assert.Assertions, o Order) {
				a.NotNil(o.PaymentDetails)
				a.Equal(testPaymentDetails.PaymentID, o.PaymentDetails.PaymentID)
			},
		},

		{
			name: "replace tx details",
			op:   ReplaceOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"txDetails"}},
			value: OrderPaid{
				TxHash: &Hash{0x04, 0x05, 0x06},
			},
			expected: func(a *assert.Assertions, o Order) {
				a.NotNil(o.TxDetails)
				a.Equal(&Hash{0x04, 0x05, 0x06}, o.TxDetails.TxHash)
			},
		},
	}

	var patcher Patcher
	patcher.validator = validate

	type vectorEntry struct {
		After   Order
		Patch   Patch
		Encoded []byte
		Hash    []byte
	}

	type vector struct {
		Before  Order
		Encoded []byte
		Hash    []byte
		Patches []vectorEntry
	}
	var vectors vector
	var err error
	vectors.Before = testOrder()
	vectors.Encoded, err = Marshal(vectors.Before)
	require.NoError(t, err)
	vectors.Hash = hash(vectors.Encoded)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			a := assert.New(t)

			order := testOrder()

			patch := createPatch(tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			err := patcher.Order(&order, decodedPatch)
			r.NoError(err)
			tc.expected(a, order)

			var v = vectorEntry{
				After: order,
				Patch: patch,
			}
			v.Encoded, err = Marshal(order)
			require.NoError(t, err)
			v.Hash = hash(v.Encoded)
			vectors.Patches = append(vectors.Patches, v)
		})
	}

	if !t.Failed() {
		encoded, err := Marshal(vectors)
		require.NoError(t, err)
		os.WriteFile("vectors_patch_order.cbor", encoded, 0644)
		encoded, err = json.MarshalIndent(vectors, "", "  ")
		require.NoError(t, err)
		os.WriteFile("vectors_patch_order.json", encoded, 0644)
	}
}

func TestPatchOrderErrors(t *testing.T) {
	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    any
		errMatch string
	}{
		{
			name:     "unsupported op",
			op:       IncrementOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"items"}},
			errMatch: "unsupported op: increment",
		},
		{
			name:     "unsupported field",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"invalid"}},
			errMatch: "unsupported field: invalid",
		},
		{
			name:     "invalid item index",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"items", "999"}},
			value:    OrderedItem{},
			errMatch: "index out of bounds: 999",
		},
		{
			name:     "invalid item index format",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"items", "abc"}},
			value:    OrderedItem{},
			errMatch: "failed to convert index to int",
		},
		{
			name:     "missing address field",
			op:       AddOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"invoiceAddress"}},
			errMatch: "Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name:     "invalid address field",
			op:       AddOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"invoiceAddress", "invalid"}},
			value:    "test",
			errMatch: "unsupported field: invalid",
		},
	}

	var patcher Patcher
	patcher.validator = validate

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			r := require.New(t)

			order := testOrder()

			patch := createPatch(tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			err := patcher.Order(&order, decodedPatch)
			r.Error(err)
			a.Contains(err.Error(), tc.errMatch)
		})
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
