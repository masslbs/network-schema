// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package testhelper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"

	"github.com/ethereum/go-ethereum/common"
	"github.com/masslbs/network-schema/go/objects"
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

func CommonEthAddr(b [20]byte) common.Address {
	return common.Address(b)
}

func MassEthAddr(b [20]byte) objects.EthereumAddress {
	return objects.EthereumAddress{
		Address: CommonEthAddr(b),
	}
}

func MassEthAddrPtr(b [20]byte) *objects.EthereumAddress {
	return &objects.EthereumAddress{
		Address: CommonEthAddr(b),
	}
}

func RandomUint256() objects.Uint256 {
	var v [32]byte
	rand.Read(v[:])
	var obj objects.Uint256
	obj.SetBytes(v[:])
	return obj
}
