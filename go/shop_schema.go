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

	"github.com/go-playground/validator/v10"
	"github.com/ipfs/go-cid"
)

type ErrBytesTooShort struct {
	Want, Got uint
}

func (err ErrBytesTooShort) Error() string {
	return fmt.Sprintf("not enough bytes. Expected %d but got %d", err.Want, err.Got)
}

/*
BASE TYPES
*/

type ObjectId = uint64

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

// make sure these types implement encoding.BinaryUnmarshaler
var (
	_ encoding.BinaryUnmarshaler = (*Signature)(nil)
	_ encoding.BinaryUnmarshaler = (*PublicKey)(nil)
	_ encoding.BinaryUnmarshaler = (*Hash)(nil)
	_ encoding.BinaryUnmarshaler = (*EthereumAddress)(nil)
)

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
The complete Shop state
*/
type Shop struct {
	Tags     Tags   `validate:"nonEmptyMapKeys"`
	Orders   Orders `validate:"nonEmptyMapKeys"`
	Accounts Accounts
	Listings Listings `validate:"nonEmptyMapKeys"`
	Manifest Manifest `validate:"required"`
}

type Accounts map[EthereumAddress]Account

type Listings map[ObjectId]Listing

type Tags map[string]Tag

type Orders map[ObjectId]Order

/*
/*
Tags schema
*/
type Tag struct {
	Name       string `validate:"required,notblank"`
	ListingIds []ObjectId
}

/*
The Manifest schema
*/
type Manifest struct {
	// shop metadata lives in the NFT
	ShopId Uint256 `validate:"required"`
	// maps payee names to payee objects
	Payees Payees `validate:"nonEmptyMapKeys"`
	// TODO: should we add a name field to the acceptedCurrencies object?
	AcceptedCurrencies ChainAddresses `validate:"required,gt=0"`
	// the currency listings are priced in
	PricingCurrency ChainAddress    `validate:"required"`
	ShippingRegions ShippingRegions `cbor:",omitempty" validate:"nonEmptyMapKeys"`
}

type Payees map[string]Payee

type ShippingRegions map[string]ShippingRegion

type ChainAddresses []ChainAddress

type ShippingRegion struct {
	Country        string
	Postcode       string
	City           string
	PriceModifiers map[string]PriceModifier `cbor:",omitempty" validate:"nonEmptyMapKeys"`
}

// ListingViewState represents the publication state of a listing
type PriceModifier priceModifierHack

// using pointers here to express optionality clearer
// TODO: add validate:"either_or=ModificationPrecents,ModificationAbsolute"`
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
	ID        ObjectId        `validate:"required,gt=0"`
	Price     Uint256         `validate:"required"`
	Metadata  ListingMetadata `validate:"required"`
	ViewState ListingViewState
	Options   ListingOptions `cbor:",omitempty" validate:"nonEmptyMapKeys"`
	// one for each combination of variations
	StockStatuses []ListingStockStatus `cbor:",omitempty"`
}

type ListingOptions map[string]ListingOption

type ListingStockStatus listingStockStatusHack

type listingStockStatusHack struct {
	VariationIDs []string // list of variation map keys
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
	Title      string            `validate:"required,notblank"`
	Variations ListingVariations `cbor:",omitempty" validate:"nonEmptyMapKeys"`
}

type ListingVariations map[string]ListingVariation

// ListingVariation represents a variation of a product option
// It's ID is the map key it's associated with
type ListingVariation struct {
	// VariationInfo is the metadata of the variation: for example if the option is Color
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
	if i >= uint(maxListingViewState) {
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
	ID              ObjectId        `validate:"required,gt=0"`
	Items           OrderedItems    `validate:"required"`
	State           OrderState      `validate:"required"`
	InvoiceAddress  *AddressDetails `cbor:",omitempty"`
	ShippingAddress *AddressDetails `cbor:",omitempty"`
	CanceledAt      *time.Time      `cbor:",omitempty"`
	ChosenPayee     *Payee          `cbor:",omitempty"`
	ChosenCurrency  *ChainAddress   `cbor:",omitempty"`
	PaymentDetails  *PaymentDetails `cbor:",omitempty"`
	TxDetails       *OrderPaid      `cbor:",omitempty"`
}

type OrderedItems []OrderedItem

func OrderValidation(sl validator.StructLevel) {
	order := sl.Current().Interface().(Order)
	switch order.State {
	case OrderStatePaid:
		if order.TxDetails == nil {
			sl.ReportError(order.TxDetails, "TxDetails", "TxDetails", "required", "")
		}
		fallthrough
	case OrderStateUnpaid:
		if order.PaymentDetails == nil {
			sl.ReportError(order.PaymentDetails, "PaymentDetails", "PaymentDetails", "required", "")
		}
		fallthrough
	case OrderStateCommited:
		if order.ChosenPayee == nil {
			sl.ReportError(order.ChosenPayee, "ChosenPayee", "ChosenPayee", "required", "")
		}
		if order.ChosenCurrency == nil {
			sl.ReportError(order.ChosenCurrency, "ChosenCurrency", "ChosenCurrency", "required", "")
		}
		if order.InvoiceAddress == nil && order.ShippingAddress == nil {
			sl.ReportError(order.InvoiceAddress, "InvoiceAddress", "InvoiceAddress", "either_or", "")
			sl.ReportError(order.ShippingAddress, "ShippingAddress", "ShippingAddress", "either_or", "")
		}
	case OrderStateCanceled:
		if order.CanceledAt == nil {
			sl.ReportError(order.CanceledAt, "CanceledAt", "CanceledAt", "required", "")
		}
	case OrderStateOpen:
		// noop
	default:
		sl.ReportError(order.State, "State", "State", fmt.Sprintf("invalid order state: %d", order.State), "")
	}
}

// OrderedItem represents an item in an order
type OrderedItem struct {
	ListingID    ObjectId `validate:"required,gt=0"`
	VariationIDs []string `cbor:",omitempty"`
	Quantity     uint32   `validate:"required,gt=0"`
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
	Name         string  `validate:"required,notblank"`
	Address1     string  `validate:"required,notblank"`
	Address2     string  `cbor:",omitempty"`
	City         string  `validate:"required,notblank"`
	PostalCode   string  `cbor:",omitempty"` // Malta does use postal codes
	Country      string  `validate:"required,notblank"`
	EmailAddress string  `validate:"required,email"`
	PhoneNumber  *string `cbor:",omitempty" validate:"omitempty,e164"`
}

type PaymentDetails struct {
	PaymentID     Hash
	Total         Uint256
	ListingHashes []cid.Cid `validate:"required,gt=0"`
	TTL           uint64    `validate:"required,gt=0"` // The time to live in block
	ShopSignature Signature
}

type OrderPaid struct {
	TxHash    *Hash `cbor:",omitempty"`
	BlockHash Hash  `cbor:",omitempty"`
}

func (op *OrderPaid) UnmarshalCBOR(data []byte) error {
	var tmp struct {
		TxHash    *Hash `cbor:",omitempty"`
		BlockHash *Hash `cbor:",omitempty"`
	}
	dec := DefaultDecoder(bytes.NewReader(data))
	err := dec.Decode(&tmp)
	if err != nil {
		return err
	}
	if tmp.BlockHash == nil {
		return fmt.Errorf("BlockHash must be set")
	}
	op.BlockHash = *tmp.BlockHash
	op.TxHash = tmp.TxHash
	return nil
}

// the CBOR library does not know how to encode custom map types.
// so we need to cast them a bit.

func (a Accounts) MarshalCBOR() ([]byte, error) {
	return Marshal(map[EthereumAddress]Account(a))
}
func (s ShippingRegions) MarshalCBOR() ([]byte, error) {
	return Marshal(map[string]ShippingRegion(s))
}
func (p Payees) MarshalCBOR() ([]byte, error) {
	return Marshal(map[string]Payee(p))
}
func (l Listings) MarshalCBOR() ([]byte, error) {
	return Marshal(map[uint64]Listing(l))
}
func (lis ListingVariations) MarshalCBOR() ([]byte, error) {
	return Marshal(map[string]ListingVariation(lis))
}
func (lis ListingOptions) MarshalCBOR() ([]byte, error) {
	return Marshal(map[string]ListingOption(lis))
}
func (t Tags) MarshalCBOR() ([]byte, error) {
	return Marshal(map[string]Tag(t))
}
func (o Orders) MarshalCBOR() ([]byte, error) {
	return Marshal(map[uint64]Order(o))
}
