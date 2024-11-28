// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestPatchAdd(t *testing.T) {
	var (
		err error
		buf bytes.Buffer
		r   = require.New(t)
		enc = DefaultEncoder(&buf)
	)

	var createListing PatchTx
	createListing.Op = AddOp
	createListing.Path = []any{"listing", ObjectId(1)}

	lis := testListing()
	createListing.Value = lis

	err = enc.Encode(createListing)
	r.NoError(err)

	opData := buf.Bytes()
	t.Log("OP encoded:")
	t.Log("\n" + pretty(opData))

	dec := DefaultDecoder(bytes.NewReader(opData))
	var rxOp PatchRx
	err = dec.Decode(&rxOp)
	r.NoError(err)
	r.Equal("listing", rxOp.Path[0])
	r.NoError(validate.Struct(rxOp))

	dec = DefaultDecoder(bytes.NewReader(rxOp.Value))
	var rxLis Listing
	err = dec.Decode(&rxLis)
	r.NoError(err)

	t.Logf("listing received: %+v", rxLis)
	r.EqualValues(lis, rxLis)
}

func TestPatchReplace(t *testing.T) {
	r := require.New(t)

	var replaceOp PatchTx
	replaceOp.Op = ReplaceOp
	replaceOp.Path = []any{"listing", ObjectId(1)}

	var partial struct {
		Price     Uint256
		ViewState ListingViewState
	}
	partial.Price = *big.NewInt(66666)
	partial.ViewState = ListingViewStateDeleted

	replaceOp.Value = partial

	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)
	err := enc.Encode(replaceOp)
	r.NoError(err)
	t.Log("replaceOp encoded:")
	t.Log("\n" + pretty(buf.Bytes()))

	var receivedOp PatchRx
	dec := DefaultDecoder(bytes.NewReader(buf.Bytes()))
	err = dec.Decode(&receivedOp)
	r.NoError(err)
	r.NoError(validate.Struct(receivedOp))
	r.Equal(ReplaceOp, receivedOp.Op)

	// listing.apply(patch) sketch
	r.Equal([]any{"listing", ObjectId(1)}, receivedOp.Path)
	lis := testListing()
	r.Equal(lis.ID, receivedOp.Path[1])

	// we know it's a listing so we use ListingPartial
	var partialRx ListingPartial
	dec = DefaultDecoder(bytes.NewReader(receivedOp.Value))
	err = dec.Decode(&partialRx)
	r.NoError(err)
	t.Logf("partial:\n%s", spew.Sdump(partialRx))

	err = lis.Patch(&partialRx)
	r.NoError(err)
	t.Logf("listing:\n%s", spew.Sdump(lis))
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
	return lis
}

func (existing *Listing) Patch(partial *ListingPartial) error {
	if partial.Price != nil {
		existing.Price = *partial.Price
	}
	if partial.Metadata != nil {
		// TODO: this would overwrite the existing metadata
		// TODO: we also need ListingMetadataPartial... :S
		existing.Metadata = *partial.Metadata
	}
	if partial.ViewState != nil {
		existing.ViewState = *partial.ViewState
	}
	if partial.Options != nil {
		for k, v := range partial.Options {
			existing.Options[k] = v
		}
	}
	if partial.StockStatuses != nil {
		existing.StockStatuses = partial.StockStatuses
	}

	// Validate the resulting struct
	if err := validate.Struct(existing); err != nil {
		return fmt.Errorf("validation failed after patch: %w", err)
	}

	return nil
}
