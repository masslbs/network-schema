package main

import (
	"bytes"
	"math/big"
	"slices"
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

func TestMissingField(t *testing.T) {
	r := require.New(t)

	type FakeListing struct {
		Price Uint256
		// looks like a listing but has no metadata
		// Metadata  ListingMetadata
		ViewState ListingViewState
	}
	var fl FakeListing
	twentythree := big.NewInt(230000)
	fl.Price = *twentythree
	fl.ViewState = ListingViewStateDeleted
	t.Log("FakeListing:\n" + pretty(fl))
	testData := dump(fl)

	keys, err := MapKeys(fl)
	r.NoError(err)
	r.Len(keys, 2)
	r.False(slices.Contains(keys, "Metadata"))

	var lis Listing
	err = Decode(&lis, testData)
	r.Error(err)
}

func TestCreateOp(t *testing.T) {
	var (
		err error
		buf bytes.Buffer
		r   = require.New(t)
		enc = DefaultEncoder(&buf)
	)

	var createListing Patch
	createListing.Op = AddOp
	createListing.Path = []any{"listing", NewObjectId(238000)}

	var lis Listing
	lis.Metadata.Title = "test Listing"
	lis.Metadata.Description = "short desc"
	lis.Metadata.Images = []string{"https://http.cat/images/100.jpg"}
	price := big.NewInt(12345)
	lis.Price = *price

	err = enc.Encode(lis)
	r.NoError(err)
	lisBytes := buf.Bytes()
	buf.Reset()
	createListing.Value = lisBytes

	err = enc.Encode(createListing)
	r.NoError(err)

	opData := buf.Bytes()
	t.Log("OP encoded:")
	//t.Error(hex.EncodeToString(opData))
	t.Log("\n" + pretty(createListing))

	var rxOp Patch
	dec := DefaultDecoder(bytes.NewReader(opData))
	err = dec.Decode(&rxOp)
	r.NoError(err)
	r.Equal("listing", rxOp.Path[0])
	var rxLis Listing

	decLis := DefaultDecoder(bytes.NewReader(rxOp.Value))
	err = decLis.Decode(&rxLis)
	r.NoError(err)

	t.Logf("listing received: %+v", rxLis)
	r.EqualValues(lis, rxLis)
}
