// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

/*
https://datatracker.ietf.org/doc/html/rfc6902/
This is a modified version of JSON Patch (rfc6902). We first constraint the operations to "add", "replace", "remove" and then we add the following operations
- increment - Increments a number by a specified value. Only valid if the value of the path is a number.,
- decrement - Decrements a number by a specified value. Only valid if the value of the path is a number.,
*/

package schema

import (
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-playground/validator/v10"
)

type Patch struct {
	Op    OpString        `validate:"oneof=add replace remove increment decrement"`
	Path  PatchPath       `validate:"required"`
	Value cbor.RawMessage `validate:"required,gt=0"`
}

// TODO: type change to number instead of string?
type OpString string

const (
	AddOp       OpString = "add"
	ReplaceOp   OpString = "replace"
	RemoveOp    OpString = "remove"
	IncrementOp OpString = "increment"
	DecrementOp OpString = "decrement"
)

// PatchPath encodes as an opaque array of [type, id, fields...]
// This utility helps getting type and id in the expected types for Go.
type PatchPath struct {
	Type ObjectType `validate:"required,notblank"`

	// one-of, element 1 of the array
	//
	// manifest: nil
	// account: EthereumAddress
	// listing: ObjectId
	// order: ObjectId
	// tag: TagName
	ObjectID  *ObjectId
	AccountID *EthereumAddress
	TagName   *string

	Fields []string // extra fields
}

func (p PatchPath) MarshalCBOR() ([]byte, error) {
	var path = make([]any, len(p.Fields)+2)
	path[0] = p.Type
	if p.ObjectID != nil {
		path[1] = *p.ObjectID
	} else if p.AccountID != nil {
		path[1] = *p.AccountID
	} else if p.TagName != nil {
		path[1] = *p.TagName
	} else {
		path[1] = nil
	}
	for i, field := range p.Fields {
		path[i+2] = field
	}
	return cbor.Marshal(path)
}

func (p *PatchPath) UnmarshalCBOR(data []byte) error {
	var path []any
	err := cbor.Unmarshal(data, &path)
	if err != nil {
		return err
	}
	if len(path) < 2 {
		return fmt.Errorf("invalid patch path: %v - need at least type and id", path)
	}
	p.Type = ObjectType(path[0].(string))
	if !p.Type.IsValid() {
		return fmt.Errorf("invalid object type: %s", path[0].(string))
	}
	// path[0] can be a uint64, an ethereum address or nil (for manifest)
	switch tv := path[1].(type) {
	case uint64:
		p.ObjectID = &tv
	case []byte:
		var addr EthereumAddress
		n := copy(addr[:], tv)
		if n != EthereumAddressSize {
			return fmt.Errorf("invalid ethereum address: %d != %d", n, EthereumAddressSize)
		}
		p.AccountID = &addr
	case string:
		p.TagName = &tv
	case nil:
		// needs to be manifest type
		if p.Type != ObjectTypeManifest {
			return fmt.Errorf("invalid path type: %s", p.Type)
		}
	default:
		return fmt.Errorf("invalid id type: %T", tv)
	}

	// validate contextual type<>id values
	switch p.Type {
	case ObjectTypeManifest:
		if p.ObjectID != nil {
			return fmt.Errorf("manifest patch should not have an id")
		}
		if p.AccountID != nil {
			return fmt.Errorf("manifest patch should not have an account id")
		}
		if p.TagName != nil {
			return fmt.Errorf("manifest patch should not have a tag name")
		}
	case ObjectTypeAccount:
		if p.AccountID == nil {
			return fmt.Errorf("account patch needs an id")
		}
		if p.ObjectID != nil {
			return fmt.Errorf("account patch should not have an object id")
		}
		if p.TagName != nil {
			return fmt.Errorf("account patch should not have a tag name")
		}
	case ObjectTypeListing, ObjectTypeOrder:
		if p.ObjectID == nil {
			return fmt.Errorf("listing/order patch needs an id")
		}
		if p.AccountID != nil {
			return fmt.Errorf("listing/order patch should not have an account id")
		}
		if p.TagName != nil {
			return fmt.Errorf("listing/order patch should not have a tag name")
		}
	case ObjectTypeTag:
		if p.TagName == nil {
			return fmt.Errorf("tag patch needs a tag name")
		}
		if p.ObjectID != nil {
			return fmt.Errorf("tag patch should not have an object id")
		}
		if p.AccountID != nil {
			return fmt.Errorf("tag patch should not have an account id")
		}
	default:
		return fmt.Errorf("unsupported path type: %s", p.Type)
	}

	// the rest of the path is the fields
	p.Fields = make([]string, len(path)-2)
	for i, field := range path[2:] {
		p.Fields[i] = field.(string)
	}
	return nil
}

// TODO: could change to number instead of string..?
type ObjectType string

const (
	ObjectTypeManifest ObjectType = "manifest"
	ObjectTypeAccount  ObjectType = "account"
	ObjectTypeListing  ObjectType = "listing"
	ObjectTypeOrder    ObjectType = "order"
	ObjectTypeTag      ObjectType = "tag"
)

func (obj *ObjectType) UnmarshalCBOR(data []byte) error {
	var s string
	err := cbor.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	var newVal = ObjectType(s)
	if !newVal.IsValid() {
		return fmt.Errorf("invalid object type: %s", s)
	}
	*obj = newVal
	return nil
}

func (obj ObjectType) IsValid() bool {
	return obj == ObjectTypeManifest || obj == ObjectTypeAccount || obj == ObjectTypeListing || obj == ObjectTypeOrder || obj == ObjectTypeTag
}

type Patcher struct {
	validator *validator.Validate
}

// Patch applies a patch to a Shop, routing to the appropriate sub-patcher based on path.Type
func (p *Patcher) Shop(in *Shop, patch Patch) error {
	var err error
	switch patch.Path.Type {
	case ObjectTypeManifest:
		err = p.Manifest(&in.Manifest, patch)
	case ObjectTypeAccount:
		err = p.Accounts(in.Accounts, patch)
	case ObjectTypeListing:
		err = p.Listings(in.Listings, patch)
	case ObjectTypeOrder:
		err = p.Orders(in.Orders, patch)
	case ObjectTypeTag:
		err = p.Tags(in.Tags, patch)
	default:
		return fmt.Errorf("unsupported path type: %s", patch.Path.Type)
	}
	return err
}

// Manifest patches a Manifest object using the new PatchField approach
func (p *Patcher) Manifest(in *Manifest, patch Patch) error {
	if patch.Path.Type != ObjectTypeManifest {
		return fmt.Errorf("invalid path type: %s", patch.Path.Type)
	}

	// For empty fields, handle entire manifest replacement
	if len(patch.Path.Fields) == 0 {
		switch patch.Op {
		case ReplaceOp:
			var newManifest Manifest
			if err := Unmarshal(patch.Value, &newManifest); err != nil {
				return fmt.Errorf("failed to unmarshal manifest: %w", err)
			}
			*in = newManifest
		default:
			return fmt.Errorf("unsupported op for entire manifest: %s", patch.Op)
		}
	} else {
		fmt.Printf("patching manifest: %s %v %x\n", patch.Op, patch.Path.Fields, patch.Value)
		// Use PatchField for field-specific operations
		if err := PatchField(in, patch.Op, patch.Path.Fields, patch.Value); err != nil {
			return err
		}
	}
	fmt.Println("validating manifest")
	return p.validator.Struct(in)
}

// Listings patches a Listings collection or individual Listing
func (p *Patcher) Listings(listings Listings, patch Patch) error {
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("listing patch needs an ID")
	}
	objID := *patch.Path.ObjectID
	lis, ok := listings.Get(objID)

	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if ok {
				return fmt.Errorf("listing %d already exists", objID)
			}
			if err := Unmarshal(patch.Value, &lis); err != nil {
				return fmt.Errorf("failed to unmarshal new listing: %w", err)
			}
		} else {
			if !ok {
				return fmt.Errorf("listing %d not found", objID)
			}
			if err := PatchField(&lis, patch.Op, patch.Path.Fields, patch.Value); err != nil {
				return fmt.Errorf("patching listing %d: %w", objID, err)
			}
		}
		if err := listings.Insert(objID, lis); err != nil {
			return fmt.Errorf("failed to insert listing %d: %w", objID, err)
		}

	case ReplaceOp:
		if !ok {
			return fmt.Errorf("listing %d not found", objID)
		}
		if len(patch.Path.Fields) == 0 {
			if err := Unmarshal(patch.Value, &lis); err != nil {
				return fmt.Errorf("failed to unmarshal replacement listing: %w", err)
			}
		} else {
			if err := PatchField(&lis, patch.Op, patch.Path.Fields, patch.Value); err != nil {
				return fmt.Errorf("patching listing %d: %w", objID, err)
			}
		}
		if err := listings.Insert(objID, lis); err != nil {
			return fmt.Errorf("failed to update listing %d: %w", objID, err)
		}

	case RemoveOp:
		if !ok {
			return fmt.Errorf("listing %d not found", objID)
		}
		if len(patch.Path.Fields) == 0 {
			if err := listings.Delete(objID); err != nil {
				return fmt.Errorf("failed to delete listing %d: %w", objID, err)
			}
		} else {
			if err := PatchField(&lis, patch.Op, patch.Path.Fields, patch.Value); err != nil {
				return fmt.Errorf("patching listing %d: %w", objID, err)
			}
			if err := listings.Insert(objID, lis); err != nil {
				return fmt.Errorf("failed to update listing %d: %w", objID, err)
			}
		}

	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}

	return p.validator.Struct(&lis)
}

// Orders patches an Orders collection or individual Order
func (p *Patcher) Orders(orders Orders, patch Patch) error {
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("order patch needs an ID")
	}
	objID := *patch.Path.ObjectID
	ord, ok := orders.Get(objID)

	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if ok {
				return fmt.Errorf("order %d already exists", objID)
			}
			if err := Unmarshal(patch.Value, &ord); err != nil {
				return fmt.Errorf("failed to unmarshal new order: %w", err)
			}
		} else {
			if !ok {
				return fmt.Errorf("order %d not found", objID)
			}
			if err := PatchField(&ord, patch.Op, patch.Path.Fields, patch.Value); err != nil {
				return fmt.Errorf("patching order %d: %w", objID, err)
			}
		}
		if err := orders.Insert(objID, ord); err != nil {
			return fmt.Errorf("failed to insert order %d: %w", objID, err)
		}

	case ReplaceOp:
		if !ok {
			return fmt.Errorf("order %d not found", objID)
		}
		if len(patch.Path.Fields) == 0 {
			if err := Unmarshal(patch.Value, &ord); err != nil {
				return fmt.Errorf("failed to unmarshal replacement order: %w", err)
			}
		} else {
			if err := PatchField(&ord, patch.Op, patch.Path.Fields, patch.Value); err != nil {
				return fmt.Errorf("patching order %d: %w", objID, err)
			}
		}
		if err := orders.Insert(objID, ord); err != nil {
			return fmt.Errorf("failed to update order %d: %w", objID, err)
		}

	case RemoveOp:
		if !ok {
			return fmt.Errorf("order %d not found", objID)
		}
		if len(patch.Path.Fields) == 0 {
			if err := orders.Delete(objID); err != nil {
				return fmt.Errorf("failed to delete order %d: %w", objID, err)
			}
		} else {
			if err := PatchField(&ord, patch.Op, patch.Path.Fields, patch.Value); err != nil {
				return fmt.Errorf("patching order %d: %w", objID, err)
			}
			if err := orders.Insert(objID, ord); err != nil {
				return fmt.Errorf("failed to update order %d: %w", objID, err)
			}
		}

	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}

	return p.validator.Struct(&ord)
}

// Accounts patches an Accounts collection or individual Account
func (p *Patcher) Accounts(accounts Accounts, patch Patch) error {
	if patch.Path.AccountID == nil {
		return fmt.Errorf("account patch needs an ID")
	}
	accID := *patch.Path.AccountID
	acc, ok := accounts.Trie.Get(accID[:])

	switch patch.Op {
	case AddOp:
		if ok {
			return fmt.Errorf("account %s already exists", accID)
		}
		if len(patch.Path.Fields) > 0 {
			return fmt.Errorf("cannot add fields to non-existent account")
		}
		if err := Unmarshal(patch.Value, &acc); err != nil {
			return fmt.Errorf("failed to unmarshal account: %w", err)
		}
		if err := accounts.Trie.Insert(accID[:], acc); err != nil {
			return fmt.Errorf("failed to insert account %s: %w", accID, err)
		}

	case RemoveOp:
		if !ok {
			return fmt.Errorf("account %s not found", accID)
		}
		if len(patch.Path.Fields) == 0 {
			if err := accounts.Trie.Delete(accID[:]); err != nil {
				return fmt.Errorf("failed to delete account %s: %w", accID, err)
			}
		} else {
			if err := PatchField(&acc, patch.Op, patch.Path.Fields, patch.Value); err != nil {
				return fmt.Errorf("patching account %s: %w", accID, err)
			}
			if err := accounts.Trie.Insert(accID[:], acc); err != nil {
				return fmt.Errorf("failed to update account %s: %w", accID, err)
			}
		}

	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}

	return p.validator.Struct(&acc)
}

// Tags patches a Tags collection or individual Tag
func (p *Patcher) Tags(tags Tags, patch Patch) error {
	if patch.Path.TagName == nil {
		return fmt.Errorf("tag patch needs a name")
	}
	tagName := *patch.Path.TagName
	tag, ok := tags.Trie.Get([]byte(tagName))

	switch patch.Op {
	case AddOp:
		if ok {
			return fmt.Errorf("tag %s already exists", tagName)
		}
		if len(patch.Path.Fields) > 0 {
			return fmt.Errorf("cannot add fields to non-existent tag")
		}
		if err := Unmarshal(patch.Value, &tag); err != nil {
			return fmt.Errorf("failed to unmarshal tag: %w", err)
		}
		if err := tags.Trie.Insert([]byte(tagName), tag); err != nil {
			return fmt.Errorf("failed to insert tag %s: %w", tagName, err)
		}

	case RemoveOp:
		if !ok {
			return fmt.Errorf("tag %s not found", tagName)
		}
		if len(patch.Path.Fields) == 0 {
			if err := tags.Trie.Delete([]byte(tagName)); err != nil {
				return fmt.Errorf("failed to delete tag %s: %w", tagName, err)
			}
		} else {
			if err := PatchField(&tag, patch.Op, patch.Path.Fields, patch.Value); err != nil {
				return fmt.Errorf("patching tag %s: %w", tagName, err)
			}
			if err := tags.Trie.Insert([]byte(tagName), tag); err != nil {
				return fmt.Errorf("failed to update tag %s: %w", tagName, err)
			}
		}

	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}

	return p.validator.Struct(&tag)
}
