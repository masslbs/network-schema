// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"

	"github.com/go-playground/validator/v10"

	"github.com/masslbs/network-schema/go/objects"
)

// Patcher is a type that applies patches to a shop
type Patcher struct {
	validator *validator.Validate
	shop      *objects.Shop // Add reference to shop for lookups
}

// NewPatcher creates a new Patcher
func NewPatcher(v *validator.Validate, shop *objects.Shop) *Patcher {
	return &Patcher{validator: v, shop: shop}
}

// ApplyPatch applies a patch to the shop
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
