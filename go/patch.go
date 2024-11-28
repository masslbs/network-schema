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
	Type   string   `validate:"required,notblank"`
	ID     ObjectId `validate:"required,gt=0"`
	Fields []string
}

func (pp *PatchPath) UnmarshalCBOR(data []byte) error {
	var v []any
	err := Unmarshal(data, &v)
	if err != nil {
		return err
	}
	if len(v) < 2 {
		return fmt.Errorf("PatchPath must have at least two elements [type, id]")
	}
	pp.Type = v[0].(string)
	pp.ID = v[1].(ObjectId)
	for _, field := range v[2:] {
		pp.Fields = append(pp.Fields, field.(string))
	}
	return nil
}

func (pp PatchPath) MarshalCBOR() ([]byte, error) {
	path := make([]any, 2+len(pp.Fields))
	path[0] = pp.Type
	path[1] = pp.ID
	for i, field := range pp.Fields {
		path[2+i] = field
	}
	return Marshal(path)
}
