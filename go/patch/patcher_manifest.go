// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"
	"strconv"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/masslbs/network-schema/go/objects"
)

func (p *Patcher) patchManifest(patch Patch) error {
	if patch.Path.Type != ObjectTypeManifest {
		return fmt.Errorf("invalid path type: %s", patch.Path.Type)
	}

	switch patch.Op {
	case ReplaceOp:
		return p.replaceManifestField(patch)
	case AddOp:
		return p.addManifestField(patch)
	case RemoveOp:
		return p.removeManifestField(patch)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
}

func (p *Patcher) replaceManifestField(patch Patch) error {
	if len(patch.Path.Fields) == 0 {
		// Replace entire manifest
		var newManifest objects.Manifest
		if err := masscbor.Unmarshal(patch.Value, &newManifest); err != nil {
			return fmt.Errorf("failed to unmarshal manifest: %w", err)
		}
		if err := p.validator.Struct(newManifest); err != nil {
			return err
		}
		p.shop.Manifest = newManifest
		return nil
	}

	switch patch.Path.Fields[0] {
	case "shopId":
		var value objects.Uint256
		if err := masscbor.Unmarshal(patch.Value, &value); err != nil {
			return fmt.Errorf("failed to unmarshal shopId: %w", err)
		}
		p.shop.Manifest.ShopID = value

	case "pricingCurrency":
		var c objects.ChainAddress
		if err := masscbor.Unmarshal(patch.Value, &c); err != nil {
			return fmt.Errorf("failed to unmarshal pricingCurrency: %w", err)
		}
		p.shop.Manifest.PricingCurrency = c

	case "payees":
		// Replace entire payees map or a single payee
		if len(patch.Path.Fields) == 1 {
			var payees objects.Payees
			if err := masscbor.Unmarshal(patch.Value, &payees); err != nil {
				return fmt.Errorf("failed to unmarshal payees: %w", err)
			}
			p.shop.Manifest.Payees = payees
		} else if len(patch.Path.Fields) == 2 {
			payeeName := patch.Path.Fields[1]
			var payee objects.Payee
			if err := masscbor.Unmarshal(patch.Value, &payee); err != nil {
				return fmt.Errorf("failed to unmarshal payee: %w", err)
			}
			if _, exists := p.shop.Manifest.Payees[payeeName]; !exists {
				return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
			}
			p.shop.Manifest.Payees[payeeName] = payee
		} else {
			return fmt.Errorf("invalid payees path")
		}

	case "shippingRegions":
		// Replace entire shippingRegions or just one region
		if len(patch.Path.Fields) == 1 {
			var regions objects.ShippingRegions
			if err := masscbor.Unmarshal(patch.Value, &regions); err != nil {
				return fmt.Errorf("failed to unmarshal shippingRegions: %w", err)
			}
			p.shop.Manifest.ShippingRegions = regions
		} else if len(patch.Path.Fields) == 2 {
			regionName := patch.Path.Fields[1]
			var region objects.ShippingRegion
			if err := masscbor.Unmarshal(patch.Value, &region); err != nil {
				return fmt.Errorf("failed to unmarshal shippingRegion: %w", err)
			}
			if _, exists := p.shop.Manifest.ShippingRegions[regionName]; !exists {
				return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
			}
			p.shop.Manifest.ShippingRegions[regionName] = region
		} else {
			return fmt.Errorf("invalid shippingRegions path")
		}

	case "acceptedCurrencies":
		// Replace entire acceptedCurrencies or a single index
		if len(patch.Path.Fields) == 1 {
			var currencies objects.ChainAddresses
			if err := masscbor.Unmarshal(patch.Value, &currencies); err != nil {
				return fmt.Errorf("failed to unmarshal acceptedCurrencies: %w", err)
			}
			p.shop.Manifest.AcceptedCurrencies = currencies
		} else if len(patch.Path.Fields) == 2 {
			i, err := strconv.Atoi(patch.Path.Fields[1])
			if err != nil {
				return fmt.Errorf("invalid acceptedCurrencies index: %w", err)
			}
			if i < 0 || i >= len(p.shop.Manifest.AcceptedCurrencies) {
				return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
			}
			var currency objects.ChainAddress
			if err := masscbor.Unmarshal(patch.Value, &currency); err != nil {
				return fmt.Errorf("failed to unmarshal accepted currency: %w", err)
			}
			p.shop.Manifest.AcceptedCurrencies[i] = currency
		} else {
			return fmt.Errorf("invalid acceptedCurrencies path")
		}

	default:
		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[0])
	}

	return p.validator.Struct(p.shop.Manifest)
}

func (p *Patcher) addManifestField(patch Patch) error {
	if len(patch.Path.Fields) == 0 {
		return fmt.Errorf("cannot add to root manifest")
	}

	switch patch.Path.Fields[0] {
	case "payees":
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid payees path")
		}
		payeeName := patch.Path.Fields[1]
		if _, exists := p.shop.Manifest.Payees[payeeName]; exists {
			return fmt.Errorf("payee %s already exists", payeeName)
		}
		var payee objects.Payee
		if err := masscbor.Unmarshal(patch.Value, &payee); err != nil {
			return fmt.Errorf("failed to unmarshal payee: %w", err)
		}
		p.shop.Manifest.Payees[payeeName] = payee
	case "shippingRegions":
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid shippingRegions path")
		}
		regionName := patch.Path.Fields[1]
		if _, exists := p.shop.Manifest.ShippingRegions[regionName]; exists {
			return fmt.Errorf("shipping region %s already exists", regionName)
		}
		var region objects.ShippingRegion
		if err := masscbor.Unmarshal(patch.Value, &region); err != nil {
			return fmt.Errorf("failed to unmarshal shipping region: %w", err)
		}
		p.shop.Manifest.ShippingRegions[regionName] = region
	case "acceptedCurrencies":
		// Handle array insert/append for acceptedCurrencies
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid acceptedCurrencies path")
		}
		index := patch.Path.Fields[1]

		var newCurrency objects.ChainAddress
		if err := masscbor.Unmarshal(patch.Value, &newCurrency); err != nil {
			return fmt.Errorf("failed to unmarshal new currency: %w", err)
		}
		// If index == "-", append to the end
		if index == "-" {
			p.shop.Manifest.AcceptedCurrencies = append(p.shop.Manifest.AcceptedCurrencies, newCurrency)
		} else {
			// Otherwise, parse insertion index
			i, err := strconv.Atoi(index)
			if err != nil {
				return fmt.Errorf("invalid acceptedCurrencies index: %w", err)
			}
			if i < 0 || i > len(p.shop.Manifest.AcceptedCurrencies) {
				return fmt.Errorf("index out of bounds: %d", i)
			}
			// Insert at position i
			ac := p.shop.Manifest.AcceptedCurrencies
			ac = append(ac[:i], append([]objects.ChainAddress{newCurrency}, ac[i:]...)...)
			p.shop.Manifest.AcceptedCurrencies = ac
		}
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
	}

	return p.validator.Struct(p.shop.Manifest)
}

func (p *Patcher) removeManifestField(patch Patch) error {
	if len(patch.Path.Fields) == 0 {
		return fmt.Errorf("cannot remove root manifest")
	}

	switch patch.Path.Fields[0] {
	case "payees":
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid payees path")
		}
		payeeName := patch.Path.Fields[1]
		if p.shop.Manifest.Payees == nil {
			return fmt.Errorf("payees map not initialized")
		}
		if _, exists := p.shop.Manifest.Payees[payeeName]; !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
		}
		delete(p.shop.Manifest.Payees, payeeName)

	case "shippingRegions":
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid shippingRegions path")
		}
		regionName := patch.Path.Fields[1]
		if p.shop.Manifest.ShippingRegions == nil {
			return fmt.Errorf("shippingRegions map not initialized")
		}
		if _, exists := p.shop.Manifest.ShippingRegions[regionName]; !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
		}
		delete(p.shop.Manifest.ShippingRegions, regionName)

	case "acceptedCurrencies":
		// Handle array removal from acceptedCurrencies
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid acceptedCurrencies path")
		}
		i, err := strconv.Atoi(patch.Path.Fields[1])
		if err != nil {
			return fmt.Errorf("invalid acceptedCurrencies index: %w", err)
		}
		if i < 0 || i >= len(p.shop.Manifest.AcceptedCurrencies) {
			return fmt.Errorf("index out of bounds: %d", i)
		}
		ac := p.shop.Manifest.AcceptedCurrencies
		p.shop.Manifest.AcceptedCurrencies = append(ac[:i], ac[i+1:]...)

	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
	}

	return p.validator.Struct(p.shop.Manifest)
}
