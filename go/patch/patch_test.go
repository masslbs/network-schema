// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/google/go-cmp/cmp/cmpopts"

	masscbor "github.com/masslbs/network-schema/go/cbor"
	"github.com/masslbs/network-schema/go/internal/testhelper"
	"github.com/masslbs/network-schema/go/objects"
	"github.com/peterldowns/testy/assert"
)

func TestPatchObjectIDs(t *testing.T) {
	testAddr := testMassEthAddr([20]byte{0xff})

	type testCase struct {
		Path Path
	}

	goodTestCases := []testCase{
		{Path: Path{Type: ObjectTypeManifest}},
		{Path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1)}},
		{Path: Path{Type: ObjectTypeAccount, AccountAddr: &testAddr}},
		{Path: Path{Type: ObjectTypeTag, TagName: testhelper.Strptr("test-tag")}},
		{Path: Path{Type: ObjectTypeInventory, ObjectID: testhelper.Uint64ptr(1)}},
	}
	for idx, tc := range goodTestCases {
		t.Run(fmt.Sprintf("good-%d", idx), func(t *testing.T) {
			data, err := masscbor.Marshal(tc.Path)
			assert.Nil(t, err)

			var got Path
			err = masscbor.Unmarshal(data, &got)
			assert.Nil(t, err)
			// unset fields for convenience
			got.Fields = nil
			assert.Equal(t, tc.Path, got)
		})
	}

	badTestCases := []testCase{
		// missing required field
		{Path: Path{Type: ObjectTypeListing}},
		{Path: Path{Type: ObjectTypeAccount}},
		{Path: Path{Type: ObjectTypeTag}},

		// wrong contextual field
		{Path: Path{Type: ObjectTypeManifest, ObjectID: testhelper.Uint64ptr(1)}},
		{Path: Path{Type: ObjectTypeManifest, AccountAddr: &testAddr}},

		{Path: Path{Type: ObjectTypeAccount, ObjectID: testhelper.Uint64ptr(1)}},
		{Path: Path{Type: ObjectTypeAccount, TagName: testhelper.Strptr("test-tag")}},

		{Path: Path{Type: ObjectTypeTag, ObjectID: testhelper.Uint64ptr(1)}},
		{Path: Path{Type: ObjectTypeTag, AccountAddr: &testAddr}},

		{Path: Path{Type: ObjectTypeOrder, TagName: testhelper.Strptr("test-tag")}},
	}
	var testPath TestPath
	for idx, tc := range badTestCases {
		t.Run(fmt.Sprintf("bad-%d", idx), func(t *testing.T) {
			data, err := masscbor.Marshal(tc.Path)
			assert.Error(t, err)
			assert.Equal(t, nil, data)

			testPath.Path = tc.Path
			data, err = masscbor.Marshal(testPath)
			// t.Logf("cbor diag:\n%s", pretty(data))
			assert.Nil(t, err)

			var got Path
			err = masscbor.Unmarshal(data, &got)
			t.Logf("got: %+v", got)
			if tc.Path.Type != ObjectTypeManifest { // we cant tell these from paths elements
				assert.Error(t, err) // decoding should fail
			}
		})
	}
}

type TestPath struct {
	Path
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
	patch := Patch{
		Op:   OpString(t.Name()),
		Path: Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1), Fields: []any{"Options", "color", "Variations", "pink"}},
	}

	data, err := masscbor.Marshal(patch)
	assert.Nil(t, err)
	// t.Log("CBOR encoded:")
	// t.Log("\n" + pretty(data))

	var rxPatch Patch
	err = masscbor.Unmarshal(data, &rxPatch)
	assert.Nil(t, err)
	assert.Equal(t, ObjectTypeListing, rxPatch.Path.Type)
	assert.Equal(t, objects.ObjectID(1), *rxPatch.Path.ObjectID)
	assert.Equal(t, []any{"Options", "color", "Variations", "pink"}, rxPatch.Path.Fields)

	var testPath struct {
		Path []any
	}
	err = masscbor.Unmarshal(data, &testPath)
	assert.Nil(t, err)
	assert.Equal(t, []any{"Listings", uint64(1), "Options", "color", "Variations", "pink"}, testPath.Path)
}

func TestPatchAdd(t *testing.T) {
	var (
		err error
		buf bytes.Buffer
		enc = masscbor.DefaultEncoder(&buf)
	)

	var createListing Patch
	createListing.Op = AddOp
	createListing.Path = Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(1)}

	_, lis := newTestListing()
	createListing.Value, err = masscbor.Marshal(lis)
	assert.Nil(t, err)
	err = enc.Encode(createListing)
	assert.Nil(t, err)

	opData := buf.Bytes()
	// t.Log("OP encoded:")
	// t.Log("\n" + pretty(opData))

	dec := masscbor.DefaultDecoder(bytes.NewReader(opData))
	var rxOp Patch
	err = dec.Decode(&rxOp)
	assert.Nil(t, err)
	assert.Equal(t, ObjectTypeListing, rxOp.Path.Type)
	assert.Equal(t, objects.ObjectID(1), *rxOp.Path.ObjectID)
	assert.Nil(t, validate.Struct(rxOp))

	dec = masscbor.DefaultDecoder(bytes.NewReader(rxOp.Value))
	var rxLis objects.Listing
	err = dec.Decode(&rxLis)
	assert.Nil(t, err)

	t.Logf("listing received: %+v", rxLis)
	assert.Equal(t, lis, rxLis, ignoreBigInts)
	assert.True(t, lis.Price.Cmp(&rxLis.Price) == 0)
}

var ignoreBigInts = cmpopts.IgnoreUnexported(*big.NewInt(0))
