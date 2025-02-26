// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

/*
https://datatracker.ietf.org/doc/html/rfc6902/
This is a modified version of JSON Patch (rfc6902). We first constraint the operations to "add", "replace", "remove" and then we add the following operations
- increment - Increments a number by a specified value. Only valid if the value of the path is a number.,
- decrement - Decrements a number by a specified value. Only valid if the value of the path is a number.,
*/

package patch

import (
	"fmt"
	"time"

	"github.com/fxamacker/cbor/v2"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/masslbs/network-schema/go/objects"
)

// To validate a patchset, construct the merkle tree of the patches and validate the root hash.
type SignedPatchSet struct {
	// The header of the patch set
	Header PatchSetHeader `validate:"required"` // TODO: dive doesn't work?

	// The signature of the header, containing the merkle root of the patches
	Signature objects.Signature `validate:"required,gt=0,dive"`

	Patches []Patch `validate:"required,gt=0,dive"`
}

type PatchSetHeader struct {
	// The nonce must be unique for each event a keycard creates.
	// The sequence values need to increase monotonicly.
	KeyCardNonce uint64 `validate:"required,gt=0"`

	// Every signed event must be tied to a shop id. This allow the
	// event to processed outside the context of the currenct connection.
	ShopID objects.Uint256 `validate:"required"`

	// The time when this event was created.
	// The relay should reject any events from the future
	Timestamp time.Time `validate:"required"`

	// The merkle root of the patches
	RootHash objects.Hash `validate:"required"`
}

type Patch struct {
	Op    OpString        `validate:"required,oneof=add replace remove increment decrement"`
	Path  PatchPath       `validate:"required"`
	Value cbor.RawMessage `validate:"required,gt=0"`
}

func (p Patch) Serialize() ([]byte, error) {
	return masscbor.Marshal(p)
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
	ObjectID    *objects.ObjectId
	AccountAddr *objects.EthereumAddress
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
		return fmt.Errorf("patch.path: invalid object type: %v", path[0])
	}
	p.Type = ObjectType(objType)
	if !p.Type.IsValid() {
		return fmt.Errorf("patch.path: invalid object type: %s", objType)
	}
	path = path[1:] // slice of type
	// path[0] can be a uint64, an ethereum address or empty (for manifest)
	switch p.Type {
	case ObjectTypeManifest:
		// noop
	case ObjectTypeAccount:
		if len(path) < 1 {
			return fmt.Errorf("patch.path: invalid ethereum address: %w", err)
		}
		data, ok := path[0].([]byte)
		if !ok {
			return fmt.Errorf("patch.path: invalid ethereum address: %v", path[0])
		}
		if len(data) != objects.EthereumAddressSize {
			return fmt.Errorf("patch.path: invalid ethereum address size: %d != %d", len(data), objects.EthereumAddressSize)
		}
		var addr objects.EthereumAddress
		copy(addr[:], data)
		p.AccountAddr = &addr
	case ObjectTypeOrder, ObjectTypeListing:
		if len(path) < 1 {
			return fmt.Errorf("patch.path: invalid object id: %w", err)
		}
		id, ok := path[0].(uint64)
		if !ok {
			return fmt.Errorf("patch.path: invalid object id: %v", path[0])
		}
		objId := objects.ObjectId(id)
		p.ObjectID = &objId
	case ObjectTypeTag:
		if len(path) < 1 {
			return fmt.Errorf("patch.path: invalid tag name: %w", err)
		}
		tagName, ok := path[0].(string)
		if !ok {
			return fmt.Errorf("patch.path: invalid tag name: %v", path[0])
		}
		p.TagName = &tagName
	default:
		return fmt.Errorf("patch.path: invalid id type: %s", path[0])
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
