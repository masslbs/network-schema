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
	"strings"
	"time"

	"github.com/fission-codes/go-car-mirror/ipld"
)

type ErrBytesTooShort struct {
	Want, Got uint
}

func (err ErrBytesTooShort) Error() string {
	return fmt.Sprintf("not enough bytes. Expected %d but got %d", err.Want, err.Got)
}

func Decode[T any](data []byte) (T, error) {
	var t T
	dec := DefaultDecoder(bytes.NewReader(data))
	err := dec.Decode(&t)
	return t, err
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
	ChainID uint64 `validate:"required,gt=0"`
	// when repsenting an ERC20 the zero address is used native currency
	Address EthereumAddress
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
	ShopId Uint256 `validate:"required"`
	// maps payee names to payee objects
	Payees map[string]Payee `validate:"required,dive,keys,required,notblank,endkeys,required"`
	// TODO: should we add a name field to the acceptedCurrencies object?
	AcceptedCurrencies []ChainAddress `validate:"required,dive,required"`
	// the currency listings are priced in
	PricingCurrency ChainAddress              `validate:"required"`
	ShippingRegions map[string]ShippingRegion `cbor:",omitempty" validate:"dive,keys,required,notblank,endkeys,required"`
}

type ShippingRegion struct {
	Country        string
	Postcode       string
	City           string
	PriceModifiers map[string]PriceModifier `cbor:",omitempty" validate:"dive,keys,required,notblank,endkeys,required"`
}

// ListingViewState represents the publication state of a listing
type PriceModifier priceModifierHack

// using pointers here to express optionality clearer
type priceModifierHack struct {
	// one of the following should be set
	// this is multiplied with the sub-total before being divided by 100.
	ModificationPrecents *Uint256              `cbor:",omitempty"`
	ModificationAbsolute *ModificationAbsolute `cbor:",omitempty"`
}

func (pm *PriceModifier) UnmarshalCBOR(data []byte) error {
	var pm2 priceModifierHack
	dec := DefaultDecoder(bytes.NewReader(data))
	err := dec.Decode(&pm2)
	if err != nil {
		return err
	}
	if pm2.ModificationPrecents == nil && pm2.ModificationAbsolute == nil {
		return fmt.Errorf("one of ModificationPrecents or ModificationAbsolute must be set")
	}
	*pm = PriceModifier(pm2)
	return nil
}

type ModificationAbsolute struct {
	Amount Uint256 `validate:"required"`
	Plus   bool    // false means subtract
}

/*
Listing schema
*/
type Listing struct {
	Price     Uint256          `validate:"required"`
	Metadata  ListingMetadata  `validate:"required"`
	ViewState ListingViewState `validate:"required"`
	// TODO: how do we enforce sorting these? maybe maps only..?
	Options map[string]ListingOption `cbor:",omitempty" validate:"dive,keys,required,endkeys,required"`
	// one for each combination of variations
	StockStatuses []ListingStockStatus `cbor:",omitempty"`
}

type ListingStockStatus listingStockStatusHack

type listingStockStatusHack struct {
	VariationIDs []uint64
	// one of the following should be set
	InStock           *bool      `cbor:",omitempty"`
	ExpectedInStockBy *time.Time `cbor:",omitempty"`
}

func (ls *ListingStockStatus) UnmarshalCBOR(data []byte) error {
	var ls2 listingStockStatusHack
	dec := DefaultDecoder(bytes.NewReader(data))
	err := dec.Decode(&ls2)
	if err != nil {
		return err
	}
	// TODO: maybe add validate:"either_or=InStock,ExpectedInStockBy"`
	if ls2.InStock == nil && ls2.ExpectedInStockBy == nil {
		return fmt.Errorf("one of InStock or ExpectedInStockBy must be set")
	}
	*ls = ListingStockStatus(ls2)
	return nil
}

// ListingMetadata represents listing information
type ListingMetadata struct {
	Title       string   `validate:"required,notblank"`
	Description string   `validate:"required,notblank"`
	Images      []string `cbor:",omitempty"`
}

// ListingOption represents a product option
type ListingOption struct {
	// the title of the option (like Color, Size, etc.)
	Title      string                      `validate:"required,notblank"`
	Variations map[string]ListingVariation `cbor:",omitempty" validate:"dive,keys,required,endkeys,required"`
}

// ListingVariation represents a variation of a product option
type ListingVariation struct {
	// the metadata of the variation: for example if the option is Color
	// then the title might be "Red"
	VariationInfo ListingMetadata `validate:"required"`
	PriceModifier PriceModifier   `cbor:",omitempty"`
	SKU           string          `cbor:",omitempty"`
}

// ListingViewState represents the publication state of a listing
type ListingViewState uint

const (
	ListingViewStateUnspecified ListingViewState = iota
	ListingViewStatePublished
	ListingViewStateDeleted

	maxListingViewState
)

func (s *ListingViewState) UnmarshalCBOR(data []byte) error {
	dec := DefaultDecoder(bytes.NewReader(data))
	var i uint
	err := dec.Decode(&i)
	if err != nil {
		return err
	}
	if i == uint(ListingViewStateUnspecified) || i >= uint(maxListingViewState) {
		return fmt.Errorf("invalid listing view state: %d", i)
	}
	*s = ListingViewState(i)
	return nil
}

/*
Account schema
*/
type Account struct {
	KeyCards []PublicKey `validate:"required,gt=0"`
	Guest    bool
}

/*
Oder Schema
*/
type Order struct {
	Items           []OrderedItem   `validate:"required,gt=0"`
	State           OrderState      `validate:"required"`
	InvoiceAddress  *AddressDetails `cbor:",omitempty"`
	ShippingAddress *AddressDetails `cbor:",omitempty"`
	CanceledAt      *time.Time      `cbor:",omitempty"`
	ChosenPayee     *Payee          `cbor:",omitempty"`
	ChosenCurrency  *ChainAddress   `cbor:",omitempty"`
	PaymentDetails  *PaymentDetails `cbor:",omitempty"`
	TxDetails       *OrderPaid      `cbor:",omitempty"`
}

// OrderedItem represents an item in an order
type OrderedItem struct {
	ListingID    ObjectId `validate:"required"`
	VariationIDs []ObjectId
	Quantity     uint32 `validate:"required,gt=0"`
}

// OrderState represents the possible states of an order
type OrderState uint

const (
	OrderStateUnspecified OrderState = iota
	OrderStateOpen
	OrderStateCanceled
	OrderStateCommited
	OrderStateUnpaid
	OrderStatePaid

	maxOrderState
)

func (s *OrderState) UnmarshalCBOR(data []byte) error {
	dec := DefaultDecoder(bytes.NewReader(data))
	var i uint
	err := dec.Decode(&i)
	if err != nil {
		return err
	}
	if i == uint(OrderStateUnspecified) || i >= uint(maxOrderState) {
		return fmt.Errorf("invalid order state: %d", i)
	}
	*s = OrderState(i)
	return nil
}

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
	ListingHashes []ipld.Cid `validate:"required,gt=0"`
	TTL           uint64     `validate:"required,gt=0"` // The time to live in block
	ShopSignature Signature
}

// TODO: add either_or=TxHash,BlockHash
type OrderPaid struct {
	TxHash    Hash `cbor:"1,keyasint"`
	BlockHash Hash `cbor:"2,keyasint"`
}
