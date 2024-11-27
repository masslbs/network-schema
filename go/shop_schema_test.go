// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
)

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

func init() {
	validate = DefaultValidator()
}

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
	dec := DefaultDecoder(bytes.NewReader(shortData))
	err = dec.Decode(&msg)
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
	dec := DefaultDecoder(bytes.NewReader(testData))
	err = dec.Decode(&lis)
	r.NoError(err)
	r.Error(validate.Struct(lis))

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
	dec = DefaultDecoder(bytes.NewReader(missingDescData))
	err = dec.Decode(&lis2)
	r.NoError(err)
	r.Error(validate.Struct(lis2))
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
	lis.ViewState = ListingViewStatePublished
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

	dec = DefaultDecoder(bytes.NewReader(rxOp.Value))
	err = dec.Decode(&rxLis)
	r.NoError(err)

	t.Logf("listing received: %+v", rxLis)
	r.EqualValues(lis, rxLis)
}

func TestCreateAllTypes(t *testing.T) {
	r := require.New(t)

	bigId := big.NewInt(12345)

	vanillaEth := addrFromHex(1, "0x0000000000000000000000000000000000000000")
	cases := []struct {
		typ any
	}{

		{Manifest{
			ShopId: *bigId,
			Payees: map[string]Payee{
				"ethereum": {
					CallAsContract: true,
					Address:        addrFromHex(1, "0x1234567890123456789012345678901234567890"),
				},
			},
			AcceptedCurrencies: []ChainAddress{vanillaEth},
			PricingCurrency:    vanillaEth,
			ShippingRegions: map[string]ShippingRegion{
				"default": {
					PriceModifiers: map[string]PriceModifier{
						"discount": {
							ModificationPrecents: big.NewInt(95),
						},
						"static": {
							ModificationAbsolute: &ModificationAbsolute{
								Amount: *big.NewInt(161),
								Plus:   false,
							},
						},
					},
				},
			},
		}},

		{Account{
			KeyCards: []PublicKey{[32]byte{1, 2, 3}},
			Guest:    true,
		}},

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
							VariationInfo: ListingMetadata{
								Title:       "Rot",
								Description: "short desc",
							},
							PriceModifier: PriceModifier{
								ModificationPrecents: big.NewInt(95),
							},
						},
						"G": {
							VariationInfo: ListingMetadata{
								Title:       "Grün",
								Description: "short desc",
							},
							PriceModifier: PriceModifier{
								ModificationAbsolute: &ModificationAbsolute{
									Amount: *big.NewInt(161),
									Plus:   false,
								},
							},
						},
						"B": {
							VariationInfo: ListingMetadata{
								Title:       "Blau",
								Description: "short desc",
							},
						},
					},
				},
			},
		}},

		{Tag{
			Name:       "test",
			ListingIds: []ObjectId{*bigId},
		}},

		// TODO: need to add validation based on the state
		{Order{
			Items: []OrderedItem{{
				ListingID: ObjectId(*bigId),
				Quantity:  1,
			}},
			State: OrderStateOpen,
		}},
	}

	var buf bytes.Buffer
	for _, c := range cases {
		r.NoError(validate.Struct(c.typ))
		buf.Reset()
		enc := DefaultEncoder(&buf)
		err := enc.Encode(c.typ)
		r.NoError(err)

		testData := buf.Bytes()
		t.Logf("encoded %T:\n%s", c.typ, pretty(testData))

		var decoded any
		switch c.typ.(type) {
		case Manifest:
			decoded, err = decode[Manifest](testData)
		case Listing:
			decoded, err = decode[Listing](testData)
		case Tag:
			decoded, err = decode[Tag](testData)
		case Account:
			decoded, err = decode[Account](testData)
		case Order:
			decoded, err = decode[Order](testData)
		default:
			t.Fatalf("unknown type: %T", c.typ)
		}
		r.NoError(err)
		r.NoError(validate.Struct(decoded))
		r.EqualValues(c.typ, decoded)
	}
}

func decode[T any](data []byte) (T, error) {
	var t T
	dec := DefaultDecoder(bytes.NewReader(data))
	err := dec.Decode(&t)
	return t, err
}
