// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	clone "github.com/huandu/go-clone/generic"
	"github.com/peterldowns/testy/assert"

	"github.com/masslbs/network-schema/go/internal/testhelper"
	massmmr "github.com/masslbs/network-schema/go/mmr"
	"github.com/masslbs/network-schema/go/objects"
)

const withVariations = false

// Defines the structure of a vector file.
type vectorFileOkay struct {
	Signer struct {
		Address objects.EthereumAddress
		Secret  []byte
	}
	Snapshots []vectorEntryOkay

	PatchSet SignedPatchSet
}
type vectorEntryOkay struct {
	Name   string
	Before vectorSnapshot
	After  vectorSnapshot
}
type vectorSnapshot struct {
	Value   objects.Shop
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

var biggestObjectID = objects.ObjectID(math.MaxUint64)

// This vector exercises the mutations of the shop object.
// Mutations of objects in the shop (listing, order, etc) are tested separately.
// The vectors file is constructed slightly differently to the other vectors files.
// Instead of starting with the same state every time ("Start" value),
// we keep the same state for all the patches.
func TestGenerateVectorsShopOkay(t *testing.T) {
	var shopIDBytes [32]byte
	rand.Read(shopIDBytes[:])
	shopID := objects.Uint256{}
	shopID.SetBytes(shopIDBytes[:])
	t.Log("shop ID: ", shopID.String())
	var (
		otherPayeeMetadata = objects.PayeeMetadata{
			CallAsContract: true,
		}

		testAcc1Addr = testMassEthAddr([20]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90})
		testAcc1     = objects.Account{
			KeyCards: []objects.PublicKey{
				testPubKey(1),
				testPubKey(2),
				testPubKey(3),
			},
			Guest: false,
		}
		guestAccAddr objects.EthereumAddress // zero address
		testAcc2     = objects.Account{
			KeyCards: []objects.PublicKey{
				testPubKey(4),
			},
			Guest: true,
		}

		testVAT = objects.PriceModifier{
			ModificationPrecents: new(objects.Uint256).SetUint64(19),
		}
		testGermany = objects.ShippingRegion{
			Country: "Germany",
			PriceModifiers: map[string]objects.PriceModifier{
				"VAT": testVAT,
				"DHL Local": {
					ModificationAbsolute: &objects.ModificationAbsolute{
						Amount: *big.NewInt(500), // TODO: assuming 2 decimals
						Plus:   true,
					},
				},
			},
		}
		testOther = objects.ShippingRegion{
			Country: "",
			PriceModifiers: map[string]objects.PriceModifier{
				"DHL International": {
					ModificationAbsolute: &objects.ModificationAbsolute{
						Amount: *big.NewInt(4200), // TODO: assuming 2 decimals
						Plus:   true,
					},
				},
			},
		}
	)

	// inline function to scope over the variables
	testShop := func() objects.Shop {
		s := objects.NewShop(23)
		s.Manifest = objects.Manifest{
			ShopID: shopID,
			Payees: objects.Payees{
				testAddr.ChainID: {
					testAddr.Address: {
						CallAsContract: false,
					},
					testAddr2.Address: {
						CallAsContract: true,
					},
				},
			},

			AcceptedCurrencies: objects.ChainAddresses{
				1: {
					testEth.Address:  {},
					testUsdc.Address: {},
				},
			},
			PricingCurrency: testUsdc,
			ShippingRegions: objects.ShippingRegions{
				"other": testOther,
			},
		}

		return s
	}
	_, testListing := newTestListing()

	var vectors vectorFileOkay

	kp := initVectors(t, &vectors, shopID)

	var state = testShop()
	var testCases = []struct {
		Name  string
		Op    OpString
		Path  Path
		Value []byte
		Check func(*testing.T, objects.Shop)
	}{
		// manifest
		{
			Name:  "add-payee",
			Op:    AddOp,
			Path:  Path{Type: ObjectTypeManifest, Fields: []any{"Payees", 1, testAddr3.Bytes()}},
			Value: mustEncode(t, otherPayeeMetadata),
			Check: func(t *testing.T, state objects.Shop) {
				assert.Equal(t, otherPayeeMetadata, state.Manifest.Payees[1][testAddr3.Address])
			},
		},
		{
			Name:  "add-shipping-region",
			Op:    AddOp,
			Path:  Path{Type: ObjectTypeManifest, Fields: []any{"ShippingRegions", "germany"}},
			Value: mustEncode(t, testGermany),
			Check: func(t *testing.T, state objects.Shop) {
				assert.Equal(t, testGermany, state.Manifest.ShippingRegions["germany"], ignoreBigInts)
			},
		},
		{
			Name: "remove-payee",
			Op:   RemoveOp,
			Path: Path{Type: ObjectTypeManifest, Fields: []any{"Payees", 1, testAddr2.Bytes()}},
			Check: func(t *testing.T, state objects.Shop) {
				assert.Equal(t, 2, len(state.Manifest.Payees[1]))
				_, has := state.Manifest.Payees[1][testAddr2.Address]
				assert.False(t, has)
				_, has = state.Manifest.Payees[1][testAddr.Address]
				assert.True(t, has)
				_, has = state.Manifest.Payees[1][testAddr3.Address]
				assert.True(t, has)
			},
		},
		// accounts
		{
			Name:  "add-account",
			Op:    AddOp,
			Path:  Path{Type: ObjectTypeAccount, AccountAddr: &testAcc1Addr},
			Value: mustEncode(t, testAcc1),
			Check: func(t *testing.T, state objects.Shop) {
				acc, ok := state.Accounts.Get(testAcc1Addr.Address[:])
				assert.True(t, ok)
				assert.Equal(t, testAcc1, acc)
				assert.Equal(t, 3, len(acc.KeyCards))
			},
		},
		{
			Name:  "append-keycard",
			Op:    AppendOp,
			Path:  Path{Type: ObjectTypeAccount, AccountAddr: &testAcc1Addr, Fields: []any{"KeyCards"}},
			Value: mustEncode(t, testPubKey(5)),
			Check: func(t *testing.T, state objects.Shop) {
				acc, ok := state.Accounts.Get(testAcc1Addr.Address[:])
				assert.True(t, ok)
				assert.Equal(t, 4, len(acc.KeyCards))
				assert.Equal(t, testPubKey(5), acc.KeyCards[3])
			},
		},
		{
			Name: "remove-keycard",
			Op:   RemoveOp,
			Path: Path{Type: ObjectTypeAccount, AccountAddr: &testAcc1Addr, Fields: []any{"KeyCards", 1}},
			Check: func(t *testing.T, state objects.Shop) {
				acc, ok := state.Accounts.Get(testAcc1Addr.Address[:])
				assert.True(t, ok)
				assert.Equal(t, 3, len(acc.KeyCards))
				assert.Equal(t, testPubKey(1), acc.KeyCards[0])
				assert.Equal(t, testPubKey(3), acc.KeyCards[1])
			},
		},
		{
			Name:  "add-guest-account",
			Op:    AddOp,
			Path:  Path{Type: ObjectTypeAccount, AccountAddr: &guestAccAddr},
			Value: mustEncode(t, testAcc2),
			Check: func(t *testing.T, state objects.Shop) {
				assert.Equal(t, 2, state.Accounts.Size())
				acc, ok := state.Accounts.Get(guestAccAddr.Address[:])
				assert.True(t, ok)
				assert.Equal(t, testAcc2, acc)
				assert.Equal(t, 2, state.Accounts.Size())
			},
		},
		// listing
		{
			Name: "add-listing",
			Op:   AddOp,
			Path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(23)},
			Value: mustEncode(t, objects.Listing{
				ID:        23,
				Price:     *big.NewInt(230000),
				ViewState: objects.ListingViewStatePublished,
				Metadata: objects.ListingMetadata{
					Title:       "test",
					Description: "test",
				},
			}),
		},
		{
			Name:  "append-listing-image",
			Op:    AppendOp,
			Path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(23), Fields: []any{"Metadata", "Images"}},
			Value: mustEncode(t, "https://http.cat/images/100.jpg"),
		},
		{
			Name:  "replace-listing-image",
			Op:    ReplaceOp,
			Path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(23), Fields: []any{"Metadata", "Images", 0}},
			Value: mustEncode(t, "https://http.cat/images/200.jpg"),
		},
		{
			Name: "add-listing2",
			Op:   AddOp,
			Path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(42)},
			Value: mustEncode(t, objects.Listing{
				ID:        42,
				ViewState: objects.ListingViewStateUnspecified,
				Price:     *big.NewInt(230000),
				Metadata: objects.ListingMetadata{
					Title:       "test23",
					Description: "test23",
				},
			}),
		},
		{
			Name: "replace-listing",
			Op:   ReplaceOp,
			Path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(42)},
			Value: mustEncode(t, objects.Listing{
				ID:        42,
				ViewState: objects.ListingViewStatePublished,
				Price:     *big.NewInt(420000),
				Metadata: objects.ListingMetadata{
					Title:       "test42",
					Description: "test42",
				},
			}),
		},
		{
			Name: "add-deleted-listing",
			Op:   AddOp,
			Path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(666)},
			Value: mustEncode(t, objects.Listing{
				ID:        666,
				ViewState: objects.ListingViewStateDeleted,
				Price:     *big.NewInt(666000),
				Metadata: objects.ListingMetadata{
					Title:       "trash",
					Description: "trash",
				},
			}),
		},
		{
			Name: "remove-deleted-listing",
			Op:   RemoveOp,
			Path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(666)},
		},
		{
			Name: "biggest item ID",
			Op:   AddOp,
			Path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(math.MaxUint64)},
			Value: mustEncode(t, objects.Listing{
				ID:    biggestObjectID,
				Price: *big.NewInt(1234567890),
				Metadata: objects.ListingMetadata{
					Title:       "biggest itemID",
					Description: strconv.FormatUint(math.MaxUint64, 10),
				},
			}),
		},
		// Tags
		{
			Name:  "add-tag",
			Op:    AddOp,
			Path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr("test-tag")},
			Value: mustEncode(t, objects.Tag{Name: "test-tag"}),
		},
		{
			Name:  "append-listing-to-tag",
			Op:    AppendOp,
			Path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr("test-tag"), Fields: []any{"ListingIDs"}},
			Value: mustEncode(t, objects.ObjectID(23)),
		},
		{
			Name:  "append-listing-to-tag2",
			Op:    AppendOp,
			Path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr("test-tag"), Fields: []any{"ListingIDs"}},
			Value: mustEncode(t, objects.ObjectID(42)),
		},
		// orders
		{
			Name: "add-order",
			Op:   AddOp,
			Path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(math.MaxUint64 - 1)},
			Value: mustEncode(t, objects.Order{
				ID:    math.MaxUint64 - 1,
				State: objects.OrderStateOpen,
				Items: []objects.OrderedItem{
					{ListingID: objects.ObjectID(23), Quantity: 1},
				},
			}),
		},
		{
			Name: "add-order2",
			Op:   AddOp,
			Path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(math.MaxUint64 - 2)},
			Value: mustEncode(t, objects.Order{
				ID:    math.MaxUint64 - 2,
				State: objects.OrderStateOpen,
				Items: []objects.OrderedItem{
					{ListingID: objects.ObjectID(42), Quantity: 1},
				},
			}),
		},
		{
			Name: "remove-order",
			Op:   RemoveOp,
			Path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(math.MaxUint64 - 2)},
		},
		{
			Name: "biggest order ID",
			Op:   AddOp,
			Path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(math.MaxUint64)},
			Value: mustEncode(t, objects.Order{
				ID:    biggestObjectID,
				State: objects.OrderStateOpen,
				Items: []objects.OrderedItem{
					{ListingID: biggestObjectID, Quantity: 789},
				},
			}),
		},
		// inventory
		{
			Name:  "add-inventory",
			Op:    AddOp,
			Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(23)},
			Value: mustEncode(t, uint64(100)),
		},
		{
			Name:  "replace-inventory",
			Op:    ReplaceOp,
			Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(23)},
			Value: mustEncode(t, uint64(24)),
		},
		{
			Name:  "add-inventory-to-be-deleted",
			Op:    AddOp,
			Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(42)},
			Value: mustEncode(t, uint64(100)),
		},
		{
			Name: "remove-inventory",
			Op:   RemoveOp,
			Path: Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(42)},
		},
		{
			Name:  "biggest item ID",
			Op:    AddOp,
			Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(math.MaxUint64)},
			Value: mustEncode(t, uint64(100)),
		},
	}

	// Add variation-specific test cases if enabled
	if withVariations {
		testCases = append(testCases, []struct {
			Name  string
			Op    OpString
			Path  Path
			Value []byte
			Check func(*testing.T, objects.Shop)
		}{
			{
				Name:  "add-listing",
				Op:    AddOp,
				Path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(9000)},
				Value: mustEncode(t, testListing),
			},
			{
				Name: "add-size-option",
				Op:   AddOp,
				Path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(9000), Fields: []any{"Options", "size"}},
				Value: mustEncode(t, objects.ListingOption{
					Title: "Sizes",
					Variations: objects.ListingVariations{
						"m":  {VariationInfo: objects.ListingMetadata{Title: "M", Description: "Medium"}},
						"l":  {VariationInfo: objects.ListingMetadata{Title: "L", Description: "Large"}},
						"xl": {VariationInfo: objects.ListingMetadata{Title: "XL", Description: "X-Large"}},
					},
				}),
			},
			{
				Name:  "add-inventory-variation",
				Op:    AddOp,
				Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(9000), Fields: []any{"r", "xl"}},
				Value: mustEncode(t, uint64(23)),
			},
			{
				Name:  "add-inventory-variation2",
				Op:    AddOp,
				Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(9000), Fields: []any{"b", "m"}},
				Value: mustEncode(t, uint64(42)),
			},
			{
				Name: "remove-inventory-variation",
				Op:   RemoveOp,
				Path: Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(9000), Fields: []any{"r", "xl"}},
			},
		}...)
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
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
				Hash:    mustHashState(t, beforeState),
			}

			patcher := NewPatcher(validate, &state)
			err := patcher.ApplyPatch(patch)
			assert.Nil(t, err)

			if testCase.Check != nil {
				testCase.Check(t, state)
			}

			afterState := clone.Clone(state)
			afterEncoded := mustEncode(t, afterState)
			entry.After = vectorSnapshot{
				Value:   afterState,
				Encoded: afterEncoded,
				Hash:    mustHashState(t, afterState),
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)
			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, patch)
		})
	}

	kp.TestSignPatchSet(t, &vectors.PatchSet)

	writeVectors(t, vectors)
}

func TestGenerateVectorsInventoryOkay(t *testing.T) {
	var (
		shopIDBytes [32]byte
		shopID      objects.Uint256
		vectors     vectorFileOkay

		state, testListing = newTestListing()
	)
	rand.Read(shopIDBytes[:])
	shopID.SetBytes(shopIDBytes[:])

	state.Manifest.ShopID = shopID

	kp := initVectors(t, &vectors, shopID)

	_, rmListing := newTestListing()
	rmListing.ID = 42

	var testCases = []struct {
		Name  string
		Op    OpString
		Path  Path
		Value []byte
		Check func(t *testing.T, i objects.Inventory)
	}{
		// inventory
		{
			Name:  "add-inventory",
			Op:    AddOp,
			Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(testListing.ID)},
			Value: mustEncode(t, uint64(100)),
		},
		{
			Name:  "replace-inventory",
			Op:    ReplaceOp,
			Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(testListing.ID)},
			Value: mustEncode(t, uint64(24)),
		},
		{
			Name:  "add listing to be deleted",
			Op:    AddOp,
			Path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(42)},
			Value: mustEncode(t, rmListing),
		},
		{
			Name:  "add-inventory-to-be-deleted",
			Op:    AddOp,
			Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(42)},
			Value: mustEncode(t, uint64(100)),
		},
		{
			Name: "remove-inventory",
			Op:   RemoveOp,
			Path: Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(42)},
		},
		{
			Name:  "increment-inventory",
			Op:    IncrementOp,
			Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(testListing.ID)},
			Value: mustEncode(t, uint64(42)),
			Check: func(t *testing.T, i objects.Inventory) {
				count, has := i.Get(testListing.ID, nil)
				assert.True(t, has)
				assert.Equal(t, uint64(24+42), count)
			},
		},
		{
			Name:  "decrement-inventory",
			Op:    DecrementOp,
			Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(testListing.ID)},
			Value: mustEncode(t, uint64(42)),
			Check: func(t *testing.T, i objects.Inventory) {
				count, has := i.Get(testListing.ID, nil)
				assert.True(t, has)
				assert.Equal(t, uint64(24), count)
			},
		},
	}

	// Add variation-specific test cases if enabled
	if withVariations {
		testCases = append(testCases, []struct {
			Name  string
			Op    OpString
			Path  Path
			Value []byte
			Check func(t *testing.T, i objects.Inventory)
		}{
			{
				Name:  "add-inventory-variation",
				Op:    AddOp,
				Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(9000), Fields: []any{"r", "xl"}},
				Value: mustEncode(t, uint64(23)),
			},
			{
				Name:  "add-inventory-variation2",
				Op:    AddOp,
				Path:  Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(9000), Fields: []any{"b", "m"}},
				Value: mustEncode(t, uint64(42)),
			},
			{
				Name: "remove-inventory-variation",
				Op:   RemoveOp,
				Path: Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(9000), Fields: []any{"r", "xl"}},
			},
		}...)
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
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
				Hash:    mustHashState(t, beforeState),
			}

			patcher := NewPatcher(validate, &state)
			err := patcher.ApplyPatch(patch)
			assert.Nil(t, err)

			if testCase.Check != nil {
				testCase.Check(t, state.Inventory)
			}

			afterState := clone.Clone(state)
			afterEncoded := mustEncode(t, afterState)
			entry.After = vectorSnapshot{
				Value:   afterState,
				Encoded: afterEncoded,
				Hash:    mustHashState(t, afterState),
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)
			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, patch)
		})
	}

	// sign the patchset
	kp.TestSignPatchSet(t, &vectors.PatchSet)

	writeVectors(t, vectors)
}

func newTestManifest() objects.Shop {
	s := objects.NewShop(666)
	var shopID objects.Uint256
	var shopIDBuf = make([]byte, 32)
	for i := range shopIDBuf {
		if i%2 == 0 {
			continue
		}
		shopIDBuf[i] = byte(0xff)
	}
	shopID.SetBytes(shopIDBuf)
	s.Manifest = objects.Manifest{
		ShopID: shopID,
		Payees: objects.Payees{
			1337: {
				testAddr123: {
					CallAsContract: false,
				},
			},
		},
		AcceptedCurrencies: objects.ChainAddresses{
			1337: {
				zeroAddress: {},
				testAddr123: {},
			},
			1: {
				zeroAddress: {},
			},
		},
		PricingCurrency: objects.ChainAddress{
			ChainID:         1337,
			EthereumAddress: testMassEthAddr([20]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
		},
		ShippingRegions: map[string]objects.ShippingRegion{
			"default": {
				Country: "DE",
			},
		},
	}
	return s
}

func TestGenerateVectorsManifestOkay(t *testing.T) {
	testCurrency := objects.ChainAddress{
		ChainID:         1337,
		EthereumAddress: testMassEthAddr([20]byte{0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00, 0xff, 0x00}),
	}

	testManifest := newTestManifest().Manifest

	// Test addresses for payees
	testChainID := uint64(1337)
	testAddr := testCommonEthAddr([20]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44})

	testCases := []struct {
		name     string
		op       OpString
		path     Path
		value    any
		expected func(*testing.T, objects.Manifest)
	}{
		{
			name:  "replace manifest",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeManifest},
			value: testManifest,
			expected: func(t *testing.T, m objects.Manifest) {
				assert.Equal(t, testManifest, m, ignoreBigInts)
			},
		},
		// simple field mutations
		{
			name:  "replace pricing currency",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeManifest, Fields: []any{"PricingCurrency"}},
			value: testCurrency,
			expected: func(t *testing.T, m objects.Manifest) {
				assert.Equal(t, testCurrency, m.PricingCurrency)
			},
		},

		// map mutations for accepted currencies
		{
			name:  "add accepted currency",
			op:    AddOp,
			path:  Path{Type: ObjectTypeManifest, Fields: []any{"AcceptedCurrencies", 1337, testCurrency.Bytes()}},
			value: struct{}{},
			expected: func(t *testing.T, m objects.Manifest) {
				addrsForChain, has := m.AcceptedCurrencies[1337]
				assert.True(t, has)
				assert.Equal(t, 3, len(addrsForChain))
				_, exists := addrsForChain[testCurrency.Address]
				assert.True(t, exists)
			},
		},
		{
			name: "remove accepted currency",
			op:   RemoveOp,
			path: Path{Type: ObjectTypeManifest, Fields: []any{"AcceptedCurrencies", 1337, testAddr123.Bytes()}},
			expected: func(t *testing.T, m objects.Manifest) {
				addrsForChain, has := m.AcceptedCurrencies[1337]
				assert.True(t, has)
				assert.Equal(t, 1, len(addrsForChain))
				_, exists := addrsForChain[zeroAddress]
				assert.True(t, exists)
			},
		},

		// map mutations
		{
			name:  "replace payee",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeManifest, Fields: []any{"Payees", 1337, testAddr123.Bytes()}},
			value: objects.PayeeMetadata{CallAsContract: true},
			expected: func(t *testing.T, m objects.Manifest) {
				assert.True(t, m.Payees[testChainID][testAddr123].CallAsContract)
			},
		},
		{
			name:  "add a payee",
			op:    AddOp,
			path:  Path{Type: ObjectTypeManifest, Fields: []any{"Payees", 1337, testAddr.Bytes()}},
			value: objects.PayeeMetadata{CallAsContract: true},
			expected: func(t *testing.T, m objects.Manifest) {
				assert.True(t, m.Payees[testChainID][testAddr].CallAsContract)
			},
		},
		{
			name: "remove a payee",
			op:   RemoveOp,
			path: Path{Type: ObjectTypeManifest, Fields: []any{"Payees", 1337, testAddr123.Bytes()}},
			expected: func(t *testing.T, m objects.Manifest) {
				_, ok := m.Payees[testChainID][testAddr123]
				assert.False(t, ok)
			},
		},
		{
			name:  "add a shipping region",
			op:    AddOp,
			path:  Path{Type: ObjectTypeManifest, Fields: []any{"ShippingRegions", "germany"}},
			value: objects.ShippingRegion{Country: "DE"},
			expected: func(t *testing.T, m objects.Manifest) {
				assert.Equal(t, 2, len(m.ShippingRegions))
				assert.Equal(t, "DE", m.ShippingRegions["germany"].Country)
			},
		},
		{
			name:  "replace a shipping region",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeManifest, Fields: []any{"ShippingRegions", "default"}},
			value: objects.ShippingRegion{Country: "DE"},
			expected: func(t *testing.T, m objects.Manifest) {
				assert.Equal(t, "DE", m.ShippingRegions["default"].Country)
			},
		},
		{
			name: "remove a shipping region",
			op:   RemoveOp,
			path: Path{Type: ObjectTypeManifest, Fields: []any{"ShippingRegions", "default"}},
			expected: func(t *testing.T, m objects.Manifest) {
				assert.Equal(t, 0, len(m.ShippingRegions))
				_, ok := m.ShippingRegions["default"]
				assert.False(t, ok)
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
		Hash:    mustHashState(t, shop),
	}

	kp := initVectors(t, &vectors, shop.Manifest.ShopID)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shop := newTestManifest()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			assert.Equal(t, tc.op, decodedPatch.Op)

			patcher := NewPatcher(validate, &shop)
			err = patcher.ApplyPatch(decodedPatch)
			assert.Nil(t, err)
			tc.expected(t, shop.Manifest)

			var entry vectorEntryOkay
			entry.Name = t.Name()
			entry.Before = before

			encoded := mustEncode(t, shop)
			entry.After = vectorSnapshot{
				Value:   shop,
				Encoded: encoded,
				Hash:    mustHashState(t, shop),
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)

			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, patch)
		})
	}

	// sign the patchset
	kp.TestSignPatchSet(t, &vectors.PatchSet)

	writeVectors(t, vectors)
}

func TestGenerateVectorsManifestError(t *testing.T) {
	var zeroAddr common.Address

	testCases := []struct {
		name     string
		op       OpString
		path     Path
		value    any
		errMatch string
	}{
		{
			name:     "unsupported op",
			op:       IncrementOp,
			path:     Path{Type: ObjectTypeManifest, Fields: []any{"Payees"}},
			errMatch: "unsupported op: increment",
		},
		{
			name:     "unsupported field",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeManifest, Fields: []any{"invalid"}},
			errMatch: "unsupported field: invalid",
		},
		{
			name:     "replace non-existent payee",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeManifest, Fields: []any{"Payees", 1, zeroAddr.Bytes()}},
			value:    objects.Payee{},
			errMatch: "object type=Manifest with fields=[Payees 1 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]] not found",
		},
		{
			name:     "remove non-existent payee",
			op:       RemoveOp,
			path:     Path{Type: ObjectTypeManifest, Fields: []any{"Payees", 1, zeroAddr.Bytes()}},
			errMatch: "object type=Manifest with fields=[Payees 1 [0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]] not found",
		},
		{
			name:     "replace non-existent shipping region",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeManifest, Fields: []any{"ShippingRegions", "nonexistent"}},
			value:    objects.ShippingRegion{},
			errMatch: "object type=Manifest with fields=[ShippingRegions nonexistent] not found",
		},
		{
			name:     "remove non-existent shipping region",
			op:       RemoveOp,
			path:     Path{Type: ObjectTypeManifest, Fields: []any{"ShippingRegions", "nonexistent"}},
			errMatch: "object type=Manifest with fields=[ShippingRegions nonexistent] not found",
		},
		{
			name:     "invalid index for acceptedCurrencies",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeManifest, Fields: []any{"AcceptedCurrencies", 999, testAddr123.Bytes()}},
			value:    objects.ChainAddress{},
			errMatch: "object type=Manifest with fields=[AcceptedCurrencies 999 [1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20]] not found",
		},
		{
			name:     "invalid value type for pricingCurrency",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeManifest, Fields: []any{"PricingCurrency"}},
			value:    "not a chain address",
			errMatch: "failed to unmarshal PricingCurrency:",
		},
	}

	var vectors vectorFileError

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shop := newTestManifest()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			encodedManifest := mustEncode(t, shop)

			patcher := NewPatcher(validate, &shop)
			err := patcher.ApplyPatch(decodedPatch)
			assert.Error(t, err)

			t.Logf("err: %s", err.Error())
			assert.True(t, strings.Contains(err.Error(), tc.errMatch))

			var entry vectorEntryError
			entry.Name = t.Name()
			entry.Patch = patch
			entry.Before = vectorSnapshot{
				Value:   shop,
				Encoded: encodedManifest,
				Hash:    mustHashState(t, shop),
			}
			entry.Error = err.Error()
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	writeVectors(t, vectors)
}

func newTestListing() (objects.Shop, objects.Listing) {
	var lis objects.Listing
	lis.ID = 1
	lis.ViewState = objects.ListingViewStatePublished
	lis.Metadata.Title = "test Listing"
	lis.Metadata.Description = "short desc"
	lis.Metadata.Images = []string{"https://http.cat/images/100.jpg"}
	price := big.NewInt(12345)
	lis.Price = *price
	lis.Options = objects.ListingOptions{
		"color": {
			Title: "Color",
			Variations: map[string]objects.ListingVariation{
				"r": {
					VariationInfo: objects.ListingMetadata{
						Title:       "Red",
						Description: "Red color",
					},
				},
				"b": {
					VariationInfo: objects.ListingMetadata{
						Title:       "Blue",
						Description: "Blue color",
					},
				},
			},
		},
	}
	lis.StockStatuses = []objects.ListingStockStatus{
		{
			VariationIDs: []string{"r"},
			InStock:      testhelper.Boolptr(true),
		},
	}
	s := newTestManifest()
	err := s.Listings.Insert(lis.ID, lis)
	if err != nil {
		panic(err)
	}
	return s, lis
}

func TestGenerateVectorsListingOkay(t *testing.T) {
	testColorOption := objects.ListingOption{
		Title: "Color",
		Variations: map[string]objects.ListingVariation{
			"pink": {
				VariationInfo: objects.ListingMetadata{
					Title:       "Pink",
					Description: "Pink color",
				},
			},
			"orange": {
				VariationInfo: objects.ListingMetadata{
					Title:       "Orange",
					Description: "Orange color",
				},
			},
		},
	}

	testSizeOption := objects.ListingOption{
		Title: "Size",
		Variations: map[string]objects.ListingVariation{
			"s": {
				VariationInfo: objects.ListingMetadata{
					Title:       "Small",
					Description: "Small size",
				},
			},
			"m": {
				VariationInfo: objects.ListingMetadata{
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
		Hash:    mustHashState(t, shop),
	}

	kp := initVectors(t, &vectors, shop.Manifest.ShopID)

	testCases := []struct {
		name     string
		op       OpString
		path     Path
		value    any
		expected func(*testing.T, objects.Listing)
	}{
		{
			name:  "create full listing",
			op:    AddOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(23)},
			value: testListing,
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, testListing, l, ignoreBigInts)
			},
		},
		{
			name:  "replace price",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Price"}},
			value: *big.NewInt(66666),
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, *big.NewInt(66666), l.Price, ignoreBigInts)
			},
		},
		{
			name:  "replace description",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Metadata", "Description"}},
			value: "new description",
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, "new description", l.Metadata.Description)
			},
		},
		{
			name:  "replace whole metadata",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Metadata"}},
			value: testListing.Metadata,
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, testListing.Metadata, l.Metadata)
			},
		},
		{
			name:  "append an image",
			op:    AppendOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Metadata", "Images"}},
			value: "https://http.cat/images/200.jpg",
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, 2, len(l.Metadata.Images))
				assert.Equal(t, "https://http.cat/images/100.jpg", l.Metadata.Images[0])
				assert.Equal(t, "https://http.cat/images/200.jpg", l.Metadata.Images[1])
			},
		},
		{
			name:  "prepend an image",
			op:    AddOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Metadata", "Images", 0}},
			value: "https://http.cat/images/200.jpg",
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, 2, len(l.Metadata.Images))
				assert.Equal(t, "https://http.cat/images/200.jpg", l.Metadata.Images[0])
				assert.Equal(t, "https://http.cat/images/100.jpg", l.Metadata.Images[1])
			},
		},
		{
			name:  "replace all images",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Metadata", "Images"}},
			value: []string{"https://http.cat/images/300.jpg", "https://http.cat/images/400.jpg"},
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, []string{"https://http.cat/images/300.jpg", "https://http.cat/images/400.jpg"}, l.Metadata.Images)
			},
		},
		{
			name: "remove an image",
			op:   RemoveOp,
			path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Metadata", "Images", 0}},
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, []string{}, l.Metadata.Images)
			},
		},
		{
			name:  "replace view state",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"ViewState"}},
			value: objects.ListingViewStatePublished,
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, objects.ListingViewStatePublished, l.ViewState)
			},
		},

		{
			name: "append a stock status",
			op:   AppendOp,
			path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"StockStatuses"}},
			value: objects.ListingStockStatus{
				VariationIDs: []string{"m"},
				InStock:      testhelper.Boolptr(true),
			},
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, 2, len(l.StockStatuses))
				stockStatus := l.StockStatuses[1]
				assert.Equal(t, []string{"m"}, stockStatus.VariationIDs)
				assert.True(t, *stockStatus.InStock)
			},
		},
		{
			name: "prepend a stock status",
			op:   AddOp,
			path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"StockStatuses", 0}},
			value: objects.ListingStockStatus{
				VariationIDs: []string{"m"},
				InStock:      testhelper.Boolptr(true),
			},
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, 2, len(l.StockStatuses))
				stockStatus := l.StockStatuses[0]
				assert.Equal(t, []string{"m"}, stockStatus.VariationIDs)
				assert.True(t, *stockStatus.InStock)
			},
		},
		{
			name: "replace stock status",
			op:   ReplaceOp,
			path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"StockStatuses", 0}},
			value: objects.ListingStockStatus{
				VariationIDs: []string{"m"},
				InStock:      testhelper.Boolptr(false),
			},
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, 1, len(l.StockStatuses))
				stockStatus := l.StockStatuses[0]
				assert.Equal(t, []string{"m"}, stockStatus.VariationIDs)
				assert.False(t, *stockStatus.InStock)
			},
		},
		{
			name:  "replace expectedInStockBy",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"StockStatuses", 0, "ExpectedInStockBy"}},
			value: testTimeFuture,
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, nil, l.StockStatuses[0].InStock)
				assert.Equal(t, testTimeFuture, *l.StockStatuses[0].ExpectedInStockBy)
			},
		},
		{
			name:  "replace inStock",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"StockStatuses", 0, "InStock"}},
			value: true,
			expected: func(t *testing.T, l objects.Listing) {
				assert.True(t, *l.StockStatuses[0].InStock)
				assert.Equal(t, nil, l.StockStatuses[0].ExpectedInStockBy)
			},
		},
		{
			name: "remove stock status",
			op:   RemoveOp,
			path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"StockStatuses", 0}},
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, 0, len(l.StockStatuses))
			},
		},

		// map manipulation of Options
		{
			name:  "add an option",
			op:    AddOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "size"}},
			value: testSizeOption,
			expected: func(t *testing.T, l objects.Listing) {
				sizeOption, ok := l.Options["size"]
				assert.True(t, ok)
				assert.Equal(t, testSizeOption, sizeOption)
			},
		},
		{
			name:  "replace one option",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "color"}},
			value: testColorOption,
			expected: func(t *testing.T, l objects.Listing) {
				colorOption, ok := l.Options["color"]
				assert.True(t, ok)
				assert.Equal(t, testColorOption, colorOption)
			},
		},
		{
			name: "replace whole options",
			op:   ReplaceOp,
			path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options"}},
			value: objects.ListingOptions{
				"color": testColorOption,
			},
			expected: func(t *testing.T, l objects.Listing) {
				colorOption, ok := l.Options["color"]
				assert.True(t, ok)
				assert.Equal(t, testColorOption, colorOption)
			},
		},
		{
			name: "remove an option",
			op:   RemoveOp,
			path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "color"}},
			expected: func(t *testing.T, l objects.Listing) {
				_, ok := l.Options["color"]
				assert.False(t, ok)
			},
		},
		{
			name:  "add a variation to an option",
			op:    AddOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "color", "Variations", "pink"}},
			value: testColorOption.Variations["pink"],
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, testColorOption.Variations["pink"], l.Options["color"].Variations["pink"])
			},
		},
		{
			name:  "replace title of an option",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "color", "Title"}},
			value: "FARBE",
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, "FARBE", l.Options["color"].Title)
			},
		},
		{
			name:  "replace variations of an option",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "color", "Variations", "b"}},
			value: testColorOption.Variations["pink"],
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, testColorOption.Variations["pink"], l.Options["color"].Variations["b"])
			},
		},
		{
			name:  "replace one variation's info",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "color", "Variations", "b", "VariationInfo"}},
			value: testColorOption.Variations["pink"].VariationInfo,
			expected: func(t *testing.T, l objects.Listing) {
				assert.Equal(t, testColorOption.Variations["pink"].VariationInfo, l.Options["color"].Variations["b"].VariationInfo)
			},
		},
		{
			name: "remove a variation from an option",
			op:   RemoveOp,
			path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "color", "Variations", "b"}},
			expected: func(t *testing.T, l objects.Listing) {
				_, ok := l.Options["color"].Variations["b"]
				assert.False(t, ok)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shop, _ = newTestListing()

			// round trip to make sure we can encode/decode the patch
			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			assert.Equal(t, tc.op, decodedPatch.Op)

			patcher := NewPatcher(validate, &shop)
			err := patcher.ApplyPatch(decodedPatch)
			assert.Nil(t, err)
			lis, ok := shop.Listings.Get(*patch.Path.ObjectID)
			assert.True(t, ok)
			tc.expected(t, lis)

			var entry vectorEntryOkay
			entry.Name = t.Name()
			entry.Before = before
			encoded := mustEncode(t, shop)
			entry.After = vectorSnapshot{
				Value:   shop,
				Encoded: encoded,
				Hash:    mustHashState(t, shop),
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)
			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, patch)
		})
	}
	// sign the patchset
	kp.TestSignPatchSet(t, &vectors.PatchSet)

	writeVectors(t, vectors)
}

func TestGenerateVectorsListingError(t *testing.T) {
	testCases := []struct {
		name     string
		op       OpString
		path     Path
		value    any
		errMatch string
	}{
		{
			name:     "invalid field path",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"invalid"}},
			value:    "test",
			errMatch: "fields=[invalid] not found",
		},
		{
			name:     "remove non-existent metadata field",
			op:       RemoveOp,
			path:     Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Metadata", "Nonexistent"}},
			errMatch: "fields=[Metadata Nonexistent] not found",
		},
		{
			name:     "invalid array index",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"StockStatuses", 999}},
			value:    objects.ListingStockStatus{},
			errMatch: "fields=[StockStatuses 999] not found",
		},
		{
			name:     "remove non-existent stock status",
			op:       RemoveOp,
			path:     Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"StockStatuses", 999}},
			errMatch: "fields=[StockStatuses 999] not found",
		},
		{
			name:     "invalid value type for price",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Price"}},
			value:    "not a number",
			errMatch: "failed to unmarshal Price:",
		},
		{
			name:     "invalid value type for viewState",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"ViewState"}},
			value:    123,
			errMatch: "failed to unmarshal ViewState:",
		},
		{
			name:     "remove non-existent option",
			op:       RemoveOp,
			path:     Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "nonexistent"}},
			errMatch: "fields=[Options nonexistent] not found",
		},
		{
			name:     "replace non-existent variation on an option",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "color", "Variations", "nonexistent"}},
			errMatch: "fields=[Options color Variations nonexistent] not found",
		},
		{
			name:     "remove non-existent variation from an option",
			op:       RemoveOp,
			path:     Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "color", "Variations", "nonexistent"}},
			errMatch: "fields=[Options color Variations nonexistent] not found",
		},
	}

	var vectors vectorFileError

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shop, _ := newTestListing()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			encodedBefore := mustEncode(t, shop)

			patcher := NewPatcher(validate, &shop)
			err := patcher.ApplyPatch(decodedPatch)
			assert.NotEqual(t, nil, err)
			t.Logf("error: %v", err)
			assert.True(t, strings.Contains(err.Error(), tc.errMatch))

			var entry vectorEntryError
			entry.Name = t.Name()
			entry.Patch = patch
			entry.Error = tc.errMatch
			entry.Before = vectorSnapshot{
				Value:   shop,
				Encoded: encodedBefore,
				Hash:    mustHashState(t, shop),
			}
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	writeVectors(t, vectors)
}

func newTestTag() (objects.Shop, objects.Tag) {
	s := newTestManifest()
	t := objects.Tag{
		Name: "test",
		ListingIDs: []objects.ObjectID{
			1,
			2,
			3,
		},
	}
	err := s.Tags.Insert(t.Name, t)
	check(err)

	testListing1 := objects.Listing{
		ID:        23,
		Price:     *big.NewInt(1000),
		Metadata:  objects.ListingMetadata{Title: "Test Listing 23", Description: "A test listing"},
		ViewState: objects.ListingViewState(0),
	}
	err = s.Listings.Insert(testListing1.ID, testListing1)
	check(err)

	testListing2 := objects.Listing{
		ID:        42,
		Price:     *big.NewInt(2000),
		Metadata:  objects.ListingMetadata{Title: "Test Listing 42", Description: "Another test listing"},
		ViewState: objects.ListingViewState(0),
	}
	err = s.Listings.Insert(testListing2.ID, testListing2)
	check(err)

	return s, t
}

func TestGenerateVectorsTagOkay(t *testing.T) {
	var testTagName = "test"

	var vectors vectorFileOkay
	shop, _ := newTestTag()
	encodedBefore := mustEncode(t, shop)

	kp := initVectors(t, &vectors, shop.Manifest.ShopID)

	var before = vectorSnapshot{
		Value:   shop,
		Encoded: encodedBefore,
		Hash:    mustHashState(t, shop),
	}

	testCases := []struct {
		name     string
		op       OpString
		path     Path
		value    any
		expected func(*testing.T, objects.Tag)
	}{
		// rename
		{
			name:  "rename tag",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr(testTagName), Fields: []any{"Name"}},
			value: "New Name",
			expected: func(t *testing.T, tag objects.Tag) {
				assert.Equal(t, "New Name", tag.Name)
			},
		},
		// add listing
		{
			name:  "add listing to tag",
			op:    AppendOp,
			path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr(testTagName), Fields: []any{"ListingIDs"}},
			value: objects.ObjectID(23),
			expected: func(t *testing.T, tag objects.Tag) {
				assert.Equal(t, 4, len(tag.ListingIDs))
				assert.Equal(t, objects.ObjectID(23), tag.ListingIDs[3])
			},
		},
		// remove listing
		{
			name: "remove listing from tag",
			op:   RemoveOp,
			path: Path{Type: ObjectTypeTag, TagName: testhelper.Strptr(testTagName), Fields: []any{"ListingIDs", 0}},
			expected: func(t *testing.T, tag objects.Tag) {
				assert.Equal(t, 2, len(tag.ListingIDs))
				assert.Equal(t, objects.ObjectID(2), tag.ListingIDs[0])
				assert.Equal(t, objects.ObjectID(3), tag.ListingIDs[1])
			},
		},
		// replace listing ID
		{
			name:  "replace listing ID",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr(testTagName), Fields: []any{"ListingIDs", 0}},
			value: objects.ObjectID(42),
			expected: func(t *testing.T, tag objects.Tag) {
				assert.Equal(t, 3, len(tag.ListingIDs))
				assert.Equal(t, objects.ObjectID(42), tag.ListingIDs[0])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shop, _ := newTestTag()

			// round trip to make sure we can encode/decode the patch
			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			assert.Equal(t, tc.op, decodedPatch.Op)

			patcher := NewPatcher(validate, &shop)
			err := patcher.ApplyPatch(decodedPatch)
			assert.Nil(t, err)
			tag, ok := shop.Tags.Get(testTagName)
			assert.True(t, ok)
			tc.expected(t, tag)

			var entry vectorEntryOkay
			entry.Name = t.Name()
			entry.Before = before
			encoded := mustEncode(t, shop)
			entry.After = vectorSnapshot{
				Value:   shop,
				Encoded: encoded,
				Hash:    mustHashState(t, shop),
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)
			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, patch)
		})
	}

	// sign the patch set
	kp.TestSignPatchSet(t, &vectors.PatchSet)

	writeVectors(t, vectors)
}

func TestGenerateVectorsTagError(t *testing.T) {
	testCases := []struct {
		name  string
		op    OpString
		path  Path
		value any
		error string
	}{
		// can only replace name, not add or remove
		{
			name:  "add to name",
			op:    AddOp,
			path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr("test"), Fields: []any{"name"}},
			value: "New Name",
			error: "unsupported field: name",
		},
		{
			name:  "remove from name",
			op:    RemoveOp,
			path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr("test"), Fields: []any{"name"}},
			value: "New Name",
			error: "unsupported field: name",
		},
		// listingIds need numerical index
		{
			name:  "add listing to tag",
			op:    AddOp,
			path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr("test"), Fields: []any{"listingIds", "helloWorld"}},
			value: objects.ObjectID(23),
			error: "unsupported field: helloWorld",
		},
		{
			name:  "remove listing from tag",
			op:    RemoveOp,
			path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr("test"), Fields: []any{"listingIds", "helloWorld"}},
			value: objects.ObjectID(23),
			error: "unsupported field: helloWorld",
		},
		{
			name:  "replace listing in tag",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr("test"), Fields: []any{"listingIds", "helloWorld"}},
			value: objects.ObjectID(42),
			error: "unsupported field: helloWorld",
		},
		// cant remove or replace non-existent listing
		{
			name:  "remove non-existent listing from tag",
			op:    RemoveOp,
			path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr("test"), Fields: []any{"listingIds", "999"}},
			value: objects.ObjectID(42),
			error: "index out of bounds: 999",
		},
		{
			name:  "replace non-existent listing in tag",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeTag, TagName: testhelper.Strptr("test"), Fields: []any{"listingIds", "999"}},
			value: objects.ObjectID(42),
			error: "index out of bounds: 999",
		},
	}

	var vectors vectorFileError

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shop, _ := newTestTag()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			encodedBefore := mustEncode(t, shop)

			patcher := NewPatcher(validate, &shop)
			err := patcher.ApplyPatch(decodedPatch)
			assert.NotEqual(t, nil, err)

			var entry vectorEntryError
			entry.Name = t.Name()
			entry.Patch = patch
			entry.Error = tc.error
			entry.Before = vectorSnapshot{
				Value:   shop,
				Encoded: encodedBefore,
				Hash:    mustHashState(t, shop),
			}
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	writeVectors(t, vectors)
}

var otherCurrency = objects.ChainAddress{
	ChainID:         1338,
	EthereumAddress: testMassEthAddr([20]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00}),
}

func newTestOrder() (objects.Shop, objects.Order) {
	s := newTestManifest()

	s.Manifest.Payees[1338] = map[common.Address]objects.PayeeMetadata{
		common.HexToAddress("0x1122334455667788990011223344556677889900"): {
			CallAsContract: true,
		},
	}

	s.Manifest.AcceptedCurrencies[1338] = map[common.Address]struct{}{
		otherCurrency.Address: {},
	}

	testListing1 := objects.Listing{
		ID:        5555,
		Price:     *big.NewInt(1000),
		Metadata:  objects.ListingMetadata{Title: "Test Listing 5555", Description: "A test listing"},
		ViewState: objects.ListingViewState(0),
	}
	err := s.Listings.Insert(testListing1.ID, testListing1)
	check(err)
	testListing2 := objects.Listing{
		ID:        5556,
		Price:     *big.NewInt(2000),
		Metadata:  objects.ListingMetadata{Title: "Test Listing 5556", Description: "A test listing"},
		ViewState: objects.ListingViewStatePublished,
	}
	err = s.Listings.Insert(testListing2.ID, testListing2)
	check(err)
	listing5557 := objects.Listing{
		ID:        5557,
		Price:     *big.NewInt(3000),
		Metadata:  objects.ListingMetadata{Title: "Test Listing 5557", Description: "Another test listing"},
		ViewState: objects.ListingViewStatePublished,
	}
	err = s.Listings.Insert(listing5557.ID, listing5557)

	o := objects.Order{
		ID:    666,
		State: objects.OrderStateOpen,
		Items: []objects.OrderedItem{
			{
				ListingID: 5555,
				Quantity:  23,
			},
			{
				ListingID: 5556,
				Quantity:  1,
			},
		},
		InvoiceAddress: &objects.AddressDetails{
			Name:         "John Doe",
			Address1:     "123 Main St",
			City:         "Anytown",
			PostalCode:   "12345",
			Country:      "US",
			EmailAddress: "john.doe@example.com",
		},
	}
	err = s.Orders.Insert(o.ID, o)
	check(err)

	o2 := clone.Clone(o)
	o2.ID = 667
	o2.State = objects.OrderStateCommitted
	o2.Items[0].Quantity = 55
	o2.Items[1] = objects.OrderedItem{
		ListingID: 5557,
		Quantity:  100,
	}
	o2.InvoiceAddress.Name = "Jane Doe"
	err = s.Orders.Insert(o2.ID, o2)
	check(err)

	return s, o
}

func TestGenerateVectorsOrderOkay(t *testing.T) {

	testPaymentDetails := objects.PaymentDetails{
		PaymentID: objects.Hash{0x01, 0x02, 0x03},
		Total:     *big.NewInt(1234567890),
		ListingHashes: [][]byte{
			testhelper.TestHash(5),
			testhelper.TestHash(6),
			testhelper.TestHash(7),
		},
		TTL:           100,
		ShopSignature: objects.Signature{0xff},
	}

	var vectors vectorFileOkay
	shop, _ := newTestOrder()
	encodedBefore := mustEncode(t, shop)
	var before = vectorSnapshot{
		Value:   shop,
		Encoded: encodedBefore,
		Hash:    mustHashState(t, shop),
	}

	kp := initVectors(t, &vectors, shop.Manifest.ShopID)

	testCases := []struct {
		name     string
		op       OpString
		path     Path
		value    any
		expected func(*testing.T, objects.Order)
	}{
		// item ops
		// ========

		{
			name:  "replace quantity of an item",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"Items", 0, "Quantity"}},
			value: uint32(42),
			expected: func(t *testing.T, o objects.Order) {
				assert.Equal(t, uint32(42), o.Items[0].Quantity)
				// ensure other order is not affected
				otherOrder, ok := shop.Orders.Get(667)
				assert.True(t, ok)
				assert.Equal(t, uint32(55), otherOrder.Items[0].Quantity)
				assert.Equal(t, "Jane Doe", otherOrder.InvoiceAddress.Name)
			},
		},
		{
			name: "append an item to an order",
			op:   AppendOp,
			path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"Items"}},
			value: objects.OrderedItem{
				ListingID: 5555,
				Quantity:  23,
			},
			expected: func(t *testing.T, o objects.Order) {
				assert.Equal(t, 3, len(o.Items))
				assert.Equal(t, uint32(23), o.Items[2].Quantity)
				assert.Equal(t, objects.ObjectID(5555), o.Items[2].ListingID)
			},
		},
		{
			name:  "increment item quantity",
			op:    IncrementOp,
			path:  Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"Items", 0, "Quantity"}},
			value: uint32(10),
			expected: func(t *testing.T, o objects.Order) {
				assert.Equal(t, uint32(33), o.Items[0].Quantity) // 23 + 10
				// ensure other order is not affected
				otherOrder, ok := shop.Orders.Get(667)
				assert.True(t, ok)
				assert.Equal(t, uint32(55), otherOrder.Items[0].Quantity)
			},
		},
		{
			name:  "decrement item quantity",
			op:    DecrementOp,
			path:  Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"Items", 0, "Quantity"}},
			value: uint32(5),
			expected: func(t *testing.T, o objects.Order) {
				assert.Equal(t, uint32(18), o.Items[0].Quantity) // 23 - 5
				// ensure other order is not affected
				otherOrder, ok := shop.Orders.Get(667)
				assert.True(t, ok)
				assert.Equal(t, uint32(55), otherOrder.Items[0].Quantity)
			},
		},
		{
			name: "remove an item from an order",
			op:   RemoveOp,
			path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"Items", 0}},
			expected: func(t *testing.T, o objects.Order) {
				assert.Equal(t, 1, len(o.Items))
			},
		},
		{
			name:  "remove all items from an order",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"Items"}},
			value: []objects.OrderedItem{},
			expected: func(t *testing.T, o objects.Order) {
				assert.Equal(t, 0, len(o.Items))
			},
		},

		// add ops
		// =======
		{
			name:  "replace invoice address name",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"InvoiceAddress", "Name"}},
			value: "John Doe",
			expected: func(t *testing.T, o objects.Order) {
				assert.Equal(t, nil, o.ShippingAddress)
				assert.Equal(t, "123 Main St", o.InvoiceAddress.Address1)
				assert.Equal(t, "Anytown", o.InvoiceAddress.City)
				assert.Equal(t, "12345", o.InvoiceAddress.PostalCode)
				assert.Equal(t, "John Doe", o.InvoiceAddress.Name)
			},
		},
		{
			name: "set shipping address",
			op:   AddOp,
			path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"ShippingAddress"}},
			value: &objects.AddressDetails{
				Name:         "Jane Doe",
				Address1:     "321 Other St",
				City:         "Othertown",
				PostalCode:   "67890",
				Country:      "US",
				EmailAddress: "jane.doe@example.com",
			},
			expected: func(t *testing.T, o objects.Order) {
				assert.NotEqual(t, nil, o.ShippingAddress)
				assert.Equal(t, "321 Other St", o.ShippingAddress.Address1)
				assert.Equal(t, "Othertown", o.ShippingAddress.City)
				assert.Equal(t, "67890", o.ShippingAddress.PostalCode)
			},
		},
		{
			name: "remove invoice address",
			op:   RemoveOp,
			path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"InvoiceAddress"}},
			expected: func(t *testing.T, o objects.Order) {
				assert.Equal(t, nil, o.InvoiceAddress)
			},
		},
		{
			name: "choose payee",
			op:   AddOp,
			path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"ChosenPayee"}},
			value: objects.Payee{
				Address: otherCurrency,
			},
			expected: func(t *testing.T, o objects.Order) {
				assert.NotEqual(t, nil, o.ChosenPayee)
				assert.Equal(t, otherCurrency, o.ChosenPayee.Address)
			},
		},
		{
			name:  "choose currency",
			op:    AddOp,
			path:  Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"ChosenCurrency"}},
			value: otherCurrency,
			expected: func(t *testing.T, o objects.Order) {
				assert.NotEqual(t, nil, o.ChosenCurrency)
				assert.Equal(t, otherCurrency, *o.ChosenCurrency)
			},
		},
		{
			name:  "add payment details",
			op:    AddOp,
			path:  Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"PaymentDetails"}},
			value: testPaymentDetails,
			expected: func(t *testing.T, o objects.Order) {
				assert.NotEqual(t, nil, o.PaymentDetails)
				assert.Equal(t, testPaymentDetails.PaymentID, o.PaymentDetails.PaymentID)
				assert.Equal(t, testPaymentDetails.Total, o.PaymentDetails.Total, ignoreBigInts)
				assert.Equal(t, testPaymentDetails.ListingHashes, o.PaymentDetails.ListingHashes)
				assert.Equal(t, testPaymentDetails.TTL, o.PaymentDetails.TTL)
				assert.Equal(t, testPaymentDetails.ShopSignature, o.PaymentDetails.ShopSignature)
			},
		},
		{
			name: "add tx details",
			op:   AddOp,
			path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"TxDetails"}},
			value: objects.OrderPaid{
				TxHash: &objects.Hash{0x01, 0x02, 0x03},
			},
			expected: func(t *testing.T, o objects.Order) {
				assert.NotEqual(t, nil, o.TxDetails)
				assert.Equal(t, &objects.Hash{0x01, 0x02, 0x03}, o.TxDetails.TxHash)
			},
		},

		// replace ops
		// ===========
		{
			name:  "replace items with empty array",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"Items"}},
			value: []objects.OrderedItem{},
			expected: func(t *testing.T, o objects.Order) {
				assert.Equal(t, 0, len(o.Items))
			},
		},
		{
			name: "replace payee",
			op:   ReplaceOp,
			path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"ChosenPayee"}},
			value: objects.Payee{
				Address: objects.ChainAddress{
					ChainID:         1338,
					EthereumAddress: testMassEthAddr([20]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00}),
				},
			},
			expected: func(t *testing.T, o objects.Order) {
				assert.NotEqual(t, nil, o.ChosenPayee)
				assert.Equal(t, objects.ChainAddress{
					ChainID:         1338,
					EthereumAddress: testMassEthAddr([20]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00}),
				}, o.ChosenPayee.Address)
			},
		},
		{
			name: "replace currency",
			op:   ReplaceOp,
			path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"ChosenCurrency"}},
			value: objects.ChainAddress{
				ChainID:         1338,
				EthereumAddress: testMassEthAddr([20]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00}),
			},
			expected: func(t *testing.T, o objects.Order) {
				assert.NotEqual(t, nil, o.ChosenCurrency)
				assert.Equal(t, &objects.ChainAddress{
					ChainID:         1338,
					EthereumAddress: testMassEthAddr([20]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0x00}),
				}, o.ChosenCurrency)
			},
		},

		{
			name:  "replace payment details",
			op:    ReplaceOp,
			path:  Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"PaymentDetails"}},
			value: testPaymentDetails,
			expected: func(t *testing.T, o objects.Order) {
				assert.NotEqual(t, nil, o.PaymentDetails)
				assert.Equal(t, testPaymentDetails.PaymentID, o.PaymentDetails.PaymentID)
			},
		},

		{
			name: "replace tx details",
			op:   ReplaceOp,
			path: Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"TxDetails"}},
			value: objects.OrderPaid{
				TxHash: &objects.Hash{0x04, 0x05, 0x06},
			},
			expected: func(t *testing.T, o objects.Order) {
				assert.NotEqual(t, nil, o.TxDetails)
				assert.Equal(t, &objects.Hash{0x04, 0x05, 0x06}, o.TxDetails.TxHash)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shop, _ := newTestOrder()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			patcher := NewPatcher(validate, &shop)
			err := patcher.ApplyPatch(decodedPatch)
			assert.Nil(t, err)
			order, ok := shop.Orders.Get(*patch.Path.ObjectID)
			assert.True(t, ok)
			tc.expected(t, order)

			encodedAfter := mustEncode(t, shop)
			var entry = vectorEntryOkay{
				Name:   tc.name,
				Before: before,
				After: vectorSnapshot{
					Value:   shop,
					Encoded: encodedAfter,
					Hash:    mustHashState(t, shop),
				},
			}
			vectors.Snapshots = append(vectors.Snapshots, entry)
			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, patch)
		})
	}

	// sign the patchset
	kp.TestSignPatchSet(t, &vectors.PatchSet)

	writeVectors(t, vectors)
}

func TestGenerateVectorsOrderError(t *testing.T) {
	testCases := []struct {
		name     string
		op       OpString
		path     Path
		value    any
		errMatch string
	}{
		{
			name:     "unsupported path",
			op:       IncrementOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"State"}},
			errMatch: "incr/decr only works on path: [Items, x, Quantity]",
		},
		{
			name:     "unsupported path",
			op:       DecrementOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"State"}},
			errMatch: "incr/decr only works on path: [Items, x, Quantity]",
		},
		{
			name:     "unsupported field",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"Invalid"}},
			errMatch: "fields=[Invalid] not found",
		},
		{
			name:     "invalid item index",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"Items", 999}},
			value:    objects.OrderedItem{},
			errMatch: "fields=[Items 999] not found",
		},
		{
			name:     "invalid item index format",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"Items", "abc"}},
			value:    objects.OrderedItem{},
			errMatch: "fields=[Items abc] not found",
		},
		{
			name:     "missing address field",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"InvoiceAddress"}},
			errMatch: "Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name:     "invalid address field",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(666), Fields: []any{"InvoiceAddress", "Invalid"}},
			value:    "test",
			errMatch: "fields=[InvoiceAddress Invalid] not found",
		},
		{
			name:     "add item after commit",
			op:       AddOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(667), Fields: []any{"Items"}},
			value:    objects.OrderedItem{ListingID: 5555, Quantity: 1},
			errMatch: errCannotModdifyCommittedOrder.Error(),
		},
		{
			name:     "replace item after commit",
			op:       ReplaceOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(667), Fields: []any{"Items", 0}},
			value:    objects.OrderedItem{ListingID: 5555, Quantity: 2},
			errMatch: errCannotModdifyCommittedOrder.Error(),
		},
		{
			name:     "increment quantity after commit",
			op:       IncrementOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(667), Fields: []any{"Items", 0, "Quantity"}},
			value:    1,
			errMatch: errCannotModdifyCommittedOrder.Error(),
		},
		{
			name:     "decrement quantity after commit",
			op:       DecrementOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(667), Fields: []any{"Items", 0, "Quantity"}},
			value:    1,
			errMatch: errCannotModdifyCommittedOrder.Error(),
		},
		{
			name:     "remove item after commit",
			op:       RemoveOp,
			path:     Path{Type: ObjectTypeOrder, ObjectID: testhelper.Uint64ptr(667), Fields: []any{"Items", 0}},
			errMatch: errCannotModdifyCommittedOrder.Error(),
		},
	}

	var vectors vectorFileError

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shop, _ := newTestOrder()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := mustEncode(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)
			encodedBefore := mustEncode(t, shop)

			patcher := NewPatcher(validate, &shop)
			err := patcher.ApplyPatch(decodedPatch)
			assert.Error(t, err)
			t.Logf("error: %v", err)
			assert.True(t, strings.Contains(err.Error(), tc.errMatch))

			var entry vectorEntryError
			entry.Name = t.Name()
			entry.Patch = patch
			entry.Error = err.Error()
			entry.Before = vectorSnapshot{
				Value:   shop,
				Encoded: encodedBefore,
				Hash:    mustHashState(t, shop),
			}
			vectors.Patches = append(vectors.Patches, entry)
		})
	}

	writeVectors(t, vectors)
}

func (patch Path) MarshalJSON() ([]byte, error) {
	path := []any{patch.Type}
	var either bool
	if patch.ObjectID != nil {
		path = append(path, *patch.ObjectID)
		either = true
	} else if patch.AccountAddr != nil {
		path = append(path, *patch.AccountAddr)
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
		RootHash objects.Hash
		Proofs   []massmmr.Proof
	}

	_, listing := newTestListing()
	encodedListing := mustEncode(t, listing)

	vectors := []testCase{
		{
			Name: "SinglePatch",
			Patches: []Patch{
				{
					Op: AddOp,
					Path: Path{
						Type:     ObjectTypeListing,
						ObjectID: testhelper.Uint64ptr(1),
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
					Path: Path{
						Type:     ObjectTypeListing,
						ObjectID: testhelper.Uint64ptr(1),
					},
					Value: encodedListing,
				},
				{
					Op: AddOp,
					Path: Path{
						Type:     ObjectTypeListing,
						ObjectID: testhelper.Uint64ptr(2),
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
					Path: Path{
						Type:     ObjectTypeListing,
						ObjectID: testhelper.Uint64ptr(1),
					},
					Value: encodedListing,
				},
				{
					Op: AddOp,
					Path: Path{
						Type:     ObjectTypeListing,
						ObjectID: testhelper.Uint64ptr(2),
					},
					Value: encodedListing,
				},
				{
					Op: ReplaceOp,
					Path: Path{
						Type:     ObjectTypeListing,
						ObjectID: testhelper.Uint64ptr(1),
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
					Path: Path{
						Type:     ObjectTypeListing,
						ObjectID: testhelper.Uint64ptr(4),
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
					Path: Path{
						Type:     ObjectTypeListing,
						ObjectID: testhelper.Uint64ptr(5),
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
					Path: Path{
						Type:     ObjectTypeListing,
						ObjectID: testhelper.Uint64ptr(16),
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
			var tree massmmr.VerifierTree
			tc.RootHash, tree, err = RootHash(tc.Patches)
			assert.Nil(t, err)

			// Generate and store proofs for each patch
			tc.Proofs = make([]massmmr.Proof, len(tc.Patches))
			for i := range tc.Patches {
				proof, err := tree.MakeProof(uint64(i))
				assert.Nil(t, err)
				assert.NotEqual(t, nil, proof)
				tc.Proofs[i] = *proof
				err = tree.VerifyProof(*proof)
				assert.Nil(t, err)
			}
		})
	}

	// Write test vectors to file
	writeVectors(t, vectors)
}

func mustHashState(t *testing.T, s objects.Shop) []byte {
	hash, err := s.Hash()
	assert.Nil(t, err)
	return hash[:]
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
