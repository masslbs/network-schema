// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"encoding"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/fission-codes/go-car-mirror/ipld"
	"github.com/fxamacker/cbor/v2"
)

type ErrBytesTooShort struct {
	Want, Got uint
}

func (err ErrBytesTooShort) Error() string {
	return fmt.Sprintf("not enough bytes. Expected %d but got %d", err.Want, err.Got)
}

type ErrRequiredFieldMissing struct {
	Want string
}

func (err ErrRequiredFieldMissing) Error() string {
	return fmt.Sprintf("missing required field %s", err.Want)
}

func Decode[V any](v V, data []byte) error {
	reqFields := requiredFields(v)
	var keys map[string]cbor.RawMessage

	// TODO: we could use maxDepth1 here
	dec := DefaultDecoder(bytes.NewReader(data))
	err := dec.Decode(&keys)
	if err != nil {
		return fmt.Errorf("data is not a map")
	}
	for _, r := range reqFields {
		_, has := keys[r]
		if !has {
			return ErrRequiredFieldMissing{r}
		}
	}
	// TODO: instead we could copy keys into v using reflection
	dec = DefaultDecoder(bytes.NewReader(data))
	return dec.Decode(&v)
}

// all fields that are not marked with cbor:",omitempty" are required
func requiredFields(v any) []string {
	tv := reflect.TypeOf(v)
	if tv.Kind() == reflect.Pointer {
		tv = tv.Elem()
	}
	if tv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("not a struct: %T Kind: %s", v, tv.Kind()))
	}
	fields := make([]string, 0, tv.NumField())
	for i := 0; i < tv.NumField(); i++ {
		field := tv.Field(i)
		tag := field.Tag.Get("cbor")
		// fmt.Printf("field: %s tag: %s\n", field.Name, tag)
		if strings.Contains(tag, ",omitempty") {
			continue
		}
		fields = append(fields, field.Name)
	}
	// TODO: cryptix was lazy during test writing and did not sort the fields
	sort.Strings(fields)
	return fields
}

/*
BASE TYPES
*/

type ObjectId = big.Int

// Signature represents a cryptographic signature
const SignatureSize = 64

type Signature [SignatureSize]byte

func (val *Signature) UnmarshalBinary(data []byte) error {
	if n := uint(len(data)); n != SignatureSize {
		return ErrBytesTooShort{Want: SignatureSize, Got: n}
	}
	copy(val[:], data)
	return nil
}

var _ encoding.BinaryUnmarshaler = (*Signature)(nil)

// PublicKey represents a public key
const PublicKeySize = 32

type PublicKey [PublicKeySize]byte

func (val *PublicKey) UnmarshalBinary(data []byte) error {
	if n := uint(len(data)); n != PublicKeySize {
		return ErrBytesTooShort{Want: PublicKeySize, Got: n}
	}
	copy(val[:], data)
	return nil
}

// Hash represents a cryptographic hash
const HashSize = 32

type Hash [HashSize]byte

func (val *Hash) UnmarshalBinary(data []byte) error {
	if n := uint(len(data)); n != HashSize {
		return ErrBytesTooShort{Want: HashSize, Got: n}
	}
	copy(val[:], data)
	return nil
}

// EthereumAddress represents an Ethereum address
const EthereumAddressSize = 20

type EthereumAddress [EthereumAddressSize]byte

func (val *EthereumAddress) UnmarshalBinary(data []byte) error {
	if n := uint(len(data)); n != EthereumAddressSize {
		return ErrBytesTooShort{Want: EthereumAddressSize, Got: n}
	}
	copy(val[:], data)
	return nil
}

// Uint256 represents a 256-bit unsigned integer
type Uint256 = big.Int

// An ethereum address with a chain ID attached
type ChainAddress struct {
	ChainID uint64
	// when repsenting an ERC20 the zero address is used native currency
	Address EthereumAddress
}

func (ca *ChainAddress) UnmarshalBinary(data []byte) error {
	return Decode(ca, data)
}

func addrFromHex(chain uint64, hexAddr string) ChainAddress {
	addr := ChainAddress{ChainID: chain}
	hexAddr = strings.TrimPrefix(hexAddr, "0x")
	decoded, err := hex.DecodeString(hexAddr)
	check(err)
	n := copy(addr.Address[:], decoded)
	if n != EthereumAddressSize {
		panic(fmt.Sprintf("copy failed: %d != %d", n, EthereumAddressSize))
	}
	return addr
}

// Payee represents a payment recipient
type Payee struct {
	Address ChainAddress

	// controls how the payment is reaches the payee.
	// true:  forwarded via pay() method
	// false: normal transfer
	// See also:
	// https://github.com/masslbs/contracts/
	// commit: 377aba24796e029945696350db581ec1f65da657
	// file: src/IPayments.sol#L90-L95.
	CallAsContract bool
}

/*
The Manifest schema
*/
type Manifest struct {
	// shop metadata lives in the NFT
	ShopId Uint256
	// maps payee names to payee objects
	Payees map[string]Payee
	// TODO: should we add a name field to the acceptedCurrencies object?
	AcceptedCurrencies []ChainAddress
	// the currency listings are priced in
	PricingCurrency ChainAddress
	ShippingRegions map[string]ShippingRegion `cbor:",omitempty"`
}

type ShippingRegion shippingRegionHack

type shippingRegionHack struct {
	Country        string
	Postcode       string
	City           string
	PriceModifiers map[string]PriceModifier `cbor:",omitempty"`
}

func (sr *ShippingRegion) UnmarshalCBOR(data []byte) error {
	var tmp shippingRegionHack
	err := Decode(&tmp, data)
	if err != nil {
		return err
	}
	*sr = ShippingRegion(tmp)
	return nil
}

// ListingViewState represents the publication state of a listing
type PriceModifier struct {
	// one of the following should be set
	// this is multiplied with the sub-total before being divided by 100.
	ModificationPrecents Uint256              `cbor:",omitempty"`
	ModificationAbsolute ModificationAbsolute `cbor:",omitempty"`
}

type ModificationAbsolute struct {
	Amount Uint256
	Plus   bool // false means subtract
}

/*
Listing schema
*/
type Listing struct {
	Price     Uint256
	Metadata  ListingMetadata
	ViewState ListingViewState
	// TODO: how do we enforce sorting these? maybe maps only..?
	Options map[string]ListingOption `cbor:",omitempty"`
	// one for each combination of variations
	StockStatuses []ListingStockStatus `cbor:",omitempty"`
}

type ListingStockStatus listingStockStatusHack

type listingStockStatusHack struct {
	VariationIDs []uint64
	// one of the following should be set
	InStock           bool      `cbor:",omitempty"`
	ExpectedInStockBy time.Time `cbor:",omitempty"`
}

func (l *ListingStockStatus) UnmarshalCBOR(data []byte) error {
	var tmp listingStockStatusHack
	err := Decode(&tmp, data)
	if err != nil {
		return err
	}
	*l = ListingStockStatus(tmp)
	return nil
}

// ListingMetadata represents listing information
type ListingMetadata listingMetadataHack

type listingMetadataHack struct {
	Title       string
	Description string
	Images      []string `cbor:",omitempty"`
}

func (l *ListingMetadata) UnmarshalCBOR(data []byte) error {
	var tmp listingMetadataHack
	err := Decode(&tmp, data)
	if err != nil {
		return err
	}
	*l = ListingMetadata(tmp)
	return nil
}

// ListingOption represents a product option
type ListingOption listingOptionHack

type listingOptionHack struct {
	// the title of the option (like Color, Size, etc.)
	Title      string
	Variations map[string]ListingVariation `cbor:",omitempty"`
}

func (lo *ListingOption) UnmarshalCBOR(data []byte) error {
	var tmp listingOptionHack
	err := Decode(&tmp, data)
	if err != nil {
		return err
	}
	*lo = ListingOption(tmp)
	return nil
}

// ListingVariation represents a variation of a product option
type ListingVariation struct {
	// the metadata of the variation: for example if the option is Color
	// then the title might be "Red"
	VariationInfo ListingMetadata
	PriceModifier PriceModifier
	SKU           string
}

// ListingViewState represents the publication state of a listing
type ListingViewState int

const (
	ListingViewStateUnspecified ListingViewState = iota
	ListingViewStatePublished
	ListingViewStateDeleted

	maxListingViewState
)

func (s *ListingViewState) UnmarshalCBOR(data []byte) error {
	dec := DefaultDecoder(bytes.NewReader(data))
	var i int
	err := dec.Decode(&i)
	if err != nil {
		return err
	}
	if i < 0 || int(i) >= int(maxListingViewState) {
		return fmt.Errorf("invalid listing view state: %d", i)
	}
	*s = ListingViewState(i)
	return nil
}

/*
Account schema
*/
type Account struct {
	KeyCards []PublicKey
	Guest    bool
}

func (a *Account) UnmarshalCBOR(data []byte) error {
	var tmp struct {
		KeyCards []PublicKey
		Guest    bool
	}
	err := Decode(&tmp, data)
	if err != nil {
		return err
	}
	*a = Account(tmp)
	return nil
}

/*
Oder Schema
*/
type Order struct {
	Items           []OrderedItem
	State           OrderState
	InvoiceAddress  AddressDetails `cbor:",omitempty"`
	ShippingAddress AddressDetails `cbor:",omitempty"`
	CanceledAt      time.Time      `cbor:",omitempty"`
	ChosenPayee     Payee          `cbor:",omitempty"`
	ChosenCurrency  ChainAddress   `cbor:",omitempty"`
	PaymentDetails  PaymentDetails `cbor:",omitempty"`
	TxDetails       OrderPaid      `cbor:",omitempty"`
}

// OrderedItem represents an item in an order
type OrderedItem struct {
	ListingID    uint64
	VariationIDs []uint64
	Quantity     uint32
}

// OrderState represents the possible states of an order
type OrderState int

const (
	OrderStateUnspecified OrderState = 0
	OrderStateOpen        OrderState = 1
	OrderStateCanceled    OrderState = 2
	OrderStateCommited    OrderState = 3
	OrderStateUnpaid      OrderState = 4
	OrderStatePaid        OrderState = 5
)

// AddressDetails represents shipping or billing address information
type AddressDetails struct {
	Name         string
	Address1     string
	Address2     string `cbor:",omitempty"`
	City         string
	PostalCode   string `cbor:",omitempty"` // Malta does use postal codes
	Country      string
	EmailAddress string
	PhoneNumber  *string `cbor:",omitempty"`
}

type PaymentDetails struct {
	PaymentID     Hash
	Total         Uint256
	ListingHashes []ipld.Cid
	TTL           uint64 // The time to live in block
	ShopSignature Signature
}

type OrderPaid struct {
	TxHash    Hash `cbor:"1,keyasint"`
	BlockHash Hash `cbor:"2,keyasint"`
}

/*
Tags schema
*/
type Tag struct {
	Name       string
	ListingIds []uint64
}

/*
The complete Shop state
*/
type Shop struct {
	Manifest Manifest
	Listings []Listing
	Accounts map[EthereumAddress]Account
	Orders   []Order
	Tags     []Tag
}
