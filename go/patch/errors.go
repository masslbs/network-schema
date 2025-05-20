// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"fmt"

	"github.com/masslbs/network-schema/go/objects"
)

// ObjectNotFoundError is an error that occurs when an object is not found
type ObjectNotFoundError struct {
	ObjectType ObjectType
	Path       Path
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

// OutOfStockError is an error that occurs when an inventory is out of stock
type OutOfStockError struct {
	ListingID  objects.ObjectID
	Variations []string
}

// Error returns a string representation of the error
func (e OutOfStockError) Error() string {
	if len(e.Variations) > 0 {
		return fmt.Sprintf("inventory %d with variations %v out of stock", e.ListingID, e.Variations)
	}
	return fmt.Sprintf("inventory %d out of stock", e.ListingID)
}

// IndexOutOfBoundsError is an error that occurs when an index is out of bounds
type IndexOutOfBoundsError struct {
	Index    int
	MaxIndex int
}

func (e IndexOutOfBoundsError) Error() string {
	return fmt.Sprintf("index %d out of bounds for array of length %d", e.Index, e.MaxIndex)
}
