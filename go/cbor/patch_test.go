// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fxamacker/cbor/v2"
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
		{Path: PatchPath{Type: ObjectTypeAccount, AccountAddr: &testAddr}},
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
		{Path: PatchPath{Type: ObjectTypeManifest, AccountAddr: &testAddr}},

		{Path: PatchPath{Type: ObjectTypeAccount, ObjectID: uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeAccount, TagName: strptr("test-tag")}},

		{Path: PatchPath{Type: ObjectTypeTag, ObjectID: uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeTag, AccountAddr: &testAddr}},

		{Path: PatchPath{Type: ObjectTypeOrder, TagName: strptr("test-tag")}},
	}
	var testPath TestPath
	for idx, tc := range badTestCases {
		t.Run(fmt.Sprintf("bad-%d", idx), func(t *testing.T) {
			r := require.New(t)
			data, err := Marshal(tc.Path)
			r.Error(err)
			r.Nil(data)

			testPath.PatchPath = tc.Path
			data, err = Marshal(testPath)
			t.Logf("cbor diag:\n%s", pretty(data))
			r.NoError(err)
			r.NotNil(data)

			var got PatchPath
			err = Unmarshal(data, &got)
			t.Logf("got: %+v", got)
			if tc.Path.Type != ObjectTypeManifest { // we cant tell these from paths elements
				r.Error(err, "decoding should fail")
			}
		})
	}
}

type TestPath struct {
	PatchPath
}

// copy of PatchPath.MarshalCBOR for testing, removed error checking
func (p TestPath) MarshalCBOR() ([]byte, error) {
	var path []any
	path = append(path, string(p.Type))
	if p.AccountAddr != nil {
		path = append(path, *p.AccountAddr)
	}
	if p.ObjectID != nil {
		path = append(path, *p.ObjectID)
	}
	if p.TagName != nil {
		path = append(path, *p.TagName)
	}
	for _, field := range p.Fields {
		path = append(path, field)
	}
	return cbor.Marshal(path)
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

	_, lis := newTestListing()
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
