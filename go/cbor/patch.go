// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
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
	"math/big"
	"slices"
	"strconv"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-playground/validator/v10"
)

// To validate a patchset, construct the merkle tree of the patches and validate the root hash.
type SignedPatchSet struct {
	// The header of the patch set
	Header PatchSetHeader `validate:"required"` // TODO: dive doesn't work?

	// The signature of the header, containing the merkle root of the patches
	Signature Signature `validate:"required,gt=0,dive"`

	Patches []Patch `validate:"required,gt=0,dive"`
}

type PatchSetHeader struct {
	// The nonce must be unique for each event a keycard creates.
	// The sequence values need to increase monotonicly.
	KeyCardNonce uint64 `validate:"required,gt=0"`

	// Every signed event must be tied to a shop id. This allow the
	// event to processed outside the context of the currenct connection.
	ShopID Uint256 `validate:"required"`

	// The time when this event was created.
	// The relay should reject any events from the future
	Timestamp time.Time `validate:"required"`

	// The merkle root of the patches
	RootHash Hash `validate:"required"`
}

type Patch struct {
	Op    OpString        `validate:"required,oneof=add replace remove increment decrement"`
	Path  PatchPath       `validate:"required"`
	Value cbor.RawMessage `validate:"required,gt=0"`
}

func (p Patch) Serialize() ([]byte, error) {
	return Marshal(p)
}

// TODO: type change to enum/number instead of string?
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
	Type ObjectType `validate:"required,notblank"`

	// one-of, element 1 of the array
	//
	// account: EthereumAddress
	// tag: TagName as string
	// listing: ObjectId
	// order: ObjectId
	// inventory: ObjectId with optional variations as fields
	//
	// exclusion: manifest, which has no id
	ObjectID    *ObjectId
	AccountAddr *EthereumAddress
	TagName     *string

	Fields []string // extra fields
}

func (p PatchPath) MarshalCBOR() ([]byte, error) {
	var extraFields = 2 // usually we have type and id
	if p.Type == ObjectTypeManifest {
		extraFields = 1 // manifest has no id
	}
	var path = make([]any, len(p.Fields)+extraFields)
	path[0] = string(p.Type)
	switch p.Type {
	case ObjectTypeManifest:
		if p.ObjectID != nil {
			return nil, fmt.Errorf("manifest patch should not have an id")
		}
		if p.AccountAddr != nil {
			return nil, fmt.Errorf("manifest patch should not have an account id")
		}
		if p.TagName != nil {
			return nil, fmt.Errorf("manifest patch should not have a tag name")
		}
	case ObjectTypeAccount:
		if p.AccountAddr == nil {
			return nil, fmt.Errorf("account patch needs an id")
		}
		path[1] = *p.AccountAddr
	case ObjectTypeListing, ObjectTypeOrder:
		if p.ObjectID == nil {
			return nil, fmt.Errorf("listing/order patch needs an id")
		}
		path[1] = *p.ObjectID
	case ObjectTypeTag:
		if p.TagName == nil {
			return nil, fmt.Errorf("tag patch needs a tag name")
		}
		path[1] = *p.TagName
	}
	for i, field := range p.Fields {
		path[i+extraFields] = field
	}
	return cbor.Marshal(path)
}

func (p *PatchPath) UnmarshalCBOR(data []byte) error {
	var path []any
	err := cbor.Unmarshal(data, &path)
	if err != nil {
		return err
	}
	objType, ok := path[0].(string)
	if !ok {
		return fmt.Errorf("patch.path: invalid object type: %v", path[0])
	}
	p.Type = ObjectType(objType)
	if !p.Type.IsValid() {
		return fmt.Errorf("patch.path: invalid object type: %s", objType)
	}
	path = path[1:] // slice of type
	// path[0] can be a uint64, an ethereum address or empty (for manifest)
	switch p.Type {
	case ObjectTypeManifest:
		// noop
	case ObjectTypeAccount:
		if len(path) < 1 {
			return fmt.Errorf("patch.path: invalid ethereum address: %w", err)
		}
		data, ok := path[0].([]byte)
		if !ok {
			return fmt.Errorf("patch.path: invalid ethereum address: %v", path[0])
		}
		if len(data) != EthereumAddressSize {
			return fmt.Errorf("patch.path: invalid ethereum address size: %d != %d", len(data), EthereumAddressSize)
		}
		var addr EthereumAddress
		copy(addr[:], data)
		p.AccountAddr = &addr
	case ObjectTypeOrder, ObjectTypeListing:
		if len(path) < 1 {
			return fmt.Errorf("patch.path: invalid object id: %w", err)
		}
		id, ok := path[0].(uint64)
		if !ok {
			return fmt.Errorf("patch.path: invalid object id: %v", path[0])
		}
		objId := ObjectId(id)
		p.ObjectID = &objId
	case ObjectTypeTag:
		if len(path) < 1 {
			return fmt.Errorf("patch.path: invalid tag name: %w", err)
		}
		tagName, ok := path[0].(string)
		if !ok {
			return fmt.Errorf("patch.path: invalid tag name: %v", path[0])
		}
		p.TagName = &tagName
	default:
		return fmt.Errorf("patch.path: invalid id type: %s", path[0])
	}

	if p.Type != ObjectTypeManifest {
		path = path[1:] // all other types have an id
	}

	// validate contextual type<>id values
	switch p.Type {
	case ObjectTypeManifest:
		if p.ObjectID != nil {
			return fmt.Errorf("manifest patch should not have an id")
		}
		if p.AccountAddr != nil {
			return fmt.Errorf("manifest patch should not have an account id")
		}
		if p.TagName != nil {
			return fmt.Errorf("manifest patch should not have a tag name")
		}
	case ObjectTypeAccount:
		if p.AccountAddr == nil {
			return fmt.Errorf("account patch needs an id")
		}
		if p.ObjectID != nil {
			return fmt.Errorf("account patch should not have an object id")
		}
		if p.TagName != nil {
			return fmt.Errorf("account patch should not have a tag name")
		}
	case ObjectTypeListing, ObjectTypeOrder:
		if p.ObjectID == nil {
			return fmt.Errorf("listing/order patch needs an id")
		}
		if p.AccountAddr != nil {
			return fmt.Errorf("listing/order patch should not have an account id")
		}
		if p.TagName != nil {
			return fmt.Errorf("listing/order patch should not have a tag name")
		}
	case ObjectTypeTag:
		if p.TagName == nil {
			return fmt.Errorf("tag patch needs a tag name")
		}
		if len(*p.TagName) == 0 {
			return fmt.Errorf("tag name cannot be empty")
		}
		if p.ObjectID != nil {
			return fmt.Errorf("tag patch should not have an object id")
		}
		if p.AccountAddr != nil {
			return fmt.Errorf("tag patch should not have an account id")
		}
	default:
		return fmt.Errorf("unsupported path type: %s", p.Type)
	}

	// the rest of the path is the fields
	p.Fields = make([]string, len(path))
	for i, field := range path {
		p.Fields[i], ok = field.(string)
		if !ok {
			return fmt.Errorf("invalid field: %v", field)
		}
	}
	return nil
}

// TODO: could change to number instead of string..?
type ObjectType string

const (
	ObjectTypeSchemaVersion ObjectType = "schemaVersion"
	ObjectTypeManifest      ObjectType = "manifest"
	ObjectTypeAccount       ObjectType = "account"
	ObjectTypeListing       ObjectType = "listing"
	ObjectTypeOrder         ObjectType = "order"
	ObjectTypeTag           ObjectType = "tag"
	ObjectTypeInventory     ObjectType = "inventory"
)

func (obj *ObjectType) UnmarshalCBOR(data []byte) error {
	var s string
	err := cbor.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	var newVal = ObjectType(s)
	if !newVal.IsValid() {
		return fmt.Errorf("invalid object type: %s", s)
	}
	*obj = newVal
	return nil
}

func (obj ObjectType) IsValid() bool {
	return obj == ObjectTypeSchemaVersion || obj == ObjectTypeManifest || obj == ObjectTypeAccount || obj == ObjectTypeListing || obj == ObjectTypeOrder || obj == ObjectTypeTag
}

type ObjectNotFoundError struct {
	ObjectType ObjectType
	Path       PatchPath
}

func (e ObjectNotFoundError) Error() string {
	var id string
	if e.Path.ObjectID != nil {
		id = fmt.Sprintf("type=%s id=%d", e.ObjectType, *e.Path.ObjectID)
	} else if e.Path.TagName != nil {
		id = fmt.Sprintf("tag=%s", *e.Path.TagName)
	} else if e.Path.AccountAddr != nil {
		id = fmt.Sprintf("account=%s", *e.Path.AccountAddr)
	} else {
		id = fmt.Sprintf("type=%s", e.ObjectType)
	}
	if len(e.Path.Fields) > 0 {
		return fmt.Sprintf("object %s with fields=%v not found", id, e.Path.Fields)
	}
	return fmt.Sprintf("object %s not found", id)
}

// Centralize patch operations in the Patcher type
type Patcher struct {
	validator *validator.Validate
	shop      *Shop // Add reference to shop for lookups
}

func NewPatcher(v *validator.Validate, shop *Shop) *Patcher {
	return &Patcher{validator: v, shop: shop}
}

// Main entry point for applying patches
func (p *Patcher) ApplyPatch(patch Patch) error {
	switch patch.Path.Type {
	case ObjectTypeManifest:
		return p.patchManifest(patch)
	case ObjectTypeAccount:
		return p.patchAccount(patch)
	case ObjectTypeListing:
		return p.patchListing(patch)
	case ObjectTypeTag:
		return p.patchTag(patch)
	case ObjectTypeOrder:
		return p.patchOrder(patch)
	case ObjectTypeInventory:
		return p.patchInventory(&p.shop.Inventory, patch)
	default:
		return fmt.Errorf("unsupported shop object type: %s", patch.Path.Type)
	}
}

// Example of centralized tag patching with referential checks
func (p *Patcher) patchTag(patch Patch) error {
	if patch.Path.TagName == nil {
		return fmt.Errorf("tag patch needs a tag name")
	}
	tagName := *patch.Path.TagName
	tag, exists := p.shop.Tags.Get(tagName)

	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if exists {
				return fmt.Errorf("tag %s already exists", tagName)
			}
			var newTag Tag
			if err := Unmarshal(patch.Value, &newTag); err != nil {
				return fmt.Errorf("failed to unmarshal tag: %w", err)
			}
			if err := p.validator.Struct(newTag); err != nil {
				return err
			}
			// Verify all referenced listings exist
			for _, listingId := range newTag.ListingIds {
				if _, exists := p.shop.Listings.Get(listingId); !exists {
					return fmt.Errorf("listing %d referenced by tag does not exist", listingId)
				}
			}
			newTag.Name = tagName
			return p.shop.Tags.Insert(tagName, newTag)
		}

		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
		}

		// Handle adding listing IDs with referential checks
		if patch.Path.Fields[0] == "listingIds" {
			if err := p.addListingToTag(&tag, patch); err != nil {
				return err
			}
			return p.shop.Tags.Insert(tagName, tag)
		}

		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[0])

	case RemoveOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
		}

		if len(patch.Path.Fields) == 0 {
			return p.shop.Tags.Delete(tagName)
		}

		if len(patch.Path.Fields) == 2 && patch.Path.Fields[0] == "listingIds" {
			idx, err := strconv.Atoi(patch.Path.Fields[1])
			if err != nil {
				return fmt.Errorf("invalid listing index: %w", err)
			}
			if idx < 0 || idx >= len(tag.ListingIds) {
				return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
			}
			tag.ListingIds = slices.Delete(tag.ListingIds, idx, idx+1)
			return p.shop.Tags.Insert(tagName, tag)
		}

		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[0])

	case ReplaceOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
		}

		if len(patch.Path.Fields) == 0 {
			var newTag Tag
			if err := Unmarshal(patch.Value, &newTag); err != nil {
				return fmt.Errorf("failed to unmarshal tag: %w", err)
			}
			if err := p.validator.Struct(newTag); err != nil {
				return err
			}
			// Verify all referenced listings exist
			for _, listingId := range newTag.ListingIds {
				if _, exists := p.shop.Listings.Get(listingId); !exists {
					return fmt.Errorf("listing %d referenced by tag does not exist", listingId)
				}
			}
			newTag.Name = tagName
			return p.shop.Tags.Insert(tagName, newTag)
		}

		switch patch.Path.Fields[0] {
		case "name":
			var newName string
			if err := Unmarshal(patch.Value, &newName); err != nil {
				return fmt.Errorf("failed to unmarshal tag name: %w", err)
			}
			tag.Name = newName
			return p.shop.Tags.Insert(tagName, tag)

		case "listingIds":
			if len(patch.Path.Fields) != 2 {
				return fmt.Errorf("invalid listingIds path")
			}
			idx, err := strconv.Atoi(patch.Path.Fields[1])
			if err != nil {
				return fmt.Errorf("invalid listing index: %w", err)
			}
			if idx < 0 || idx >= len(tag.ListingIds) {
				return ObjectNotFoundError{ObjectType: ObjectTypeTag, Path: patch.Path}
			}
			var listingId ObjectId
			if err := Unmarshal(patch.Value, &listingId); err != nil {
				return fmt.Errorf("failed to unmarshal listing ID: %w", err)
			}
			if _, exists := p.shop.Listings.Get(listingId); !exists {
				return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: PatchPath{ObjectID: &listingId}}
			}
			tag.ListingIds[idx] = listingId
			return p.shop.Tags.Insert(tagName, tag)
		}

		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[0])
	}

	return fmt.Errorf("unsupported operation: %s", patch.Op)
}

// Helper for adding listings to tags with referential checks
func (p *Patcher) addListingToTag(tag *Tag, patch Patch) error {
	if len(patch.Path.Fields) != 2 {
		return fmt.Errorf("invalid listingIds path")
	}

	var listingId ObjectId
	if err := Unmarshal(patch.Value, &listingId); err != nil {
		return fmt.Errorf("failed to unmarshal listing ID: %w", err)
	}

	// Check if listing exists before adding reference
	if _, exists := p.shop.Listings.Get(listingId); !exists {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: PatchPath{ObjectID: &listingId}}
	}

	// Handle append vs insert at index
	if patch.Path.Fields[1] == "-" {
		// Append to end
		tag.ListingIds = append(tag.ListingIds, listingId)
		return nil
	}

	// Try to parse numerical index
	idx, err := strconv.Atoi(patch.Path.Fields[1])
	if err != nil {
		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[1])
	}

	// Validate index bounds
	if idx < 0 || idx > len(tag.ListingIds) {
		return fmt.Errorf("index out of bounds: %d", idx)
	}

	// Insert at index by growing slice and shifting elements
	tag.ListingIds = append(tag.ListingIds, 0)
	copy(tag.ListingIds[idx+1:], tag.ListingIds[idx:])
	tag.ListingIds[idx] = listingId
	return nil
}

func (p *Patcher) patchManifest(patch Patch) error {
	if patch.Path.Type != ObjectTypeManifest {
		return fmt.Errorf("invalid path type: %s", patch.Path.Type)
	}

	switch patch.Op {
	case ReplaceOp:
		return p.replaceManifestField(patch)
	case AddOp:
		return p.addManifestField(patch)
	case RemoveOp:
		return p.removeManifestField(patch)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
}

func (p *Patcher) replaceManifestField(patch Patch) error {
	if len(patch.Path.Fields) == 0 {
		// Replace entire manifest
		var newManifest Manifest
		if err := Unmarshal(patch.Value, &newManifest); err != nil {
			return fmt.Errorf("failed to unmarshal manifest: %w", err)
		}
		if err := p.validator.Struct(newManifest); err != nil {
			return err
		}
		p.shop.Manifest = newManifest
		return nil
	}

	switch patch.Path.Fields[0] {
	case "shopId":
		var value Uint256
		if err := Unmarshal(patch.Value, &value); err != nil {
			return fmt.Errorf("failed to unmarshal shopId: %w", err)
		}
		p.shop.Manifest.ShopID = value

	case "pricingCurrency":
		var c ChainAddress
		if err := Unmarshal(patch.Value, &c); err != nil {
			return fmt.Errorf("failed to unmarshal pricingCurrency: %w", err)
		}
		p.shop.Manifest.PricingCurrency = c

	case "payees":
		// Replace entire payees map or a single payee
		if len(patch.Path.Fields) == 1 {
			var payees Payees
			if err := Unmarshal(patch.Value, &payees); err != nil {
				return fmt.Errorf("failed to unmarshal payees: %w", err)
			}
			p.shop.Manifest.Payees = payees
		} else if len(patch.Path.Fields) == 2 {
			payeeName := patch.Path.Fields[1]
			var payee Payee
			if err := Unmarshal(patch.Value, &payee); err != nil {
				return fmt.Errorf("failed to unmarshal payee: %w", err)
			}
			if _, exists := p.shop.Manifest.Payees[payeeName]; !exists {
				return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
			}
			p.shop.Manifest.Payees[payeeName] = payee
		} else {
			return fmt.Errorf("invalid payees path")
		}

	case "shippingRegions":
		// Replace entire shippingRegions or just one region
		if len(patch.Path.Fields) == 1 {
			var regions ShippingRegions
			if err := Unmarshal(patch.Value, &regions); err != nil {
				return fmt.Errorf("failed to unmarshal shippingRegions: %w", err)
			}
			p.shop.Manifest.ShippingRegions = regions
		} else if len(patch.Path.Fields) == 2 {
			regionName := patch.Path.Fields[1]
			var region ShippingRegion
			if err := Unmarshal(patch.Value, &region); err != nil {
				return fmt.Errorf("failed to unmarshal shippingRegion: %w", err)
			}
			if _, exists := p.shop.Manifest.ShippingRegions[regionName]; !exists {
				return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
			}
			p.shop.Manifest.ShippingRegions[regionName] = region
		} else {
			return fmt.Errorf("invalid shippingRegions path")
		}

	case "acceptedCurrencies":
		// Replace entire acceptedCurrencies or a single index
		if len(patch.Path.Fields) == 1 {
			var currencies ChainAddresses
			if err := Unmarshal(patch.Value, &currencies); err != nil {
				return fmt.Errorf("failed to unmarshal acceptedCurrencies: %w", err)
			}
			p.shop.Manifest.AcceptedCurrencies = currencies
		} else if len(patch.Path.Fields) == 2 {
			i, err := strconv.Atoi(patch.Path.Fields[1])
			if err != nil {
				return fmt.Errorf("invalid acceptedCurrencies index: %w", err)
			}
			if i < 0 || i >= len(p.shop.Manifest.AcceptedCurrencies) {
				return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
			}
			var currency ChainAddress
			if err := Unmarshal(patch.Value, &currency); err != nil {
				return fmt.Errorf("failed to unmarshal accepted currency: %w", err)
			}
			p.shop.Manifest.AcceptedCurrencies[i] = currency
		} else {
			return fmt.Errorf("invalid acceptedCurrencies path")
		}

	default:
		return fmt.Errorf("unsupported field: %s", patch.Path.Fields[0])
	}

	return p.validator.Struct(p.shop.Manifest)
}

func (p *Patcher) addManifestField(patch Patch) error {
	if len(patch.Path.Fields) == 0 {
		return fmt.Errorf("cannot add to root manifest")
	}

	switch patch.Path.Fields[0] {
	case "payees":
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid payees path")
		}
		payeeName := patch.Path.Fields[1]
		if _, exists := p.shop.Manifest.Payees[payeeName]; exists {
			return fmt.Errorf("payee %s already exists", payeeName)
		}
		var payee Payee
		if err := Unmarshal(patch.Value, &payee); err != nil {
			return fmt.Errorf("failed to unmarshal payee: %w", err)
		}
		p.shop.Manifest.Payees[payeeName] = payee
	case "shippingRegions":
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid shippingRegions path")
		}
		regionName := patch.Path.Fields[1]
		if _, exists := p.shop.Manifest.ShippingRegions[regionName]; exists {
			return fmt.Errorf("shipping region %s already exists", regionName)
		}
		var region ShippingRegion
		if err := Unmarshal(patch.Value, &region); err != nil {
			return fmt.Errorf("failed to unmarshal shipping region: %w", err)
		}
		p.shop.Manifest.ShippingRegions[regionName] = region
	case "acceptedCurrencies":
		// Handle array insert/append for acceptedCurrencies
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid acceptedCurrencies path")
		}
		index := patch.Path.Fields[1]

		var newCurrency ChainAddress
		if err := Unmarshal(patch.Value, &newCurrency); err != nil {
			return fmt.Errorf("failed to unmarshal new currency: %w", err)
		}
		// If index == "-", append to the end
		if index == "-" {
			p.shop.Manifest.AcceptedCurrencies = append(p.shop.Manifest.AcceptedCurrencies, newCurrency)
		} else {
			// Otherwise, parse insertion index
			i, err := strconv.Atoi(index)
			if err != nil {
				return fmt.Errorf("invalid acceptedCurrencies index: %w", err)
			}
			if i < 0 || i > len(p.shop.Manifest.AcceptedCurrencies) {
				return fmt.Errorf("index out of bounds: %d", i)
			}
			// Insert at position i
			ac := p.shop.Manifest.AcceptedCurrencies
			ac = append(ac[:i], append([]ChainAddress{newCurrency}, ac[i:]...)...)
			p.shop.Manifest.AcceptedCurrencies = ac
		}
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
	}

	return p.validator.Struct(p.shop.Manifest)
}

func (p *Patcher) removeManifestField(patch Patch) error {
	if len(patch.Path.Fields) == 0 {
		return fmt.Errorf("cannot remove root manifest")
	}

	switch patch.Path.Fields[0] {
	case "payees":
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid payees path")
		}
		payeeName := patch.Path.Fields[1]
		if p.shop.Manifest.Payees == nil {
			return fmt.Errorf("payees map not initialized")
		}
		if _, exists := p.shop.Manifest.Payees[payeeName]; !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
		}
		delete(p.shop.Manifest.Payees, payeeName)

	case "shippingRegions":
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid shippingRegions path")
		}
		regionName := patch.Path.Fields[1]
		if p.shop.Manifest.ShippingRegions == nil {
			return fmt.Errorf("shippingRegions map not initialized")
		}
		if _, exists := p.shop.Manifest.ShippingRegions[regionName]; !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
		}
		delete(p.shop.Manifest.ShippingRegions, regionName)

	case "acceptedCurrencies":
		// Handle array removal from acceptedCurrencies
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid acceptedCurrencies path")
		}
		i, err := strconv.Atoi(patch.Path.Fields[1])
		if err != nil {
			return fmt.Errorf("invalid acceptedCurrencies index: %w", err)
		}
		if i < 0 || i >= len(p.shop.Manifest.AcceptedCurrencies) {
			return fmt.Errorf("index out of bounds: %d", i)
		}
		ac := p.shop.Manifest.AcceptedCurrencies
		p.shop.Manifest.AcceptedCurrencies = append(ac[:i], ac[i+1:]...)

	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeManifest, Path: patch.Path}
	}

	return p.validator.Struct(p.shop.Manifest)
}

func (p *Patcher) patchAccount(patch Patch) error {
	if patch.Path.AccountAddr == nil {
		return fmt.Errorf("account patch needs an address")
	}

	acc, exists := p.shop.Accounts.Get(patch.Path.AccountAddr[:])

	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if exists {
				return fmt.Errorf("account already exists")
			}
			var newAcc Account
			if err := Unmarshal(patch.Value, &newAcc); err != nil {
				return fmt.Errorf("failed to unmarshal account: %w", err)
			}
			if err := p.validator.Struct(newAcc); err != nil {
				return err
			}
			return p.shop.Accounts.Insert(patch.Path.AccountAddr[:], newAcc)
		}
		return fmt.Errorf("add operation not supported for account fields")

	case RemoveOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeAccount, Path: patch.Path}
		}

		if len(patch.Path.Fields) == 0 {
			return p.shop.Accounts.Delete(patch.Path.AccountAddr[:])
		}

		if len(patch.Path.Fields) != 2 || patch.Path.Fields[0] != "keyCards" {
			return fmt.Errorf("can only remove from keyCards array")
		}

		i, err := strconv.Atoi(patch.Path.Fields[1])
		if err != nil {
			return fmt.Errorf("invalid keyCards index: %w", err)
		}
		if i < 0 || i >= len(acc.KeyCards) {
			return fmt.Errorf("index out of bounds: %d", i)
		}

		acc.KeyCards = append(acc.KeyCards[:i], acc.KeyCards[i+1:]...)
		return p.shop.Accounts.Insert(patch.Path.AccountAddr[:], acc)

	case ReplaceOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeAccount, Path: patch.Path}
		}
		if len(patch.Path.Fields) == 0 {
			var newAcc Account
			if err := Unmarshal(patch.Value, &newAcc); err != nil {
				return fmt.Errorf("failed to unmarshal account: %w", err)
			}
			if err := p.validator.Struct(newAcc); err != nil {
				return err
			}
			return p.shop.Accounts.Insert(patch.Path.AccountAddr[:], newAcc)
		}
		return fmt.Errorf("replace operation not supported for account fields")

	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}

}

func (p *Patcher) patchListing(patch Patch) error {
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("listing patch needs an ID")
	}
	objId := *patch.Path.ObjectID
	listing, exists := p.shop.Listings.Get(objId)

	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if exists {
				return fmt.Errorf("listing %d already exists", objId)
			}
			var newListing Listing
			if err := Unmarshal(patch.Value, &newListing); err != nil {
				return fmt.Errorf("failed to unmarshal listing: %w", err)
			}
			if err := p.validator.Struct(newListing); err != nil {
				return err
			}
			return p.shop.Listings.Insert(objId, newListing)
		}

		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}

		if err := p.addListingField(&listing, patch); err != nil {
			return err
		}
		return p.shop.Listings.Insert(objId, listing)

	case ReplaceOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		if len(patch.Path.Fields) == 0 {
			var newListing Listing
			if err := Unmarshal(patch.Value, &newListing); err != nil {
				return fmt.Errorf("failed to unmarshal listing: %w", err)
			}
			if err := p.validator.Struct(newListing); err != nil {
				return err
			}
			return p.shop.Listings.Insert(objId, newListing)
		}
		if err := p.replaceListingField(&listing, patch); err != nil {
			return err
		}
		return p.shop.Listings.Insert(objId, listing)

	case RemoveOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		if len(patch.Path.Fields) == 0 {
			referenced := false
			tagNames := []string{}
			p.shop.Tags.All(func(key []byte, tag Tag) bool {
				if slices.Contains(tag.ListingIds, objId) {
					referenced = true
					tagNames = append(tagNames, string(key))
				}
				return true
			})
			if referenced {
				return fmt.Errorf("listing %d is referenced by tags: %v", objId, tagNames)
			}
			return p.shop.Listings.Delete(objId)
		}
		if err := p.removeListingField(&listing, patch); err != nil {
			return err
		}
		return p.shop.Listings.Insert(objId, listing)
	}

	return fmt.Errorf("unsupported operation: %s", patch.Op)
}

func (p *Patcher) addListingField(listing *Listing, patch Patch) error {
	switch patch.Path.Fields[0] {
	case "metadata":
		return p.addListingMetadata(listing, patch)
	case "stockStatuses":
		return p.addListingStockStatus(listing, patch)
	case "options":
		return p.addListingOption(listing, patch)
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
}

func (p *Patcher) removeListingField(listing *Listing, patch Patch) error {
	switch patch.Path.Fields[0] {
	case "metadata":
		return p.removeListingMetadata(listing, patch)
	case "stockStatuses":
		return p.removeListingStockStatus(listing, patch)
	case "options":
		return p.removeListingOption(listing, patch)
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
}

func (p *Patcher) replaceListingField(listing *Listing, patch Patch) error {
	switch patch.Path.Fields[0] {
	case "metadata":
		return p.replaceListingMetadata(listing, patch)
	case "stockStatuses":
		return p.replaceListingStockStatuses(listing, patch)
	case "options":
		return p.replaceListingOptions(listing, patch)
	case "price":
		var newPrice big.Int
		if err := Unmarshal(patch.Value, &newPrice); err != nil {
			return fmt.Errorf("failed to unmarshal price: %w", err)
		}
		listing.Price = newPrice
	case "viewState":
		var v ListingViewState
		if err := Unmarshal(patch.Value, &v); err != nil {
			return fmt.Errorf("failed to unmarshal viewState: %w", err)
		}
		listing.ViewState = v
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
	return nil
}

func (p *Patcher) addListingMetadata(listing *Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid metadata path")
	}
	switch patch.Path.Fields[1] {
	case "images":
		if len(patch.Path.Fields) < 3 {
			return fmt.Errorf("invalid images path")
		}
		index := patch.Path.Fields[2]

		var newImage string
		if err := Unmarshal(patch.Value, &newImage); err != nil {
			return fmt.Errorf("failed to unmarshal image: %w", err)
		}

		if index == "-" {
			listing.Metadata.Images = append(listing.Metadata.Images, newImage)
		} else {
			i, err := strconv.Atoi(index)
			if err != nil {
				return fmt.Errorf("invalid images index: %w", err)
			}
			if i < 0 || i > len(listing.Metadata.Images) {
				return fmt.Errorf("index out of bounds: %d", i)
			}
			arr := listing.Metadata.Images
			arr = append(arr[:i], append([]string{newImage}, arr[i:]...)...)
			listing.Metadata.Images = arr
		}
	default:
		return fmt.Errorf("invalid metadata path")
	}
	return nil
}

func (p *Patcher) removeListingMetadata(listing *Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid metadata path")
	}
	switch patch.Path.Fields[1] {
	case "images":
		if len(patch.Path.Fields) != 3 {
			return fmt.Errorf("invalid images path")
		}
		i, err := strconv.Atoi(patch.Path.Fields[2])
		if err != nil {
			return fmt.Errorf("invalid images index: %w", err)
		}
		if i < 0 || i >= len(listing.Metadata.Images) {
			return fmt.Errorf("index out of bounds: %d", i)
		}
		arr := listing.Metadata.Images
		listing.Metadata.Images = append(arr[:i], arr[i+1:]...)
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
	return nil
}

func (p *Patcher) replaceListingMetadata(listing *Listing, patch Patch) error {
	if len(patch.Path.Fields) == 1 {
		var newMd ListingMetadata
		if err := Unmarshal(patch.Value, &newMd); err != nil {
			return fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		listing.Metadata = newMd
		return nil
	}
	switch patch.Path.Fields[1] {
	case "title":
		var val string
		if err := Unmarshal(patch.Value, &val); err != nil {
			return fmt.Errorf("failed to unmarshal title: %w", err)
		}
		listing.Metadata.Title = val
	case "description":
		var val string
		if err := Unmarshal(patch.Value, &val); err != nil {
			return fmt.Errorf("failed to unmarshal description: %w", err)
		}
		listing.Metadata.Description = val
	case "images":
		if len(patch.Path.Fields) == 2 {
			var images []string
			if err := Unmarshal(patch.Value, &images); err != nil {
				return fmt.Errorf("failed to unmarshal images: %w", err)
			}
			listing.Metadata.Images = images
			return nil
		}
		if len(patch.Path.Fields) == 3 {
			i, err := strconv.Atoi(patch.Path.Fields[2])
			if err != nil {
				return fmt.Errorf("invalid images index: %w", err)
			}
			if i < 0 || i >= len(listing.Metadata.Images) {
				return fmt.Errorf("index out of bounds: %d", i)
			}
			var val string
			if err := Unmarshal(patch.Value, &val); err != nil {
				return fmt.Errorf("failed to unmarshal image: %w", err)
			}
			listing.Metadata.Images[i] = val
			return nil
		}
		return fmt.Errorf("invalid images path")
	default:
		return fmt.Errorf("unsupported metadata field: %s", patch.Path.Fields[1])
	}
	return nil
}

func (p *Patcher) addListingStockStatus(listing *Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid stockStatuses path")
	}
	index := patch.Path.Fields[1]

	var newSS ListingStockStatus
	if err := Unmarshal(patch.Value, &newSS); err != nil {
		return fmt.Errorf("failed to unmarshal stock status: %w", err)
	}

	if index == "-" {
		listing.StockStatuses = append(listing.StockStatuses, newSS)
	} else {
		i, err := strconv.Atoi(index)
		if err != nil {
			return fmt.Errorf("invalid stockStatuses index: %w", err)
		}
		if i < 0 || i > len(listing.StockStatuses) {
			return fmt.Errorf("index out of bounds: %d", i)
		}
		slice := listing.StockStatuses
		slice = append(slice[:i], append([]ListingStockStatus{newSS}, slice[i:]...)...)
		listing.StockStatuses = slice
	}
	return nil
}

func (p *Patcher) removeListingStockStatus(listing *Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid stockStatuses path")
	}
	i, err := strconv.Atoi(patch.Path.Fields[1])
	if err != nil {
		return fmt.Errorf("invalid stockStatuses index: %w", err)
	}
	if i < 0 || i >= len(listing.StockStatuses) {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
	slice := listing.StockStatuses
	listing.StockStatuses = append(slice[:i], slice[i+1:]...)
	return nil
}

func (p *Patcher) replaceListingStockStatuses(listing *Listing, patch Patch) error {
	if len(patch.Path.Fields) == 1 {
		var statuses []ListingStockStatus
		if err := Unmarshal(patch.Value, &statuses); err != nil {
			return fmt.Errorf("failed to unmarshal stock statuses: %w", err)
		}
		listing.StockStatuses = statuses
		return nil
	}
	i, err := strconv.Atoi(patch.Path.Fields[1])
	if err != nil {
		return fmt.Errorf("invalid stockStatuses index: %w", err)
	}
	if i < 0 || i >= len(listing.StockStatuses) {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
	if len(patch.Path.Fields) == 2 {
		var ss ListingStockStatus
		if err := Unmarshal(patch.Value, &ss); err != nil {
			return fmt.Errorf("failed to unmarshal stock status: %w", err)
		}
		listing.StockStatuses[i] = ss
		return nil
	}
	switch patch.Path.Fields[2] {
	case "inStock":
		var val bool
		if err := Unmarshal(patch.Value, &val); err != nil {
			return fmt.Errorf("failed to unmarshal inStock: %w", err)
		}
		listing.StockStatuses[i].InStock = &val
		listing.StockStatuses[i].ExpectedInStockBy = nil
	case "expectedInStockBy":
		var t time.Time
		if err := Unmarshal(patch.Value, &t); err != nil {
			return fmt.Errorf("failed to unmarshal expectedInStockBy: %w", err)
		}
		listing.StockStatuses[i].ExpectedInStockBy = &t
		listing.StockStatuses[i].InStock = nil
	default:
		return fmt.Errorf("unsupported stockStatus field: %s", patch.Path.Fields[2])
	}

	return nil
}

func (p *Patcher) addListingOption(listing *Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid options path")
	}
	optionName := patch.Path.Fields[1]
	if len(patch.Path.Fields) == 2 {
		if _, exists := listing.Options[optionName]; exists {
			return fmt.Errorf("option '%s' already exists", optionName)
		}
		var opt ListingOption
		if err := Unmarshal(patch.Value, &opt); err != nil {
			return fmt.Errorf("failed to unmarshal option: %w", err)
		}
		// Check if any variation name in the new option is already used in existing options
		for newVarName := range opt.Variations {
			for existingOptName, existingOpt := range listing.Options {
				if _, ok := existingOpt.Variations[newVarName]; ok {
					return fmt.Errorf("variation name '%q' already exists under option '%q'", newVarName, existingOptName)
				}
			}
		}
		if listing.Options == nil {
			listing.Options = make(map[string]ListingOption)
		}
		listing.Options[optionName] = opt
		return nil
	} else if len(patch.Path.Fields) == 4 && patch.Path.Fields[2] == "variations" {
		varName := patch.Path.Fields[3]
		opt, exists := listing.Options[optionName]
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		if opt.Variations == nil {
			opt.Variations = make(map[string]ListingVariation)
		}
		// Check if variation ID exists under any option
		for otherOptName, existingOpt := range listing.Options {
			if _, ok := existingOpt.Variations[varName]; ok {
				return fmt.Errorf("variation name '%q' already exists under option '%q'", varName, otherOptName)
			}
		}
		var v ListingVariation
		if err := Unmarshal(patch.Value, &v); err != nil {
			return fmt.Errorf("failed to unmarshal variation: %w", err)
		}
		opt.Variations[varName] = v
		listing.Options[optionName] = opt
		return nil
	}
	return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
}

func (p *Patcher) removeListingOption(listing *Listing, patch Patch) error {
	if len(patch.Path.Fields) < 2 {
		return fmt.Errorf("invalid options path")
	}
	optionName := patch.Path.Fields[1]
	opt, exists := listing.Options[optionName]
	if !exists {
		return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
	}
	if len(patch.Path.Fields) == 2 {
		delete(listing.Options, optionName)
		return nil
	}
	if len(patch.Path.Fields) == 4 && patch.Path.Fields[2] == "variations" {
		varID := patch.Path.Fields[3]
		if _, ok := opt.Variations[varID]; !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		delete(opt.Variations, varID)
		listing.Options[optionName] = opt
		return nil
	}
	return fmt.Errorf("invalid variation path")
}

func (p *Patcher) replaceListingOptions(listing *Listing, patch Patch) error {
	if len(patch.Path.Fields) == 1 {
		var newOptions ListingOptions
		if err := Unmarshal(patch.Value, &newOptions); err != nil {
			return fmt.Errorf("failed to unmarshal options: %w", err)
		}
		listing.Options = newOptions
		return nil
	}
	optionName := patch.Path.Fields[1]
	if len(patch.Path.Fields) == 2 {
		var newOpt ListingOption
		if err := Unmarshal(patch.Value, &newOpt); err != nil {
			return fmt.Errorf("failed to unmarshal listing option: %w", err)
		}
		listing.Options[optionName] = newOpt
		return nil
	}
	if len(patch.Path.Fields) == 3 && patch.Path.Fields[2] == "title" {
		var val string
		if err := Unmarshal(patch.Value, &val); err != nil {
			return fmt.Errorf("failed to unmarshal option title: %w", err)
		}
		opt := listing.Options[optionName]
		opt.Title = val
		listing.Options[optionName] = opt
		return nil
	}
	if len(patch.Path.Fields) == 4 && patch.Path.Fields[2] == "variations" {
		varID := patch.Path.Fields[3]
		opt, ok := listing.Options[optionName]
		if !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		var newVar ListingVariation
		if err := Unmarshal(patch.Value, &newVar); err != nil {
			return fmt.Errorf("failed to unmarshal listing variation: %w", err)
		}
		if _, has := opt.Variations[varID]; !has {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		opt.Variations[varID] = newVar
		listing.Options[optionName] = opt
		return nil
	}
	if len(patch.Path.Fields) == 5 &&
		patch.Path.Fields[2] == "variations" &&
		patch.Path.Fields[4] == "variationInfo" {
		varID := patch.Path.Fields[3]
		opt, ok := listing.Options[optionName]
		if !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		v, ok := opt.Variations[varID]
		if !ok {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: patch.Path}
		}
		var newInfo ListingMetadata
		if err := Unmarshal(patch.Value, &newInfo); err != nil {
			return fmt.Errorf("failed to unmarshal listing variation info: %w", err)
		}
		v.VariationInfo = newInfo
		opt.Variations[varID] = v
		listing.Options[optionName] = opt
		return nil
	}
	return fmt.Errorf("invalid variation path")
}

func (p *Patcher) patchOrder(patch Patch) error {
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("order patch needs an ID")
	}
	objId := *patch.Path.ObjectID
	order, exists := p.shop.Orders.Get(objId)

	switch patch.Op {
	case AddOp:
		if len(patch.Path.Fields) == 0 {
			if exists {
				return fmt.Errorf("order %d already exists", objId)
			}
			var newOrder Order
			if err := Unmarshal(patch.Value, &newOrder); err != nil {
				return fmt.Errorf("failed to unmarshal order: %w", err)
			}
			if err := p.validateOrderReferences(&newOrder); err != nil {
				return err
			}
			if err := p.validator.Struct(newOrder); err != nil {
				return err
			}
			return p.shop.Orders.Insert(objId, newOrder)
		}

		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		return p.addOrderField(&order, patch)

	case ReplaceOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		if len(patch.Path.Fields) == 0 {
			var newOrder Order
			if err := Unmarshal(patch.Value, &newOrder); err != nil {
				return fmt.Errorf("failed to unmarshal order: %w", err)
			}
			if err := p.validateOrderReferences(&newOrder); err != nil {
				return err
			}
			if err := p.validator.Struct(newOrder); err != nil {
				return err
			}
			return p.shop.Orders.Insert(objId, newOrder)
		}
		return p.replaceOrderField(&order, patch)

	case RemoveOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		if len(patch.Path.Fields) == 0 {
			return p.shop.Orders.Delete(objId)
		}
		return p.removeOrderField(&order, patch)

	case IncrementOp, DecrementOp:
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}
		return p.modifyOrderQuantity(&order, patch)
	}

	return fmt.Errorf("unsupported operation: %s", patch.Op)
}

func (p *Patcher) validateOrderReferences(order *Order) error {
	for _, item := range order.Items {
		listing, exists := p.shop.Listings.Get(item.ListingID)
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: PatchPath{ObjectID: &item.ListingID}}
		}

		for _, varID := range item.VariationIDs {
			found := false
			for _, opt := range listing.Options {
				if _, exists := opt.Variations[varID]; exists {
					found = true
					break
				}
			}
			if !found {
				return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: PatchPath{ObjectID: &item.ListingID, Fields: []string{"options", varID}}}
			}
		}
	}

	if order.ChosenPayee != nil {
		found := false
		for _, payee := range p.shop.Manifest.Payees {
			if payee == *order.ChosenPayee {
				found = true
				break
			}
		}
		if !found {
			return ObjectNotFoundError{
				ObjectType: ObjectTypeOrder,
				Path:       PatchPath{Fields: []string{"chosenPayee", order.ChosenPayee.Address.String()}}}
		}
	}

	if order.ChosenCurrency != nil {
		found := false
		for _, currency := range p.shop.Manifest.AcceptedCurrencies {
			if currency == *order.ChosenCurrency {
				found = true
				break
			}
		}
		if !found {
			return ObjectNotFoundError{
				ObjectType: ObjectTypeManifest,
				Path:       PatchPath{Fields: []string{"acceptedCurrencies", order.ChosenCurrency.String()}}}
		}
	}

	return nil
}

func (p *Patcher) addOrderField(order *Order, patch Patch) error {
	if len(patch.Path.Fields) == 0 {
		return fmt.Errorf("field path required for add operation")
	}

	switch patch.Path.Fields[0] {
	case "items":
		var item OrderedItem
		if err := Unmarshal(patch.Value, &item); err != nil {
			return fmt.Errorf("failed to unmarshal order item: %w", err)
		}
		listing, exists := p.shop.Listings.Get(item.ListingID)
		if !exists {
			return ObjectNotFoundError{ObjectType: ObjectTypeListing, Path: PatchPath{ObjectID: &item.ListingID}}
		}
		for _, varID := range item.VariationIDs {
			found := false
			for _, opt := range listing.Options {
				if _, exists := opt.Variations[varID]; exists {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("variation %s not found in listing %d", varID, item.ListingID)
			}
		}
		order.Items = append(order.Items, item)
	case "shippingAddress":
		if order.ShippingAddress != nil {
			return fmt.Errorf("shipping address already set")
		}
		var shippingAddress AddressDetails
		if err := Unmarshal(patch.Value, &shippingAddress); err != nil {
			return fmt.Errorf("failed to unmarshal shipping address: %w", err)
		}
		order.ShippingAddress = &shippingAddress
	case "invoiceAddress":
		if order.InvoiceAddress != nil {
			return fmt.Errorf("invoice address already set")
		}
		var invoiceAddress AddressDetails
		if err := Unmarshal(patch.Value, &invoiceAddress); err != nil {
			return fmt.Errorf("failed to unmarshal invoice address: %w", err)
		}
		order.InvoiceAddress = &invoiceAddress
	case "paymentDetails":
		if order.PaymentDetails != nil {
			return fmt.Errorf("payment details already set")
		}
		var paymentDetails PaymentDetails
		if err := Unmarshal(patch.Value, &paymentDetails); err != nil {
			return fmt.Errorf("failed to unmarshal payment details: %w", err)
		}
		order.PaymentDetails = &paymentDetails
	case "chosenPayee":
		if order.ChosenPayee != nil {
			return fmt.Errorf("chosen payee already set")
		}
		var payee Payee
		if err := Unmarshal(patch.Value, &payee); err != nil {
			return fmt.Errorf("failed to unmarshal payee: %w", err)
		}
		order.ChosenPayee = &payee
	case "chosenCurrency":
		if order.ChosenCurrency != nil {
			return fmt.Errorf("chosen currency already set")
		}
		var currency ChainAddress
		if err := Unmarshal(patch.Value, &currency); err != nil {
			return fmt.Errorf("failed to unmarshal currency: %w", err)
		}
		order.ChosenCurrency = &currency
	case "txDetails":
		if order.TxDetails != nil {
			return fmt.Errorf("tx details already set")
		}
		var txDetails OrderPaid
		if err := Unmarshal(patch.Value, &txDetails); err != nil {
			return fmt.Errorf("failed to unmarshal tx details: %w", err)
		}
		order.TxDetails = &txDetails
	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
	}
	if err := p.validator.Struct(order); err != nil {
		return err
	}
	return p.shop.Orders.Insert(*patch.Path.ObjectID, *order)
}

func (p *Patcher) replaceOrderField(order *Order, patch Patch) error {
	nFields := len(patch.Path.Fields)

	switch patch.Path.Fields[0] {
	case "state":
		var state OrderState
		if err := Unmarshal(patch.Value, &state); err != nil {
			return fmt.Errorf("failed to unmarshal order state: %w", err)
		}
		order.State = state

	case "chosenPayee":
		var payee Payee
		if err := Unmarshal(patch.Value, &payee); err != nil {
			return fmt.Errorf("failed to unmarshal payee: %w", err)
		}
		found := false
		for _, p := range p.shop.Manifest.Payees {
			if p == payee {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("payee not found in manifest payees")
		}
		order.ChosenPayee = &payee

	case "chosenCurrency":
		var currency ChainAddress
		if err := Unmarshal(patch.Value, &currency); err != nil {
			return fmt.Errorf("failed to unmarshal currency: %w", err)
		}
		found := false
		for _, c := range p.shop.Manifest.AcceptedCurrencies {
			if c == currency {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("currency not found in accepted currencies")
		}
		order.ChosenCurrency = &currency

	case "paymentDetails":
		var details PaymentDetails
		if err := Unmarshal(patch.Value, &details); err != nil {
			return fmt.Errorf("failed to unmarshal payment details: %w", err)
		}
		order.PaymentDetails = &details

	case "txDetails":
		var details OrderPaid
		if err := Unmarshal(patch.Value, &details); err != nil {
			return fmt.Errorf("failed to unmarshal tx details: %w", err)
		}
		order.TxDetails = &details

	case "items":

		switch {
		case nFields == 1:
			// Replace entire items array
			var items []OrderedItem
			if err := Unmarshal(patch.Value, &items); err != nil {
				return fmt.Errorf("failed to unmarshal items: %w", err)
			}
			order.Items = items

		case nFields >= 2:
			// Get the item index
			index, err := strconv.Atoi(patch.Path.Fields[1])
			if err != nil {
				return fmt.Errorf("failed to convert index to int: %w", err)
			}
			if index < 0 || index >= len(order.Items) {
				return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
			}

			if nFields == 2 {
				// Replace entire item at index
				var item OrderedItem
				if err := Unmarshal(patch.Value, &item); err != nil {
					return fmt.Errorf("failed to unmarshal item: %w", err)
				}
				order.Items[index] = item
			} else if patch.Path.Fields[2] == "quantity" {
				// Replace just the quantity
				var quantity uint32
				if err := Unmarshal(patch.Value, &quantity); err != nil {
					return fmt.Errorf("failed to unmarshal quantity: %w", err)
				}
				order.Items[index].Quantity = quantity
			} else {
				return fmt.Errorf("unsupported field: %s", patch.Path.Fields[2])
			}
		default:
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}

	case "invoiceAddress":
		if order.InvoiceAddress == nil {
			return fmt.Errorf("invoice address not set")
		}

		switch {
		case nFields == 1:
			var newAddress AddressDetails
			if err := Unmarshal(patch.Value, &newAddress); err != nil {
				return fmt.Errorf("failed to unmarshal invoice address: %w", err)
			}
			order.InvoiceAddress = &newAddress
		case nFields == 2:
			switch patch.Path.Fields[1] {
			case "name":
				var newName string
				if err := Unmarshal(patch.Value, &newName); err != nil {
					return fmt.Errorf("failed to unmarshal invoice address name: %w", err)
				}
				order.InvoiceAddress.Name = newName
			default:
				return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
			}
		default:
			return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
		}

	default:
		return ObjectNotFoundError{ObjectType: ObjectTypeOrder, Path: patch.Path}
	}

	if err := p.validator.Struct(order); err != nil {
		return err
	}
	return p.shop.Orders.Insert(*patch.Path.ObjectID, *order)
}

func (p *Patcher) removeOrderField(order *Order, patch Patch) error {
	switch patch.Path.Fields[0] {
	case "items":
		if len(patch.Path.Fields) != 2 {
			return fmt.Errorf("invalid items path")
		}
		index, err := strconv.Atoi(patch.Path.Fields[1])
		if err != nil {
			return fmt.Errorf("invalid item index: %w", err)
		}
		if index < 0 || index >= len(order.Items) {
			return fmt.Errorf("item index out of bounds: %d", index)
		}

		order.Items = slices.Delete(order.Items, index, index+1)

	case "shippingAddress":
		if order.ShippingAddress == nil {
			return fmt.Errorf("shipping address not set")
		}
		order.ShippingAddress = nil

	case "invoiceAddress":
		if order.InvoiceAddress == nil {
			return fmt.Errorf("invoice address not set")
		}
		order.InvoiceAddress = nil

	default:
		return fmt.Errorf("cannot remove field: %s", patch.Path.Fields[0])
	}

	if err := p.validator.Struct(order); err != nil {
		return err
	}
	return p.shop.Orders.Insert(*patch.Path.ObjectID, *order)
}

func (p *Patcher) modifyOrderQuantity(order *Order, patch Patch) error {
	index, err := order.checkPathAndIndex(patch.Path.Fields)
	if err != nil {
		return err
	}

	var value uint32
	if err := Unmarshal(patch.Value, &value); err != nil {
		return fmt.Errorf("failed to unmarshal value: %w", err)
	}

	if patch.Op == IncrementOp {
		order.Items[index].Quantity += value
	} else {
		if value > order.Items[index].Quantity {
			return fmt.Errorf("cannot decrement below zero")
		}
		order.Items[index].Quantity -= value
	}

	if err := p.validator.Struct(order); err != nil {
		return err
	}
	return p.shop.Orders.Insert(*patch.Path.ObjectID, *order)
}

func (existing *Order) checkPathAndIndex(fields []string) (int, error) {
	if len(fields) != 3 || fields[0] != "items" || fields[2] != "quantity" {
		return 0, fmt.Errorf("incr/decr only works on path: [items, x, quantity]")
	}
	index, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, fmt.Errorf("failed to convert index to int: %w", err)
	}
	if index < 0 || index >= len(existing.Items) {
		return 0, fmt.Errorf("index out of bounds: %d", index)
	}
	return index, nil
}

func (p *Patcher) patchInventory(inventory *Inventory, patch Patch) error {

	// validate patch edits an existing listing
	if patch.Path.ObjectID == nil {
		return fmt.Errorf("inventory patch needs an id")
	}
	objId := *patch.Path.ObjectID
	lis, ok := p.shop.Listings.Get(objId)
	if !ok {
		return fmt.Errorf("listing %d not found", objId)
	}
	// if it is a variation, check that they exist
	if n := len(patch.Path.Fields); n > 0 {
		var found = uint(n)
		for _, field := range patch.Path.Fields {
			for _, opt := range lis.Options {
				_, has := opt.Variations[field]
				if has {
					found--
					break // found
				}
			}
		}
		if found > 0 {
			return fmt.Errorf("some variation of object %d not found", objId)
		}
	}

	if patch.Path.ObjectID == nil {
		return fmt.Errorf("inventory patch needs an ID")
	}
	var (
		err    error
		newVal uint64
	)
	if patch.Op != RemoveOp {
		err = cbor.Unmarshal(patch.Value, &newVal)
		if err != nil {
			return fmt.Errorf("failed to unmarshal inventory value: %w", err)
		}
	}

	current, ok := inventory.Get(objId, patch.Path.Fields)
	switch patch.Op {
	case AddOp:
		if ok {
			return fmt.Errorf("inventory %d already exists", objId)
		}
		err = inventory.Insert(objId, patch.Path.Fields, newVal)
	case RemoveOp:
		if !ok {
			return fmt.Errorf("inventory %d not found", objId)
		}
		err = inventory.Delete(objId, patch.Path.Fields)
	case ReplaceOp:
		if !ok {
			return fmt.Errorf("inventory %d not found", objId)
		}
		err = inventory.Insert(objId, patch.Path.Fields, newVal)
	case IncrementOp:
		current += newVal
		err = inventory.Insert(objId, patch.Path.Fields, current)
	case DecrementOp:
		if current < newVal {
			return fmt.Errorf("inventory %d cannot decrement below 0", objId)
		}
		current -= newVal
		err = inventory.Insert(objId, patch.Path.Fields, current)
	default:
		return fmt.Errorf("unsupported op: %s", patch.Op)
	}
	if err != nil {
		return fmt.Errorf("failed to patch(%s) inventory %d: %w", patch.Op, objId, err)
	}
	return nil
}
