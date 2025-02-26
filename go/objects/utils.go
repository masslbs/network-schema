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

func DefaultValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("notblank", validators.NotBlank)
	validate.RegisterAlias("nonEmptyMapKeys", "dive,keys,required,notblank,endkeys,required")
	validate.RegisterStructValidation(OrderValidation, Order{})
	// we cant "nonEmptyMapKeys" via struct tags, since the library cant iterate through the HAMT
	validate.RegisterStructValidation(HAMTValidation, Tags{}, Accounts{}, Listings{}, Orders{})
	return validate
}

func bytesToId(buf []byte) ObjectId {
	if len(buf) != 8 {
		panic(fmt.Sprintf("expected 8 bytes, got %d", len(buf)))
	}
	return ObjectId(binary.BigEndian.Uint64(buf))
}

func combinedIDtoBytes(id ObjectId, variations []string) []byte {
	buf := idToBytes(id)
	sort.Strings(variations)
	buf = append(buf, []byte(strings.Join(variations, "|"))...)
	return buf
}

func bytesToCombinedID(buf []byte) (ObjectId, []string) {
	id := bytesToId(buf[:8])
	variations := strings.Split(string(buf[8:]), "|")
	if len(variations) == 1 && variations[0] == "" {
		variations = []string{}
	}
	return id, variations
}
