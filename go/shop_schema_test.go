// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/ipfs/go-cid"
	mh "github.com/multiformats/go-multihash"
	"github.com/stretchr/testify/require"
)

// use a single instance of Validate, it caches struct info
var validate = DefaultValidator()

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

func TestMissingFields(t *testing.T) {
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

func TestCreateAllTypes(t *testing.T) {

	bigId := big.NewInt(12345)

	testAddress := &AddressDetails{
		Name:         "test",
		Address1:     "test",
		City:         "test",
		PostalCode:   "test",
		Country:      "test",
		EmailAddress: "test@foo.bar",
		PhoneNumber:  strptr("+21911223344"),
	}
	expectedInStockBy := time.Unix(9999999999, 0)

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
			ID:        1,
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
			StockStatuses: []ListingStockStatus{
				{
					VariationIDs: []string{"r"},
					InStock:      boolptr(true),
				},
				{
					VariationIDs: []string{"m"},
					InStock:      boolptr(false),
				},
				{
					VariationIDs:      []string{"b"},
					ExpectedInStockBy: &expectedInStockBy,
				},
			},
		}},

		{Tag{
			Name:       "test",
			ListingIds: []uint64{1, 2, 3},
		}},

		{Order{
			ID: math.MaxUint64,
			Items: []OrderedItem{{
				ListingID: 1,
				Quantity:  1,
			}},
			State: OrderStateOpen,
		}},

		{Order{
			ID: math.MaxUint64 - 1,
			Items: []OrderedItem{{
				ListingID: 1,
				Quantity:  1,
			}},
			State: OrderStateCommited,
			ChosenPayee: &Payee{
				CallAsContract: true,
				Address:        addrFromHex(1, "0x1234567890123456789012345678901234567890"),
			},
			ChosenCurrency: &vanillaEth,
			InvoiceAddress: testAddress,
		}},

		{Order{
			ID: math.MaxUint64 - 2,
			Items: []OrderedItem{{
				ListingID: 1,
				Quantity:  1,
			}},
			State: OrderStateUnpaid,
			ChosenPayee: &Payee{
				CallAsContract: true,
				Address:        addrFromHex(1, "0x1234567890123456789012345678901234567890"),
			},
			ChosenCurrency: &vanillaEth,
			InvoiceAddress: testAddress,
			PaymentDetails: &PaymentDetails{
				TTL:       1000,
				PaymentID: Hash{},
				ListingHashes: []cid.Cid{
					testHash(0),
					testHash(1),
				},
			},
		}},

		{Order{
			ID: math.MaxUint64 - 3,
			Items: []OrderedItem{{
				ListingID: 1,
				Quantity:  1,
			}},
			State: OrderStatePaid,
			ChosenPayee: &Payee{
				CallAsContract: true,
				Address:        addrFromHex(1, "0x1234567890123456789012345678901234567890"),
			},
			ChosenCurrency: &vanillaEth,
			InvoiceAddress: testAddress,
			PaymentDetails: &PaymentDetails{
				TTL:       1000,
				PaymentID: Hash{},
				ListingHashes: []cid.Cid{
					testHash(0),
					testHash(1),
				},
			},
			TxDetails: &OrderPaid{
				TxHash: &Hash{},
			},
		}},
	}

	var buf bytes.Buffer
	for i, c := range cases {
		t.Run(fmt.Sprintf("index:%d/type:%T", i, c.typ), func(t *testing.T) {
			r := require.New(t)
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
		})
	}
}

// utils

func decode[T any](data []byte) (T, error) {
	var t T
	dec := DefaultDecoder(bytes.NewReader(data))
	err := dec.Decode(&t)
	return t, err
}

func testHash(i uint) cid.Cid {
	h, err := mh.Sum([]byte(fmt.Sprintf("TEST-%d", i)), mh.SHA3, 4)
	check(err)
	// TODO: check what the codec number should be
	return cid.NewCidV1(666, h)
}

func strptr(s string) *string {
	return &s
}

func boolptr(b bool) *bool {
	return &b
}

func uint64ptr(i uint64) *uint64 {
	return &i
}
