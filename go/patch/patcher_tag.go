// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"
	"slices"
	"strconv"

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
			for _, listingId := range newTag.ListingIds {
				if _, exists := p.shop.Listings.Get(listingId); !exists {
					return fmt.Errorf("listing %d referenced by tag does not exist", listingId)
				}
			}
			newTag.Name = tagName
			return p.shop.Tags.Insert(tagName, newTag)
		}

		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
		}

		// Handle adding listing IDs with referential checks
		if patch.Path.Fields[0] == "listingIds" {
			if err := p.addListingToTag(&tag, patch); err != nil {
				return err
			}
			return p.shop.Tags.Insert(tagName, tag)
		}

		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[0])

	case RemoveOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
		}

		if len(patch.Path.Fields) == 0 {
			return p.shop.Tags.Delete(tagName)
		}

		if len(patch.Path.Fields) == 2 && patch.Path.Fields[0] == "listingIds" {
			idx, err := strconv.Atoi(patch.Path.Fields[1])
			if err != nil {
				return fmt.Errorf("invalid listing index: %w", err)
			}
			if idx < 0 || idx >= len(tag.ListingIds) {
				return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
			}
			tag.ListingIds = slices.Delete(tag.ListingIds, idx, idx+1)
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
			for _, listingId := range newTag.ListingIds {
				if _, exists := p.shop.Listings.Get(listingId); !exists {
					return fmt.Errorf("listing %d referenced by tag does not exist", listingId)
				}
			}
			newTag.Name = tagName
			return p.shop.Tags.Insert(tagName, newTag)
		}

		switch patch.Path.Fields[0] {
		case "name":
			var newName string
			if err := masscbor.Unmarshal(patch.Value, &newName); err != nil {
				return fmt.Errorf("failed to unmarshal tag name: %w", err)
			}
			tag.Name = newName
			return p.shop.Tags.Insert(tagName, tag)

		case "listingIds":
			if len(patch.Path.Fields) != 2 {
				return fmt.Errorf("invalid listingIds path")
			}
			idx, err := strconv.Atoi(patch.Path.Fields[1])
			if err != nil {
				return fmt.Errorf("invalid listing index: %w", err)
			}
			if idx < 0 || idx >= len(tag.ListingIds) {
				return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
			}
			var listingId objects.ObjectId
			if err := masscbor.Unmarshal(patch.Value, &listingId); err != nil {
				return fmt.Errorf("failed to unmarshal listing ID: %w", err)
			}
			if _, exists := p.shop.Listings.Get(listingId); !exists {
				return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: PatchPath{ObjectID: &listingId}}
			}
			tag.ListingIds[idx] = listingId
			return p.shop.Tags.Insert(tagName, tag)
		}

		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[0])
	}

	return fmt.Errorf("unsupported operation: %s", patch.Op)
}

// Helper for adding listings to tags with referential checks
func (p *Patcher) addListingToTag(tag *objects.Tag, patch Patch) error {
	if len(patch.Path.Fields) != 2 {
		return fmt.Errorf("invalid listingIds path")
	}

	var listingId objects.ObjectId
	if err := masscbor.Unmarshal(patch.Value, &listingId); err != nil {
		return fmt.Errorf("failed to unmarshal listing ID: %w", err)
	}

	// Check if listing exists before adding reference
	if _, exists := p.shop.Listings.Get(listingId); !exists {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: PatchPath{ObjectID: &listingId}}
	}

	// Handle append vs insert at index
	if patch.Path.Fields[1] == "-" {
		// Append to end
		tag.ListingIds = append(tag.ListingIds, listingId)
		return nil
	}

	// Try to parse numerical index
	idx, err := strconv.Atoi(patch.Path.Fields[1])
	if err != nil {
		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[1])
	}

	// Validate index bounds
	if idx < 0 || idx > len(tag.ListingIds) {
		return fmt.Errorf("index out of bounds: %d", idx)
	}

	// Insert at index by growing slice and shifting elements
	tag.ListingIds = append(tag.ListingIds, 0)
	copy(tag.ListingIds[idx+1:], tag.ListingIds[idx:])
	tag.ListingIds[idx] = listingId
	return nil
}
