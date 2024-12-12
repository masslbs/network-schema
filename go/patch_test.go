// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"fmt"
	"testing"

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
