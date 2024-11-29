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
	Path  []any           `validate:"required,gte=2"`
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

func (patch Patch) UnpackPath() (PatchPath, error) {
	var path PatchPath
	if len(patch.Path) < 2 {
		return PatchPath{}, fmt.Errorf("PatchPath must have at least two elements [type, id]")
	}
	path.Type = patch.Path[0].(string)
	path.ID = patch.Path[1].(ObjectId)
	for _, field := range patch.Path[2:] {
		path.Fields = append(path.Fields, field.(string))
	}
	return path, nil
}
