// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package objects

import (
	"bytes"
	"fmt"
	"time"

	masscbor "github.com/masslbs/network-schema/go/cbor"
)

// Listing represents a listed item in a shop
type Listing struct {
	ID        ObjectID        `validate:"required,gt=0"`
	Price     Uint256         `validate:"required"`
	Metadata  ListingMetadata `validate:"required"`
	ViewState ListingViewState
	Options   ListingOptions `cbor:",omitempty" validate:"nonEmptyMapKeys"`

	// one for each combination of variations
	StockStatuses []ListingStockStatus `cbor:",omitempty"`
}

// ListingOptions maps from a variation title to a listing option
type ListingOptions map[string]ListingOption

// ListingStockStatus represents the stock status of a listing
type ListingStockStatus listingStockStatusHack

type listingStockStatusHack struct {
	VariationIDs []string // list of variation map keys

	// one of the following needs to be set
	InStock           *bool      `cbor:",omitempty"`
	ExpectedInStockBy *time.Time `cbor:",omitempty"`
}

// UnmarshalCBOR implements the cbor.Unmarshaler interface
func (ls *ListingStockStatus) UnmarshalCBOR(data []byte) error {
	var ls2 listingStockStatusHack
	dec := masscbor.DefaultDecoder(bytes.NewReader(data))
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

// ListingMetadata represents information about a listing
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

// ListingVariations maps from a variation title to a listing variation
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
	// ListingViewStateUnspecified is the default state of a listing
	ListingViewStateUnspecified ListingViewState = iota
	// ListingViewStatePublished is the state of a listing that is published
	ListingViewStatePublished
	// ListingViewStateDeleted is the state of a listing that is deleted
	ListingViewStateDeleted

	maxListingViewState
)

// UnmarshalCBOR unmarshals a ListingViewState from a byte slice
func (s *ListingViewState) UnmarshalCBOR(data []byte) error {
	dec := masscbor.DefaultDecoder(bytes.NewReader(data))
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
