package hamt

import (
	"bytes"
	"io"

	"github.com/fxamacker/cbor/v2"
)

// TODO: this file is temporary, should be merged with schema/utils.go

func DefaultDecoder(rd io.Reader) (*cbor.Decoder, error) {
	opts := cbor.DecOptions{
		BinaryUnmarshaler: cbor.BinaryUnmarshalerByteString,
	}
	mode, err := opts.DecMode()
	if err != nil {
		return nil, err
	}
	return mode.NewDecoder(rd), nil
}

func Unmarshal(data []byte, v interface{}) error {
	dec, err := DefaultDecoder(bytes.NewReader(data))
	if err != nil {
		return err
	}
	return dec.Decode(v)
}

func DefaultEncoder(w io.Writer) (*cbor.Encoder, error) {
	opts := cbor.CanonicalEncOptions()
	opts.BigIntConvert = cbor.BigIntConvertShortest
	opts.Time = cbor.TimeRFC3339
	mode, err := opts.EncMode()
	if err != nil {
		return nil, err
	}
	return mode.NewEncoder(w), nil
}

func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc, err := DefaultEncoder(&buf)
	if err != nil {
		return nil, err
	}
	err = enc.Encode(v)
	return buf.Bytes(), err
}
