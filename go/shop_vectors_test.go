// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/fxamacker/cbor/v2"
	clone "github.com/huandu/go-clone/generic"
	"github.com/stretchr/testify/require"
)

func TestEncodeShopElementOrdering(t *testing.T) {
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

// This vector exercises the mutations of the shop object.
// Mutations of objects in the shop (listing, order, etc) are tested seperatly.
// The vectors file is constructed slightly differently to the other vectors files.
// Instead of starting with the same state every time (before & after),
// we keep the same state for all the patches.
func TestEncodeShop(t *testing.T) {
	r := require.New(t)

	mustEncode := func(v any) cbor.RawMessage {
		data, err := Marshal(v)
		r.NoError(err)
		return data
	}

	type vectorEntry struct {
		Patch   Patch
		After   Shop
		Encoded []byte
		Hash    []byte
	}
	type vectorFile struct {
		Start   Shop
		Encoded []byte
		Patches []vectorEntry
		Hash    []byte
	}

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
			Value: mustEncode(yetAnotherPayee),
		},
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeManifest, Fields: []string{"shippingRegions", "germany"}},
			Value: mustEncode(ShippingRegion{Country: "DE"}),
		},
		{
			Op:   "remove",
			Path: PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "with-escrow"}},
		},
		// accounts
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeAccount, AccountID: &testAcc1Addr},
			Value: mustEncode(testAcc1),
		},
		{
			Op:   "remove",
			Path: PatchPath{Type: ObjectTypeAccount, AccountID: &testAcc1Addr, Fields: []string{"keyCards", "1"}},
		},
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeAccount, AccountID: &guestAccAddr},
			Value: mustEncode(testAcc2),
		},
		// listing
		{
			Op:   "add",
			Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(23)},
			Value: mustEncode(Listing{
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
			Value: mustEncode("https://http.cat/images/100.jpg"),
		},
		{
			Op:    "replace",
			Path:  PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(23), Fields: []string{"metadata", "images", "0"}},
			Value: mustEncode("https://http.cat/images/200.jpg"),
		},
		{
			Op:   "add",
			Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(42)},
			Value: mustEncode(Listing{
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
			Value: mustEncode(Listing{
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
			Value: mustEncode(Listing{
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
			Value: mustEncode(Tag{Name: "test-tag"}),
		},
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test-tag"), Fields: []string{"listingIds", "-"}},
			Value: mustEncode(ObjectId(23)),
		},
		{
			Op:    "add",
			Path:  PatchPath{Type: ObjectTypeTag, TagName: strptr("test-tag"), Fields: []string{"listingIds", "-"}},
			Value: mustEncode(ObjectId(42)),
		},
		// orders
		{
			Op:   "add",
			Path: PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(math.MaxUint64 - 1)},
			Value: mustEncode(Order{
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
			Value: mustEncode(Order{
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

	var vectors vectorFile
	vectors.Start = testShop()
	vectors.Encoded = mustEncode(vectors.Start)
	vectors.Hash = hash(vectors.Encoded)
	var state = testShop()
	for i, patch := range patches {
		t.Run(fmt.Sprintf("patch-%d", i), func(t *testing.T) {
			r := require.New(t)
			var entry = vectorEntry{
				Patch: patch,
			}
			err := patcher.Shop(&state, patch)
			r.NoError(err)
			now := clone.Clone(state)
			entry.After = now
			entry.Encoded = mustEncode(now)
			entry.Hash = hash(entry.Encoded)
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

func (accs Accounts) MarshalJSON() ([]byte, error) {
	// Convert account/userWallet addresses to hex strings for JSON encoding
	hexAccs := make(map[string]Account, len(accs))
	for addr, acc := range accs {
		hexAccs[hex.EncodeToString(addr[:])] = acc
	}
	return json.Marshal(hexAccs)
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
