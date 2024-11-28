package schema

import (
	"fmt"
	"strconv"

	"github.com/fxamacker/cbor/v2"
)

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
	case "StockStatuses":
		if len(fields) == 1 {
			return fmt.Errorf("StockStatuses requires at least one field")
		}
		index, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("failed to convert index to int: %w", err)
		}
		if index < 0 || index >= len(existing.StockStatuses) {
			return fmt.Errorf("index out of bounds: %d", index)
		}
		var newStatus ListingStockStatus
		err = Unmarshal(value, &newStatus)
		if err != nil {
			return fmt.Errorf("failed to unmarshal stock status: %w", err)
		}
		existing.StockStatuses[index] = newStatus
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
	if fields[1] == "variations" {
		err := option.Variations.PatchReplace(fields[2:], value)
		if err != nil {
			return fmt.Errorf("failed to replace option variation: %w", err)
		}
		return nil
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
