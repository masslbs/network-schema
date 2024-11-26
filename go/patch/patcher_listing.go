// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"
	"math/big"
	"slices"
	"time"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/masslbs/network-schema/go/objects"
)

func (p *Patcher) patchListing(patch Patch) error {
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("listing patch needs an ID")
	}
	objID := *patch.Path.ObjectID
	listing, exists := p.shop.Listings.Get(objID)

	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if exists {
				return fmt.Errorf("listing %d already exists", objID)
			}
			var newListing objects.Listing
			if err := masscbor.Unmarshal(patch.Value, &newListing); err != nil {
				return fmt.Errorf("failed to unmarshal listing: %w", err)
			}
			if err := p.validator.Struct(newListing); err != nil {
				return err
			}
			return p.shop.Listings.Insert(objID, newListing)
		}

		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}

		if err := p.addListingField(&listing, patch); err != nil {
			return err
		}
		return p.shop.Listings.Insert(objID, listing)

	case AppendOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		if err := p.appendListingField(&listing, patch); err != nil {
			return err
		}
		return p.shop.Listings.Insert(objID, listing)

	case ReplaceOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		if len(patch.Path.Fields) == 0 {
			var newListing objects.Listing
			if err := masscbor.Unmarshal(patch.Value, &newListing); err != nil {
				return fmt.Errorf("failed to unmarshal listing: %w", err)
			}
			if err := p.validator.Struct(newListing); err != nil {
				return err
			}
			return p.shop.Listings.Insert(objID, newListing)
		}
		if err := p.replaceListingField(&listing, patch); err != nil {
			return err
		}
		return p.shop.Listings.Insert(objID, listing)

	case RemoveOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		if len(patch.Path.Fields) == 0 {
			referenced := false
			tagNames := []string{}
			p.shop.Tags.All(func(key []byte, tag objects.Tag) bool {
				if slices.Contains(tag.ListingIDs, objID) {
					referenced = true
					tagNames = append(tagNames, string(key))
				}
				return true
			})
			if referenced {
				return fmt.Errorf("listing %d is referenced by tags: %v", objID, tagNames)
			}
			return p.shop.Listings.Delete(objID)
		}
		if err := p.removeListingField(&listing, patch); err != nil {
			return err
		}
		return p.shop.Listings.Insert(objID, listing)
	}

	return fmt.Errorf("unsupported operation: %s", patch.Op)
}

func (p *Patcher) addListingField(listing *objects.Listing, patch Patch) error {
	switch patch.Path.Fields[0] {
	case "Metadata":
		return p.addListingMetadata(listing, patch)
	case "StockStatuses":
		return p.addListingStockStatus(listing, patch)
	case "Options":
		return p.addListingOption(listing, patch)
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
}

func (p *Patcher) appendListingField(listing *objects.Listing, patch Patch) error {
	switch patch.Path.Fields[0] {
	case "Metadata":
		if len(patch.Path.Fields) < 2 {
			return fmt.Errorf("invalid metadata path")
		}
		switch patch.Path.Fields[1] {
		case "Images":
			var image string
			if err := masscbor.Unmarshal(patch.Value, &image); err != nil {
				return fmt.Errorf("failed to unmarshal image: %w", err)
			}
			listing.Metadata.Images = append(listing.Metadata.Images, image)
			return nil
		default:
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
	case "StockStatuses":
		var stockStatus objects.ListingStockStatus
		if err := masscbor.Unmarshal(patch.Value, &stockStatus); err != nil {
			return fmt.Errorf("failed to unmarshal stock status: %w", err)
		}
		listing.StockStatuses = append(listing.StockStatuses, stockStatus)
		return nil
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
}

func (p *Patcher) removeListingField(listing *objects.Listing, patch Patch) error {
	switch patch.Path.Fields[0] {
	case "Metadata":
		return p.removeListingMetadata(listing, patch)
	case "StockStatuses":
		return p.removeListingStockStatus(listing, patch)
	case "Options":
		return p.removeListingOption(listing, patch)
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
}

func (p *Patcher) replaceListingField(listing *objects.Listing, patch Patch) error {
	switch patch.Path.Fields[0] {
	case "Metadata":
		return p.replaceListingMetadata(listing, patch)
	case "StockStatuses":
		return p.replaceListingStockStatuses(listing, patch)
	case "Options":
		return p.replaceListingOptions(listing, patch)
	case "Price":
		var newPrice big.Int
		if err := masscbor.Unmarshal(patch.Value, &newPrice); err != nil {
			return fmt.Errorf("failed to unmarshal Price: %w", err)
		}
		listing.Price = newPrice
	case "ViewState":
		var v objects.ListingViewState
		if err := masscbor.Unmarshal(patch.Value, &v); err != nil {
			return fmt.Errorf("failed to unmarshal ViewState: %w", err)
		}
		listing.ViewState = v
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
	return nil
}

func (p *Patcher) addListingMetadata(listing *objects.Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid metadata path")
	}
	switch patch.Path.Fields[1] {
	case "Images":
		if len(patch.Path.Fields) < 3 {
			return fmt.Errorf("invalid Images path")
		}
		var newImage string
		if err := masscbor.Unmarshal(patch.Value, &newImage); err != nil {
			return fmt.Errorf("failed to unmarshal image: %w", err)
		}
		index, err := indexFromAny(patch.Path.Fields[2], len(listing.Metadata.Images))
		if err != nil {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		listing.Metadata.Images = slices.Insert(listing.Metadata.Images, index, newImage)
	default:
		return fmt.Errorf("invalid metadata path")
	}
	return nil
}

func (p *Patcher) removeListingMetadata(listing *objects.Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid metadata path")
	}
	switch patch.Path.Fields[1] {
	case "Images":
		if len(patch.Path.Fields) != 3 {
			return fmt.Errorf("invalid Images path")
		}
		index, err := indexFromAny(patch.Path.Fields[2], len(listing.Metadata.Images))
		if err != nil {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		listing.Metadata.Images = slices.Delete(listing.Metadata.Images, index, index+1)
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
	return nil
}

func (p *Patcher) replaceListingMetadata(listing *objects.Listing, patch Patch) error {
	if len(patch.Path.Fields) == 1 {
		var newMd objects.ListingMetadata
		if err := masscbor.Unmarshal(patch.Value, &newMd); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		listing.Metadata = newMd
		return nil
	}
	switch patch.Path.Fields[1] {
	case "Title":
		var val string
		if err := masscbor.Unmarshal(patch.Value, &val); err != nil {
			return fmt.Errorf("failed to unmarshal Title: %w", err)
		}
		listing.Metadata.Title = val
	case "Description":
		var val string
		if err := masscbor.Unmarshal(patch.Value, &val); err != nil {
			return fmt.Errorf("failed to unmarshal Description: %w", err)
		}
		listing.Metadata.Description = val
	case "Images":
		if len(patch.Path.Fields) == 2 {
			var Images []string
			if err := masscbor.Unmarshal(patch.Value, &Images); err != nil {
				return fmt.Errorf("failed to unmarshal Images: %w", err)
			}
			listing.Metadata.Images = Images
			return nil
		}
		if len(patch.Path.Fields) == 3 {
			i, err := indexFromAny(patch.Path.Fields[2], len(listing.Metadata.Images))
			if err != nil {
				return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
			}
			var val string
			if err := masscbor.Unmarshal(patch.Value, &val); err != nil {
				return fmt.Errorf("failed to unmarshal image: %w", err)
			}
			listing.Metadata.Images[i] = val
			return nil
		}
		return fmt.Errorf("invalid Images path")
	default:
		return fmt.Errorf("unsupported metadata field: %s", patch.Path.Fields[1])
	}
	return nil
}

func (p *Patcher) addListingStockStatus(listing *objects.Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid stockStatuses path")
	}

	index, err := indexFromAny(patch.Path.Fields[1], len(listing.StockStatuses))
	if err != nil {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}

	var newSS objects.ListingStockStatus
	if err := masscbor.Unmarshal(patch.Value, &newSS); err != nil {
		return fmt.Errorf("failed to unmarshal stock status: %w", err)
	}

	listing.StockStatuses = slices.Insert(listing.StockStatuses, index, newSS)

	return nil
}

func (p *Patcher) removeListingStockStatus(listing *objects.Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid stockStatuses path")
	}
	i, err := indexFromAny(patch.Path.Fields[1], len(listing.StockStatuses))
	if err != nil {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
	listing.StockStatuses = slices.Delete(listing.StockStatuses, i, i+1)
	return nil
}

func (p *Patcher) replaceListingStockStatuses(listing *objects.Listing, patch Patch) error {
	if len(patch.Path.Fields) == 1 {
		var statuses []objects.ListingStockStatus
		if err := masscbor.Unmarshal(patch.Value, &statuses); err != nil {
			return fmt.Errorf("failed to unmarshal stock statuses: %w", err)
		}
		listing.StockStatuses = statuses
		return nil
	}
	i, err := indexFromAny(patch.Path.Fields[1], len(listing.StockStatuses))
	if err != nil {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
	if len(patch.Path.Fields) == 2 {
		var ss objects.ListingStockStatus
		if err := masscbor.Unmarshal(patch.Value, &ss); err != nil {
			return fmt.Errorf("failed to unmarshal stock status: %w", err)
		}
		listing.StockStatuses[i] = ss
		return nil
	}
	switch patch.Path.Fields[2] {
	case "InStock":
		var val bool
		if err := masscbor.Unmarshal(patch.Value, &val); err != nil {
			return fmt.Errorf("failed to unmarshal inStock: %w", err)
		}
		listing.StockStatuses[i].InStock = &val
		listing.StockStatuses[i].ExpectedInStockBy = nil
	case "ExpectedInStockBy":
		var t time.Time
		if err := masscbor.Unmarshal(patch.Value, &t); err != nil {
			return fmt.Errorf("failed to unmarshal expectedInStockBy: %w", err)
		}
		listing.StockStatuses[i].ExpectedInStockBy = &t
		listing.StockStatuses[i].InStock = nil
	default:
		return fmt.Errorf("unsupported stockStatus field: %s", patch.Path.Fields[2])
	}

	return nil
}

func (p *Patcher) addListingOption(listing *objects.Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid Options path")
	}
	optionName, ok := patch.Path.Fields[1].(string)
	if !ok {
		return fmt.Errorf("invalid option name: %T", patch.Path.Fields[1])
	}
	if len(patch.Path.Fields) == 2 {
		if _, exists := listing.Options[optionName]; exists {
			return fmt.Errorf("option '%s' already exists", optionName)
		}
		var opt objects.ListingOption
		if err := masscbor.Unmarshal(patch.Value, &opt); err != nil {
			return fmt.Errorf("failed to unmarshal option: %w", err)
		}
		// Check if any variation name in the new option is already used in existing Options
		for newVarName := range opt.Variations {
			for existingOptName, existingOpt := range listing.Options {
				if _, ok := existingOpt.Variations[newVarName]; ok {
					return fmt.Errorf("variation name '%q' already exists under option '%q'", newVarName, existingOptName)
				}
			}
		}
		if listing.Options == nil {
			listing.Options = make(map[string]objects.ListingOption)
		}
		listing.Options[optionName] = opt
		return nil
	} else if len(patch.Path.Fields) == 4 && patch.Path.Fields[2] == "Variations" {
		varName, ok := patch.Path.Fields[3].(string)
		if !ok {
			return fmt.Errorf("invalid variation name: %T", patch.Path.Fields[3])
		}
		opt, exists := listing.Options[optionName]
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		if opt.Variations == nil {
			opt.Variations = make(map[string]objects.ListingVariation)
		}
		// Check if variation ID exists under any option
		for otherOptName, existingOpt := range listing.Options {
			if _, ok := existingOpt.Variations[varName]; ok {
				return fmt.Errorf("variation name '%q' already exists under option '%q'", varName, otherOptName)
			}
		}
		var v objects.ListingVariation
		if err := masscbor.Unmarshal(patch.Value, &v); err != nil {
			return fmt.Errorf("failed to unmarshal variation: %w", err)
		}
		opt.Variations[varName] = v
		listing.Options[optionName] = opt
		return nil
	}
	return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
}

func (p *Patcher) removeListingOption(listing *objects.Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid Options path")
	}
	optionName, ok := patch.Path.Fields[1].(string)
	if !ok {
		return fmt.Errorf("invalid option name: %T", patch.Path.Fields[1])
	}
	opt, exists := listing.Options[optionName]
	if !exists {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
	if len(patch.Path.Fields) == 2 {
		delete(listing.Options, optionName)
		return nil
	}
	if len(patch.Path.Fields) == 4 && patch.Path.Fields[2] == "Variations" {
		varID, ok := patch.Path.Fields[3].(string)
		if !ok {
			return fmt.Errorf("invalid variation name: %T", patch.Path.Fields[3])
		}
		if _, ok := opt.Variations[varID]; !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		delete(opt.Variations, varID)
		listing.Options[optionName] = opt
		return nil
	}
	return fmt.Errorf("invalid variation path")
}

func (p *Patcher) replaceListingOptions(listing *objects.Listing, patch Patch) error {
	if len(patch.Path.Fields) == 1 {
		var newOptions objects.ListingOptions
		if err := masscbor.Unmarshal(patch.Value, &newOptions); err != nil {
			return fmt.Errorf("failed to unmarshal Options: %w", err)
		}
		listing.Options = newOptions
		return nil
	}
	optionName, ok := patch.Path.Fields[1].(string)
	if !ok {
		return fmt.Errorf("invalid option name: %T", patch.Path.Fields[1])
	}
	if len(patch.Path.Fields) == 2 {
		var newOpt objects.ListingOption
		if err := masscbor.Unmarshal(patch.Value, &newOpt); err != nil {
			return fmt.Errorf("failed to unmarshal listing option: %w", err)
		}
		listing.Options[optionName] = newOpt
		return nil
	}
	if len(patch.Path.Fields) == 3 && patch.Path.Fields[2] == "Title" {
		var val string
		if err := masscbor.Unmarshal(patch.Value, &val); err != nil {
			return fmt.Errorf("failed to unmarshal option Title: %w", err)
		}
		opt := listing.Options[optionName]
		opt.Title = val
		listing.Options[optionName] = opt
		return nil
	}
	if len(patch.Path.Fields) == 4 && patch.Path.Fields[2] == "Variations" {
		varID, ok := patch.Path.Fields[3].(string)
		if !ok {
			return fmt.Errorf("invalid variation name: %T", patch.Path.Fields[3])
		}
		opt, ok := listing.Options[optionName]
		if !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		var newVar objects.ListingVariation
		if err := masscbor.Unmarshal(patch.Value, &newVar); err != nil {
			return fmt.Errorf("failed to unmarshal listing variation: %w", err)
		}
		if _, has := opt.Variations[varID]; !has {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		opt.Variations[varID] = newVar
		listing.Options[optionName] = opt
		return nil
	}
	if len(patch.Path.Fields) == 5 &&
		patch.Path.Fields[2] == "Variations" &&
		patch.Path.Fields[4] == "VariationInfo" {
		varID, ok := patch.Path.Fields[3].(string)
		if !ok {
			return fmt.Errorf("invalid variation name: %T", patch.Path.Fields[3])
		}
		opt, ok := listing.Options[optionName]
		if !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		v, ok := opt.Variations[varID]
		if !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		var newInfo objects.ListingMetadata
		if err := masscbor.Unmarshal(patch.Value, &newInfo); err != nil {
			return fmt.Errorf("failed to unmarshal listing variation info: %w", err)
		}
		v.VariationInfo = newInfo
		opt.Variations[varID] = v
		listing.Options[optionName] = opt
		return nil
	}
	return fmt.Errorf("invalid variation path")
}
