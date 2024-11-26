package main

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/require"
)

func TestSignatureIncomplete(t *testing.T) {
	r := require.New(t)

	// TODO: [n]byte doesnt enforce that we actually got n bytes
	var msg struct {
		Sig Signature
	}

	// we Prepare a message that looks like a proper signature but is too short
	var short struct {
		Sig []byte
	}
	short.Sig = []byte{'x', 'x', 'x'}
	shortData, err := cbor.Marshal(short)
	r.NoError(err)
	diag(short)
	rd := bytes.NewReader(shortData)
	dec := DefaultDecoder(rd)

	err = dec.Decode(&msg)
	r.EqualValues([64]byte{}, msg.Sig)
	r.Error(err)
	r.IsType(ErrBytesTooShort{}, err)
}

func TestIncompleteField(t *testing.T) {
	r := require.New(t)

	// try missing metadata
	var lis Listing
	lis.Metadata.Title = "foo"
	diag(lis)

	type FakeListing struct {
		Price Uint256
		// looks like a listing but has no metadata
		// Metadata  ListingMetadata
		ViewState     ListingViewState
		Options       []ListingOption
		StockStatuses []ListingStockStatus
	}
	var fl FakeListing
	twentythree := big.NewInt(230000)
	fl.Price = *twentythree
	fl.ViewState = ListingViewStateDeleted
	diag(fl)
	testData := dump(fl)

	// TODO: shouldnt unmarshal with missing Metadata
	err := cbor.Unmarshal(testData, &lis)
	r.Error(err)
	t.Log("got expected error:", err)
}
