// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package objects

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"math/big"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fxamacker/cbor/v2"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/peterldowns/testy/assert"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/masslbs/network-schema/go/internal/testhelper"
)

func TestMapOrdering(t *testing.T) {
	shop := NewShop(42)

	var buf bytes.Buffer
	enc := masscbor.DefaultEncoder(&buf)
	err := enc.Encode(shop)
	assert.Nil(t, err)
	// Check field ordering in encoded shop
	// The fields should be ordered: Tags, Orders, Accounts, Listings, Manifest
	// This corresponds to the following CBOR structure:
	want := []byte{
		0xa7, // map(7)
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
		0x6d, // text(12)
		'S', 'c', 'h', 'e', 'm', 'a', 'V', 'e', 'r', 's', 'i', 'o', 'n',
		0x18, 0x2a, // unsigned(42)
	}
	got := buf.Bytes()
	gotHex := hex.EncodeToString(got)
	t.Log("Got:", gotHex)
	assert.Equal(t, len(want), len(got))
	assert.Equal(t, want, got)
}

var validate = DefaultValidator()

func TestECDAPublicKeySize(t *testing.T) {
	priv, err := crypto.GenerateKey()
	assert.Nil(t, err)
	pk := crypto.CompressPubkey(&priv.PublicKey)
	assert.Equal(t, PublicKeySize, len(pk))
}

func TestSignatureIncomplete(t *testing.T) {
	// we Prepare a message that looks like a proper signature but is too short
	var short struct {
		Sig []byte
	}
	short.Sig = []byte{'x', 'x', 'x'}
	shortData, err := cbor.Marshal(short)
	assert.Nil(t, err)

	t.Log("shortData:\n" + pretty(shortData))
	// The actual signature is 64 bytes
	var msg struct {
		Sig Signature
	}
	dec := masscbor.DefaultDecoder(bytes.NewReader(shortData))
	err = dec.Decode(&msg)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, [65]byte{}, msg.Sig)
	var wantErr ErrBytesTooShort
	assert.True(t, errors.As(err, &wantErr))
}

func TestMissingFields(t *testing.T) {
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
	enc := masscbor.DefaultEncoder(&buf)
	err := enc.Encode(fl)
	assert.Nil(t, err)
	testData := buf.Bytes()
	t.Log("FakeListing:\n" + pretty(testData))

	var lis Listing
	dec := masscbor.DefaultDecoder(bytes.NewReader(testData))
	err = dec.Decode(&lis)
	assert.Nil(t, err)
	assert.NotEqual(t, nil, validate.Struct(lis))

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
	assert.Nil(t, err)
	missingDescData := buf.Bytes()
	t.Log("missingDescData:\n" + pretty(missingDescData))

	var lis2 Listing
	dec = masscbor.DefaultDecoder(bytes.NewReader(missingDescData))
	err = dec.Decode(&lis2)
	assert.Nil(t, err)
	assert.NotEqual(t, nil, validate.Struct(lis2))
}

func TestCreateAllTypes(t *testing.T) {
	bigID := big.NewInt(12345)

	testAddress := &AddressDetails{
		Name:         "test",
		Address1:     "test",
		City:         "test",
		PostalCode:   "test",
		Country:      "test",
		EmailAddress: "test@foo.bar",
		PhoneNumber:  testhelper.Strptr("+21911223344"),
	}
	expectedInStockBy := time.Unix(9999999999, 0).UTC()

	vanillaEth := MustAddrFromHex(1, "0x0000000000000000000000000000000000000000")
	cases := []struct {
		typ any
	}{
		{Manifest{
			ShopID: *bigID,
			Payees: Payees{
				1: {
					vanillaEth.Address: {
						CallAsContract: true,
					},
				},
			},
			AcceptedCurrencies: ChainAddresses{1: {vanillaEth.Address: {}}},
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
			KeyCards: []PublicKey{[PublicKeySize]byte{1, 2, 3}},
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
							PriceModifier: &PriceModifier{
								ModificationPrecents: big.NewInt(95),
							},
						},
						"G": {
							VariationInfo: ListingMetadata{
								Title:       "Grün",
								Description: "short desc",
							},
							PriceModifier: &PriceModifier{
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
					InStock:      testhelper.Boolptr(true),
				},
				{
					VariationIDs: []string{"m"},
					InStock:      testhelper.Boolptr(false),
				},
				{
					VariationIDs:      []string{"b"},
					ExpectedInStockBy: &expectedInStockBy,
				},
			},
		}},

		{Tag{
			Name:       "test",
			ListingIDs: []uint64{1, 2, 3},
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
			State: OrderStateCommitted,
			ChosenPayee: &Payee{
				CallAsContract: true,
				Address:        MustAddrFromHex(1, "0x1234567890123456789012345678901234567890"),
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
				Address:        MustAddrFromHex(1, "0x1234567890123456789012345678901234567890"),
			},
			ChosenCurrency: &vanillaEth,
			InvoiceAddress: testAddress,
			PaymentDetails: &PaymentDetails{
				TTL:       1000,
				PaymentID: Hash{},
				ListingHashes: [][]byte{
					testhelper.TestHash(0),
					testhelper.TestHash(1),
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
				Address:        MustAddrFromHex(1, "0x1234567890123456789012345678901234567890"),
			},
			ChosenCurrency: &vanillaEth,
			InvoiceAddress: testAddress,
			PaymentDetails: &PaymentDetails{
				TTL:       1000,
				PaymentID: Hash{},
				ListingHashes: [][]byte{
					testhelper.TestHash(0),
					testhelper.TestHash(1),
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
			assert.Nil(t, validate.Struct(c.typ))
			buf.Reset()
			enc := masscbor.DefaultEncoder(&buf)
			err := enc.Encode(c.typ)
			assert.Nil(t, err)

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
			assert.Nil(t, err)
			assert.Nil(t, validate.Struct(decoded))
			assert.Equal(t, c.typ, decoded, ignoreBigInts)
		})
	}
}

var ignoreBigInts = cmpopts.IgnoreUnexported(big.Int{})

// helpers

func decode[T any](data []byte) (T, error) {
	var t T
	dec := masscbor.DefaultDecoder(bytes.NewReader(data))
	err := dec.Decode(&t)
	return t, err
}

func pretty(data []byte) string {
	if os.Getenv("PRETTY") == "" {
		return hex.EncodeToString(data)
	}
	shell := exec.Command("cbor2pretty.rb")
	shell.Stdin = bytes.NewReader(data)

	out, err := shell.CombinedOutput()
	check(err)
	return string(out)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
