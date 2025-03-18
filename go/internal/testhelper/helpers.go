// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package testhelper

import (
	"crypto/sha256"
	"encoding/binary"
)

func TestHash(i uint) []byte {
	h := sha256.New()
	binary.Write(h, binary.BigEndian, i)
	return h.Sum(nil)
}

func Strptr(s string) *string {
	return &s
}

func Boolptr(b bool) *bool {
	return &b
}

func Uint64ptr(i uint64) *uint64 {
	return &i
}
