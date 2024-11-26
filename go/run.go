package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"os/exec"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	"golang.org/x/exp/maps"
)

func MassMarketTags() cbor.TagSet {
	tags := cbor.NewTagSet()

	// Register tag for Status enum type (using tag 1000)
	tags.Add(
		cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
		reflect.TypeOf(ListingViewState(0)), // Reflect the Status type
		1000,                                // CBOR tag number for the Status enum
	)
	return tags
}

func DefaultDecoder(rd io.Reader) cbor.Decoder {
	opts := cbor.DecOptions{
		BinaryUnmarshaler: cbor.BinaryUnmarshalerByteString,
	}
	mode, err := opts.DecModeWithTags(MassMarketTags())
	check(err)
	return *mode.NewDecoder(rd)
}

func DefaultEncoder(w io.Writer) cbor.Encoder {
	opts := cbor.CanonicalEncOptions()
	opts.BigIntConvert = cbor.BigIntConvertShortest
	mode, err := opts.EncModeWithTags(MassMarketTags())
	check(err)
	return *mode.NewEncoder(w)
}

func MapKeys(val any) ([]string, error) {
	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)
	err := enc.Encode(val)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	dec := DefaultDecoder(&buf)
	err = dec.Decode(&m)
	if err != nil {
		return nil, err
	}

	return maps.Keys(m), nil
}

func main() {
	var t Tag
	t.Name = "foo"
	t.ListingIds = []uint64{1, 2, 3}
	fmt.Println("Simple Tag")
	diag(t)

	var l Listing
	l.Metadata.Title = "foo"
	price := big.NewInt(1111122222333344449)
	price = price.Mul(price, big.NewInt(999999999999999999))
	l.Price = *price
	l.ViewState = ListingViewStatePublished
	dump(l)

	fmt.Println(MapKeys(l))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func dump(val any) []byte {
	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)

	err := enc.Encode(val)
	check(err)

	fmt.Printf("CBOR of: %+v\n", val)
	data := buf.Bytes()
	fmt.Println(hex.EncodeToString(data))
	return data
}

func diag(val any) {
	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)

	err := enc.Encode(val)
	check(err)

	diagStr, err := cbor.Diagnose(buf.Bytes())
	check(err)

	//fmt.Printf("\n\nDIAG of: %+v\n", val)

	fmt.Println(diagStr)
}

func pretty(val any) string {
	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)

	err := enc.Encode(val)
	check(err)

	shell := exec.Command("cbor2pretty.rb")
	shell.Stdin = &buf

	out, err := shell.CombinedOutput()
	check(err)
	return string(out)
}
