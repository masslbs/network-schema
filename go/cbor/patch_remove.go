// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"fmt"
	"slices"
	"strconv"
)

// Manifest
// ========

func (existing *Manifest) PatchRemove(fields []string) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchRemove manifest requires at least one field")
	}
	switch fields[0] {
	case "payees":
		return existing.Payees.PatchRemove(fields[1:])
	case "shippingRegions":
		return existing.ShippingRegions.PatchRemove(fields[1:])
	case "acceptedCurrencies":
		return existing.AcceptedCurrencies.PatchRemove(fields[1:])
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
}

func (existing Payees) PatchRemove(fields []string) error {
	if len(fields) != 1 {
		return fmt.Errorf("Payees requires exactly one field")
	}
	_, has := existing[fields[0]]
	if !has {
		return fmt.Errorf("payee not found: %s", fields[0])
	}
	delete(existing, fields[0])
	return nil
}

func (existing ShippingRegions) PatchRemove(fields []string) error {
	if len(fields) != 1 {
		return fmt.Errorf("ShippingRegions requires exactly one field")
	}
	_, has := existing[fields[0]]
	if !has {
		return fmt.Errorf("shipping region not found: %s", fields[0])
	}
	delete(existing, fields[0])
	return nil
}

func (existing *ChainAddresses) PatchRemove(fields []string) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchRemove acceptedCurrencies requires at least one field")
	}
	index, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("failed to convert index to int: %w", err)
	}
	if index < 0 || index >= len(*existing) {
		return fmt.Errorf("index out of bounds: %d", index)
	}
	*existing = slices.Delete(*existing, index, index+1)
	return nil
}

// Listing
// =======

func (existing *Listing) PatchRemove(fields []string) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchRemove requires at least one field")
	}
	switch fields[0] {
	case "metadata":
		return existing.Metadata.PatchRemove(fields[1:])
	case "stockStatuses":
		// TODO: need to make []ListingStockStatus it's own type
		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("failed to convert index to int: %w", err)
		}
		if index < 0 || index >= len(existing.StockStatuses) {
			return fmt.Errorf("index out of bounds: %d", index)
		}
		existing.StockStatuses = slices.Delete(existing.StockStatuses, index, index+1)
		return nil
	case "options":
		return existing.Options.PatchRemove(fields[1:])
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
}

func (existing *ListingMetadata) PatchRemove(fields []string) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchRemove metadata requires at least one field")
	}
	switch fields[0] {
	case "images":
		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("failed to convert index to int: %w", err)
		}
		if index < 0 || index >= len(existing.Images) {
			return fmt.Errorf("index out of bounds: %d", index)
		}
		existing.Images = slices.Delete(existing.Images, index, index+1)
		return nil
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
}

func (existing *ListingOptions) PatchRemove(fields []string) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchRemove options requires at least one field")
	}
	if len(fields) == 1 {
		_, has := (*existing)[fields[0]]
		if !has {
			return fmt.Errorf("option not found: %s", fields[0])
		}
		delete(*existing, fields[0])
		return nil
	}
	option, ok := (*existing)[fields[0]]
	if !ok {
		return fmt.Errorf("option not found: %s", fields[0])
	}
	if fields[1] == "variations" {
		return option.Variations.PatchRemove(fields[2:])
	}
	return fmt.Errorf("unsupported field: %s", fields[1])
}

func (existing *ListingVariations) PatchRemove(fields []string) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchRemove variations requires at least one field")
	}
	_, has := (*existing)[fields[0]]
	if !has {
		return fmt.Errorf("variation not found: %s", fields[0])
	}
	delete(*existing, fields[0])
	return nil
}

// Tag
// ===

func (existing *Tag) PatchRemove(fields []string) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchRemove tag requires at least one field")
	}
	switch fields[0] {
	case "listingIds":
		idx, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("failed to convert index to int: %w", err)
		}
		if idx < 0 || idx >= len(existing.ListingIds) {
			return fmt.Errorf("index out of bounds: %d", idx)
		}
		existing.ListingIds = slices.Delete(existing.ListingIds, idx, idx+1)
		return nil
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
}

// Order
// =====

func (existing *Order) PatchRemove(fields []string) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchRemove requires at least one field")
	}
	switch fields[0] {
	case "items":
		return existing.Items.PatchRemove(fields[1:])
	case "invoiceAddress":
		existing.InvoiceAddress = nil
	case "shippingAddress":
		existing.ShippingAddress = nil
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
	return nil
}

func (existing *OrderedItems) PatchRemove(fields []string) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchRemove items requires at least one field")
	}
	index, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("failed to convert index to int: %w", err)
	}
	if index < 0 || index >= len(*existing) {
		return fmt.Errorf("index out of bounds: %d", index)
	}
	*existing = slices.Delete(*existing, index, index+1)
	return nil
}
