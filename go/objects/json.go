// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package objects

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// fix formatting for test vectors
// go's json encoder defaults to encode []byte as base64 encoded string

func (sig Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(sig[:]))
}

func (accs Accounts) MarshalJSON() ([]byte, error) {
	// Convert account/userWallet addresses to hex strings for JSON compatible map keys
	hexAccs := make(map[string]Account, accs.Size())
	accs.All(func(addr []byte, acc Account) bool {
		hexAccs[hex.EncodeToString(addr)] = acc
		return true
	})
	return json.Marshal(hexAccs)
}

func (lis Listings) MarshalJSON() ([]byte, error) {
	hexLis := make(map[ObjectId]Listing, lis.Size())
	lis.All(func(id []byte, lis Listing) bool {
		hexLis[bytesToId(id)] = lis
		return true
	})
	return json.Marshal(hexLis)
}

func (inv Inventory) MarshalJSON() ([]byte, error) {
	stringID := make(map[string]uint64, inv.Size())
	inv.All(func(id []byte, inv uint64) bool {
		objID, vars := bytesToCombinedID(id)
		mapKey := strconv.FormatUint(uint64(objID), 10)
		if len(vars) > 0 {
			mapKey += ":" + strings.Join(vars, "-")
		}
		stringID[mapKey] = inv
		return true
	})
	return json.Marshal(stringID)
}

// use default json encoding for ethereum addresses
func (addr EthereumAddress) MarshalJSON() ([]byte, error) {
	common := common.Address(addr)
	return json.Marshal(common)
}

func (pub PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(pub[:]))
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(h[:]))
}

func (s Shop) MarshalJSON() ([]byte, error) {
	type jsonShop struct {
		SchemaVersion uint64             `json:"SchemaVersion"`
		Manifest      Manifest           `json:"Manifest"`
		Tags          map[string]Tag     `json:"Tags"`
		Orders        map[string]Order   `json:"Orders"`
		Accounts      map[string]Account `json:"Accounts"`
		Listings      map[string]Listing `json:"Listings"`
		Inventory     map[string]uint64  `json:"Inventory"`
	}
	var js jsonShop
	js.SchemaVersion = s.SchemaVersion
	js.Manifest = s.Manifest
	js.Tags = make(map[string]Tag)
	js.Orders = make(map[string]Order)
	js.Accounts = make(map[string]Account)
	js.Listings = make(map[string]Listing)
	js.Inventory = make(map[string]uint64)
	s.Tags.All(func(name []byte, tag Tag) bool {
		js.Tags[hex.EncodeToString(name)] = tag
		return true
	})

	s.Orders.All(func(id []byte, order Order) bool {
		js.Orders[hex.EncodeToString(id)] = order
		return true
	})

	s.Accounts.All(func(addr []byte, acc Account) bool {
		js.Accounts[hex.EncodeToString(addr)] = acc
		return true
	})

	s.Listings.All(func(id []byte, lis Listing) bool {
		js.Listings[hex.EncodeToString(id)] = lis
		return true
	})

	s.Inventory.All(func(id []byte, inv uint64) bool {
		js.Inventory[hex.EncodeToString(id)] = inv
		return true
	})

	return json.Marshal(js)
}
