// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"math/big"
	"reflect"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

func TestSignatureIncomplete(t *testing.T) {
	r := require.New(t)

	var msg struct {
		Sig Signature
	}

	// we Prepare a message that looks like a proper signature but is too short
	var short struct {
		Sig []byte
	}
	short.Sig = []byte{'x', 'x', 'x'}
	shortData, err := cbor.Marshal(short)
	r.NoError(err)
	t.Log("shortData:\n" + pretty(shortData))
	rd := bytes.NewReader(shortData)
	dec := DefaultDecoder(rd)

	err = dec.Decode(&msg)
	r.EqualValues([64]byte{}, msg.Sig)
	r.Error(err)
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
							PriceModifier: OrderPriceModifier{
								ModificationPrecents: *big.NewInt(95),
							},
						},
						"G": {
							VariationInfo: ListingMetadata{Title: "Grün"},
							PriceModifier: OrderPriceModifier{
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
					Country:  "",
					Postcode: "",
					City:     "",
					PriceModifiers: map[string]OrderPriceModifier{
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
	}

	var buf bytes.Buffer
	for _, c := range cases {
		r.Equal(c.required, requiredFields(c.typ))
		buf.Reset()
		enc := DefaultEncoder(&buf)
		err := enc.Encode(c.typ)
		r.NoError(err)

		testData := buf.Bytes()
		//t.Logf("encoded %T:\n%s", c.typ, pretty(testData))

		// Create a concrete instance of the same type
		rx := reflect.New(reflect.TypeOf(c.typ)).Interface()
		// Type assert to the correct type before passing to Decode
		switch c.typ.(type) {
		case Listing:
			var l Listing
			err = Decode(&l, testData)
			rx = l
		case Manifest:
			var m Manifest
			err = Decode(&m, testData)
			rx = m
		}

		r.NoError(err)
		r.EqualValues(c.typ, rx)
	}
}
