// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
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
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-playground/validator/v10"
)

// To validate a patchset, construct the merkle tree of the patches and validate the root hash.
type SignedPatchSet struct {
	// The header of the patch set
	Header PatchSetHeader `validate:"required"` // TODO: dive doesn't work?

	// The signature of the header, containing the merkle root of the patches
	Signature Signature `validate:"required,gt=0,dive"`

	Patches []Patch `validate:"required,gt=0,dive"`
}

type PatchSetHeader struct {
	// The nonce must be unique for each event a keycard creates.
	// The sequence values need to increase monotonicly.
	KeyCardNonce uint64 `validate:"required,gt=0"`

	// Every signed event must be tied to a shop id. This allow the
	// event to processed outside the context of the currenct connection.
	ShopID Uint256 `validate:"required"`

	// The time when this event was created.
	// The relay should reject any events from the future
	Timestamp time.Time `validate:"required"`

	// The merkle root of the patches
	RootHash Hash `validate:"required"`
}

type Patch struct {
	Op    OpString        `validate:"required,oneof=add replace remove increment decrement"`
	Path  PatchPath       `validate:"required"`
	Value cbor.RawMessage `validate:"required,gt=0"`
}

func (p Patch) Serialize() ([]byte, error) {
	return Marshal(p)
}

// TODO: type change to enum/number instead of string?
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
	// account: EthereumAddress
	// tag: TagName as string
	// listing: ObjectId
	// order: ObjectId
	// inventory: ObjectId with optional variations as fields
	//
	// exclusion: manifest, which has no id
	ObjectID    *ObjectId
	AccountAddr *EthereumAddress
	TagName     *string

	Fields []string // extra fields
}

func (p PatchPath) MarshalCBOR() ([]byte, error) {
	var extraFields = 2 // usually we have type and id
	if p.Type == ObjectTypeManifest {
		extraFields = 1 // manifest has no id
	}
	var path = make([]any, len(p.Fields)+extraFields)
	path[0] = string(p.Type)
	switch p.Type {
	case ObjectTypeManifest:
		if p.ObjectID != nil {
			return nil, fmt.Errorf("manifest patch should not have an id")
		}
		if p.AccountAddr != nil {
			return nil, fmt.Errorf("manifest patch should not have an account id")
		}
		if p.TagName != nil {
			return nil, fmt.Errorf("manifest patch should not have a tag name")
		}
	case ObjectTypeAccount:
		if p.AccountAddr == nil {
			return nil, fmt.Errorf("account patch needs an id")
		}
		path[1] = *p.AccountAddr
	case ObjectTypeListing, ObjectTypeOrder:
		if p.ObjectID == nil {
			return nil, fmt.Errorf("listing/order patch needs an id")
		}
		path[1] = *p.ObjectID
	case ObjectTypeTag:
		if p.TagName == nil {
			return nil, fmt.Errorf("tag patch needs a tag name")
		}
		path[1] = *p.TagName
	}
	for i, field := range p.Fields {
		path[i+extraFields] = field
	}
	return cbor.Marshal(path)
}

func (p *PatchPath) UnmarshalCBOR(data []byte) error {
	var path []any
	err := cbor.Unmarshal(data, &path)
	if err != nil {
		return err
	}
	objType, ok := path[0].(string)
	if !ok {
		return fmt.Errorf("invalid object type: %v", path[0])
	}
	p.Type = ObjectType(objType)
	if !p.Type.IsValid() {
		return fmt.Errorf("invalid object type: %s", objType)
	}
	path = path[1:] // slice of type
	// path[0] can be a uint64, an ethereum address or empty (for manifest)
	switch p.Type {
	case ObjectTypeManifest:
		// noop
	case ObjectTypeAccount:
		if len(path) < 1 {
			return fmt.Errorf("invalid ethereum address: %w", err)
		}
		data, ok := path[0].([]byte)
		if !ok {
			return fmt.Errorf("invalid ethereum address: %v", path[0])
		}
		if len(data) != EthereumAddressSize {
			return fmt.Errorf("invalid ethereum address size: %d != %d", len(data), EthereumAddressSize)
		}
		var addr EthereumAddress
		copy(addr[:], data)
		p.AccountAddr = &addr
	case ObjectTypeOrder, ObjectTypeListing:
		if len(path) < 1 {
			return fmt.Errorf("invalid object id: %w", err)
		}
		id, ok := path[0].(uint64)
		if !ok {
			return fmt.Errorf("invalid object id: %v", path[0])
		}
		objId := ObjectId(id)
		p.ObjectID = &objId
	case ObjectTypeTag:
		if len(path) < 1 {
			return fmt.Errorf("invalid tag name: %w", err)
		}
		tagName, ok := path[0].(string)
		if !ok {
			return fmt.Errorf("invalid tag name: %v", path[0])
		}
		p.TagName = &tagName
	default:
		return fmt.Errorf("invalid id type: %s", path[0])
	}

	if p.Type != ObjectTypeManifest {
		path = path[1:] // all other types have an id
	}

	// validate contextual type<>id values
	switch p.Type {
	case ObjectTypeManifest:
		if p.ObjectID != nil {
			return fmt.Errorf("manifest patch should not have an id")
		}
		if p.AccountAddr != nil {
			return fmt.Errorf("manifest patch should not have an account id")
		}
		if p.TagName != nil {
			return fmt.Errorf("manifest patch should not have a tag name")
		}
	case ObjectTypeAccount:
		if p.AccountAddr == nil {
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
		if p.AccountAddr != nil {
			return fmt.Errorf("listing/order patch should not have an account id")
		}
		if p.TagName != nil {
			return fmt.Errorf("listing/order patch should not have a tag name")
		}
	case ObjectTypeTag:
		if p.TagName == nil {
			return fmt.Errorf("tag patch needs a tag name")
		}
		if len(*p.TagName) == 0 {
			return fmt.Errorf("tag name cannot be empty")
		}
		if p.ObjectID != nil {
			return fmt.Errorf("tag patch should not have an object id")
		}
		if p.AccountAddr != nil {
			return fmt.Errorf("tag patch should not have an account id")
		}
	default:
		return fmt.Errorf("unsupported path type: %s", p.Type)
	}

	// the rest of the path is the fields
	p.Fields = make([]string, len(path))
	for i, field := range path {
		p.Fields[i], ok = field.(string)
		if !ok {
			return fmt.Errorf("invalid field: %v", field)
		}
	}
	return nil
}

// TODO: could change to number instead of string..?
type ObjectType string

const (
	ObjectTypeSchemaVersion ObjectType = "schemaVersion"
	ObjectTypeManifest      ObjectType = "manifest"
	ObjectTypeAccount       ObjectType = "account"
	ObjectTypeListing       ObjectType = "listing"
	ObjectTypeOrder         ObjectType = "order"
	ObjectTypeTag           ObjectType = "tag"
	ObjectTypeInventory     ObjectType = "inventory"
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
	return obj == ObjectTypeSchemaVersion || obj == ObjectTypeManifest || obj == ObjectTypeAccount || obj == ObjectTypeListing || obj == ObjectTypeOrder || obj == ObjectTypeTag
}

type Patcher struct {
	validator *validator.Validate
}

func NewPatcher(v *validator.Validate) *Patcher {
	return &Patcher{validator: v}
}

func (p *Patcher) Shop(in *Shop, patch Patch) error {
	var err error
	switch patch.Path.Type {
	case ObjectTypeSchemaVersion:
		if in.SchemaVersion != 0 && patch.Op != ReplaceOp {
			return fmt.Errorf("schema version can only be replaced once it is set")
		}
		if in.SchemaVersion == 0 && patch.Op != AddOp {
			return fmt.Errorf("schema version can only be initialized once")
		}
		var newVal uint64
		err = cbor.Unmarshal(patch.Value, &newVal)
		if err != nil {
			return fmt.Errorf("failed to unmarshal schema version: %w", err)
		}
		if newVal <= in.SchemaVersion {
			return fmt.Errorf("schema version can only be incremented")
		}
		in.SchemaVersion = newVal
		return nil
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
	case ObjectTypeInventory:
		// validate patch edits an existing listing
		if patch.Path.ObjectID == nil {
			return fmt.Errorf("inventory patch needs an id")
		}
		objId := *patch.Path.ObjectID
		lis, ok := in.Listings.Get(objId)
		if !ok {
			return fmt.Errorf("listing %d not found", objId)
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
		err = p.Inventory(&in.Inventory, patch)
	default:
		return fmt.Errorf("unsupported path: %s", patch.Path.Type)
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
	if patch.Path.AccountAddr == nil {
		return fmt.Errorf("account patch needs an ID")
	}
	accID := *patch.Path.AccountAddr
	acc, ok := in.Trie.Get(accID[:])
	switch patch.Op {
	case AddOp:
		if ok {
			return fmt.Errorf("account %s already exists", accID)
		}
		err = cbor.Unmarshal(patch.Value, &acc)
		if err != nil {
			return fmt.Errorf("failed to unmarshal account: %w", err)
		}
		err = in.Trie.Insert(accID[:], acc)
		if err != nil {
			return fmt.Errorf("failed to insert account %s: %w", accID, err)
		}
	case RemoveOp:
		if !ok {
			return fmt.Errorf("account %s not found", accID)
		}
		if len(patch.Path.Fields) == 0 {
			err = in.Trie.Delete(accID[:])
			if err != nil {
				return fmt.Errorf("failed to delete account %s: %w", accID, err)
			}
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
			err = in.Trie.Insert(accID[:], acc)
			if err != nil {
				return fmt.Errorf("failed to insert account %s: %w", accID, err)
			}
		}
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	return nil
}

func (p *Patcher) Inventory(in *Inventory, patch Patch) error {
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("inventory patch needs an ID")
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
	objId := *patch.Path.ObjectID
	current, ok := in.Get(objId, patch.Path.Fields)
	switch patch.Op {
	case AddOp:
		if ok {
			return fmt.Errorf("inventory %d already exists", objId)
		}
		err = in.Insert(objId, patch.Path.Fields, newVal)
	case RemoveOp:
		if !ok {
			return fmt.Errorf("inventory %d not found", objId)
		}
		err = in.Delete(objId, patch.Path.Fields)
	case ReplaceOp:
		if !ok {
			return fmt.Errorf("inventory %d not found", objId)
		}
		err = in.Insert(objId, patch.Path.Fields, newVal)
	case IncrementOp:
		current += newVal
		err = in.Insert(objId, patch.Path.Fields, current)
	case DecrementOp:
		if current < newVal {
			return fmt.Errorf("inventory %d cannot decrement below 0", objId)
		}
		current -= newVal
		err = in.Insert(objId, patch.Path.Fields, current)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	if err != nil {
		return fmt.Errorf("failed to patch(%s) inventory %d: %w", patch.Op, objId, err)
	}
	return nil
}

func (p *Patcher) Listings(listings Listings, patch Patch) error {
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("listing patch needs an ID")
	}
	objId := *patch.Path.ObjectID
	lis, ok := listings.Get(objId)
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
		err = listings.Insert(objId, lis)
		if err != nil {
			return fmt.Errorf("failed to insert Listing %d: %w", objId, err)
		}
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
		err = listings.Insert(objId, lis)
		if err != nil {
			return fmt.Errorf("failed to insert Listing %d: %w", objId, err)
		}
	case RemoveOp:
		if !ok {
			return fmt.Errorf("listing %d not found", objId)
		}
		if len(patch.Path.Fields) == 0 {
			err := listings.Delete(objId)
			if err != nil {
				return fmt.Errorf("failed to delete Listing %d: %w", objId, err)
			}
		} else {
			err := p.Listing(&lis, patch)
			if err != nil {
				return fmt.Errorf("failed to patch Listing %d: %w", objId, err)
			}
			err = listings.Insert(objId, lis)
			if err != nil {
				return fmt.Errorf("failed to insert Listing %d: %w", objId, err)
			}
		}
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
	tag, ok := in.Get(tagName)
	var err error
	switch patch.Op {
	case AddOp:
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
		err = in.Insert(tagName, tag)
		if err != nil {
			return fmt.Errorf("failed to insert Tag %s: %w", tagName, err)
		}
	case ReplaceOp:
		if !ok {
			return fmt.Errorf("tag %s not found", tagName)
		}
		if len(patch.Path.Fields) == 0 {
			err = Unmarshal(patch.Value, &tag)
			if err != nil {
				return fmt.Errorf("failed to patch Tag %s: %w", tagName, err)
			}
		} else {
			err = p.Tag(&tag, patch)
			if err != nil {
				return fmt.Errorf("failed to patch Tag %s: %w", tagName, err)
			}
		}
		err = in.Insert(tagName, tag)
		if err != nil {
			return fmt.Errorf("failed to insert Tag %s: %w", tagName, err)
		}
	case RemoveOp:
		if !ok {
			return fmt.Errorf("tag %s not found", tagName)
		}
		if len(patch.Path.Fields) == 0 {
			err = in.Delete(tagName)
		} else {
			err = p.Tag(&tag, patch)
			if err != nil {
				return fmt.Errorf("failed to patch Tag %s: %w", tagName, err)
			}
			err = in.Insert(tagName, tag)
			if err != nil {
				return fmt.Errorf("failed to insert Tag %s: %w", tagName, err)
			}
		}
		if err != nil {
			return fmt.Errorf("failed to delete Tag %s: %w", tagName, err)
		}
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
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("order patch needs an ID")
	}
	objId := *patch.Path.ObjectID
	order, ok := in.Get(objId)
	var err error
	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if ok {
				return fmt.Errorf("order %d already exists", objId)
			}
			err = Unmarshal(patch.Value, &order)
			if err != nil {
				return fmt.Errorf("failed to unmarshal order: %w", err)
			}
		} else {
			err := p.Order(&order, patch)
			if err != nil {
				return fmt.Errorf("failed to patch Order %d: %w", objId, err)
			}
		}
		err = in.Insert(objId, order)
	case ReplaceOp:
		if !ok {
			return fmt.Errorf("order %d not found", objId)
		}
		if len(patch.Path.Fields) == 0 {
			err = Unmarshal(patch.Value, &order)
		} else {
			err = p.Order(&order, patch)
		}
		if err != nil {
			return fmt.Errorf("failed to patch Order %d: %w", objId, err)
		}
		err = in.Insert(objId, order)
	case RemoveOp:
		if !ok {
			return fmt.Errorf("order %d not found", objId)
		}
		if len(patch.Path.Fields) == 0 {
			err = in.Delete(objId)
		} else {
			err = p.Order(&order, patch)
			if err != nil {
				return fmt.Errorf("failed to patch Order %d: %w", objId, err)
			}
			err = in.Insert(objId, order)
		}
	case IncrementOp:
		err = p.Order(&order, patch)
	case DecrementOp:
		err = p.Order(&order, patch)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	if err != nil {
		return fmt.Errorf("failed to patch Order %d: %w", objId, err)
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
	case IncrementOp:
		err = in.PatchIncrement(patch.Path.Fields, patch.Value)
	case DecrementOp:
		err = in.PatchDecrement(patch.Path.Fields, patch.Value)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	if err != nil {
		return err
	}
	return p.validator.Struct(in)
}

func (existing *Order) checkPathAndIndex(fields []string) (int, error) {
	if len(fields) != 3 || fields[0] != "items" || fields[2] != "quantity" {
		return 0, fmt.Errorf("incr/decr only works on path: [items, x, quantity]")
	}
	index, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert index to int: %w", err)
	}
	if index < 0 || index >= len(existing.Items) {
		return 0, fmt.Errorf("index out of bounds: %d", index)
	}
	return index, nil
}

func (existing *Order) PatchIncrement(fields []string, data cbor.RawMessage) error {
	index, err := existing.checkPathAndIndex(fields)
	if err != nil {
		return err
	}
	var value uint32
	err = Unmarshal(data, &value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}
	existing.Items[index].Quantity += value
	return nil
}

func (existing *Order) PatchDecrement(fields []string, data cbor.RawMessage) error {
	index, err := existing.checkPathAndIndex(fields)
	if err != nil {
		return err
	}
	var value uint32
	err = Unmarshal(data, &value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}
	existing.Items[index].Quantity -= value
	return nil
}
