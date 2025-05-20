// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package objects

import (
	"testing"

	"github.com/peterldowns/testy/assert"
)

func TestCombinedID(t *testing.T) {
	id := ObjectID(1)
	buf := combinedIDtoBytes(id, nil)
	id, variations := bytesToCombinedID(buf)
	assert.Equal(t, id, ObjectID(1))
	assert.Equal(t, variations, []string{})

	// test with variations
	id = ObjectID(2)
	variations = []string{"a", "b", "c"}
	buf = combinedIDtoBytes(id, variations)
	id, variations = bytesToCombinedID(buf)
	assert.Equal(t, id, ObjectID(2))
	assert.Equal(t, variations, []string{"a", "b", "c"})
}
