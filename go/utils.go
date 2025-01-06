// SPDX-FileCopyrightText: 2024 - 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"

	"github.com/fxamacker/cbor/v2"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
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

func Unmarshal(data []byte, v interface{}) error {
	dec := DefaultDecoder(bytes.NewReader(data))
	return dec.Decode(v)
}

func DefaultEncoder(w io.Writer) *cbor.Encoder {
	opts := cbor.CanonicalEncOptions()
	opts.BigIntConvert = cbor.BigIntConvertShortest
	opts.Time = cbor.TimeRFC3339
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

func DefaultValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("notblank", validators.NotBlank)
	validate.RegisterAlias("nonEmptyMapKeys", "dive,keys,required,notblank,endkeys,required")
	validate.RegisterStructValidation(OrderValidation, Order{})
	// we cant "nonEmptyMapKeys" via struct tags, since the library cant iterate through the HAMT
	validate.RegisterStructValidation(HAMTValidation, Tags{}, Accounts{}, Listings{}, Orders{})
	return validate
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

	fmt.Println(diagStr)
}

func pretty(data []byte) string {
	if os.Getenv("PRETTY") == "" {
		return hex.EncodeToString(data)
	}
	shell := exec.Command("cbor2pretty.rb")
	shell.Stdin = bytes.NewReader(data)

	out, err := shell.CombinedOutput()
	check(err)
	return string(out)
}
