// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

// Package cbor provides utility functions for encoding and decoding CBOR data using default options.
//
// This includes:
// - canonical encoding rules
// - using RFC3339 for time values
// - tagging of time values
// - big.Int values are converted to integers, if they fit
package cbor

import (
	"bytes"
	"io"

	"github.com/fxamacker/cbor/v2"
)

// DefaultDecoder returns a new decoder for the given reader.
func DefaultDecoder(rd io.Reader) cbor.Decoder {
	opts := cbor.DecOptions{
		BinaryUnmarshaler: cbor.BinaryUnmarshalerByteString,
	}
	opts.TimeTag = cbor.DecTagRequired
	mode, err := opts.DecMode()
	check(err)
	return *mode.NewDecoder(rd)
}

// Unmarshal unmarshals the given data into the given value.
func Unmarshal(data []byte, v interface{}) error {
	dec := DefaultDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}

// DefaultEncoder returns a new encoder for the given writer.
func DefaultEncoder(w io.Writer) *cbor.Encoder {
	opts := cbor.CanonicalEncOptions()
	opts.BigIntConvert = cbor.BigIntConvertShortest
	opts.Time = cbor.TimeRFC3339
	opts.TimeTag = cbor.EncTagRequired
	mode, err := opts.EncMode()
	check(err)
	return mode.NewEncoder(w)
}

// Marshal marshals the given value into a byte slice.
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
