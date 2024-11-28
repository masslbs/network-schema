// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"StockStatuses", "-"}},
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
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"StockStatuses", "0"}},
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
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"StockStatuses", "0"}},
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
			name: "remove stock status",
			op:   RemoveOp,
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"StockStatuses", "0"}},
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lis := testListing()

			// round trip to make sure we can encode/decode the patch
			patch := createPatch(tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			r := require.New(t)
			r.Equal(tc.op, decodedPatch.Op)
			r.Equal(tc.path, decodedPatch.Path)

			a := assert.New(t)
			var err error
			switch decodedPatch.Op {
			case ReplaceOp:
				err = lis.PatchReplace(decodedPatch.Path.Fields, decodedPatch.Value)
			case AddOp:
				err = lis.PatchAdd(decodedPatch.Path.Fields, decodedPatch.Value)
			case RemoveOp:
				err = lis.PatchRemove(decodedPatch.Path.Fields)
			default:
				t.Fatalf("unsupported op: %s", decodedPatch.Op)
			}
			r.NoError(err)
			a.NoError(validate.Struct(lis))
			tc.expected(a, lis)
		})
	}
}

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
	t.Log("Patch encoded:\n" + pretty(encoded))
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
