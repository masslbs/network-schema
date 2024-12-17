// SPDX-FileCopyrightText: 2024 Mass Labs
//
// SPDX-License-Identifier: MIT

package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func openTestFile(t testing.TB, fileName string) *os.File {
	path := filepath.Join(os.Getenv("TEST_DATA_OUT"), fileName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
	}
	return file
}
