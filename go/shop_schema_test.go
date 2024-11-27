// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"math/big"
	"sort"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

func TestSignatureIncomplete(t *testing.T) {
	r := require.New(t)

	// we Prepare a message that looks like a proper signature but is too short
	var short struct {
		Sig []byte
	}
	short.Sig = []byte{'x', 'x', 'x'}
	shortData, err := cbor.Marshal(short)
	r.NoError(err)
	t.Log("shortData:\n" + pretty(shortData))
	// The actual signature is 64 bytes
	var msg struct {
		Sig Signature
	}
	err = Decode(&msg, shortData)
	r.Error(err)
	r.EqualValues([64]byte{}, msg.Sig)
	r.IsType(ErrBytesTooShort{}, err)
}

func TestMissingField(t *testing.T) {
	r := require.New(t)

	type FakeListing struct {
		Price Uint256
		// looks like a listing but has no metadata
		// Metadata  ListingMetadata
		ViewState ListingViewState
	}
	var fl FakeListing
	twentythree := big.NewInt(230000)
	fl.Price = *twentythree
	fl.ViewState = ListingViewStateDeleted

	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)
	err := enc.Encode(fl)
	r.NoError(err)
	testData := buf.Bytes()
	t.Log("FakeListing:\n" + pretty(testData))

	var lis Listing
	err = Decode(&lis, testData)
	r.Error(err)
	r.IsType(ErrRequiredFieldMissing{}, err)

	// If we set metadata, it needs description
	var missingDesc struct {
		Price    Uint256
		Metadata struct {
			Title string
		}
		ViewState ListingViewState
	}
	missingDesc.Price = *twentythree
	missingDesc.ViewState = ListingViewStatePublished
	missingDesc.Metadata.Title = "test"

	buf.Reset()
	err = enc.Encode(missingDesc)
	r.NoError(err)
	missingDescData := buf.Bytes()
	t.Log("missingDescData:\n" + pretty(missingDescData))

	var lis2 Listing
	err = Decode(&lis2, missingDescData)
	r.Error(err)
	r.IsType(ErrRequiredFieldMissing{}, err, err.Error())
}

func TestCreateOp(t *testing.T) {
	var (
		err error
		buf bytes.Buffer
		r   = require.New(t)
		enc = DefaultEncoder(&buf)
	)

	var createListing Patch
	createListing.Op = AddOp
	createListing.Path = []any{"listing", ObjectId(*big.NewInt(238000))}

	var lis Listing
	lis.Metadata.Title = "test Listing"
	lis.Metadata.Description = "short desc"
	lis.Metadata.Images = []string{"https://http.cat/images/100.jpg"}
	price := big.NewInt(12345)
	lis.Price = *price

	err = enc.Encode(lis)
	r.NoError(err)
	lisBytes := buf.Bytes()
	buf.Reset()
	createListing.Value = lisBytes

	err = enc.Encode(createListing)
	r.NoError(err)

	opData := buf.Bytes()
	t.Log("OP encoded:")
	t.Log("\n" + pretty(opData))

	var rxOp Patch
	dec := DefaultDecoder(bytes.NewReader(opData))
	err = dec.Decode(&rxOp)
	r.NoError(err)
	r.Equal("listing", rxOp.Path[0])
	var rxLis Listing

	err = Decode(&rxLis, rxOp.Value)
	r.NoError(err)

	t.Logf("listing received: %+v", rxLis)
	r.EqualValues(lis, rxLis)
}

func TestCreateAllTypes(t *testing.T) {
	r := require.New(t)

	bigId := big.NewInt(12345)

	vanillaEth := addrFromHex(1, "0x0000000000000000000000000000000000000000")
	cases := []struct {
		typ      any
		required []string
	}{
		{Listing{
			Price:     *big.NewInt(12345),
			ViewState: ListingViewStatePublished,
			Metadata: ListingMetadata{
				Title:       "test Listing",
				Description: "short desc",
				Images:      []string{"https://http.cat/images/100.jpg"},
			},
			Options: map[string]ListingOption{
				"Color": {
					Title: "Farbe",
					Variations: map[string]ListingVariation{
						"R": {
							VariationInfo: ListingMetadata{Title: "Rot"},
							PriceModifier: PriceModifier{
								ModificationPrecents: *big.NewInt(95),
							},
						},
						"G": {
							VariationInfo: ListingMetadata{Title: "Grün"},
							PriceModifier: PriceModifier{
								ModificationAbsolute: ModificationAbsolute{
									Amount: *big.NewInt(161),
									Plus:   false,
								},
							},
						},
						"B": {
							VariationInfo: ListingMetadata{Title: "Blau"},
						},
					},
				},
			},
		}, []string{"Price", "Metadata", "ViewState"}},

		{Manifest{
			ShopId: *bigId,
			Payees: map[string]Payee{
				"ethereum": {
					CallAsContract: true,
					Address:        addrFromHex(1, "0x1234567890123456789012345678901234567890"),
				},
			},
			AcceptedCurrencies: []ChainAddress{
				vanillaEth,
			},
			PricingCurrency: vanillaEth,
			ShippingRegions: map[string]ShippingRegion{
				"default": {
					PriceModifiers: map[string]PriceModifier{
						"discount": {
							ModificationPrecents: *big.NewInt(95),
						},
						"static": {
							ModificationAbsolute: ModificationAbsolute{
								Amount: *big.NewInt(161),
								Plus:   false,
							},
						},
					},
				},
			},
		}, []string{"ShopId", "Payees", "AcceptedCurrencies", "PricingCurrency"}},

		{Account{
			KeyCards: []PublicKey{[32]byte{1, 2, 3}},
			Guest:    true,
		}, []string{"KeyCards", "Guest"}},
	}

	var buf bytes.Buffer
	for _, c := range cases {
		sort.Strings(c.required) // TODO: cryptix was lazy during test writing and did not sort the fields
		r.Equal(c.required, requiredFields(c.typ))
		buf.Reset()
		enc := DefaultEncoder(&buf)
		err := enc.Encode(c.typ)
		r.NoError(err)

		testData := buf.Bytes()
		t.Logf("encoded %T:\n%s", c.typ, pretty(testData))

		var decoded any
		switch c.typ.(type) {
		case Listing:
			var l Listing
			err = Decode(&l, testData)
			decoded = l
		case Manifest:
			var m Manifest
			err = Decode(&m, testData)
			decoded = m
		case Account:
			var a Account
			err = Decode(&a, testData)
			decoded = a
		default:
			t.Fatalf("unknown type: %T", c.typ)
		}
		r.NoError(err)
		r.EqualValues(c.typ, decoded)
	}
}
