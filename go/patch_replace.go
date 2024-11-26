// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fxamacker/cbor/v2"
)

// Manifest
// ========

func (existing *Manifest) PatchReplace(fields []string, value cbor.RawMessage) error {
	switch fields[0] {
	case "payees":
		return existing.Payees.PatchReplace(fields[1:], value)
	case "shippingRegions":
		return existing.ShippingRegions.PatchReplace(fields[1:], value)
	case "acceptedCurrencies":
		return existing.AcceptedCurrencies.PatchReplace(fields[1:], value)
	case "pricingCurrency":
		var currency ChainAddress
		err := Unmarshal(value, &currency)
		if err != nil {
			return fmt.Errorf("failed to unmarshal currency: %w", err)
		}
		existing.PricingCurrency = currency
		return nil
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
}

func (existing ChainAddresses) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return Unmarshal(value, &existing)
	}
	index, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("failed to convert index to int: %w", err)
	}
	if index < 0 || index >= len(existing) {
		return fmt.Errorf("index out of bounds: %d", index)
	}
	var currency ChainAddress
	err = Unmarshal(value, &currency)
	if err != nil {
		return fmt.Errorf("failed to unmarshal currency: %w", err)
	}
	existing[index] = currency
	return nil
}

func (existing Payees) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return Unmarshal(value, &existing)
	}
	if len(fields) != 1 {
		return fmt.Errorf("Payees requires exactly one field")
	}
	_, has := existing[fields[0]]
	if !has {
		return fmt.Errorf("payee not found: %s", fields[0])
	}
	var payee Payee
	err := Unmarshal(value, &payee)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payee: %w", err)
	}
	existing[fields[0]] = payee
	return nil
}

func (existing ShippingRegions) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return Unmarshal(value, &existing)
	}
	if len(fields) != 1 {
		return fmt.Errorf("ShippingRegions requires exactly one field")
	}
	_, has := existing[fields[0]]
	if !has {
		return fmt.Errorf("shipping region not found: %s", fields[0])
	}
	var region ShippingRegion
	err := Unmarshal(value, &region)
	if err != nil {
		return fmt.Errorf("failed to unmarshal shipping region: %w", err)
	}
	existing[fields[0]] = region
	return nil
}

// Listing
// =======

func (existing *Listing) PatchReplace(fields []string, value cbor.RawMessage) error {
	switch fields[0] {
	case "price":
		var price Uint256
		err := Unmarshal(value, &price)
		if err != nil {
			return fmt.Errorf("failed to unmarshal price: %w", err)
		}
		existing.Price = price
	case "metadata":
		err := existing.Metadata.PatchReplace(fields[1:], value)
		if err != nil {
			return fmt.Errorf("failed to replace metadata: %w", err)
		}
	case "viewState":
		var viewState ListingViewState
		err := Unmarshal(value, &viewState)
		if err != nil {
			return fmt.Errorf("failed to unmarshal viewState: %w", err)
		}
		existing.ViewState = viewState
	case "options":
		err := existing.Options.PatchReplace(fields[1:], value)
		if err != nil {
			return fmt.Errorf("failed to replace options: %w", err)
		}
	case "stockStatuses":
		if len(fields) == 1 {
			return fmt.Errorf("StockStatuses requires at least one field")
		}
		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("failed to convert index to int: %w", err)
		}
		if len(fields) == 2 {
			if index < 0 || index >= len(existing.StockStatuses) {
				return fmt.Errorf("index out of bounds: %d", index)
			}
			var newStatus ListingStockStatus
			err = Unmarshal(value, &newStatus)
			if err != nil {
				return fmt.Errorf("failed to unmarshal stock status: %w", err)
			}
			existing.StockStatuses[index] = newStatus
		} else if len(fields) == 3 {
			switch fields[2] {
			case "expectedInStockBy":
				var expectedInStockBy time.Time
				err := Unmarshal(value, &expectedInStockBy)
				if err != nil {
					return fmt.Errorf("failed to unmarshal expectedInStockBy: %w", err)
				}
				existing.StockStatuses[index].InStock = nil
				existing.StockStatuses[index].ExpectedInStockBy = &expectedInStockBy
			case "inStock":
				var inStock bool
				err := Unmarshal(value, &inStock)
				if err != nil {
					return fmt.Errorf("failed to unmarshal inStock: %w", err)
				}
				existing.StockStatuses[index].InStock = &inStock
				existing.StockStatuses[index].ExpectedInStockBy = nil
			default:
				return fmt.Errorf("unsupported field: %s", fields[2])
			}
		}
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
	return nil
}

func (existing *ListingMetadata) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 { // replace the whole metadata
		return Unmarshal(value, existing)
	}
	switch fields[0] {
	case "description":
		var description string
		err := Unmarshal(value, &description)
		if err != nil {
			return fmt.Errorf("failed to unmarshal description: %w", err)
		}
		existing.Description = description
	case "title":
		var title string
		err := Unmarshal(value, &title)
		if err != nil {
			return fmt.Errorf("failed to unmarshal title: %w", err)
		}
		existing.Title = title
	case "images":
		if len(fields) == 1 {
			var images []string
			err := Unmarshal(value, &images)
			if err != nil {
				return fmt.Errorf("failed to unmarshal images: %w", err)
			}
			existing.Images = images
		} else {
			index, err := strconv.Atoi(fields[1])
			if err != nil {
				return fmt.Errorf("failed to convert index to int: %w", err)
			}
			if index < 0 || index >= len(existing.Images) {
				return fmt.Errorf("index out of bounds: %d", index)
			}
			var image string
			err = Unmarshal(value, &image)
			if err != nil {
				return fmt.Errorf("failed to unmarshal image: %w", err)
			}
			existing.Images[index] = image
		}
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
	return nil
}

func (existing *ListingOptions) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 { // replace the whole options
		return Unmarshal(value, existing)
	}
	option, ok := (*existing)[fields[0]]
	if !ok {
		return fmt.Errorf("option not found: %s", fields[0])
	}

	if len(fields) == 1 { // replace the whole option
		var newOption ListingOption
		err := Unmarshal(value, &newOption)
		if err != nil {
			return fmt.Errorf("failed to unmarshal new option: %w", err)
		}
		(*existing)[fields[0]] = newOption
		return nil
	}

	// patch a variation
	switch fields[1] {
	case "title":
		var title string
		err := Unmarshal(value, &title)
		if err != nil {
			return fmt.Errorf("failed to unmarshal title: %w", err)
		}
		option.Title = title
		(*existing)[fields[0]] = option
	case "variations":
		err := option.Variations.PatchReplace(fields[2:], value)
		if err != nil {
			return fmt.Errorf("failed to replace option variation: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported field: %s", fields[1])
	}

	return nil
}

func (existing *ListingVariations) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 { // replace the whole variations
		return Unmarshal(value, existing)
	}
	variation, ok := (*existing)[fields[0]]
	if !ok {
		return fmt.Errorf("variation not found: %s", fields[0])
	}
	if len(fields) == 1 { // replace the whole variation
		var newVariation ListingVariation
		err := Unmarshal(value, &newVariation)
		if err != nil {
			return fmt.Errorf("failed to unmarshal new variation: %w", err)
		}
		(*existing)[fields[0]] = newVariation
		return nil
	}
	switch fields[1] {
	case "variationInfo":
		var variationInfo ListingMetadata
		err := Unmarshal(value, &variationInfo)
		if err != nil {
			return fmt.Errorf("failed to unmarshal variation info: %w", err)
		}
		variation.VariationInfo = variationInfo
	case "priceModifier":
		var priceModifier PriceModifier
		err := Unmarshal(value, &priceModifier)
		if err != nil {
			return fmt.Errorf("failed to unmarshal price modifier: %w", err)
		}
		variation.PriceModifier = priceModifier
	default:
		return fmt.Errorf("unsupported field: %s", fields[1])
	}
	(*existing)[fields[0]] = variation
	return nil
}

// Tag
// ===

func (existing *Tag) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return fmt.Errorf("Tag requires at least one field")
	}
	switch fields[0] {
	case "name":
		var name string
		err := Unmarshal(value, &name)
		if err != nil {
			return fmt.Errorf("failed to unmarshal name: %w", err)
		}
		existing.Name = name
	case "listingIds":
		if len(fields) == 1 {
			// replace all listing ids
			var ids []ObjectId
			err := Unmarshal(value, &ids)
			if err != nil {
				return fmt.Errorf("failed to unmarshal ids: %w", err)
			}
			existing.ListingIds = ids
		} else {
			idx, err := strconv.Atoi(fields[1])
			if err != nil {
				return fmt.Errorf("failed to convert index to int: %w", err)
			}
			if idx < 0 || idx >= len(existing.ListingIds) {
				return fmt.Errorf("index out of bounds: %d", idx)
			}
			var id ObjectId
			err = Unmarshal(value, &id)
			if err != nil {
				return fmt.Errorf("failed to unmarshal id: %w", err)
			}
			existing.ListingIds[idx] = id
		}
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
	return nil
}

// Order
// =====

func (existing *Order) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return fmt.Errorf("Order requires at least one field")
	}
	switch fields[0] {
	case "items":
		return existing.Items.PatchReplace(fields[1:], value)
	case "chosenPayee":
		var payee Payee
		err := Unmarshal(value, &payee)
		if err != nil {
			return fmt.Errorf("failed to unmarshal payee: %w", err)
		}
		existing.ChosenPayee = &payee
	case "chosenCurrency":
		var currency ChainAddress
		err := Unmarshal(value, &currency)
		if err != nil {
			return fmt.Errorf("failed to unmarshal currency: %w", err)
		}
		existing.ChosenCurrency = &currency
	case "paymentDetails":
		var paymentDetails PaymentDetails
		err := Unmarshal(value, &paymentDetails)
		if err != nil {
			return fmt.Errorf("failed to unmarshal payment details: %w", err)
		}
		existing.PaymentDetails = &paymentDetails
	case "txDetails":
		var txDetails OrderPaid
		err := Unmarshal(value, &txDetails)
		if err != nil {
			return fmt.Errorf("failed to unmarshal tx details: %w", err)
		}
		existing.TxDetails = &txDetails
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
	return nil
}

func (existing *OrderedItems) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return Unmarshal(value, &existing)
	}
	index, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("failed to convert index to int: %w", err)
	}
	if index < 0 || index >= len(*existing) {
		return fmt.Errorf("index out of bounds: %d", index)
	}
	if len(fields) == 1 {
		return Unmarshal(value, &(*existing)[index])
	}
	switch fields[1] {
	case "quantity":
		var quantity uint32
		err := Unmarshal(value, &quantity)
		if err != nil {
			return fmt.Errorf("failed to unmarshal quantity: %w", err)
		}
		(*existing)[index].Quantity = quantity
	}
	return nil
}
