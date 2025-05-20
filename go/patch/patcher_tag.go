// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"
	"slices"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/masslbs/network-schema/go/objects"
)

// Example of centralized tag patching with referential checks
func (p *Patcher) patchTag(patch Patch) error {
	if patch.Path.TagName == nil {
		return fmt.Errorf("tag patch needs a tag name")
	}
	tagName := *patch.Path.TagName
	tag, exists := p.shop.Tags.Get(tagName)

	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if exists {
				return fmt.Errorf("tag %s already exists", tagName)
			}
			var newTag objects.Tag
			if err := masscbor.Unmarshal(patch.Value, &newTag); err != nil {
				return fmt.Errorf("failed to unmarshal tag: %w", err)
			}
			if err := p.validator.Struct(newTag); err != nil {
				return err
			}
			// Verify all referenced listings exist
			for _, listingID := range newTag.ListingIDs {
				if _, exists := p.shop.Listings.Get(listingID); !exists {
					return fmt.Errorf("listing %d referenced by tag does not exist", listingID)
				}
			}
			newTag.Name = tagName
			return p.shop.Tags.Insert(tagName, newTag)
		}

		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
		}

		// Handle adding listing IDs with referential checks
		if patch.Path.Fields[0] == "ListingIDs" {
			if err := p.addListingToTag(&tag, patch); err != nil {
				return err
			}
			return p.shop.Tags.Insert(tagName, tag)
		}

		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[0])

	case AppendOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
		}

		if len(patch.Path.Fields) != 1 || patch.Path.Fields[0] != "ListingIDs" {
			return fmt.Errorf("unsupported field for append operation: %v", patch.Path.Fields)
		}

		var listingID objects.ObjectID
		if err := masscbor.Unmarshal(patch.Value, &listingID); err != nil {
			return fmt.Errorf("failed to unmarshal listing ID: %w", err)
		}

		// Verify the referenced listing exists
		if _, exists := p.shop.Listings.Get(listingID); !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: Path{ObjectID: &listingID}}
		}

		// Check if the listing ID is already in the tag
		if slices.Contains(tag.ListingIDs, listingID) {
			return fmt.Errorf("listing %d is already in tag %s", listingID, tagName)
		}

		tag.ListingIDs = append(tag.ListingIDs, listingID)
		return p.shop.Tags.Insert(tagName, tag)

	case RemoveOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
		}

		if len(patch.Path.Fields) == 0 {
			return p.shop.Tags.Delete(tagName)
		}

		if len(patch.Path.Fields) == 2 && patch.Path.Fields[0] == "ListingIDs" {
			idx, err := indexFromAny(patch.Path.Fields[1], len(tag.ListingIDs))
			if err != nil {
				return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
			}
			tag.ListingIDs = slices.Delete(tag.ListingIDs, idx, idx+1)
			return p.shop.Tags.Insert(tagName, tag)
		}

		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[0])

	case ReplaceOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
		}

		if len(patch.Path.Fields) == 0 {
			var newTag objects.Tag
			if err := masscbor.Unmarshal(patch.Value, &newTag); err != nil {
				return fmt.Errorf("failed to unmarshal tag: %w", err)
			}
			if err := p.validator.Struct(newTag); err != nil {
				return err
			}
			// Verify all referenced listings exist
			for _, listingID := range newTag.ListingIDs {
				if _, exists := p.shop.Listings.Get(listingID); !exists {
					return fmt.Errorf("listing %d referenced by tag does not exist", listingID)
				}
			}
			newTag.Name = tagName
			return p.shop.Tags.Insert(tagName, newTag)
		}

		switch patch.Path.Fields[0] {
		case "Name":
			var newName string
			if err := masscbor.Unmarshal(patch.Value, &newName); err != nil {
				return fmt.Errorf("failed to unmarshal tag Name: %w", err)
			}
			tag.Name = newName
			return p.shop.Tags.Insert(tagName, tag)

		case "ListingIDs":
			if len(patch.Path.Fields) != 2 {
				return fmt.Errorf("invalid ListingIDs path")
			}
			idx, err := indexFromAny(patch.Path.Fields[1], len(tag.ListingIDs))
			if err != nil {
				return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
			}
			var listingID objects.ObjectID
			if err := masscbor.Unmarshal(patch.Value, &listingID); err != nil {
				return fmt.Errorf("failed to unmarshal listing ID: %w", err)
			}
			if _, exists := p.shop.Listings.Get(listingID); !exists {
				return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: Path{ObjectID: &listingID}}
			}
			tag.ListingIDs[idx] = listingID
			return p.shop.Tags.Insert(tagName, tag)
		}

		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[0])
	}

	return fmt.Errorf("unsupported operation: %s", patch.Op)
}

// Helper for adding listings to tags with referential checks
func (p *Patcher) addListingToTag(tag *objects.Tag, patch Patch) error {
	if len(patch.Path.Fields) != 2 {
		return fmt.Errorf("invalid ListingIDs path")
	}

	var listingID objects.ObjectID
	if err := masscbor.Unmarshal(patch.Value, &listingID); err != nil {
		return fmt.Errorf("failed to unmarshal listing ID: %w", err)
	}

	// Check if listing exists before adding reference
	if _, exists := p.shop.Listings.Get(listingID); !exists {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: Path{ObjectID: &listingID}}
	}

	idx, err := indexFromAny(patch.Path.Fields[1], len(tag.ListingIDs))
	if err != nil {
		return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
	}

	tag.ListingIDs = slices.Insert(tag.ListingIDs, idx, listingID)
	return nil
}
