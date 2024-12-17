// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatchObjectIDs(t *testing.T) {
	testAddr := addrFromHex(1, "0x1234567890123456789012345678901234567890").Address

	type testCase struct {
		Path PatchPath
	}

	goodTestCases := []testCase{
		{Path: PatchPath{Type: ObjectTypeManifest}},
		{Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeAccount, AccountID: &testAddr}},
		{Path: PatchPath{Type: ObjectTypeTag, TagName: strptr("test-tag")}},
	}
	for idx, tc := range goodTestCases {
		t.Run(fmt.Sprintf("good-%d", idx), func(t *testing.T) {
			r := require.New(t)
			data, err := Marshal(tc.Path)
			r.NoError(err)

			var got PatchPath
			err = Unmarshal(data, &got)
			r.NoError(err)
			// unset fields for convenience
			got.Fields = nil
			r.EqualValues(tc.Path, got)
		})
	}

	badTestCases := []testCase{
		// missing required field
		{Path: PatchPath{Type: ObjectTypeListing}},
		{Path: PatchPath{Type: ObjectTypeAccount}},
		{Path: PatchPath{Type: ObjectTypeTag}},

		// wrong contextual field
		{Path: PatchPath{Type: ObjectTypeManifest, ObjectID: uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeManifest, AccountID: &testAddr}},

		{Path: PatchPath{Type: ObjectTypeAccount, ObjectID: uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeAccount, TagName: strptr("test-tag")}},

		{Path: PatchPath{Type: ObjectTypeTag, ObjectID: uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeTag, AccountID: &testAddr}},

		{Path: PatchPath{Type: ObjectTypeOrder, TagName: strptr("test-tag")}},
	}
	for idx, tc := range badTestCases {
		t.Run(fmt.Sprintf("bad-%d", idx), func(t *testing.T) {
			r := require.New(t)
			data, err := Marshal(tc.Path)
			r.NoError(err)

			var got PatchPath
			err = Unmarshal(data, &got)
			r.Error(err)
		})
	}
}

func TestPatchPath(t *testing.T) {
	r := require.New(t)
	patch := Patch{
		Op:   OpString(t.Name()),
		Path: PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color", "variations", "pink"}},
	}

	data, err := Marshal(patch)
	r.NoError(err)
	t.Log("CBOR encoded:")
	t.Log("\n" + pretty(data))

	var rxPatch Patch
	err = Unmarshal(data, &rxPatch)
	r.NoError(err)
	r.Equal(ObjectTypeListing, rxPatch.Path.Type)
	r.EqualValues(ObjectId(1), *rxPatch.Path.ObjectID)
	r.Equal([]string{"options", "color", "variations", "pink"}, rxPatch.Path.Fields)

	var testPath struct {
		Path []any
	}
	err = Unmarshal(data, &testPath)
	r.NoError(err)
	r.Equal([]any{"listing", uint64(1), "options", "color", "variations", "pink"}, testPath.Path)
}

func TestPatchAdd(t *testing.T) {
	var (
		err error
		buf bytes.Buffer
		r   = require.New(t)
		enc = DefaultEncoder(&buf)
	)

	var createListing Patch
	createListing.Op = AddOp
	createListing.Path = PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1)}

	lis := testListing()
	createListing.Value, err = Marshal(lis)
	r.NoError(err)
	err = enc.Encode(createListing)
	r.NoError(err)

	opData := buf.Bytes()
	t.Log("OP encoded:")
	t.Log("\n" + pretty(opData))

	dec := DefaultDecoder(bytes.NewReader(opData))
	var rxOp Patch
	err = dec.Decode(&rxOp)
	r.NoError(err)
	r.Equal(ObjectTypeListing, rxOp.Path.Type)
	r.EqualValues(ObjectId(1), *rxOp.Path.ObjectID)
	r.NoError(validate.Struct(rxOp))

	dec = DefaultDecoder(bytes.NewReader(rxOp.Value))
	var rxLis Listing
	err = dec.Decode(&rxLis)
	r.NoError(err)

	t.Logf("listing received: %+v", rxLis)
	r.EqualValues(lis, rxLis)
}

func TestPatchListingErrors(t *testing.T) {
	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    interface{}
		errMatch string
	}{
		{
			name:     "invalid field path",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"invalid"}},
			value:    "test",
			errMatch: "unsupported field: invalid",
		},
		{
			name:     "remove non-existent metadata field",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"metadata", "nonexistent"}},
			errMatch: "unsupported field: nonexistent",
		},
		{
			name:     "invalid array index",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"stockStatuses", "999"}},
			value:    ListingStockStatus{},
			errMatch: "index out of bounds: 999",
		},
		{
			name:     "remove non-existent stock status",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"stockStatuses", "999"}},
			errMatch: "index out of bounds: 999",
		},
		{
			name:     "invalid value type for price",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"price"}},
			value:    "not a number",
			errMatch: "failed to unmarshal price:",
		},
		{
			name:     "invalid value type for viewState",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"viewState"}},
			value:    123,
			errMatch: "failed to unmarshal viewState:",
		},
		{
			name:     "remove non-existent option",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "nonexistent"}},
			errMatch: "option not found: nonexistent",
		},
		{
			name:     "replace non-existent variation on an option",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color", "variations", "nonexistent"}},
			errMatch: "variation not found: nonexistent",
		},
		{
			name:     "remove non-existent variation from an option",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeListing, ObjectID: uint64ptr(1), Fields: []string{"options", "color", "variations", "nonexistent"}},
			errMatch: "variation not found: nonexistent",
		},
	}

	var patcher Patcher
	patcher.validator = validate

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lis := testListing()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			err := patcher.Listing(&lis, decodedPatch)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMatch)
		})
	}
}

func TestPatchManifestErrors(t *testing.T) {

	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    any
		errMatch string
	}{
		{
			name:     "unsupported op",
			op:       IncrementOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees"}},
			errMatch: "unsupported op: increment",
		},
		{
			name:     "unsupported field",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"invalid"}},
			errMatch: "unsupported field: invalid",
		},
		{
			name:     "replace non-existent payee",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "nonexistent"}},
			value:    Payee{},
			errMatch: "payee not found: nonexistent",
		},
		{
			name:     "remove non-existent payee",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"payees", "nonexistent"}},
			errMatch: "payee not found: nonexistent",
		},
		{
			name:     "replace non-existent shipping region",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"shippingRegions", "nonexistent"}},
			value:    ShippingRegion{},
			errMatch: "shipping region not found: nonexistent",
		},
		{
			name:     "remove non-existent shipping region",
			op:       RemoveOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"shippingRegions", "nonexistent"}},
			errMatch: "shipping region not found: nonexistent",
		},
		{
			name:     "invalid index for acceptedCurrencies",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"acceptedCurrencies", "999"}},
			value:    ChainAddress{},
			errMatch: "index out of bounds: 999",
		},
		{
			name:     "invalid value type for pricingCurrency",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeManifest, Fields: []string{"pricingCurrency"}},
			value:    "not a chain address",
			errMatch: "failed to unmarshal currency:",
		},
	}

	var patcher Patcher
	patcher.validator = validate

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			r := require.New(t)

			manifest := testManifest()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			err := patcher.Manifest(&manifest, decodedPatch)
			r.Error(err)
			a.Contains(err.Error(), tc.errMatch)
		})
	}
}

func TestPatchOrderErrors(t *testing.T) {
	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    any
		errMatch string
	}{
		{
			name:     "unsupported op",
			op:       IncrementOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"items"}},
			errMatch: "unsupported op: increment",
		},
		{
			name:     "unsupported field",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"invalid"}},
			errMatch: "unsupported field: invalid",
		},
		{
			name:     "invalid item index",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"items", "999"}},
			value:    OrderedItem{},
			errMatch: "index out of bounds: 999",
		},
		{
			name:     "invalid item index format",
			op:       ReplaceOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"items", "abc"}},
			value:    OrderedItem{},
			errMatch: "failed to convert index to int",
		},
		{
			name:     "missing address field",
			op:       AddOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"invoiceAddress"}},
			errMatch: "Field validation for 'Name' failed on the 'required' tag",
		},
		{
			name:     "invalid address field",
			op:       AddOp,
			path:     PatchPath{Type: ObjectTypeOrder, ObjectID: uint64ptr(1), Fields: []string{"invoiceAddress", "invalid"}},
			value:    "test",
			errMatch: "unsupported field: invalid",
		},
	}

	var patcher Patcher
	patcher.validator = validate

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := assert.New(t)
			r := require.New(t)

			order := testOrder()

			patch := createPatch(t, tc.op, tc.path, tc.value)
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			err := patcher.Order(&order, decodedPatch)
			r.Error(err)
			a.Contains(err.Error(), tc.errMatch)
		})
	}
}
