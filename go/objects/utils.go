// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package objects

import (
	"encoding/binary"
	"fmt"
	"sort"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
)

// DefaultValidator returns a new validator
func DefaultValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("notblank", validators.NotBlank)
	validate.RegisterAlias("nonEmptyMapKeys", "dive,keys,required,notblank,endkeys,required")
	validate.RegisterStructValidation(OrderValidation, Order{})
	// we cant "nonEmptyMapKeys" via struct tags, since the library cant iterate through the HAMT
	validate.RegisterStructValidation(HAMTValidation, Tags{}, Accounts{}, Listings{}, Orders{})
	return validate
}

func bytesToID(buf []byte) ObjectID {
	if len(buf) != 8 {
		panic(ErrBytesTooShort{Want: 8, Got: uint(len(buf))})
	}
	return ObjectID(binary.BigEndian.Uint64(buf))
}

// combinedIDtoBytes converts an ObjectID and a list of variations to a byte slice
func combinedIDtoBytes(id ObjectID, variations []string) []byte {
	buf := idToBytes(id)
	sort.Strings(variations)
	buf = append(buf, []byte(strings.Join(variations, "|"))...)
	return buf
}

func bytesToCombinedID(buf []byte) (ObjectID, []string) {
	if len(buf) < 8 {
		panic(ErrBytesTooShort{Want: 8, Got: uint(len(buf))})
	}
	id := bytesToID(buf[:8])
	variations := strings.Split(string(buf[8:]), "|")
	if len(variations) == 1 && variations[0] == "" {
		variations = []string{}
	}
	return id, variations
}

// ErrBytesTooShort is an error that occurs when a byte slice is too short
type ErrBytesTooShort struct {
	Want, Got uint
}

// Error returns a string representation of the error
func (err ErrBytesTooShort) Error() string {
	return fmt.Sprintf("not enough bytes. Expected %d but got %d", err.Want, err.Got)
}
