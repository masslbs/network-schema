// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/masslbs/network-schema/go/objects"
)

type ObjectNotFoundError struct {
	ObjectType ObjectType
	Path       PatchPath
}

func (e ObjectNotFoundError) Error() string {
	var id string
	if e.Path.ObjectID != nil {
		id = fmt.Sprintf("type=%s id=%d", e.ObjectType, *e.Path.ObjectID)
	} else if e.Path.TagName != nil {
		id = fmt.Sprintf("tag=%s", *e.Path.TagName)
	} else if e.Path.AccountAddr != nil {
		addr := e.Path.AccountAddr.Address
		id = fmt.Sprintf("account=%s", addr.Hex())
	} else {
		id = fmt.Sprintf("type=%s", e.ObjectType)
	}
	if len(e.Path.Fields) > 0 {
		return fmt.Sprintf("object %s with fields=%v not found", id, e.Path.Fields)
	}
	return fmt.Sprintf("object %s not found", id)
}

// Centralize patch operations in the Patcher type
type Patcher struct {
	validator *validator.Validate
	shop      *objects.Shop // Add reference to shop for lookups
}

func NewPatcher(v *validator.Validate, shop *objects.Shop) *Patcher {
	return &Patcher{validator: v, shop: shop}
}

// Main entry point for applying patches
func (p *Patcher) ApplyPatch(patch Patch) error {
	if !(patch.Path.Type == ObjectTypeSchemaVersion ||
		patch.Path.Type == ObjectTypeManifest) {
		// only accept writes if we have a valid manifest
		if err := p.validator.Struct(p.shop.Manifest); err != nil {
			return fmt.Errorf("invalid manifest: %w", err)
		}
	}

	switch patch.Path.Type {
	case ObjectTypeManifest:
		return p.patchManifest(patch)
	case ObjectTypeAccount:
		return p.patchAccount(patch)
	case ObjectTypeListing:
		return p.patchListing(patch)
	case ObjectTypeTag:
		return p.patchTag(patch)
	case ObjectTypeOrder:
		return p.patchOrder(patch)
	case ObjectTypeInventory:
		return p.patchInventory(&p.shop.Inventory, patch)
	default:
		return fmt.Errorf("unsupported shop object type: %s", patch.Path.Type)
	}
}
