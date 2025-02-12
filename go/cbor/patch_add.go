// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"fmt"
	"net/url"
	"slices"
	"strconv"

	"github.com/fxamacker/cbor/v2"
)

// Manifest
// ========

func (existing *Manifest) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchAdd manifest requires at least one field")
	}
	switch fields[0] {
	case "payees":
		return existing.Payees.PatchAdd(fields[1:], value)
	case "acceptedCurrencies":
		return existing.AcceptedCurrencies.PatchAdd(fields[1:], value)
	case "shippingRegions":
		return existing.ShippingRegions.PatchAdd(fields[1:], value)
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
}

func (existing *ChainAddresses) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) < 1 {
		return fmt.Errorf("PatchAdd acceptedCurrencies requires at least one field")
	}
	var currency ChainAddress
	err := Unmarshal(value, &currency)
	if err != nil {
		return fmt.Errorf("failed to unmarshal currency: %w", err)
	}
	switch fields[0] {
	case "-": // append to the list
		*existing = append(*existing, currency)
	default:
		index, err := strconv.Atoi(fields[0])
		if err != nil {
			return fmt.Errorf("failed to convert index to int: %w", err)
		}
		*existing = slices.Insert(*existing, index, currency)
	}
	return nil
}
func (existing Payees) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) != 1 {
		return fmt.Errorf("PatchAdd payees requires exactly one field")
	}
	_, ok := existing[fields[0]]
	if ok {
		return fmt.Errorf("payee already exists: %s", fields[0])
	}
	var payee Payee
	err := Unmarshal(value, &payee)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payee: %w", err)
	}
	existing[fields[0]] = payee
	return nil
}

func (existing ShippingRegions) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) != 1 {
		return fmt.Errorf("PatchAdd shippingRegions requires exactly one field")
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

func (existing *Listing) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		var newListing Listing
		err := Unmarshal(value, &newListing)
		if err != nil {
			return fmt.Errorf("failed to unmarshal listing: %w", err)
		}
		*existing = newListing
		return nil
	}
	switch fields[0] {
	case "metadata":
		return existing.Metadata.PatchAdd(fields[1:], value)
	case "options":
		return existing.Options.PatchAdd(fields[1:], value)
	case "stockStatuses":
		var stockStatus ListingStockStatus
		err := Unmarshal(value, &stockStatus)
		if err != nil {
			return fmt.Errorf("failed to unmarshal stock status: %w", err)
		}
		switch fields[1] {
		case "-": // append to the list
			existing.StockStatuses = append(existing.StockStatuses, stockStatus)
			return nil
		default:
			index, err := strconv.Atoi(fields[1])
			if err != nil {
				return fmt.Errorf("failed to convert index to int: %w", err)
			}
			if index < 0 || index >= len(existing.StockStatuses) {
				return fmt.Errorf("index out of bounds: %d", index)
			}
			existing.StockStatuses = slices.Insert(existing.StockStatuses, index, stockStatus)
			return nil
		}
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
}

func (existing *ListingMetadata) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchAdd metadata requires at least one field")
	}
	switch fields[0] {
	case "images":
		var image string
		err := Unmarshal(value, &image)
		if err != nil {
			return fmt.Errorf("failed to unmarshal image: %w", err)
		}
		_, err = url.Parse(image)
		if err != nil {
			return fmt.Errorf("invalid image URL: %w", err)
		}
		switch fields[1] {
		case "-": // append to the list
			existing.Images = append(existing.Images, image)
		default:
			index, err := strconv.Atoi(fields[1])
			if err != nil {
				return fmt.Errorf("failed to convert index to int: %w", err)
			}
			if index < 0 || index >= len(existing.Images) {
				return fmt.Errorf("index out of bounds: %d", index)
			}
			existing.Images = slices.Insert(existing.Images, index, image)
		}
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
	return nil
}

func (existing *ListingOptions) PatchAdd(fields []string, value cbor.RawMessage) error {
	switch n := len(fields); n {
	case 0:
		return fmt.Errorf("PatchAdd options requires at least one field")
	case 1:
		// check if it's a new option
		_, ok := (*existing)[fields[0]]
		if ok {
			return fmt.Errorf("option already exists: %v", fields)
		}
		var option ListingOption
		err := Unmarshal(value, &option)
		if err != nil {
			return fmt.Errorf("failed to unmarshal option: %w", err)
		}
		(*existing)[fields[0]] = option
		return nil
	default:
		// adding a variation
		option, ok := (*existing)[fields[0]]
		if !ok {
			return fmt.Errorf("option not found: %v", fields)
		}
		if fields[1] != "variations" {
			return fmt.Errorf("unsupported field: %s", fields[1])
		}
		return option.Variations.PatchAdd(fields[2:], value)
	}
}

func (existing *ListingVariations) PatchAdd(fields []string, value cbor.RawMessage) error {
	switch n := len(fields); n {
	case 1: // replace the whole variations
		var variation ListingVariation
		err := Unmarshal(value, &variation)
		if err != nil {
			return fmt.Errorf("failed to unmarshal variation: %w", err)
		}
		(*existing)[fields[0]] = variation
		return nil
	default:
		return fmt.Errorf("PatchAdd variations requires at least one field got %d", n)
	}
}

// Tag
// ====

func (existing *Tag) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchAdd tag requires at least one field")
	}
	switch fields[0] {
	case "listingIds":
		if fields[1] == "-" {
			// append
			var id ObjectId
			err := Unmarshal(value, &id)
			if err != nil {
				return fmt.Errorf("failed to unmarshal id: %w", err)
			}
			existing.ListingIds = append(existing.ListingIds, id)
			return nil
		} else {
			index, err := strconv.Atoi(fields[1])
			if err != nil {
				return fmt.Errorf("failed to convert index to int: %w", err)
			}
			var id ObjectId
			err = Unmarshal(value, &id)
			if err != nil {
				return fmt.Errorf("failed to unmarshal id: %w", err)
			}
			existing.ListingIds = slices.Insert(existing.ListingIds, index, id)
		}
		return nil
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
}

// Order
// =====

func (existing *Order) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return Unmarshal(value, &existing)
	}
	switch fields[0] {
	case "items":
		return existing.Items.PatchAdd(fields[1:], value)
	case "invoiceAddress":
		if len(fields) == 1 {
			var tmp AddressDetails
			err := Unmarshal(value, &tmp)
			if err != nil {
				return fmt.Errorf("failed to unmarshal address details: %w", err)
			}
			existing.InvoiceAddress = &tmp
			return nil
		}
		return existing.InvoiceAddress.PatchAdd(fields[1:], value)
	case "shippingAddress":
		if len(fields) == 1 {
			var tmp AddressDetails
			err := Unmarshal(value, &tmp)
			if err != nil {
				return fmt.Errorf("failed to unmarshal address details: %w", err)
			}
			existing.ShippingAddress = &tmp
			return nil
		}
		return existing.ShippingAddress.PatchAdd(fields[1:], value)
	case "chosenPayee":
		var payee Payee
		err := Unmarshal(value, &payee)
		if err != nil {
			return fmt.Errorf("failed to unmarshal payee: %w", err)
		}
		existing.ChosenPayee = &payee
		return nil
	case "chosenCurrency":
		var currency ChainAddress
		err := Unmarshal(value, &currency)
		if err != nil {
			return fmt.Errorf("failed to unmarshal currency: %w", err)
		}
		existing.ChosenCurrency = &currency
		return nil
	case "paymentDetails":
		var paymentDetails PaymentDetails
		err := Unmarshal(value, &paymentDetails)
		if err != nil {
			return fmt.Errorf("failed to unmarshal payment details: %w", err)
		}
		existing.PaymentDetails = &paymentDetails
		return nil
	case "txDetails":
		var txDetails OrderPaid
		err := Unmarshal(value, &txDetails)
		if err != nil {
			return fmt.Errorf("failed to unmarshal tx details: %w", err)
		}
		existing.TxDetails = &txDetails
		return nil
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
}

func (existing *OrderedItems) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return Unmarshal(value, &existing)
	}
	if fields[0] == "-" {
		var item OrderedItem
		err := Unmarshal(value, &item)
		if err != nil {
			return fmt.Errorf("failed to unmarshal item: %w", err)
		}
		*existing = append(*existing, item)
		return nil
	}
	index, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("failed to convert index to int: %w", err)
	}
	if index < 0 || index >= len(*existing) {
		return fmt.Errorf("index out of bounds: %d", index)
	}
	var item OrderedItem
	err = Unmarshal(value, &item)
	if err != nil {
		return fmt.Errorf("failed to unmarshal item: %w", err)
	}
	(*existing)[index] = item
	return nil
}

func (existing *AddressDetails) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return fmt.Errorf("AddressDetails requires at least one field")
	}
	var strValue string
	err := Unmarshal(value, &strValue)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}
	switch fields[0] {
	case "name":
		existing.Name = strValue
	case "address1":
		existing.Address1 = strValue
	case "address2":
		existing.Address2 = strValue
	case "city":
		existing.City = strValue
	case "postalCode":
		existing.PostalCode = strValue
	case "country":
		existing.Country = strValue
	case "emailAddress":
		existing.EmailAddress = strValue
	case "phoneNumber":
		existing.PhoneNumber = &strValue
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
	return nil
}
