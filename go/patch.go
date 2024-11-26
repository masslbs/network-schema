/*
https://datatracker.ietf.org/doc/html/rfc6902/
This is a modified version of JSON Patch (rfc6902). We first constraint the operations to "add", "replace", "remove" and then we add the following operations
- increment - Increments a number by a specified value. Only valid if the value of the path is a number.,
- decrement - Decrements a number by a specified value. Only valid if the value of the path is a number.,
*/

package main

type Write struct {
	Patchs []Patch
}

type Patch struct {
	Op   OpString
	Path string
	// ??
	Value map[string]interface{}
}

type OpString string

const (
    AddOp OpString = "add"
    ReplaceOp OpString = "replace"
    RemoveOp OpString = "remove"
    IncrementOp OpString = "increment"
    DecrementOp OpString = "decrement"
)
