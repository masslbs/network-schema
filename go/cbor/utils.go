// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"io"

	"github.com/fxamacker/cbor/v2"
)

func MassMarketTags() cbor.TagSet {
	tags := cbor.NewTagSet()

	// Register tag for Status enum type (using tag 1000)
	// tags.Add(
	// 	cbor.TagOptions{EncTag: cbor.EncTagRequired, DecTag: cbor.DecTagRequired},
	// 	reflect.TypeOf(ListingViewState(0)),
	// 	1000,
	// )
	return tags
}

func DefaultDecoder(rd io.Reader) cbor.Decoder {
	opts := cbor.DecOptions{
		BinaryUnmarshaler: cbor.BinaryUnmarshalerByteString,
	}
	opts.TimeTag = cbor.DecTagRequired
	mode, err := opts.DecModeWithTags(MassMarketTags())
	check(err)
	return *mode.NewDecoder(rd)
}

func Unmarshal(data []byte, v interface{}) error {
	dec := DefaultDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}

func DefaultEncoder(w io.Writer) *cbor.Encoder {
	opts := cbor.CanonicalEncOptions()
	opts.BigIntConvert = cbor.BigIntConvertShortest
	opts.Time = cbor.TimeRFC3339
	opts.TimeTag = cbor.EncTagRequired
	mode, err := opts.EncModeWithTags(MassMarketTags())
	check(err)
	return mode.NewEncoder(w)
}

func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := DefaultEncoder(&buf)
	err := enc.Encode(v)
	return buf.Bytes(), err
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
