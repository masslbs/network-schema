// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/masslbs/network-schema/go/internal/testhelper"
	"github.com/masslbs/network-schema/go/objects"
)

func TestPatchObjectIDs(t *testing.T) {
	testAddr := testMassEthAddr([20]byte{0xff})

	type testCase struct {
		Path PatchPath
	}

	goodTestCases := []testCase{
		{Path: PatchPath{Type: ObjectTypeManifest}},
		{Path: PatchPath{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeAccount, AccountAddr: &testAddr}},
		{Path: PatchPath{Type: ObjectTypeTag, TagName: testhelper.Strptr("test-tag")}},
		{Path: PatchPath{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(1)}},
	}
	for idx, tc := range goodTestCases {
		t.Run(fmt.Sprintf("good-%d", idx), func(t *testing.T) {
			r := require.New(t)
			data, err := masscbor.Marshal(tc.Path)
			r.NoError(err)

			var got PatchPath
			err = masscbor.Unmarshal(data, &got)
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
		{Path: PatchPath{Type: ObjectTypeManifest, ObjectID: testhelper.Uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeManifest, AccountAddr: &testAddr}},

		{Path: PatchPath{Type: ObjectTypeAccount, ObjectID: testhelper.Uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeAccount, TagName: testhelper.Strptr("test-tag")}},

		{Path: PatchPath{Type: ObjectTypeTag, ObjectID: testhelper.Uint64ptr(1)}},
		{Path: PatchPath{Type: ObjectTypeTag, AccountAddr: &testAddr}},

		{Path: PatchPath{Type: ObjectTypeOrder, TagName: testhelper.Strptr("test-tag")}},
	}
	var testPath TestPath
	for idx, tc := range badTestCases {
		t.Run(fmt.Sprintf("bad-%d", idx), func(t *testing.T) {
			r := require.New(t)
			data, err := masscbor.Marshal(tc.Path)
			r.Error(err)
			r.Nil(data)

			testPath.PatchPath = tc.Path
			data, err = masscbor.Marshal(testPath)
			// t.Logf("cbor diag:\n%s", pretty(data))
			r.NoError(err)
			r.NotNil(data)

			var got PatchPath
			err = masscbor.Unmarshal(data, &got)
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
		Path: PatchPath{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []string{"options", "color", "variations", "pink"}},
	}

	data, err := masscbor.Marshal(patch)
	r.NoError(err)
	// t.Log("CBOR encoded:")
	// t.Log("\n" + pretty(data))

	var rxPatch Patch
	err = masscbor.Unmarshal(data, &rxPatch)
	r.NoError(err)
	r.Equal(ObjectTypeListing, rxPatch.Path.Type)
	r.EqualValues(objects.ObjectId(1), *rxPatch.Path.ObjectID)
	r.Equal([]string{"options", "color", "variations", "pink"}, rxPatch.Path.Fields)

	var testPath struct {
		Path []any
	}
	err = masscbor.Unmarshal(data, &testPath)
	r.NoError(err)
	r.Equal([]any{"listing", uint64(1), "options", "color", "variations", "pink"}, testPath.Path)
}

func TestPatchAdd(t *testing.T) {
	var (
		err error
		buf bytes.Buffer
		r   = require.New(t)
		enc = masscbor.DefaultEncoder(&buf)
	)

	var createListing Patch
	createListing.Op = AddOp
	createListing.Path = PatchPath{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1)}

	_, lis := newTestListing()
	createListing.Value, err = masscbor.Marshal(lis)
	r.NoError(err)
	err = enc.Encode(createListing)
	r.NoError(err)

	opData := buf.Bytes()
	// t.Log("OP encoded:")
	// t.Log("\n" + pretty(opData))

	dec := masscbor.DefaultDecoder(bytes.NewReader(opData))
	var rxOp Patch
	err = dec.Decode(&rxOp)
	r.NoError(err)
	r.Equal(ObjectTypeListing, rxOp.Path.Type)
	r.EqualValues(objects.ObjectId(1), *rxOp.Path.ObjectID)
	r.NoError(validate.Struct(rxOp))

	dec = masscbor.DefaultDecoder(bytes.NewReader(rxOp.Value))
	var rxLis objects.Listing
	err = dec.Decode(&rxLis)
	r.NoError(err)

	t.Logf("listing received: %+v", rxLis)
	r.EqualValues(lis, rxLis)
}
