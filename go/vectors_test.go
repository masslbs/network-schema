// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
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
	shop := NewShop()

	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)
	err := enc.Encode(shop)
	r.NoError(err)

	// Check field ordering in encoded shop
	// The fields should be ordered: Tags, Orders, Accounts, Listings, Manifest
	// This corresponds to the following CBOR structure:
	want := []byte{
		0xa6, // map(6)
		0x64, // text(4)
		'T', 'a', 'g', 's',
		0x82, 0x00, 0xf6, // empty hamt
		0x66, // text(6)
		'O', 'r', 'd', 'e', 'r', 's',
		0x82, 0x00, 0xf6, // empty hamt
		0x68, // text(8)
		'A', 'c', 'c', 'o', 'u', 'n', 't', 's',
		0x82, 0x00, 0xf6, // empty hamt
		0x68, // text(8)
		'L', 'i', 's', 't', 'i', 'n', 'g', 's',
		0x82, 0x00, 0xf6, // empty hamt
		0x68, // text(8)
		'M', 'a', 'n', 'i', 'f', 'e', 's', 't',
		0xa4, // map(4)
		0x66, // text(6);
		'P', 'a', 'y', 'e', 'e', 's',
		0xf6, // primitive(22)
		0x66, // text(6)
		'S', 'h', 'o', 'p', 'I', 'D',
		0x00, // unsigned(0)
		0x6f, // text(15)
		'P', 'r', 'i', 'c', 'i', 'n', 'g', 'C', 'u', 'r', 'r', 'e', 'n', 'c', 'y',
		0xa2, // map(2)
		0x67, // text(7)
		'A', 'd', 'd', 'r', 'e', 's', 's',
		0x54, // bytes(20)
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x67, // text(7)
		'C', 'h', 'a', 'i', 'n', 'I', 'D',
		0x00, // unsigned(0)
		0x72, // text(18)
		'A', 'c', 'c', 'e', 'p', 't', 'e', 'd', 'C', 'u', 'r', 'r', 'e', 'n', 'c', 'i', 'e', 's',
		0xf6, // primitive(22)
		0x69, // text(8)
		'I', 'n', 'v', 'e', 'n', 't', 'o', 'r', 'y',
		0x82, 0x00, 0xf6, // empty hamt
	}
	got := buf.Bytes()
	gotHex := hex.EncodeToString(got)
	t.Log("Got:", gotHex)
	r.Equal(len(want), len(got), "encoded shop should be the same length")

	r.Equal(want, got)
}

// Defines the structure of a vector file.
type vectorFileOkay struct {
	Signer struct {
		Address EthereumAddress
		Secret  []byte
	}
	Snapshots []vectorEntryOkay

	PatchSet  PatchSet
	Signature Signature // signature of the patchset, using the signer's secret
}
type vectorEntryOkay struct {
	Name   string
	Before vectorSnapshot
	After  vectorSnapshot
}
type vectorSnapshot struct {
	Value   Shop
	Encoded []byte
	Hash    []byte
}
type vectorFileError struct {
	Patches []vectorEntryError
}
type vectorEntryError struct {
	Name   string
	Patch  Patch
	Before vectorSnapshot
	Error  string
}

var kcNonce uint64 = 23

// This vector exercises the mutations of the shop object.
// Mutations of objects in the shop (listing, order, etc) are tested seperatly.
// The vectors file is constructed slightly differently to the other vectors files.
// Instead of starting with the same state every time ("Start" value),
// we keep the same state for all the patches.
func TestGenerateVectorsShopOkay(t *testing.T) {
	r := require.New(t)

	var shopIdBytes [32]byte
	rand.Read(shopIdBytes[:])
	shopId := Uint256{}
	shopId.SetBytes(shopIdBytes[:])
	t.Log("shop ID: ", shopId.String())
	var (
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
		guestAccAddr EthereumAddress // zero address
		testAcc2     = Account{
			KeyCards: []PublicKey{
				testPubKey(4),
			},
			Guest: true,
		}

		testVAT = PriceModifier{
			ModificationPrecents: new(Uint256).SetUint64(19),
		}
		testGermany = ShippingRegion{
			Country: "Germany",
			PriceModifiers: map[string]PriceModifier{
				"VAT": testVAT,
				"DHL Local": {
					ModificationAbsolute: &ModificationAbsolute{
						Amount: *big.NewInt(500), // TODO: assuming 2 decimals
						Plus:   true,
					},
				},
			},
		}
		testOther = ShippingRegion{
			Country: "",
			PriceModifiers: map[string]PriceModifier{
				"DHL International": {
					ModificationAbsolute: &ModificationAbsolute{
						Amount: *big.NewInt(4200), // TODO: assuming 2 decimals
						Plus:   true,
					},
				},
			},
		}
	)

	// inline function to scope over the variables
	testShop := func() Shop {
		s := NewShop()
		s.Manifest = Manifest{
			ShopID: shopId,
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
			ShippingRegions: ShippingRegions{
				"other": testOther,
			},
		}
		return s
	}
	_, testListing := newTestListing()

	var patcher Patcher
	patcher.validator = validate

	var vectors vectorFileOkay

	kp := initVectors(t, &vectors, shopId)

	var state = testShop()
	var testCases = []struct {
		Name  string
		Op    OpString
		Path  PatchPath
		Value []byte
		Check func(*require.Assertions, Shop)
	}{
		// manifest
		{
			Name:  "add-payee",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "yet-another-payee"}},
			Value: mustEncode(t, yetAnotherPayee),
			Check: func(r *require.Assertions, state Shop) {
				r.Equal(yetAnotherPayee, state.Manifest.Payees["yet-another-payee"])
			},
		},
		{
			Name:  "add-shipping-region",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"shippingRegions", "germany"}},
			Value: mustEncode(t, testGermany),
			Check: func(r *require.Assertions, state Shop) {
				r.Equal(testGermany, state.Manifest.ShippingRegions["germany"])
			},
		},
		{
			Name: "remove-payee",
			Op:   RemoveOp,
			Path: PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "with-escrow"}},
			Check: func(r *require.Assertions, state Shop) {
				r.Equal(2, len(state.Manifest.Payees))
				_, has := state.Manifest.Payees["with-escrow"]
				r.False(has)
				_, has = state.Manifest.Payees["default"]
				r.True(has)
				_, has = state.Manifest.Payees["yet-another-payee"]
				r.True(has)
			},
		},
		// accounts
		{
			Name:  "add-account",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeAccount, AccountID: &testAcc1Addr},
			Value: mustEncode(t, testAcc1),
			Check: func(r *require.Assertions, state Shop) {
				acc, ok := state.Accounts.Get(testAcc1Addr[:])
				r.True(ok)
				r.Equal(testAcc1, acc)
			},
		},
		{
			Name: "remove-keycard",
			Op:   RemoveOp,
			Path: PatchPath{Type: ObjectTypeAccount, AccountID: &testAcc1Addr, Fields: []string{"keyCards", "1"}},
			Check: func(r *require.Assertions, state Shop) {
				acc, ok := state.Accounts.Get(testAcc1Addr[:])
				r.True(ok)
				r.Len(acc.KeyCards, 2)
			},
		},
		{
			Name:  "add-guest-account",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeAccount, AccountID: &guestAccAddr},
			Value: mustEncode(t, testAcc2),
			Check: func(r *require.Assertions, state Shop) {
				r.Equal(2, state.Accounts.Size())
				acc, ok := state.Accounts.Get(guestAccAddr[:])
				r.True(ok)
				r.Equal(testAcc2, acc)
			},
		},
		// listing
		{
			Name: "add-listing",
			Op:   AddOp,
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
			Name:  "add-listing-image",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(23), Fields: []string{"metadata", "images", "-"}},
			Value: mustEncode(t, "https://http.cat/images/100.jpg"),
		},
		{
			Name:  "replace-listing-image",
			Op:    ReplaceOp,
			Path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(23), Fields: []string{"metadata", "images", "0"}},
			Value: mustEncode(t, "https://http.cat/images/200.jpg"),
		},
		{
			Name: "add-listing2",
			Op:   AddOp,
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
			Name: "replace-listing",
			Op:   ReplaceOp,
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
			Name: "add-deleted-listing",
			Op:   AddOp,
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
			Name: "remove-deleted-listing",
			Op:   RemoveOp,
			Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(666)},
		},
		// Tags
		{
			Name:  "add-tag",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test-tag")},
			Value: mustEncode(t, Tag{Name: "test-tag"}),
		},
		{
			Name:  "add-listing-to-tag",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test-tag"), Fields: []string{"listingIds", "-"}},
			Value: mustEncode(t, ObjectId(23)),
		},
		{
			Name:  "add-listing-to-tag2",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test-tag"), Fields: []string{"listingIds", "-"}},
			Value: mustEncode(t, ObjectId(42)),
		},
		// orders
		{
			Name: "add-order",
			Op:   AddOp,
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
			Name: "add-order2",
			Op:   AddOp,
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
			Name: "remove-order",
			Op:   RemoveOp,
			Path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(math.MaxUint64 - 2)},
		},
		// inventory
		{
			Name:  "add-inventory",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(23)},
			Value: mustEncode(t, uint64(100)),
		},
		{
			Name:  "replace-inventory",
			Op:    ReplaceOp,
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(23)},
			Value: mustEncode(t, uint64(24)),
		},
		{
			Name:  "add-inventory-to-be-deleted",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(42)},
			Value: mustEncode(t, uint64(100)),
		},
		{ // inject item with variations
			Name:  "add-listing",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(9000)},
			Value: mustEncode(t, testListing),
		},
		{
			Name: "add-size-option",
			Op:   AddOp,
			Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(9000), Fields: []string{"options", "size"}},
			Value: mustEncode(t, ListingOption{
				Title: "Sizes",
				Variations: ListingVariations{
					"m":  {VariationInfo: ListingMetadata{Title: "M", Description: "Medium"}},
					"l":  {VariationInfo: ListingMetadata{Title: "L", Description: "Large"}},
					"xl": {VariationInfo: ListingMetadata{Title: "XL", Description: "X-Large"}},
				},
			}),
		},
		{
			Name:  "add-inventory-variation",
			Op:    AddOp, // adds a variation to the inventory for id 1
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(9000), Fields: []string{"r", "xl"}},
			Value: mustEncode(t, uint64(23)),
		},
		{
			Name:  "add-inventory-variation2",
			Op:    AddOp, // adds a variation to the inventory for id 1
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(9000), Fields: []string{"b", "m"}},
			Value: mustEncode(t, uint64(42)),
		},
		{
			Name: "remove-inventory",
			Op:   RemoveOp,
			Path: PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(42)},
		},
		{
			Name: "remove-inventory-variation",
			Op:   RemoveOp,
			Path: PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(9000), Fields: []string{"r", "xl"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := require.New(t)
			patch := Patch{
				Op:    testCase.Op,
				Path:  testCase.Path,
				Value: testCase.Value,
			}
			var entry = vectorEntryOkay{
				Name: t.Name(),
			}

			// we need to clone the state because the patcher mutates the state
			// and we want to keep the original state for the before value for serialization

			beforeState := clone.Clone(state)
			beforeEncoded := mustEncode(t, beforeState)
			entry.Before = vectorSnapshot{
				Value:   beforeState,
				Encoded: beforeEncoded,
				Hash:    hash(beforeEncoded),
			}

			err := patcher.Shop(&state, patch)
			r.NoError(err)

			if testCase.Check != nil {
				testCase.Check(r, state)
			}

			afterState := clone.Clone(state)
			afterEncoded := mustEncode(t, afterState)
			entry.After = vectorSnapshot{
				Value:   afterState,
				Encoded: afterEncoded,
				Hash:    hash(afterEncoded),
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)
			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, patch)
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

	// sign the patchset
	patchSetEncoded := mustEncode(t, vectors.PatchSet)
	vectors.Signature = kp.TestSign(t, patchSetEncoded)

	writeVectors(t, vectors)
}

func TestGenerateVectorsInventoryOkay(t *testing.T) {
	r := require.New(t)

	var (
		shopIdBytes [32]byte
		shopId      Uint256
		patcher     Patcher
		vectors     vectorFileOkay

		state, testListing = newTestListing()
	)
	rand.Read(shopIdBytes[:])
	shopId.SetBytes(shopIdBytes[:])

	state.Manifest.ShopID = shopId

	patcher.validator = validate

	kp := initVectors(t, &vectors, shopId)

	_, rmListing := newTestListing()
	rmListing.ID = 42

	var testCases = []struct {
		Name  string
		Op    OpString
		Path  PatchPath
		Value []byte
		Check func(*require.Assertions, Inventory)
	}{
		// inventory
		{
			Name:  "add-inventory",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(testListing.ID)},
			Value: mustEncode(t, uint64(100)),
		},
		{
			Name:  "replace-inventory",
			Op:    ReplaceOp,
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(testListing.ID)},
			Value: mustEncode(t, uint64(24)),
		},
		{
			Name:  "add listing to be deleted",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(42)},
			Value: mustEncode(t, rmListing),
		},
		{
			Name:  "add-inventory-to-be-deleted",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(42)},
			Value: mustEncode(t, uint64(100)),
		},
		{ // inject item with variations
			Name:  "add-listing",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(9000)},
			Value: mustEncode(t, testListing),
		},
		{
			Name: "add-size-option",
			Op:   AddOp,
			Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(9000), Fields: []string{"options", "size"}},
			Value: mustEncode(t, ListingOption{
				Title: "Sizes",
				Variations: ListingVariations{
					"m":  {VariationInfo: ListingMetadata{Title: "M", Description: "Medium"}},
					"l":  {VariationInfo: ListingMetadata{Title: "L", Description: "Large"}},
					"xl": {VariationInfo: ListingMetadata{Title: "XL", Description: "X-Large"}},
				},
			}),
		},
		{
			Name:  "add-inventory-variation",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(9000), Fields: []string{"r", "xl"}},
			Value: mustEncode(t, uint64(23)),
		},
		{
			Name:  "add-inventory-variation2",
			Op:    AddOp,
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(9000), Fields: []string{"b", "m"}},
			Value: mustEncode(t, uint64(42)),
		},
		{
			Name: "remove-inventory",
			Op:   RemoveOp,
			Path: PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(42)},
		},
		{
			Name: "remove-inventory-variation",
			Op:   RemoveOp,
			Path: PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(9000), Fields: []string{"r", "xl"}},
		},
		{
			Name:  "increment-inventory",
			Op:    IncrementOp,
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(9000), Fields: []string{"b", "m"}},
			Value: mustEncode(t, uint64(42)),
			Check: func(r *require.Assertions, i Inventory) {
				count, has := i.Get(9000, []string{"b", "m"})
				r.True(has)
				r.Equal(uint64(2*42), count)
			},
		},
		{
			Name:  "decrement-inventory",
			Op:    DecrementOp,
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(9000), Fields: []string{"b", "m"}},
			Value: mustEncode(t, uint64(42)),
			Check: func(r *require.Assertions, i Inventory) {
				count, has := i.Get(9000, []string{"b", "m"})
				r.True(has)
				r.Equal(uint64(42), count)
			},
		},
		{
			Name: "add-variation-for-next-test",
			Op:   AddOp,
			Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(9000), Fields: []string{"options", "size", "variations", "s"}},
			Value: mustEncode(t, ListingVariation{
				VariationInfo: ListingMetadata{Title: "S", Description: "Small"},
			}),
			Check: func(r *require.Assertions, i Inventory) {
				count, has := i.Get(9000, []string{"r", "s"})
				r.False(has)
				r.Equal(uint64(0), count)
			},
		},
		{
			Name:  "increment-from-zero",
			Op:    IncrementOp,
			Path:  PatchPath{Type: ObjectTypeInventory, ObjectID: uint64ptr(9000), Fields: []string{"r", "s"}},
			Value: mustEncode(t, uint64(123)),
			Check: func(r *require.Assertions, i Inventory) {
				count, has := i.Get(9000, []string{"r", "s"})
				r.True(has)
				r.Equal(uint64(123), count)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			r := require.New(t)
			patch := Patch{
				Op:    testCase.Op,
				Path:  testCase.Path,
				Value: testCase.Value,
			}
			var entry = vectorEntryOkay{
				Name: t.Name(),
			}

			// we need to clone the state because the patcher mutates the state
			// and we want to keep the original state for the before value for serialization

			beforeState := clone.Clone(state)
			beforeEncoded := mustEncode(t, beforeState)
			entry.Before = vectorSnapshot{
				Value:   beforeState,
				Encoded: beforeEncoded,
				Hash:    hash(beforeEncoded),
			}

			err := patcher.Shop(&state, patch)
			r.NoError(err)

			if testCase.Check != nil {
				testCase.Check(r, state.Inventory)
			}

			afterState := clone.Clone(state)
			afterEncoded := mustEncode(t, afterState)
			entry.After = vectorSnapshot{
				Value:   afterState,
				Encoded: afterEncoded,
				Hash:    hash(afterEncoded),
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)
			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, patch)
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

	// sign the patchset
	patchSetEncoded := mustEncode(t, vectors.PatchSet)
	vectors.Signature = kp.TestSign(t, patchSetEncoded)

	writeVectors(t, vectors)
}

func newTestManifest() Shop {
	s := NewShop()
	s.Manifest = Manifest{
		ShopID: *big.NewInt(1),
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
	return s
}

func TestGenerateVectorsManifestOkay(t *testing.T) {
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
	var vectors vectorFileOkay

	shop := newTestManifest()
	shopEncoded := mustEncode(t, shop)
	// we use the same before for all test cases
	var before = vectorSnapshot{
		Value:   shop,
		Encoded: shopEncoded,
		Hash:    hash(shopEncoded),
	}

	kp := initVectors(t, &vectors, shop.Manifest.ShopID)

	var patcher Patcher
	patcher.validator = validate
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shop := newTestManifest()
			r := require.New(t)
			a := assert.New(t)

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			r.Equal(tc.op, decodedPatch.Op)

			err = patcher.Shop(&shop, decodedPatch)
			r.NoError(err)
			tc.expected(a, shop.Manifest)

			var entry vectorEntryOkay
			entry.Name = t.Name()
			entry.Before = before

			encoded := mustEncode(t, shop)
			entry.After = vectorSnapshot{
				Value:   shop,
				Encoded: encoded,
				Hash:    hash(encoded),
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)

			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, patch)
		})
	}

	// sign the patchset
	patchSetEncoded := mustEncode(t, vectors.PatchSet)
	vectors.Signature = kp.TestSign(t, patchSetEncoded)

	writeVectors(t, vectors)
}

func TestGenerateVectorsManifestError(t *testing.T) {

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

	var vectors vectorFileError

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			r := require.New(t)

			shop := newTestManifest()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			encodedManifest := mustEncode(t, shop)

			err := patcher.Shop(&shop, decodedPatch)
			r.Error(err)
			a.Contains(err.Error(), tc.errMatch)

			var entry vectorEntryError
			entry.Name = t.Name()
			entry.Patch = patch
			entry.Before = vectorSnapshot{
				Value:   shop,
				Encoded: encodedManifest,
				Hash:    hash(encodedManifest),
			}
			entry.Error = err.Error()
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	writeVectors(t, vectors)
}

func newTestListing() (Shop, Listing) {
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
	s := newTestManifest()
	err := s.Listings.Insert(lis.ID, lis)
	check(err)
	return s, lis
}

func TestGenerateVectorsListingOkay(t *testing.T) {
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

	shop, testListing := newTestListing()

	var vectors vectorFileOkay
	var before = vectorSnapshot{
		Value:   shop,
		Encoded: mustEncode(t, shop),
		Hash:    hash(mustEncode(t, shop)),
	}

	var patcher Patcher
	patcher.validator = validate

	kp := initVectors(t, &vectors, shop.Manifest.ShopID)

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
			path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(23)},
			value: testListing,
			expected: func(a *assert.Assertions, l Listing) {
				a.EqualValues(testListing, l)
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
			value: testListing.Metadata,
			expected: func(a *assert.Assertions, l Listing) {
				a.Equal(testListing.Metadata, l.Metadata)
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

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			a := assert.New(t)

			shop, _ = newTestListing()

			// round trip to make sure we can encode/decode the patch
			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			r.Equal(tc.op, decodedPatch.Op)

			err := patcher.Shop(&shop, decodedPatch)
			r.NoError(err)
			lis, ok := shop.Listings.Get(*patch.Path.ObjectID)
			r.True(ok)
			tc.expected(a, lis)

			var entry vectorEntryOkay
			entry.Name = t.Name()
			entry.Before = before
			encoded := mustEncode(t, shop)
			entry.After = vectorSnapshot{
				Value:   shop,
				Encoded: encoded,
				Hash:    hash(encoded),
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)
		})
	}
	// sign the patchset
	patchSetEncoded := mustEncode(t, vectors.PatchSet)
	vectors.Signature = kp.TestSign(t, patchSetEncoded)

	writeVectors(t, vectors)
}

func TestGenerateVectorsListingError(t *testing.T) {
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

	var vectors vectorFileError

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shop, _ := newTestListing()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			encodedBefore := mustEncode(t, shop)

			err := patcher.Shop(&shop, decodedPatch)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMatch)

			var entry vectorEntryError
			entry.Name = t.Name()
			entry.Patch = patch
			entry.Error = tc.errMatch
			entry.Before = vectorSnapshot{
				Value:   shop,
				Encoded: encodedBefore,
				Hash:    hash(encodedBefore),
			}
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	writeVectors(t, vectors)
}

func newTestTag() (Shop, Tag) {
	s := newTestManifest()
	t := Tag{
		Name: "test",
		ListingIds: []ObjectId{
			1,
			2,
			3,
		},
	}
	err := s.Tags.Insert(t.Name, t)
	check(err)
	return s, t
}

func TestGenerateVectorsTagOkay(t *testing.T) {
	var testTagName = "test"

	var patcher Patcher
	patcher.validator = validate

	var vectors vectorFileOkay
	shop, _ := newTestTag()
	encodedBefore := mustEncode(t, shop)

	kp := initVectors(t, &vectors, shop.Manifest.ShopID)

	var before = vectorSnapshot{
		Value:   shop,
		Encoded: encodedBefore,
		Hash:    hash(encodedBefore),
	}

	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    interface{}
		expected func(*assert.Assertions, Tag)
	}{
		// rename
		{
			name:  "rename tag",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeTag, TagName: strptr(testTagName), Fields: []string{"name"}},
			value: "New Name",
			expected: func(a *assert.Assertions, t Tag) {
				a.Equal("New Name", t.Name)
			},
		},
		// add listing
		{
			name:  "add listing to tag",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeTag, TagName: strptr(testTagName), Fields: []string{"listingIds", "-"}},
			value: ObjectId(23),
			expected: func(a *assert.Assertions, t Tag) {
				if !a.Len(t.ListingIds, 4) {
					return
				}
				a.Equal(ObjectId(23), t.ListingIds[3])
			},
		},
		// remove listing
		{
			name: "remove listing from tag",
			op:   RemoveOp,
			path: PatchPath{Type: ObjectTypeTag, TagName: strptr(testTagName), Fields: []string{"listingIds", "0"}},
			expected: func(a *assert.Assertions, t Tag) {
				if !a.Len(t.ListingIds, 2) {
					return
				}
				a.Equal(ObjectId(2), t.ListingIds[0])
				a.Equal(ObjectId(3), t.ListingIds[1])
			},
		},
		// replace listing ID
		{
			name:  "replace listing ID",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeTag, TagName: strptr(testTagName), Fields: []string{"listingIds", "0"}},
			value: ObjectId(42),
			expected: func(a *assert.Assertions, t Tag) {
				if !a.Len(t.ListingIds, 3) {
					return
				}
				a.Equal(ObjectId(42), t.ListingIds[0])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			a := assert.New(t)

			shop, _ := newTestTag()

			// round trip to make sure we can encode/decode the patch
			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			r.Equal(tc.op, decodedPatch.Op)

			err := patcher.Shop(&shop, decodedPatch)
			r.NoError(err)
			tag, ok := shop.Tags.Get(testTagName)
			r.True(ok)
			tc.expected(a, tag)

			var entry vectorEntryOkay
			entry.Name = t.Name()
			entry.Before = before
			encoded := mustEncode(t, tag)
			entry.After = vectorSnapshot{
				Value:   shop,
				Encoded: encoded,
				Hash:    hash(encoded),
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)
			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, patch)
		})
	}

	// sign the patch set
	patchSetEncoded := mustEncode(t, vectors.PatchSet)
	vectors.Signature = kp.TestSign(t, patchSetEncoded)

	writeVectors(t, vectors)
}

func TestGenerateVectorsTagError(t *testing.T) {
	testCases := []struct {
		name  string
		op    OpString
		path  PatchPath
		value interface{}
		error string
	}{
		// can only replace name, not add or remove
		{
			name:  "add to name",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test"), Fields: []string{"name"}},
			value: "New Name",
			error: "unsupported field: name",
		},
		{
			name:  "remove from name",
			op:    RemoveOp,
			path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test"), Fields: []string{"name"}},
			value: "New Name",
			error: "unsupported field: name",
		},
		// listingIds need numerical index
		{
			name:  "add listing to tag",
			op:    AddOp,
			path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test"), Fields: []string{"listingIds", "helloWorld"}},
			value: ObjectId(23),
			error: "unsupported field: helloWorld",
		},
		{
			name:  "remove listing from tag",
			op:    RemoveOp,
			path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test"), Fields: []string{"listingIds", "helloWorld"}},
			value: ObjectId(23),
			error: "unsupported field: helloWorld",
		},
		{
			name:  "replace listing in tag",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test"), Fields: []string{"listingIds", "helloWorld"}},
			value: ObjectId(42),
			error: "unsupported field: helloWorld",
		},
		// cant remove or replace non-existent listing
		{
			name:  "remove non-existent listing from tag",
			op:    RemoveOp,
			path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test"), Fields: []string{"listingIds", "999"}},
			value: ObjectId(42),
			error: "index out of bounds: 999",
		},
		{
			name:  "replace non-existent listing in tag",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test"), Fields: []string{"listingIds", "999"}},
			value: ObjectId(42),
			error: "index out of bounds: 999",
		},
	}

	var patcher Patcher
	patcher.validator = validate

	var vectors vectorFileError

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shop, _ := newTestTag()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			encodedBefore := mustEncode(t, shop)

			err := patcher.Shop(&shop, decodedPatch)
			require.Error(t, err)

			var entry vectorEntryError
			entry.Name = t.Name()
			entry.Patch = patch
			entry.Error = tc.error
			entry.Before = vectorSnapshot{
				Value:   shop,
				Encoded: encodedBefore,
				Hash:    hash(encodedBefore),
			}
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	writeVectors(t, vectors)
}

func newTestOrder() (Shop, Order) {
	s := newTestManifest()
	o := Order{
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
	err := s.Orders.Insert(o.ID, o)
	check(err)
	return s, o
}

func TestGenerateVectorsOrderOkay(t *testing.T) {

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

	var patcher Patcher
	patcher.validator = validate

	var vectors vectorFileOkay
	shop, _ := newTestOrder()
	encodedBefore := mustEncode(t, shop)
	var before = vectorSnapshot{
		Value:   shop,
		Encoded: encodedBefore,
		Hash:    hash(encodedBefore),
	}

	kp := initVectors(t, &vectors, shop.Manifest.ShopID)

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
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"items", "0", "quantity"}},
			value: uint32(42),
			expected: func(a *assert.Assertions, o Order) {
				a.Equal(uint32(42), o.Items[0].Quantity)
			},
		},
		{
			name: "add an item to an order",
			op:   AddOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"items", "-"}},
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
			name:  "increment item quantity",
			op:    IncrementOp,
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"items", "0", "quantity"}},
			value: uint32(10),
			expected: func(a *assert.Assertions, o Order) {
				a.Equal(uint32(33), o.Items[0].Quantity) // 23 + 10
			},
		},
		{
			name:  "decrement item quantity",
			op:    DecrementOp,
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"items", "0", "quantity"}},
			value: uint32(5),
			expected: func(a *assert.Assertions, o Order) {
				a.Equal(uint32(18), o.Items[0].Quantity) // 23 - 5
			},
		},
		{
			name: "remove an item from an order",
			op:   RemoveOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"items", "0"}},
			expected: func(a *assert.Assertions, o Order) {
				a.Len(o.Items, 0)
			},
		},
		{
			name:  "remove all items from an order",
			op:    ReplaceOp,
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"items"}},
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
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"invoiceAddress", "name"}},
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
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"shippingAddress"}},
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
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"invoiceAddress"}},
			expected: func(a *assert.Assertions, o Order) {
				a.Nil(o.InvoiceAddress)
			},
		},
		{
			name: "choose payee",
			op:   AddOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"chosenPayee"}},
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
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"chosenCurrency"}},
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
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"paymentDetails"}},
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
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"txDetails"}},
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
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"chosenPayee"}},
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
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"chosenCurrency"}},
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
			path:  PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"paymentDetails"}},
			value: testPaymentDetails,
			expected: func(a *assert.Assertions, o Order) {
				a.NotNil(o.PaymentDetails)
				a.Equal(testPaymentDetails.PaymentID, o.PaymentDetails.PaymentID)
			},
		},

		{
			name: "replace tx details",
			op:   ReplaceOp,
			path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"txDetails"}},
			value: OrderPaid{
				TxHash: &Hash{0x04, 0x05, 0x06},
			},
			expected: func(a *assert.Assertions, o Order) {
				a.NotNil(o.TxDetails)
				a.Equal(&Hash{0x04, 0x05, 0x06}, o.TxDetails.TxHash)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			a := assert.New(t)

			shop, _ := newTestOrder()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			err := patcher.Shop(&shop, decodedPatch)
			r.NoError(err)
			order, ok := shop.Orders.Get(*patch.Path.ObjectID)
			r.True(ok)
			tc.expected(a, order)

			encodedAfter := mustEncode(t, shop)
			var entry = vectorEntryOkay{
				Name:   tc.name,
				Before: before,
				After: vectorSnapshot{
					Value:   shop,
					Encoded: encodedAfter,
					Hash:    hash(encodedAfter),
				},
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)
			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, patch)
		})
	}

	// sign the patchset
	patchSetEncoded := mustEncode(t, vectors.PatchSet)
	vectors.Signature = kp.TestSign(t, patchSetEncoded)

	writeVectors(t, vectors)
}

func TestGenerateVectorsOrderError(t *testing.T) {
	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    any
		errMatch string
	}{
		{
			name:     "unsupported path",
			op:       IncrementOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"state"}},
			errMatch: "incr/decr only works on path: [items, x, quantity]",
		},
		{
			name:     "unsupported path",
			op:       DecrementOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"state"}},
			errMatch: "incr/decr only works on path: [items, x, quantity]",
		},
		{
			name:     "unsupported field",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"invalid"}},
			errMatch: "unsupported field: invalid",
		},
		{
			name:     "invalid item index",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"items", "999"}},
			value:    OrderedItem{},
			errMatch: "index out of bounds: 999",
		},
		{
			name:     "invalid item index format",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"items", "abc"}},
			value:    OrderedItem{},
			errMatch: "failed to convert index to int",
		},
		{
			name:     "missing address field",
			op:       AddOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"invoiceAddress"}},
			errMatch: "Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name:     "invalid address field",
			op:       AddOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(666), Fields: []string{"invoiceAddress", "invalid"}},
			value:    "test",
			errMatch: "unsupported field: invalid",
		},
	}

	var patcher Patcher
	patcher.validator = validate

	var vectors vectorFileError

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			r := require.New(t)

			shop, _ := newTestOrder()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			encodedBefore := mustEncode(t, shop)

			err := patcher.Shop(&shop, decodedPatch)
			r.Error(err)
			a.Contains(err.Error(), tc.errMatch)

			var entry vectorEntryError
			entry.Name = t.Name()
			entry.Patch = patch
			entry.Error = err.Error()
			entry.Before = vectorSnapshot{
				Value:   shop,
				Encoded: encodedBefore,
				Hash:    hash(encodedBefore),
			}
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	writeVectors(t, vectors)
}
