// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/masslbs/network-schema/go/objects"
)

func (p *Patcher) patchInventory(inventory *objects.Inventory, patch Patch) error {
	// validate patch edits an existing listing
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("inventory patch needs an id")
	}
	objID := *patch.Path.ObjectID
	lis, ok := p.shop.Listings.Get(objID)
	if !ok {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}

	strFields, err := fieldsToStringArray(patch.Path.Fields)
	if err != nil {
		return fmt.Errorf("failed to convert fields to string array: %w", err)
	}

	// if it is a variation, check that they exist
	if n := len(patch.Path.Fields); n > 0 {
		var found = uint(n)
		for _, field := range strFields {
			for _, opt := range lis.Options {
				_, has := opt.Variations[field]
				if has {
					found--
					break // found
				}
			}
		}
		if found > 0 {
			return fmt.Errorf("some variation of object %d not found", objID)
		}
	}

	var newVal uint64
	if patch.Op != RemoveOp {
		err = cbor.Unmarshal(patch.Value, &newVal)
		if err != nil {
			return fmt.Errorf("failed to unmarshal inventory value: %w", err)
		}
	}

	current, ok := inventory.Get(objID, strFields)
	switch patch.Op {
	case AddOp:
		if ok {
			return fmt.Errorf("inventory %d already exists", objID)
		}
		err = inventory.Insert(objID, strFields, newVal)
	case RemoveOp:
		if !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeInventory, Path: patch.Path}
		}
		err = inventory.Delete(objID, strFields)
	case ReplaceOp:
		if !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeInventory, Path: patch.Path}
		}
		err = inventory.Insert(objID, strFields, newVal)
	case IncrementOp:
		current += newVal
		err = inventory.Insert(objID, strFields, current)
	case DecrementOp:
		if current < newVal {
			return OutOfStockError{ListingID: objID, Variations: strFields}
		}
		current -= newVal
		err = inventory.Insert(objID, strFields, current)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	if err != nil {
		return fmt.Errorf("failed to patch(%s) inventory %d: %w", patch.Op, objID, err)
	}
	return nil
}
