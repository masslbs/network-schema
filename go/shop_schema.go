package main

import (
	"math/big"
	"time"

	"github.com/fission-codes/go-car-mirror/ipld"
)

/*
BASE TYPES
*/

// Signature represents a cryptographic signature
type Signature [64]byte

// PublicKey represents a public key
type PublicKey [32]byte

// Hash represents a cryptographic hash
type Hash [32]byte

// EthereumAddress represents an Ethereum address
type EthereumAddress [20]byte

// Uint256 represents a 256-bit unsigned integer
type Uint256 big.Int

// An ethereum address with a chain ID attached
type ChainAddress struct {
	ChainID uint64
	// when repsenting an ERC20 the zero address is used native currency
	Address EthereumAddress
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
The Shop schema
*/
type Shop struct {
	Manifest Manifest
	Listings []Listing
	Accounts map[EthereumAddress]Account
	Orders   []Order
	Tags     []Tag
}

/*
The Manifest schema
*/
type Manifest struct {
	// shop metadata lives in the NFT
	ShopId Uint256
	// maps payee names to payee objects
	Payee map[string]Payee
	// TODO: should we add a name field to the acceptedCurrencies object?
	AcceptedCurrencies []ChainAddress
	// the currency listings are priced in
	PricingCurrency ChainAddress
	ShippingRegions map[string]ShippingRegion
}

type ShippingRegion struct {
	/* the location

	   the region for an order is picked by successivly matching fields.
	   empty-string values match everything / act as catch-all's.

	   therefore this can be used to say "only on this city" for pickups.
	   Or, for an international region, all three fields should be empty.

	   TODO: need a country map and dropdowns for matching to work
	*/
	Country        string
	Postcode       string
	City           string
	PriceModifiers map[string]OrderPriceModifier
}

// ListingViewState represents the publication state of a listing
type OrderPriceModifier struct {
	// one of the following should be set
	// this is multiplied with the sub-total before being divided by 100.
	ModificationPrecents Uint256 `cbor:",omitempty"`
	ModificationAbsolute Uint256 `cbor:",omitempty"`
}

/*
Listing schema
*/
type Listing struct {
	Price     Uint256
	Metadata  ListingMetadata
	ViewState ListingViewState
	Options   []ListingOption
	// one for each combination of variations
	StockStatuses []ListingStockStatus
}

type ListingStockStatus struct {
	VariationIDs []uint64
	// one of the following should be set
	InStock           bool      `cbor:",omitempty"`
	ExpectedInStockBy time.Time `cbor:",omitempty"`
}

// ListingMetadata represents listing information
type ListingMetadata struct {
	Title       string
	Description string
	Images      []string
}

// ListingOption represents a product option
type ListingOption struct {
	// the title of the option (like Color, Size, etc.)
	Title      string
	Variations []ListingVariation
}

// ListingVariation represents a variation of a product option
type ListingVariation struct {
	// the metadata of the variation: for example if the option is Color
	// then the title might be "Red"
	VariationInfo ListingMetadata
	PriceDiffSign bool
	PriceDiff     Uint256
	SKU           string
}

// ListingViewState represents the publication state of a listing
type ListingViewState int

const (
	ListingViewStateUnspecified ListingViewState = 0
	ListingViewStatePublished   ListingViewState = 1
	ListingViewStateDeleted     ListingViewState = 2
)

/*
Account schema
*/
type Account struct {
	KeyCards []PublicKey
	Guest    bool
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
	Name        string
	ListingIds []uint64
}
