// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package objects

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCombinedID(t *testing.T) {
	id := ObjectId(1)
	buf := combinedIDtoBytes(id, nil)
	id, variations := bytesToCombinedID(buf)
	require.Equal(t, id, ObjectId(1))
	require.Equal(t, variations, []string{})

	// test with variations
	id = ObjectId(2)
	variations = []string{"a", "b", "c"}
	buf = combinedIDtoBytes(id, variations)
	id, variations = bytesToCombinedID(buf)
	require.Equal(t, id, ObjectId(2))
	require.Equal(t, variations, []string{"a", "b", "c"})
}
