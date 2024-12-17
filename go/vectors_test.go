// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	clone "github.com/huandu/go-clone/generic"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapOrdering(t *testing.T) {
	r := require.New(t)
	var shop Shop
	shop.Accounts = make(Accounts)
	shop.Listings = make(Listings)
	shop.Orders = make(Orders)
	shop.Tags = make(Tags)

	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)
	err := enc.Encode(shop)
	r.NoError(err)

	// Check field ordering in encoded shop
	// The fields should be ordered: Tags, Orders, Accounts, Listings, Manifest
	// This corresponds to the following CBOR structure:
	want := []byte{
		0xa5, // map(5)
		0x64, // text(4)
		'T', 'a', 'g', 's',
		0xa0, // map(0)
		0x66, // text(6)
		'O', 'r', 'd', 'e', 'r', 's',
		0xa0, // map(0)
		0x68, // text(8)
		'A', 'c', 'c', 'o', 'u', 'n', 't', 's',
		0xa0, // map(0)
		0x68, // text(8)
		'L', 'i', 's', 't', 'i', 'n', 'g', 's',
		0xa0, // map(0)
		0x68, // text(8)
		'M', 'a', 'n', 'i', 'f', 'e', 's', 't',
		0xa5, // map(5)
		0x66, 0x50, 0x61, 0x79, 0x65, 0x65, 0x73, 0xf6, 0x66, 0x53, 0x68, 0x6f, 0x70, 0x49, 0x64, 0x00, 0x6f, 0x50, 0x72, 0x69, 0x63, 0x69, 0x6e, 0x67, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0xa2, 0x67, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x54, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x67, 0x43, 0x68, 0x61, 0x69, 0x6e, 0x49, 0x44, 0x00, 0x6f, 0x53, 0x68, 0x69, 0x70, 0x70, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x73, 0xf6, 0x72, 0x41, 0x63, 0x63, 0x65, 0x70, 0x74, 0x65, 0x64, 0x43, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x69, 0x65, 0x73, 0xf6,
	}
	r.Equal(len(want), buf.Len(), "encoded shop should be the same length")

	got := buf.Bytes()
	r.Equal(want, got, "encoded shop fields should be in alphabetical order")
}

// Defines the structure of a vector file.
type vectorFile[T any] struct {
	Patches []vectorEntry[T]
}
type vectorEntry[T any] struct {
	Name          string
	Patch         Patch
	Before, After vectorSnapshot[T]
}
type vectorSnapshot[T any] struct {
	Value   T
	Encoded []byte
	Hash    []byte
}

// This vector exercises the mutations of the shop object.
// Mutations of objects in the shop (listing, order, etc) are tested seperatly.
// The vectors file is constructed slightly differently to the other vectors files.
// Instead of starting with the same state every time ("Start" value),
// we keep the same state for all the patches.
func TestGenerateVectorsShop(t *testing.T) {
	r := require.New(t)

	var (
		shopId   Uint256      = *big.NewInt(23)
		testAddr ChainAddress = addrFromHex(1, "0x1234567890123456789012345678901234567890")
		testEth  ChainAddress = addrFromHex(1, "0x0000000000000000000000000000000000000000")
		testUsdc ChainAddress = addrFromHex(1, "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

		yetAnotherPayee = Payee{
			Address:        testAddr,
			CallAsContract: true,
		}

		testAcc1Addr = EthereumAddress{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90}
		testAcc1     = Account{
			KeyCards: []PublicKey{
				testPubKey(1),
				testPubKey(2),
				testPubKey(3),
			},
			Guest: false,
		}
		guestAccAddr EthereumAddress
		testAcc2     = Account{
			KeyCards: []PublicKey{
				testPubKey(4),
			},
			Guest: true,
		}
	)

	testShop := func() Shop {
		s := Shop{}
		s.Manifest = Manifest{
			ShopId: shopId,
			Payees: Payees{
				"default": {
					Address:        testAddr,
					CallAsContract: false,
				},
				"with-escrow": {
					Address:        testAddr,
					CallAsContract: true,
				},
			},
			AcceptedCurrencies: ChainAddresses{
				testEth,
				testUsdc,
			},
			PricingCurrency: testUsdc,
		}
		s.Manifest.ShippingRegions = make(ShippingRegions)
		s.Accounts = make(Accounts)
		s.Listings = make(Listings)
		s.Tags = make(Tags)
		s.Orders = make(Orders)
		return s
	}

	var patches = []Patch{
		// manifest
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "yet-another-payee"}},
			Value: mustEncode(t, yetAnotherPayee),
		},
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"shippingRegions", "germany"}},
			Value: mustEncode(t, ShippingRegion{Country: "DE"}),
		},
		{
			Op:   "remove",
			Path: PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "with-escrow"}},
		},
		// accounts
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeAccount, AccountID: &testAcc1Addr},
			Value: mustEncode(t, testAcc1),
		},
		{
			Op:   "remove",
			Path: PatchPath{Type: ObjectTypeAccount, AccountID: &testAcc1Addr, Fields: []string{"keyCards", "1"}},
		},
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeAccount, AccountID: &guestAccAddr},
			Value: mustEncode(t, testAcc2),
		},
		// listing
		{
			Op:   "add",
			Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(23)},
			Value: mustEncode(t, Listing{
				ID:        23,
				Price:     *big.NewInt(230000),
				ViewState: ListingViewStatePublished,
				Metadata: ListingMetadata{
					Title:       "test",
					Description: "test",
				},
			}),
		},
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(23), Fields: []string{"metadata", "images", "-"}},
			Value: mustEncode(t, "https://http.cat/images/100.jpg"),
		},
		{
			Op:    "replace",
			Path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(23), Fields: []string{"metadata", "images", "0"}},
			Value: mustEncode(t, "https://http.cat/images/200.jpg"),
		},
		{
			Op:   "add",
			Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(42)},
			Value: mustEncode(t, Listing{
				ID:        42,
				ViewState: ListingViewStateUnspecified,
				Price:     *big.NewInt(230000),
				Metadata: ListingMetadata{
					Title:       "test23",
					Description: "test23",
				},
			}),
		},
		{
			Op:   "replace",
			Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(42)},
			Value: mustEncode(t, Listing{
				ID:        42,
				ViewState: ListingViewStatePublished,
				Price:     *big.NewInt(420000),
				Metadata: ListingMetadata{
					Title:       "test42",
					Description: "test42",
				},
			}),
		},
		{
			Op:   "add",
			Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(666)},
			Value: mustEncode(t, Listing{
				ID:        666,
				ViewState: ListingViewStateDeleted,
				Price:     *big.NewInt(666000),
				Metadata: ListingMetadata{
					Title:       "trash",
					Description: "trash",
				},
			}),
		},
		{
			Op:   "remove",
			Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(666)},
		},
		// Tags
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test-tag")},
			Value: mustEncode(t, Tag{Name: "test-tag"}),
		},
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test-tag"), Fields: []string{"listingIds", "-"}},
			Value: mustEncode(t, ObjectId(23)),
		},
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test-tag"), Fields: []string{"listingIds", "-"}},
			Value: mustEncode(t, ObjectId(42)),
		},
		// orders
		{
			Op:   "add",
			Path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(math.MaxUint64 - 1)},
			Value: mustEncode(t, Order{
				ID:    math.MaxUint64 - 1,
				State: OrderStateOpen,
				Items: []OrderedItem{
					{ListingID: ObjectId(23), Quantity: 1},
				},
			}),
		},
		{
			Op:   "add",
			Path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(math.MaxUint64 - 2)},
			Value: mustEncode(t, Order{
				ID:    math.MaxUint64 - 2,
				State: OrderStateOpen,
				Items: []OrderedItem{
					{ListingID: ObjectId(42), Quantity: 1},
				},
			}),
		},
		{
			Op:   "remove",
			Path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(math.MaxUint64 - 2)},
		},
	}

	var patcher Patcher
	patcher.validator = validate

	var vectors vectorFile[Shop]

	var state = testShop()
	for i, patch := range patches {
		t.Run(fmt.Sprintf("patch-%d", i), func(t *testing.T) {
			r := require.New(t)
			var entry = vectorEntry[Shop]{
				Name:  t.Name(),
				Patch: patch,
			}

			// we need to clone the state because the patcher mutates the state
			// and we want to keep the original state for the before value for serialization

			beforeState := clone.Clone(state)
			beforeEncoded := mustEncode(t, beforeState)
			entry.Before = vectorSnapshot[Shop]{
				Value:   beforeState,
				Encoded: beforeEncoded,
				Hash:    hash(beforeEncoded),
			}

			err := patcher.Shop(&state, patch)
			r.NoError(err)

			afterState := clone.Clone(state)
			afterEncoded := mustEncode(t, afterState)
			entry.After = vectorSnapshot[Shop]{
				Value:   afterState,
				Encoded: afterEncoded,
				Hash:    hash(afterEncoded),
			}
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	// check all maps are non-empty
	r.NotEmpty(state.Manifest.Payees)
	r.NotEmpty(state.Manifest.ShippingRegions)
	r.NotEmpty(state.Manifest.AcceptedCurrencies)
	r.NotEmpty(state.Manifest.PricingCurrency)
	r.NotEmpty(state.Accounts)
	r.NotEmpty(state.Listings)
	r.NotEmpty(state.Tags)
	r.NotEmpty(state.Orders)

	if !t.Failed() {
		var buf bytes.Buffer
		enc := DefaultEncoder(&buf)
		err := enc.Encode(vectors)
		r.NoError(err)
		written := int64(buf.Len())

		f := openTestFile(t, "vectors_patch_shop.cbor")
		defer f.Close()
		n, err := buf.WriteTo(f)
		r.NoError(err)
		r.EqualValues(n, written)

		f = openTestFile(t, "vectors_patch_shop.json")
		r.NoError(err)
		jsonEnc := json.NewEncoder(f)
		jsonEnc.SetIndent("", "  ")
		err = jsonEnc.Encode(vectors)
		r.NoError(err)
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

func TestGenerateVectorsManifest(t *testing.T) {
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

	var err error
	var vectors vectorFile[Manifest]

	// we use the same before for all test cases
	var before = vectorSnapshot[Manifest]{
		Value:   testManifest(),
		Encoded: mustEncode(t, testManifest()),
		Hash:    hash(mustEncode(t, testManifest())),
	}

	var patcher Patcher
	patcher.validator = validate
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manifest := testManifest()
			r := require.New(t)
			a := assert.New(t)

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			r.Equal(tc.op, decodedPatch.Op)

			err = patcher.Manifest(&manifest, decodedPatch)
			r.NoError(err)
			tc.expected(a, manifest)

			var entry vectorEntry[Manifest]
			entry.Name = t.Name()
			entry.Patch = patch
			entry.Before = before

			encoded := mustEncode(t, manifest)
			entry.After = vectorSnapshot[Manifest]{
				Value:   manifest,
				Encoded: encoded,
				Hash:    hash(encoded),
			}
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	if !t.Failed() {
		tempFile := openTestFile(t, "vectors_patch_manifest.json")
		jsonEnc := json.NewEncoder(tempFile)
		jsonEnc.SetIndent("", "  ")
		err = jsonEnc.Encode(vectors)
		require.NoError(t, err)
		require.NoError(t, tempFile.Close())
		tempFile = openTestFile(t, "vectors_patch_manifest.cbor")
		enc := DefaultEncoder(tempFile)
		err = enc.Encode(vectors)
		require.NoError(t, err)
		require.NoError(t, tempFile.Close())
	}
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

func TestGenerateVectorsListing(t *testing.T) {
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
			expected: func(a *assert.Assertions, l Listing) {
				a.EqualValues(testListing(), l)
			},
		},
		{
			name:  "replace price",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"price"}},
			value: *big.NewInt(66666),
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(*big.NewInt(66666), l.Price)
			},
		},
		{
			name:  "replace description",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"metadata", "description"}},
			value: "new description",
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal("new description", l.Metadata.Description)
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

	var vectors vectorFile[Listing]
	var err error
	var before = vectorSnapshot[Listing]{
		Value:   testListing(),
		Encoded: mustEncode(t, testListing()),
		Hash:    hash(mustEncode(t, testListing())),
	}

	var patcher Patcher
	patcher.validator = validate
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			a := assert.New(t)

			lis := testListing()

			// round trip to make sure we can encode/decode the patch
			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			r.Equal(tc.op, decodedPatch.Op)

			err := patcher.Listing(&lis, decodedPatch)
			r.NoError(err)
			tc.expected(a, lis)

			var entry vectorEntry[Listing]
			entry.Name = t.Name()
			entry.Patch = patch
			entry.Before = before
			encoded := mustEncode(t, lis)
			entry.After = vectorSnapshot[Listing]{
				Value:   lis,
				Encoded: encoded,
				Hash:    hash(encoded),
			}
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	if !t.Failed() {
		tempFile := openTestFile(t, "vectors_patch_listing.json")
		jsonEnc := json.NewEncoder(tempFile)
		jsonEnc.SetIndent("", "  ")
		err = jsonEnc.Encode(vectors)
		require.NoError(t, err)
		tempFile.Close()
		tempFile = openTestFile(t, "vectors_patch_listing.cbor")
		cborEnc := DefaultEncoder(tempFile)
		err = cborEnc.Encode(vectors)
		require.NoError(t, err)
		tempFile.Close()
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

func TestGenerateVectorsOrder(t *testing.T) {

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

	var vectors vectorFile[Order]
	encodedBefore := mustEncode(t, testOrder())
	var before = vectorSnapshot[Order]{
		Value:   testOrder(),
		Encoded: encodedBefore,
		Hash:    hash(encodedBefore),
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			a := assert.New(t)

			order := testOrder()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			err := patcher.Order(&order, decodedPatch)
			r.NoError(err)
			tc.expected(a, order)

			encodedAfter := mustEncode(t, order)
			var entry = vectorEntry[Order]{
				Name:   tc.name,
				Before: before,
				Patch:  patch,
				After: vectorSnapshot[Order]{
					Value:   order,
					Encoded: encodedAfter,
					Hash:    hash(encodedAfter),
				},
			}
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	if !t.Failed() {
		encoded, err := Marshal(vectors)
		require.NoError(t, err)
		f := openTestFile(t, "vectors_patch_order.cbor")
		defer f.Close()
		f.Write(encoded)

		encoded, err = json.MarshalIndent(vectors, "", "  ")
		require.NoError(t, err)
		f = openTestFile(t, "vectors_patch_order.json")
		defer f.Close()
		f.Write(encoded)

	}
}
