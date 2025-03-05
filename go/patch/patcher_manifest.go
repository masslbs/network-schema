// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
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

var bigZero = big.NewInt(0)

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
		if p.shop.Manifest.ShopID.Cmp(bigZero) != 0 && p.shop.Manifest.ShopID.Cmp(&newManifest.ShopID) != 0 {
			return fmt.Errorf("shopId mismatch")
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
		// Replace entire payees map or a specific entry
		if len(patch.Path.Fields) == 1 {
			var payees objects.Payees
			if err := masscbor.Unmarshal(patch.Value, &payees); err != nil {
				return fmt.Errorf("failed to unmarshal payees: %w", err)
			}
			p.shop.Manifest.Payees = payees
		} else if len(patch.Path.Fields) == 3 {
			// Format: payees/chainId/hexEthAddr
			chainIDStr := patch.Path.Fields[1]
			hexEthAddr := patch.Path.Fields[2]

			chainID, err := strconv.ParseUint(chainIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid chain ID: %w", err)
			}

			ethAddr := common.HexToAddress(hexEthAddr)

			// Initialize the chain map if it doesn't exist
			if p.shop.Manifest.Payees == nil {
				p.shop.Manifest.Payees = make(objects.Payees)
			}
			if p.shop.Manifest.Payees[chainID] == nil {
				p.shop.Manifest.Payees[chainID] = make(map[common.Address]objects.PayeeMetadata)
			}

			// Check if the payee exists
			if _, exists := p.shop.Manifest.Payees[chainID][ethAddr]; !exists {
				return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
			}

			var metadata objects.PayeeMetadata
			if err := masscbor.Unmarshal(patch.Value, &metadata); err != nil {
				return fmt.Errorf("failed to unmarshal payee metadata: %w", err)
			}
			p.shop.Manifest.Payees[chainID][ethAddr] = metadata
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
		// Replace entire acceptedCurrencies or a specific entry
		if len(patch.Path.Fields) == 1 {
			var currencies objects.ChainAddresses
			if err := masscbor.Unmarshal(patch.Value, &currencies); err != nil {
				return fmt.Errorf("failed to unmarshal acceptedCurrencies: %w", err)
			}
			p.shop.Manifest.AcceptedCurrencies = currencies
		} else if len(patch.Path.Fields) == 3 {
			// Format: acceptedCurrencies/chainId/hexEthAddr
			chainIDStr := patch.Path.Fields[1]
			hexEthAddr := patch.Path.Fields[2]

			chainID, err := strconv.ParseUint(chainIDStr, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid chain ID: %w", err)
			}

			ethAddr := common.HexToAddress(hexEthAddr)

			// Initialize the chain map if it doesn't exist
			if p.shop.Manifest.AcceptedCurrencies == nil {
				p.shop.Manifest.AcceptedCurrencies = make(objects.ChainAddresses)
			}
			if p.shop.Manifest.AcceptedCurrencies[chainID] == nil {
				p.shop.Manifest.AcceptedCurrencies[chainID] = make(map[common.Address]struct{})
			}

			// Check if the currency exists
			if _, exists := p.shop.Manifest.AcceptedCurrencies[chainID][ethAddr]; !exists {
				return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
			}

			// For accepted currencies, we just need to ensure it exists (empty struct)
			p.shop.Manifest.AcceptedCurrencies[chainID][ethAddr] = struct{}{}
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
		if len(patch.Path.Fields) != 3 {
			return fmt.Errorf("invalid payees path")
		}

		// Format: payees/chainId/hexEthAddr
		chainIDStr := patch.Path.Fields[1]
		hexEthAddr := patch.Path.Fields[2]

		chainID, err := strconv.ParseUint(chainIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid chain ID: %w", err)
		}

		ethAddr := common.HexToAddress(hexEthAddr)

		// Initialize the chain map if it doesn't exist
		if p.shop.Manifest.Payees == nil {
			p.shop.Manifest.Payees = make(objects.Payees)
		}
		if p.shop.Manifest.Payees[chainID] == nil {
			p.shop.Manifest.Payees[chainID] = make(map[common.Address]objects.PayeeMetadata)
		}

		if _, exists := p.shop.Manifest.Payees[chainID][ethAddr]; exists {
			return fmt.Errorf("payee %s already exists for chain %d", hexEthAddr, chainID)
		}

		var metadata objects.PayeeMetadata
		if err := masscbor.Unmarshal(patch.Value, &metadata); err != nil {
			return fmt.Errorf("failed to unmarshal payee metadata: %w", err)
		}
		p.shop.Manifest.Payees[chainID][ethAddr] = metadata

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
		if len(patch.Path.Fields) != 3 {
			return fmt.Errorf("invalid acceptedCurrencies path")
		}

		// Format: acceptedCurrencies/chainId/hexEthAddr
		chainIDStr := patch.Path.Fields[1]
		hexEthAddr := patch.Path.Fields[2]

		chainID, err := strconv.ParseUint(chainIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid chain ID: %w", err)
		}

		ethAddr := common.HexToAddress(hexEthAddr)

		// Initialize the chain map if it doesn't exist
		if p.shop.Manifest.AcceptedCurrencies == nil {
			p.shop.Manifest.AcceptedCurrencies = make(objects.ChainAddresses)
		}
		if p.shop.Manifest.AcceptedCurrencies[chainID] == nil {
			p.shop.Manifest.AcceptedCurrencies[chainID] = make(map[common.Address]struct{})
		}

		// Check if the currency already exists
		if _, exists := p.shop.Manifest.AcceptedCurrencies[chainID][ethAddr]; exists {
			return fmt.Errorf("currency %s already exists for chain %d", hexEthAddr, chainID)
		}

		p.shop.Manifest.AcceptedCurrencies[chainID][ethAddr] = struct{}{}

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
		if len(patch.Path.Fields) != 3 {
			return fmt.Errorf("invalid payees path")
		}

		// Format: payees/chainId/hexEthAddr
		chainIDStr := patch.Path.Fields[1]
		hexEthAddr := patch.Path.Fields[2]

		chainID, err := strconv.ParseUint(chainIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid chain ID: %w", err)
		}

		ethAddr := common.HexToAddress(hexEthAddr)

		if p.shop.Manifest.Payees == nil || p.shop.Manifest.Payees[chainID] == nil {
			return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
		}

		if _, exists := p.shop.Manifest.Payees[chainID][ethAddr]; !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
		}

		delete(p.shop.Manifest.Payees[chainID], ethAddr)

		// If the chain map is now empty, remove it too
		if len(p.shop.Manifest.Payees[chainID]) == 0 {
			delete(p.shop.Manifest.Payees, chainID)
		}

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
		if len(patch.Path.Fields) != 3 {
			return fmt.Errorf("invalid acceptedCurrencies path")
		}

		// Format: acceptedCurrencies/chainId/hexEthAddr
		chainIDStr := patch.Path.Fields[1]
		hexEthAddr := patch.Path.Fields[2]

		chainID, err := strconv.ParseUint(chainIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid chain ID: %w", err)
		}

		ethAddr := common.HexToAddress(hexEthAddr)

		if p.shop.Manifest.AcceptedCurrencies == nil || p.shop.Manifest.AcceptedCurrencies[chainID] == nil {
			return fmt.Errorf("acceptedCurrencies map not initialized for chain %d", chainID)
		}

		if _, exists := p.shop.Manifest.AcceptedCurrencies[chainID][ethAddr]; !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
		}

		delete(p.shop.Manifest.AcceptedCurrencies[chainID], ethAddr)

		// If the chain map is now empty, remove it too
		if len(p.shop.Manifest.AcceptedCurrencies[chainID]) == 0 {
			delete(p.shop.Manifest.AcceptedCurrencies, chainID)
		}

	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
	}

	return p.validator.Struct(p.shop.Manifest)
}
