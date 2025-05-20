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
)

// fix formatting for test vectors
// go's json encoder defaults to encode []byte as base64 encoded string

// MarshalJSON encodes a signature as a base64 encoded string
func (sig Signature) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(sig[:]))
}

// MarshalJSON encodes an Accounts object as a JSON object
func (accs Accounts) MarshalJSON() ([]byte, error) {
	// Convert account/userWallet addresses to hex strings for JSON compatible map keys
	hexAccs := make(map[string]Account, accs.Size())
	accs.All(func(addr []byte, acc Account) bool {
		hexAccs[hex.EncodeToString(addr)] = acc
		return true
	})
	return json.Marshal(hexAccs)
}

// MarshalJSON encodes a Listings object as a JSON object
func (lis Listings) MarshalJSON() ([]byte, error) {
	hexLis := make(map[ObjectID]Listing, lis.Size())
	lis.All(func(id []byte, lis Listing) bool {
		hexLis[bytesToID(id)] = lis
		return true
	})
	return json.Marshal(hexLis)
}

// MarshalJSON encodes an Inventory object as a JSON object
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

// MarshalJSON encodes an Ethereum address as a JSON string
func (addr EthereumAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(addr.Address)
}

// MarshalJSON encodes a ChainAddress as a JSON object
func (addr ChainAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ChainID int    `json:"ChainID"`
		Address string `json:"Address"`
	}{
		ChainID: int(addr.ChainID),
		Address: addr.Address.Hex(),
	})
}

// MarshalJSON encodes a PublicKey as a base64 encoded string
func (pub PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(pub[:]))
}

// MarshalJSON encodes a Hash as a base64 encoded string
func (h Hash) MarshalJSON() ([]byte, error) {
	return json.Marshal(base64.StdEncoding.EncodeToString(h[:]))
}

// MarshalJSON encodes a Shop object as a JSON object
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
