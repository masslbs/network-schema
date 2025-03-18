// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package objects

import "github.com/ethereum/go-ethereum/common"

// Manifest defines metadata needed to operate a shop
type Manifest struct {
	// shop metadata lives in the NFT
	ShopID             Uint256 `validate:"required"`
	Payees             Payees
	AcceptedCurrencies ChainAddresses
	PricingCurrency    ChainAddress    // the currency listings are priced in
	ShippingRegions    ShippingRegions `json:",omitempty"`
}

// PayeeMetadata stores additional metadata about a payee
type PayeeMetadata struct {
	CallAsContract bool
}

// Payees maps from chain id to a map of addresses to payee metadata
type Payees map[uint64]map[common.Address]PayeeMetadata

// ChainAddresses maps from chain id to a map of addresses to a boolean
type ChainAddresses map[uint64]map[common.Address]struct{}

// ShippingRegions maps from a nickname to a shipping region
type ShippingRegions map[string]ShippingRegion
