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
	"github.com/fxamacker/cbor/v2"
)

type Write struct {
	Patches []PatchTx
}

type PatchRx struct {
	Op    OpString        `validate:"oneof=add replace remove increment decrement"`
	Path  []any           `validate:"required,gte=2"`
	Value cbor.RawMessage `validate:"required,gt=0"`
}

type PatchTx struct {
	Op    OpString `validate:"oneof=add replace remove increment decrement"`
	Path  []any    `validate:"required,gte=2"`
	Value any      `validate:"required"`
}

type OpString string

const (
	AddOp       OpString = "add"
	ReplaceOp   OpString = "replace"
	RemoveOp    OpString = "remove"
	IncrementOp OpString = "increment"
	DecrementOp OpString = "decrement"
)
