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
	"slices"
	"strconv"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-playground/validator/v10"
)

type Write struct {
	Patches []Patch
}

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
	if err != nil {
		return err
	}
	return p.validator.Struct(in)
}

func (p *Patcher) Manifest(in *Manifest, patch Patch) error {
	var err error
	if patch.Path.Type != ObjectTypeManifest {
		return fmt.Errorf("invalid path type: %s", patch.Path.Type)
	}
	switch patch.Op {
	case ReplaceOp:
		err = in.PatchReplace(patch.Path.Fields, patch.Value)
	case AddOp:
		err = in.PatchAdd(patch.Path.Fields, patch.Value)
	case RemoveOp:
		err = in.PatchRemove(patch.Path.Fields)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	if err != nil {
		return err
	}
	return p.validator.Struct(in)
}

func (p *Patcher) Accounts(in Accounts, patch Patch) error {
	var err error
	if patch.Path.Type != ObjectTypeAccount {
		return fmt.Errorf("invalid path type: %s", patch.Path.Type)
	}
	if patch.Path.AccountID == nil {
		return fmt.Errorf("account patch needs an ID")
	}
	accID := *patch.Path.AccountID
	// TODO: ipld dag-cbor might make these complicated...
	// mapKey := hex.EncodeToString(accID[:])
	mapKey := accID
	acc, ok := in[mapKey]
	switch patch.Op {
	case AddOp:
		if ok {
			return fmt.Errorf("account %s already exists", accID)
		}
		err = cbor.Unmarshal(patch.Value, &acc)
		if err != nil {
			return fmt.Errorf("failed to unmarshal account: %w", err)
		}
		in[mapKey] = acc
	case RemoveOp:
		if !ok {
			return fmt.Errorf("account %s not found", accID)
		}
		if len(patch.Path.Fields) == 0 {
			delete(in, mapKey)
		} else {
			if patch.Path.Fields[0] != "keyCards" {
				return fmt.Errorf("unsupported field: %s", patch.Path.Fields[0])
			}
			idx, err := strconv.Atoi(patch.Path.Fields[1])
			if err != nil {
				return fmt.Errorf("failed to convert index to int: %w", err)
			}
			if idx < 0 || idx >= len(acc.KeyCards) {
				return fmt.Errorf("index out of bounds: %d", idx)
			}
			acc.KeyCards = slices.Delete(acc.KeyCards, idx, idx+1)
		}
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	return nil
}

func (p *Patcher) Listings(listings Listings, patch Patch) error {
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("listing patch needs an ID")
	}
	objId := *patch.Path.ObjectID
	lis, ok := listings[objId]
	switch patch.Op {
	case AddOp:
		var err error
		if len(patch.Path.Fields) == 0 {
			if ok {
				return fmt.Errorf("listing %d already exists", objId)
			}
			err = Unmarshal(patch.Value, &lis)
		} else {
			if !ok {
				return fmt.Errorf("listing %d not found", objId)
			}
			err = p.Listing(&lis, patch)
		}
		if err != nil {
			return fmt.Errorf("failed to patch Listing %d: %w", objId, err)
		}
		listings[objId] = lis
	case ReplaceOp:
		if !ok {
			return fmt.Errorf("listing %d not found", objId)
		}
		var err error
		if len(patch.Path.Fields) == 0 {
			err = Unmarshal(patch.Value, &lis)
		} else {
			err = p.Listing(&lis, patch)
		}
		if err != nil {
			return fmt.Errorf("failed to patch Listing %d: %w", objId, err)
		}
		listings[objId] = lis
	case RemoveOp:
		if !ok {
			return fmt.Errorf("listing %d not found", objId)
		}
		delete(listings, objId)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	return nil
}

func (p *Patcher) Listing(in *Listing, patch Patch) error {
	var err error
	if patch.Path.Type != ObjectTypeListing {
		return fmt.Errorf("invalid path type: %s", patch.Path.Type)
	}
	switch patch.Op {
	case ReplaceOp:
		err = in.PatchReplace(patch.Path.Fields, patch.Value)
	case AddOp:
		err = in.PatchAdd(patch.Path.Fields, patch.Value)
	case RemoveOp:
		err = in.PatchRemove(patch.Path.Fields)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	if err != nil {
		return err
	}
	return p.validator.Struct(in)
}

func (p *Patcher) Tags(in Tags, patch Patch) error {
	if patch.Path.TagName == nil {
		return fmt.Errorf("tag patch needs a tag name")
	}

	tagName := *patch.Path.TagName
	tag, ok := in[tagName]
	switch patch.Op {
	case AddOp:
		var err error
		if len(patch.Path.Fields) == 0 {
			if ok {
				return fmt.Errorf("tag %s already exists", tagName)
			}
			err = Unmarshal(patch.Value, &tag)
		} else {
			if !ok {
				return fmt.Errorf("tag %s not found", tagName)
			}
			err = p.Tag(&tag, patch)
		}
		if err != nil {
			return fmt.Errorf("failed to patch Tag %s: %w", tagName, err)
		}
		in[tagName] = tag
	case RemoveOp:
		if !ok {
			return fmt.Errorf("tag %s not found", tagName)
		}
		delete(in, tagName)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	return nil
}

func (p *Patcher) Tag(in *Tag, patch Patch) error {
	var err error
	if patch.Path.Type != ObjectTypeTag {
		return fmt.Errorf("invalid path type: %s", patch.Path.Type)
	}
	if patch.Path.TagName == nil {
		return fmt.Errorf("tag patch needs a tag name")
	}
	switch patch.Op {
	case AddOp:
		err = in.PatchAdd(patch.Path.Fields, patch.Value)
	case ReplaceOp:
		err = in.PatchReplace(patch.Path.Fields, patch.Value)
	case RemoveOp:
		err = in.PatchRemove(patch.Path.Fields)
	}
	if err != nil {
		return fmt.Errorf("failed to patch Tag %s: %w", *patch.Path.TagName, err)
	}
	return nil
}

func (p *Patcher) Orders(in Orders, patch Patch) error {
	// needs an ID
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("order patch needs an ID")
	}
	objId := *patch.Path.ObjectID
	order, ok := in[objId]
	switch patch.Op {
	case AddOp:
		if ok {
			return fmt.Errorf("order %d already exists", objId)
		}
		err := p.Order(&order, patch)
		if err != nil {
			return fmt.Errorf("failed to patch Order %d: %w", objId, err)
		}
		in[objId] = order
	case RemoveOp:
		if !ok {
			return fmt.Errorf("order %d not found", objId)
		}
		delete(in, objId)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	return nil
}

func (p *Patcher) Order(in *Order, patch Patch) error {
	var err error
	if patch.Path.Type != ObjectTypeOrder {
		return fmt.Errorf("invalid path type: %s", patch.Path.Type)
	}
	switch patch.Op {
	case AddOp:
		err = in.PatchAdd(patch.Path.Fields, patch.Value)
	case ReplaceOp:
		err = in.PatchReplace(patch.Path.Fields, patch.Value)
	case RemoveOp:
		err = in.PatchRemove(patch.Path.Fields)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	if err != nil {
		return err
	}
	return p.validator.Struct(in)
}
