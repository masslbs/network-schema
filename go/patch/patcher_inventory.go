// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/masslbs/network-schema/go/objects"
)

type OutOfStockError struct {
	listingId  objects.ObjectId
	variations []string
}

func (e OutOfStockError) Error() string {
	if len(e.variations) > 0 {
		return fmt.Sprintf("inventory %d with variations %v out of stock", e.listingId, e.variations)
	}
	return fmt.Sprintf("inventory %d out of stock", e.listingId)
}

func (p *Patcher) patchInventory(inventory *objects.Inventory, patch Patch) error {
	// validate patch edits an existing listing
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("inventory patch needs an id")
	}
	objId := *patch.Path.ObjectID
	lis, ok := p.shop.Listings.Get(objId)
	if !ok {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}

	// if it is a variation, check that they exist
	if n := len(patch.Path.Fields); n > 0 {
		var found = uint(n)
		for _, field := range patch.Path.Fields {
			for _, opt := range lis.Options {
				_, has := opt.Variations[field]
				if has {
					found--
					break // found
				}
			}
		}
		if found > 0 {
			return fmt.Errorf("some variation of object %d not found", objId)
		}
	}

	var (
		err    error
		newVal uint64
	)
	if patch.Op != RemoveOp {
		err = cbor.Unmarshal(patch.Value, &newVal)
		if err != nil {
			return fmt.Errorf("failed to unmarshal inventory value: %w", err)
		}
	}

	current, ok := inventory.Get(objId, patch.Path.Fields)
	switch patch.Op {
	case AddOp:
		if ok {
			return fmt.Errorf("inventory %d already exists", objId)
		}
		err = inventory.Insert(objId, patch.Path.Fields, newVal)
	case RemoveOp:
		if !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeInventory, Path: patch.Path}
		}
		err = inventory.Delete(objId, patch.Path.Fields)
	case ReplaceOp:
		if !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeInventory, Path: patch.Path}
		}
		err = inventory.Insert(objId, patch.Path.Fields, newVal)
	case IncrementOp:
		current += newVal
		err = inventory.Insert(objId, patch.Path.Fields, current)
	case DecrementOp:
		if current < newVal {
			return OutOfStockError{listingId: objId, variations: patch.Path.Fields}
		}
		current -= newVal
		err = inventory.Insert(objId, patch.Path.Fields, current)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	if err != nil {
		return fmt.Errorf("failed to patch(%s) inventory %d: %w", patch.Op, objId, err)
	}
	return nil
}
