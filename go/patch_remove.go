// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"fmt"
	"strconv"
)

func (existing *Listing) PatchRemove(fields []string) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchRemove requires at least one field")
	}
	switch fields[0] {
	case "metadata":
		return existing.Metadata.PatchRemove(fields[1:])
	case "StockStatuses":
		// TODO: need to make []ListingStockStatus it's own type
		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("failed to convert index to int: %w", err)
		}
		if index < 0 || index >= len(existing.StockStatuses) {
			return fmt.Errorf("index out of bounds: %d", index)
		}
		existing.StockStatuses = append(existing.StockStatuses[:index], existing.StockStatuses[index+1:]...)
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
		existing.Images = append(existing.Images[:index], existing.Images[index+1:]...)
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
