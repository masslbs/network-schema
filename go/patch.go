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

type Write struct {
	Patches []Patch
}

type Patch struct {
	Op    OpString        `validate:"oneof=add replace remove increment decrement"`
	Path  PatchPath       `validate:"required"`
	Value cbor.RawMessage `validate:"required,gt=0"`
}

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
	Type   string `validate:"required,notblank"`
	ID     ObjectId
	Fields []string
}

func (p PatchPath) MarshalCBOR() ([]byte, error) {
	var path = make([]any, len(p.Fields)+2)
	path[0] = p.Type
	path[1] = p.ID
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
	p.Type = path[0].(string)
	p.ID = path[1].(ObjectId)
	p.Fields = make([]string, len(path)-2)
	for i, field := range path[2:] {
		p.Fields[i] = field.(string)
	}
	return nil
}

type Patcher struct {
	validator *validator.Validate
}

func (p *Patcher) Manifest(in *Manifest, patch Patch) error {
	var err error
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

func (p *Patcher) Listing(in *Listing, patch Patch) error {
	var err error
	switch patch.Op {
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
