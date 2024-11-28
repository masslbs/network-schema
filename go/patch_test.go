// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

func TestPatchAdd(t *testing.T) {
	var (
		err error
		buf bytes.Buffer
		r   = require.New(t)
		enc = DefaultEncoder(&buf)
	)

	var createListing Patch
	createListing.Op = AddOp
	createListing.Path = PatchPath{Type: "listing", ID: 1}

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
	r.Equal("listing", rxOp.Path.Type)
	r.Equal(ObjectId(1), rxOp.Path.ID)
	r.NoError(validate.Struct(rxOp))

	dec = DefaultDecoder(bytes.NewReader(rxOp.Value))
	var rxLis Listing
	err = dec.Decode(&rxLis)
	r.NoError(err)

	t.Logf("listing received: %+v", rxLis)
	r.EqualValues(lis, rxLis)
}

func TestPatchListing(t *testing.T) {
	r := require.New(t)

	testColorOption := ListingOption{
		Title: "Color",
		Variations: map[string]ListingVariation{
			"pink": {
				ID: 333,
				VariationInfo: ListingMetadata{
					Title:       "Pink",
					Description: "Pink color",
				},
			},
			"orange": {
				ID: 444,
				VariationInfo: ListingMetadata{
					Title:       "Orange",
					Description: "Orange color",
				},
			},
		},
	}

	testSizeOption := ListingOption{
		Title: "Size",
		Variations: map[string]ListingVariation{
			"s": {
				ID: 33,
				VariationInfo: ListingMetadata{
					Title:       "Small",
					Description: "Small size",
				},
			},
			"m": {
				ID: 44,
				VariationInfo: ListingMetadata{
					Title:       "Medium",
					Description: "Medium size",
				},
			},
		},
	}

	testCases := []struct {
		name     string
		op       OpString
		path     PatchPath
		value    interface{}
		expected func(*require.Assertions, Listing)
	}{
		{
			name:  "replace price",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"price"}},
			value: *big.NewInt(66666),
			expected: func(r *require.Assertions, l Listing) {
				r.Equal(*big.NewInt(66666), l.Price)
			},
		},
		{
			name:  "replace description",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"metadata", "description"}},
			value: "new description",
			expected: func(r *require.Assertions, l Listing) {
				r.Equal("new description", l.Metadata.Description)
			},
		},
		{
			name:  "replace whole metadata",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"metadata"}},
			value: testListing().Metadata,
			expected: func(r *require.Assertions, l Listing) {
				r.Equal(testListing().Metadata, l.Metadata)
			},
		},
		{
			name:  "replace view state",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"viewState"}},
			value: ListingViewStatePublished,
			expected: func(r *require.Assertions, l Listing) {
				r.Equal(ListingViewStatePublished, l.ViewState)
			},
		},
		// map manipulation of Options
		{
			name:  "add an option",
			op:    AddOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "size"}},
			value: testSizeOption,
			expected: func(r *require.Assertions, l Listing) {
				sizeOption, ok := l.Options["size"]
				r.True(ok)
				r.Equal(testSizeOption, sizeOption)
			},
		},
		{
			name:  "replace one option",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "color"}},
			value: testColorOption,
			expected: func(r *require.Assertions, l Listing) {
				colorOption, ok := l.Options["color"]
				r.True(ok)
				r.Equal(testColorOption, colorOption)
			},
		},
		{
			name: "replace whole options",
			op:   ReplaceOp,
			path: PatchPath{Type: "listing", ID: 1, Fields: []string{"options"}},
			value: ListingOptions{
				"color": testColorOption,
			},
			expected: func(r *require.Assertions, l Listing) {
				colorOption, ok := l.Options["color"]
				r.True(ok)
				r.Equal(testColorOption, colorOption)
			},
		},

		{
			name:  "replace variation of an option",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "color", "variations", "b"}},
			value: testColorOption.Variations["pink"],
			expected: func(r *require.Assertions, l Listing) {
				r.Equal(testColorOption.Variations["pink"], l.Options["color"].Variations["b"])
			},
		},
		{
			name:  "replace variation info",
			op:    ReplaceOp,
			path:  PatchPath{Type: "listing", ID: 1, Fields: []string{"options", "color", "variations", "b", "variationInfo"}},
			value: testColorOption.Variations["pink"].VariationInfo,
			expected: func(r *require.Assertions, l Listing) {
				r.Equal(testColorOption.Variations["pink"].VariationInfo, l.Options["color"].Variations["b"].VariationInfo)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lis := testListing()
			patch := createPatch(tc.op, tc.path, tc.value)
			// round trip to make sure we can encode/decode the patch
			encodedPatch := encodePatch(t, patch)
			decodedPatch := decodePatch(t, encodedPatch)

			r.Equal(tc.op, decodedPatch.Op)
			r.Equal(tc.path, decodedPatch.Path)

			switch decodedPatch.Op {
			case ReplaceOp:
				err := lis.PatchReplace(decodedPatch.Path.Fields, decodedPatch.Value)
				r.NoError(err)
			case AddOp:
				err := lis.PatchAdd(decodedPatch.Path.Fields, decodedPatch.Value)
				r.NoError(err)
			}
			tc.expected(r, lis)
		})
	}
}

func createPatch(op OpString, path PatchPath, value interface{}) Patch {
	encodedValue, err := Marshal(value)
	if err != nil {
		panic(err)
	}
	return Patch{
		Op:    op,
		Path:  path,
		Value: encodedValue,
	}
}

func encodePatch(t *testing.T, patch Patch) []byte {
	encoded, err := Marshal(patch)
	require.NoError(t, err)
	t.Log("Patch encoded:\n" + pretty(encoded))
	return encoded
}

func decodePatch(t *testing.T, encoded []byte) Patch {
	var decoded Patch
	dec := DefaultDecoder(bytes.NewReader(encoded))
	err := dec.Decode(&decoded)
	require.NoError(t, err)
	require.NoError(t, validate.Struct(decoded))
	return decoded
}

func testListing() Listing {
	var lis Listing
	lis.ID = 1
	lis.ViewState = ListingViewStatePublished
	lis.Metadata.Title = "test Listing"
	lis.Metadata.Description = "short desc"
	lis.Metadata.Images = []string{"https://http.cat/images/100.jpg"}
	price := big.NewInt(12345)
	lis.Price = *price
	lis.Options = ListingOptions{
		"color": {
			Title: "Color",
			Variations: map[string]ListingVariation{
				"r": {
					ID: 1,
					VariationInfo: ListingMetadata{
						Title:       "Red",
						Description: "Red color",
					},
				},
				"b": {
					ID: 2,
					VariationInfo: ListingMetadata{
						Title:       "Blue",
						Description: "Blue color",
					},
				},
			},
		},
	}
	return lis
}

func (existing *Listing) PatchAdd(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 {
		return fmt.Errorf("PatchAdd requires at least one field")
	}
	switch fields[0] {
	case "options":
		assert(len(fields) == 2, "PatchAdd options requires exactly two fields")
		var option ListingOption
		err := Unmarshal(value, &option)
		if err != nil {
			return fmt.Errorf("failed to unmarshal option: %w", err)
		}
		existing.Options[fields[1]] = option
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
	return nil
}

func (existing *Listing) PatchReplace(fields []string, value cbor.RawMessage) error {
	switch fields[0] {
	case "price":
		var price Uint256
		err := Unmarshal(value, &price)
		if err != nil {
			return fmt.Errorf("failed to unmarshal price: %w", err)
		}
		existing.Price = price
	case "metadata":
		err := existing.Metadata.PatchReplace(fields[1:], value)
		if err != nil {
			return fmt.Errorf("failed to replace metadata: %w", err)
		}
	case "viewState":
		var viewState ListingViewState
		err := Unmarshal(value, &viewState)
		if err != nil {
			return fmt.Errorf("failed to unmarshal viewState: %w", err)
		}
		existing.ViewState = viewState
	case "options":
		err := existing.Options.PatchReplace(fields[1:], value)
		if err != nil {
			return fmt.Errorf("failed to replace options: %w", err)
		}
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
	// Validate the resulting struct
	if err := validate.Struct(existing); err != nil {
		return fmt.Errorf("validation failed after patch: %w", err)
	}

	return nil
}

func (existing *ListingMetadata) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 { // replace the whole metadata
		return Unmarshal(value, existing)
	}
	switch fields[0] {
	case "description":
		var description string
		err := Unmarshal(value, &description)
		if err != nil {
			return fmt.Errorf("failed to unmarshal description: %w", err)
		}
		existing.Description = description
	default:
		return fmt.Errorf("unsupported field: %s", fields[0])
	}
	return nil
}

func (existing *ListingOptions) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 { // replace the whole options
		return Unmarshal(value, existing)
	}
	option, ok := (*existing)[fields[0]]
	if !ok {
		return fmt.Errorf("option not found: %s", fields[0])
	}

	if len(fields) == 1 { // replace the whole option
		var newOption ListingOption
		err := Unmarshal(value, &newOption)
		if err != nil {
			return fmt.Errorf("failed to unmarshal new option: %w", err)
		}
		(*existing)[fields[0]] = newOption
		return nil
	}

	// patch a variation
	err := option.Variations.PatchReplace(fields[2:], value)
	if err != nil {
		return fmt.Errorf("failed to replace option variation: %w", err)
	}

	return nil
}

func (existing *ListingVariations) PatchReplace(fields []string, value cbor.RawMessage) error {
	if len(fields) == 0 { // replace the whole variations
		return Unmarshal(value, existing)
	}
	variation, ok := (*existing)[fields[0]]
	if !ok {
		return fmt.Errorf("variation not found: %s", fields[0])
	}
	if len(fields) == 1 { // replace the whole variation
		var newVariation ListingVariation
		err := Unmarshal(value, &newVariation)
		if err != nil {
			return fmt.Errorf("failed to unmarshal new variation: %w", err)
		}
		(*existing)[fields[0]] = newVariation
		return nil
	}
	switch fields[1] {
	case "variationInfo":
		var variationInfo ListingMetadata
		err := Unmarshal(value, &variationInfo)
		if err != nil {
			return fmt.Errorf("failed to unmarshal variation info: %w", err)
		}
		variation.VariationInfo = variationInfo
	case "priceModifier":
		var priceModifier PriceModifier
		err := Unmarshal(value, &priceModifier)
		if err != nil {
			return fmt.Errorf("failed to unmarshal price modifier: %w", err)
		}
		variation.PriceModifier = priceModifier
	default:
		return fmt.Errorf("unsupported field: %s", fields[1])
	}
	(*existing)[fields[0]] = variation
	return nil
}
