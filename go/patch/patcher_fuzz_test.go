// SPDX-FileCopyrightText: 2025 Mass Labs
//
// SPDX-License-Identifier: MIT

package patch

import (
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/fxamacker/cbor/v2"
	clone "github.com/huandu/go-clone/generic"

	"github.com/masslbs/network-schema/go/objects"
)

// TestVectorType represents the different types of test vector files
type TestVectorType string

const (
	VectorShop      TestVectorType = "ShopOkay"
	VectorInventory TestVectorType = "InventoryOkay"
	VectorManifest  TestVectorType = "ManifestOkay"
	VectorListing   TestVectorType = "ListingOkay"
	VectorTag       TestVectorType = "TagOkay"
	VectorOrder     TestVectorType = "OrderOkay"
)

// loadTestVectors loads test vectors from the test data files
func loadTestVectors(t testing.TB, vectorType TestVectorType) *vectorFileOkay {
	// Define vector file path
	filename := string(vectorType) + ".cbor"

	// Check a few common paths for the testdata directory
	possiblePaths := []string{
		filepath.Join("vectors", filename),
		filepath.Join("..", "vectors", filename),
		filepath.Join("..", "..", "vectors", filename),
		filepath.Join(os.Getenv("TEST_DATA_OUT"), "vectors", filename),
	}

	var data []byte
	var err error
	var foundPath string

	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			foundPath = path
			break
		}
	}

	if err != nil {
		t.Logf("Warning: Could not load test vectors from %s. Some test cases will be skipped: %v", filename, err)
		return nil
	}

	t.Logf("Loading test vectors from: %s", foundPath)

	var vectors vectorFileOkay
	err = cbor.Unmarshal(data, &vectors)
	if err != nil {
		t.Logf("Warning: Failed to unmarshal test vectors: %v", err)
		return nil
	}

	return &vectors
}

// FuzzPatcherSimple tests that the Patcher doesn't panic with arbitrary input
func FuzzPatcherSimple(f *testing.F) {

	// Add basic seeds for different operations and types
	f.Add([]byte("add"), []byte("Manifest"), []byte{0x01})
	f.Add([]byte("replace"), []byte("Accounts"), []byte{0x01, 0x02})
	f.Add([]byte("remove"), []byte("Listings"), []byte{0x03, 0x04})
	f.Add([]byte("append"), []byte("Orders"), []byte{0x05, 0x06})
	f.Add([]byte("increment"), []byte("Tags"), []byte{0x07, 0x08})
	f.Add([]byte("decrement"), []byte("Inventory"), []byte{0x09, 0x0A})

	// Invalid operations
	f.Add([]byte("invalid_op"), []byte("Manifest"), []byte{0x0B})

	// Different object types with specific formats
	f.Add([]byte("add"), []byte("SchemaVersion"), []byte{0x01})
	f.Add([]byte("replace"), []byte("Manifest"), []byte{0x02})

	// Invalid object types
	f.Add([]byte("add"), []byte("InvalidType"), []byte{0x0C})

	// Empty values
	f.Add([]byte("add"), []byte("Manifest"), []byte{})

	// Add seeds from test vectors
	var shopStates []*objects.Shop
	// Load test vectors to use as corpus
	testVectorTypes := []TestVectorType{
		VectorShop,
		VectorInventory,
		VectorManifest,
		VectorListing,
		VectorTag,
		VectorOrder,
	}
	for _, vectorType := range testVectorTypes {
		v := loadTestVectors(f, vectorType)
		if v == nil {
			continue
		}
		for _, snapshot := range v.Snapshots {
			shop := clone.Clone(snapshot.Before.Value)
			shopStates = append(shopStates, &shop)
		}
		for _, patch := range v.PatchSet.Patches {
			// Extract op, type, and value from patch
			opBytes, err := cbor.Marshal(string(patch.Op))
			if err != nil {
				continue
			}

			typeBytes, err := cbor.Marshal(string(patch.Path.Type))
			if err != nil {
				continue
			}

			valueBytes := patch.Value
			if len(valueBytes) == 0 {
				valueBytes = []byte{0x01}
			}

			// Add to corpus
			f.Add(opBytes, typeBytes, []byte(valueBytes))
		}
	}

	v := objects.DefaultValidator()

	// Run the fuzzer
	f.Fuzz(func(t *testing.T, opBytes, typeBytes, valueBytes []byte) {
		// Get a random shop state from our collection instead of always creating a new minimal one
		var shop *objects.Shop
		if len(shopStates) == 0 {
			// Fallback to basic shop if no states loaded
			s := objects.NewShop(23)
			shop = &s
		} else {
			// Select a random shop state from our loaded test vectors
			shopIndex := rand.Uint64N(uint64(len(shopStates)))
			shop = shopStates[shopIndex]
		}

		// Create and register custom validators

		// Create patcher
		patcher := NewPatcher(v, shop)

		// Create a patch from the fuzz inputs
		patch := Patch{}

		// Try to set Op - should be a string from the set of valid ops
		var op string
		if err := cbor.Unmarshal(opBytes, &op); err == nil {
			patch.Op = OpString(op)
		} else {
			// If unmarshal fails, just use a random valid op
			validOps := []OpString{AddOp, ReplaceOp, RemoveOp, AppendOp, IncrementOp, DecrementOp}
			if len(opBytes) > 0 {
				patch.Op = validOps[int(opBytes[0])%len(validOps)]
			} else {
				patch.Op = AddOp
			}
		}

		// Try to set Type
		var objType string
		if err := cbor.Unmarshal(typeBytes, &objType); err == nil {
			patch.Path.Type = ObjectType(objType)
		} else {
			// If unmarshal fails, just use a random valid type
			validTypes := []ObjectType{
				ObjectTypeSchemaVersion,
				ObjectTypeManifest,
				ObjectTypeAccount,
				ObjectTypeListing,
				ObjectTypeOrder,
				ObjectTypeTag,
				ObjectTypeInventory,
			}
			if len(typeBytes) > 0 {
				patch.Path.Type = validTypes[int(typeBytes[0])%len(validTypes)]
			} else {
				patch.Path.Type = ObjectTypeManifest
			}
		}

		// Set appropriate ID fields based on type
		switch patch.Path.Type {
		case ObjectTypeAccount:
			var addr objects.EthereumAddress
			if len(valueBytes) >= objects.EthereumAddressSize {
				copy(addr.Address[:], valueBytes[:objects.EthereumAddressSize])
			}
			patch.Path.AccountAddr = &addr
		case ObjectTypeListing, ObjectTypeOrder, ObjectTypeInventory:
			var id uint64 = 1 // Default ID
			if len(valueBytes) >= 8 {
				// Extract a uint64 from the first 8 bytes
				for i := 0; i < 8 && i < len(valueBytes); i++ {
					id = (id << 8) | uint64(valueBytes[i])
				}
				// Ensure ID is not zero as the validation might reject it
				if id == 0 {
					id = 1
				}
			}
			objID := objects.ObjectID(id)
			patch.Path.ObjectID = &objID

			// For inventory, we might need variation fields too
			if patch.Path.Type == ObjectTypeInventory && len(valueBytes) > 8 {
				// Add a simple variation array if there's more data
				patch.Path.Fields = []any{string(valueBytes[8:])}
			}
		case ObjectTypeTag:
			tagName := "tag"
			if len(valueBytes) > 0 {
				// Use the bytes as a tag name if possible
				tagName = string(valueBytes)
				// Ensure tag name is not empty
				if tagName == "" {
					tagName = "tag"
				}
			}
			patch.Path.TagName = &tagName
		}

		// Set Value - this is any CBOR data
		// Ensure the Value field has at least some content
		if len(valueBytes) == 0 {
			valueBytes = []byte{0x01} // default non-empty value
		}
		patch.Value = valueBytes

		// Expected errors are fine, we're just making sure it doesn't panic
		err := patcher.ApplyPatch(patch)
		if err != nil {
			// This is expected in many cases, so we just log it
			t.Logf("Patcher.ApplyPatch returned error: %v", err)
		}
	})
}
