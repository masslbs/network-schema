package schema

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

func (existing *Listing) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchAdd requires at least one field")
	}
	switch fields[0] {
	case "options":
		return existing.Options.PatchAdd(fields[1:], value)
	case "StockStatuses":
		var stockStatus ListingStockStatus
		err := Unmarshal(value, &stockStatus)
		if err != nil {
			return fmt.Errorf("failed to unmarshal stock status: %w", err)
		}
		existing.StockStatuses = append(existing.StockStatuses, stockStatus)
		return nil
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
}

func (existing *ListingOptions) PatchAdd(fields []string, value cbor.RawMessage) error {
	switch n := len(fields); n {
	case 0:
		return fmt.Errorf("PatchAdd options requires at least one field")
	case 2: // replace the whole options
		var option ListingOption
		err := Unmarshal(value, &option)
		if err != nil {
			return fmt.Errorf("failed to unmarshal option: %w", err)
		}
		(*existing)[fields[1]] = option
		return nil
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
