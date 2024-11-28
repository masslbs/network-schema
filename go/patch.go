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
	Path  PatchPath       `validate:"required,gte=2"`
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

type PatchPath []any

func (pp *PatchPath) UnmarshalCBOR(data []byte) error {
	var v []any
	err := Unmarshal(data, &v)
	if err != nil {
		return err
	}
	if len(v) < 2 {
		return fmt.Errorf("PatchPath must have at least two elements [type, id]")
	}
	*pp = v
	return nil
}

func (p PatchPath) Type() string {
	assert(len(p) > 0, "PatchPath must have at least one element")
	first, ok := p[0].(string)
	if !ok {
		return ""
	}
	return first
}

func (p PatchPath) ID() ObjectId {
	assert(len(p) > 1, "PatchPath must have at least two elements")
	id, ok := p[1].(ObjectId)
	if !ok {
		return 0
	}
	return id
}

func (p PatchPath) Fields() []string {
	assert(len(p) > 1, "PatchPath must have at least two elements")
	var fields []string
	for _, field := range p[2:] {
		fields = append(fields, field.(string))
	}
	return fields
}

func assert(condition bool, msg string) {
	if !condition {
		panic(msg)
	}
}
