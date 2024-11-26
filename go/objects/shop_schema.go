// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

// Package objects contains the type definitions and encodingfunctions for the shop schema
package objects

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/go-playground/validator/v10"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	hamt "github.com/masslbs/network-schema/go/hamt"
)

// Shop represents a shop and all its contents
type Shop struct {
	// SchemaVersion is the version of the shop schema. Can only be increased.
	SchemaVersion uint64 `validate:"required,gt=0"`
	Tags          Tags
	Orders        Orders
	Accounts      Accounts
	Listings      Listings
	Manifest      Manifest `validate:"required"`
	Inventory     Inventory
}

// NewShop creates a new shop
func NewShop(version uint64) Shop {
	s := Shop{}
	s.SchemaVersion = version
	s.Accounts.Trie = hamt.NewTrie[Account]()
	s.Listings.Trie = hamt.NewTrie[Listing]()
	s.Orders.Trie = hamt.NewTrie[Order]()
	s.Tags.Trie = hamt.NewTrie[Tag]()
	s.Inventory.Trie = hamt.NewTrie[uint64]()
	return s
}

// HAMTValidation validates the HAMT types
func HAMTValidation(sl validator.StructLevel) {
	hamt := sl.Current().Interface()
	val := sl.Validator()
	switch tval := hamt.(type) {
	case Accounts:
		tval.All(func(key []byte, value Account) bool {
			if len(key) != EthereumAddressSize {
				sl.ReportError(value, "key", "key", "tooShort", "")
				return true
			}
			err := val.Struct(value)
			if err != nil {
				sl.ReportError(value, string(key), "value", err.Error(), "")
			}
			return true
		})
	case Listings:
		tval.All(func(key []byte, value Listing) bool {
			if len(key) < 8 {
				sl.ReportError(value, "key", "key", "tooShort", "")
				return true
			}
			id := bytesToID(key)
			if id == 0 {
				sl.ReportError(value, "key", "key", "notZero", "")
				return true
			}
			err := val.Struct(value)
			if err != nil {
				sl.ReportError(value, string(key), "value", err.Error(), "")
			}
			return true
		})
	case Orders:
		tval.All(func(key []byte, value Order) bool {
			if len(key) < 8 {
				sl.ReportError(value, "key", "key", "tooShort", "")
				return true
			}
			id := bytesToID(key)
			if id == 0 {
				sl.ReportError(value, "key", "key", "notZero", "")
				return true
			}
			err := val.Struct(value)
			if err != nil {
				sl.ReportError(value, string(key), "value", err.Error(), "")
			}
			return true
		})
	case Tags:
		tval.All(func(key []byte, value Tag) bool {
			if len(key) == 0 {
				sl.ReportError(value, "key", "key", "tooShort", "")
				return true
			}
			err := val.Struct(value)
			if err != nil {
				sl.ReportError(value, string(key), "value", err.Error(), "")
			}
			return true
		})
	case Inventory:
		tval.All(func(key []byte, value uint64) bool {
			if len(key) < 8 {
				sl.ReportError(value, "key", "key", "tooShort", "")
				return true
			}
			id := bytesToID(key)
			if id == 0 {
				sl.ReportError(value, "key", "key", "notZero", "")
				return true
			}
			return true
		})
	default:
		panic(fmt.Sprintf("unknown hamt type: %T", tval))
	}
}

// Hash hashes the shop
func (s *Shop) Hash() (Hash, error) {
	var err error

	// the hash is calculated by using the hashes of the hamts
	var hashedShop struct {
		SchemaVersion uint64
		Manifest      Manifest
		Tags          []byte
		Orders        []byte
		Accounts      []byte
		Listings      []byte
		Inventory     []byte
	}

	// copy the shop data into the hashedShop struct
	hashedShop.SchemaVersion = s.SchemaVersion
	hashedShop.Manifest = s.Manifest

	// hash all the hamts
	hashedShop.Tags, err = s.Tags.Hash()
	if err != nil {
		return Hash{}, err
	}
	hashedShop.Orders, err = s.Orders.Hash()
	if err != nil {
		return Hash{}, err
	}
	hashedShop.Accounts, err = s.Accounts.Hash()
	if err != nil {
		return Hash{}, err
	}
	hashedShop.Listings, err = s.Listings.Hash()
	if err != nil {
		return Hash{}, err
	}
	hashedShop.Inventory, err = s.Inventory.Hash()
	if err != nil {
		return Hash{}, err
	}
	// finally, hash the whole thing
	h := sha256.New()
	// var buf bytes.Buffer
	// w := io.MultiWriter(h, &buf)
	err = masscbor.DefaultEncoder(h).Encode(hashedShop)
	if err != nil {
		return Hash{}, err
	}
	// fmt.Println("\n\ndebug:\n")
	// fmt.Println(hex.EncodeToString(buf.Bytes()))
	return Hash(h.Sum(nil)), nil
}

// Tag represents a tag
type Tag struct {
	Name       string `validate:"required,notblank"`
	ListingIDs []ObjectID
}

// Account represents an account
type Account struct {
	KeyCards []PublicKey
	Guest    bool
}

// HAMT wrapper types for key typing

// Accounts is a HAMT of accounts
type Accounts struct {
	*hamt.Trie[Account]
}

// Listings is a HAMT of listings
type Listings struct {
	*hamt.Trie[Listing]
}

// Get gets a listing from the HAMT
func (l *Listings) Get(id ObjectID) (Listing, bool) {
	buf := idToBytes(id)
	lis, ok := l.Trie.Get(buf)
	return lis, ok
}

// Has checks if a listing exists in the HAMT
func (l *Listings) Has(id ObjectID) bool {
	buf := idToBytes(id)
	_, ok := l.Trie.Get(buf)
	return ok
}

// Insert inserts a listing into the HAMT
func (l *Listings) Insert(id ObjectID, lis Listing) error {
	buf := idToBytes(id)
	return l.Trie.Insert(buf, lis)
}

// Delete deletes a listing from the HAMT
func (l *Listings) Delete(id ObjectID) error {
	buf := idToBytes(id)
	return l.Trie.Delete(buf)
}

// idToBytes converts an ObjectID to a byte slice
func idToBytes(id ObjectID) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(id))
	return buf
}

// Tags is a HAMT of tags
type Tags struct {
	*hamt.Trie[Tag]
}

// Get gets a tag from the HAMT
func (t *Tags) Get(name string) (Tag, bool) {
	buf := []byte(name)
	tag, ok := t.Trie.Get(buf)
	return tag, ok
}

// Has checks if a tag exists in the HAMT
func (t *Tags) Has(name string) bool {
	_, ok := t.Get(name)
	return ok
}

// Insert inserts a tag into the HAMT
func (t *Tags) Insert(name string, tag Tag) error {
	buf := []byte(name)
	return t.Trie.Insert(buf, tag)
}

// Delete deletes a tag from the HAMT
func (t *Tags) Delete(name string) error {
	buf := []byte(name)
	return t.Trie.Delete(buf)
}

// Orders is a HAMT of orders
type Orders struct {
	*hamt.Trie[Order]
}

// Get gets an order from the HAMT
func (l *Orders) Get(id ObjectID) (Order, bool) {
	buf := idToBytes(id)
	val, ok := l.Trie.Get(buf)
	return val, ok
}

// Insert inserts an order into the HAMT
func (l *Orders) Insert(id ObjectID, val Order) error {
	buf := idToBytes(id)
	return l.Trie.Insert(buf, val)
}

// Delete deletes an order from the HAMT
func (l *Orders) Delete(id ObjectID) error {
	buf := idToBytes(id)
	return l.Trie.Delete(buf)
}

// Inventory is a HAMT of inventory
type Inventory struct {
	*hamt.Trie[uint64]
}

// Get gets an inventory item from the HAMT
func (l *Inventory) Get(id ObjectID, variations []string) (uint64, bool) {
	buf := combinedIDtoBytes(id, variations)
	val, ok := l.Trie.Get(buf)
	return val, ok
}

// Insert inserts an inventory item into the HAMT
func (l *Inventory) Insert(id ObjectID, variations []string, val uint64) error {
	buf := combinedIDtoBytes(id, variations)
	return l.Trie.Insert(buf, val)
}

// Delete deletes an inventory item from the HAMT
func (l *Inventory) Delete(id ObjectID, variations []string) error {
	buf := combinedIDtoBytes(id, variations)
	return l.Trie.Delete(buf)
}
