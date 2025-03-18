package patch

import (
	"math/big"
	"testing"

	clone "github.com/huandu/go-clone/generic"
	"github.com/peterldowns/testy/assert"

	"github.com/masslbs/network-schema/go/internal/testhelper"
	"github.com/masslbs/network-schema/go/objects"
)

// TestUserFlowVectors creates test vectors that simulate complete interactions
// between users and a shop, from browsing to checkout to order processing.
func TestGenerateVectorsUserFlows(t *testing.T) {
	// Define user stories (complete interaction flows)
	userStories := []struct {
		name  string
		story func(*testing.T, *vectorFileOkay)
	}{
		{"SimpleShoppingTrip", simpleShoppingTripStory},
		{"ShoppingTripWithVariations", shoppingTripStoryWithVariations},
		// {"OrderCancellationFlow", orderCancellationStory},
	}

	for _, story := range userStories {
		t.Run(story.name, func(t *testing.T) {
			// Each story writes its own vector file
			var vectors vectorFileOkay
			story.story(t, &vectors)
			writeVectors(t, vectors)
		})
	}
}

// A simple user story: user browses shop, adds item to cart, checks out
func simpleShoppingTripStory(t *testing.T, vectors *vectorFileOkay) {
	// Setup the empty shop
	shop := objects.NewShop(42)
	shopID := testRandomUint256()
	shop.Manifest.ShopID = shopID

	// Initialize vectors
	kp := initVectors(t, vectors, shopID)

	// Step 1: Shop owner sets up the shop
	steps := []struct {
		name     string
		patch    Patch
		validate func(*testing.T, objects.Shop)
	}{
		{
			name: "SetupShopBasics",
			patch: createPatch(t, ReplaceOp, Path{Type: ObjectTypeManifest}, objects.Manifest{
				ShopID: shopID,
				Payees: objects.Payees{
					1: {testAddr.Address: {CallAsContract: false}},
				},
				AcceptedCurrencies: objects.ChainAddresses{
					1: {testEth.Address: {}, testUsdc.Address: {}},
				},
				PricingCurrency: testUsdc,
				ShippingRegions: objects.ShippingRegions{
					"Default": {Country: "DE"},
				},
			}),
		},

		// Step 2: Shop owner adds a simple product
		{
			name: "AddTshirtListing",
			patch: createPatch(t, AddOp, Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(101)}, objects.Listing{
				ID:        101,
				Price:     *big.NewInt(1999), // $19.99
				ViewState: objects.ListingViewStatePublished,
				Metadata: objects.ListingMetadata{
					Title:       "T-Shirt",
					Description: "A great cotton t-shirt",
					Images:      []string{"https://example.com/tshirt.jpg"},
				},
			}),
		},

		// Step 3: Update inventory
		{
			name: "SetInventory",
			patch: createPatch(t, AddOp, Path{
				Type:     ObjectTypeInventory,
				ObjectID: testhelper.Uint64ptr(101),
			}, uint64(50)),
		},

		// Step 4: Customer creates an account (or uses a guest account)
		{
			name: "CreateCustomerAccount",
			patch: createPatch(t, AddOp, Path{
				Type:        ObjectTypeAccount,
				AccountAddr: testMassEthAddrPtr([20]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90}),
			}, objects.Account{
				KeyCards: []objects.PublicKey{testPubKey(42)},
				Guest:    true,
			}),
		},

		// Step 5: Customer creates an order
		{
			name: "CreateOrder",
			patch: createPatch(t, AddOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
			}, objects.Order{
				ID:    5001,
				State: objects.OrderStateOpen,
				Items: []objects.OrderedItem{
					{
						ListingID: 101,
						Quantity:  1,
					},
				},
				InvoiceAddress: &objects.AddressDetails{
					Name:         "Jane Doe",
					Address1:     "123 Main St",
					City:         "Anytown",
					PostalCode:   "12345",
					Country:      "US",
					EmailAddress: "jane@example.com",
				},
				ShippingAddress: &objects.AddressDetails{
					Name:         "Jane Doe",
					Address1:     "123 Main St",
					City:         "Anytown",
					PostalCode:   "12345",
					Country:      "US",
					EmailAddress: "jane@example.com",
				},
			}),
			validate: func(t *testing.T, s objects.Shop) {
				order, found := s.Orders.Get(5001)
				assert.True(t, found)
				assert.Equal(t, objects.OrderStateOpen, order.State)
			},
		},

		// Step 6: Order gets committed with payment details
		{
			name: "CommitOrder",
			patch: createPatch(t, ReplaceOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"State"},
			}, objects.OrderStateCommitted),
		},

		{
			name: "ChoosePaymentChannel1",
			patch: createPatch(t, AddOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"ChosenPayee"},
			}, objects.Payee{
				Address:        testAddr,
				CallAsContract: false,
			}),
		},
		{
			name: "ChoosePaymentChannel2",
			patch: createPatch(t, AddOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"ChosenCurrency"},
			}, testUsdc),
		},
		{
			name: "ChoosePaymentChannel3",
			patch: createPatch(t, ReplaceOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"State"},
			}, objects.OrderStatePaymentChosen),
		},

		// Step 7: Add payment details
		{
			name: "AddPaymentDetails",
			patch: createPatch(t, ReplaceOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"PaymentDetails"},
			}, objects.PaymentDetails{
				PaymentID: objects.Hash{0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37},
				ListingHashes: [][]byte{
					{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
				Total: *big.NewInt(1999),
				TTL:   3600,
			}),
		},
		{
			name: "ReadyToPay",
			patch: createPatch(t, ReplaceOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"State"},
			}, objects.OrderStateUnpaid),
		},

		// Step 8: relay found payment hash and updates order state
		{
			name: "RelayFoundPaymentHash1",
			patch: createPatch(t, AddOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"TxDetails"},
			}, objects.OrderPaid{
				TxHash:    &objects.Hash{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd},
				BlockHash: objects.Hash{0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37},
			}),
		},
		{
			name: "RelayFoundPaymentHash2",
			patch: createPatch(t, ReplaceOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"State"},
			}, objects.OrderStatePaid),
		},

		// Step 9: decrement inventory
		{
			name: "DecrementInventory",
			patch: createPatch(t, DecrementOp, Path{
				Type:     ObjectTypeInventory,
				ObjectID: testhelper.Uint64ptr(101),
			}, uint64(1)),
		},
	}

	// Execute all steps in the story
	beforeState := clone.Clone(shop)
	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			// Capture before state

			beforeEncoded := mustEncode(t, beforeState)
			before := vectorSnapshot{
				Value:   beforeState,
				Encoded: beforeEncoded,
				Hash:    mustHashState(t, beforeState),
			}

			// Apply patch
			patcher := NewPatcher(validate, &shop)
			err := patcher.ApplyPatch(step.patch)
			assert.Nil(t, err)

			// Optional validation
			if step.validate != nil {
				step.validate(t, shop)
			}

			// Capture after state
			afterState := clone.Clone(shop)
			afterEncoded := mustEncode(t, afterState)
			after := vectorSnapshot{
				Value:   afterState,
				Encoded: afterEncoded,
				Hash:    mustHashState(t, afterState),
			}

			// Add to vectors
			vectors.Snapshots = append(vectors.Snapshots, vectorEntryOkay{
				Name:   step.name,
				Before: before,
				After:  after,
			})
			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, step.patch)

			// iterate on the state per patch
			beforeState = clone.Clone(afterState)
		})
	}

	// Sign the patch set
	kp.TestSignPatchSet(t, &vectors.PatchSet)
}

// A simple user story: user browses shop, adds item to cart, checks out
func shoppingTripStoryWithVariations(t *testing.T, vectors *vectorFileOkay) {
	// Setup the empty shop
	shop := objects.NewShop(42)
	shopID := testRandomUint256()
	shop.Manifest.ShopID = shopID

	// Initialize vectors
	kp := initVectors(t, vectors, shopID)

	// Step 1: Shop owner sets up the shop
	steps := []struct {
		name     string
		patch    Patch
		validate func(*testing.T, objects.Shop)
	}{
		{
			name: "SetupShopBasics",
			patch: createPatch(t, ReplaceOp, Path{Type: ObjectTypeManifest}, objects.Manifest{
				ShopID: shopID,
				Payees: objects.Payees{
					1: {testAddr.Address: {CallAsContract: false}},
				},
				AcceptedCurrencies: objects.ChainAddresses{
					1: {testEth.Address: {}, testUsdc.Address: {}},
				},
				PricingCurrency: testUsdc,
				ShippingRegions: objects.ShippingRegions{
					"US": {Country: "United States"},
				},
			}),
			validate: func(t *testing.T, s objects.Shop) {
				assert.Equal(t, 1, len(s.Manifest.ShippingRegions))
			},
		},

		// Step 2: Shop owner adds products
		{
			name: "AddTshirtListing",
			patch: createPatch(t, AddOp, Path{Type: ObjectTypeListing, ObjectID: testhelper.Uint64ptr(101)}, objects.Listing{
				ID:        101,
				Price:     *big.NewInt(1999), // $19.99
				ViewState: objects.ListingViewStatePublished,
				Metadata: objects.ListingMetadata{
					Title:       "T-Shirt",
					Description: "A great cotton t-shirt",
					Images:      []string{"https://example.com/tshirt.jpg"},
				},
				Options: objects.ListingOptions{
					"size": {
						Title: "Size",
						Variations: map[string]objects.ListingVariation{
							"s": {VariationInfo: objects.ListingMetadata{Title: "Small", Description: "A small t-shirt"}},
							"m": {VariationInfo: objects.ListingMetadata{Title: "Medium", Description: "A medium t-shirt"}},
							"l": {VariationInfo: objects.ListingMetadata{Title: "Large", Description: "A large t-shirt"}},
						},
					},
					"color": {
						Title: "Color",
						Variations: map[string]objects.ListingVariation{
							"red":  {VariationInfo: objects.ListingMetadata{Title: "Red", Description: "A red t-shirt"}},
							"blue": {VariationInfo: objects.ListingMetadata{Title: "Blue", Description: "A blue t-shirt"}},
						},
					},
				},
			}),
		},

		// Step 3: Update inventory
		{
			name: "SetRedMediumInventory",
			patch: createPatch(t, AddOp, Path{
				Type:     ObjectTypeInventory,
				ObjectID: testhelper.Uint64ptr(101),
				Fields:   []any{"red", "m"},
			}, uint64(50)),
		},

		// Step 4: Customer creates an account (or uses a guest account)
		{
			name: "CreateCustomerAccount",
			patch: createPatch(t, AddOp, Path{
				Type:        ObjectTypeAccount,
				AccountAddr: testMassEthAddrPtr([20]byte{0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90, 0x12, 0x34, 0x56, 0x78, 0x90}),
			}, objects.Account{
				KeyCards: []objects.PublicKey{testPubKey(42)},
				Guest:    true,
			}),
		},

		// Step 5: Customer creates an order
		{
			name: "CreateOrder",
			patch: createPatch(t, AddOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
			}, objects.Order{
				ID:    5001,
				State: objects.OrderStateOpen,
				Items: []objects.OrderedItem{
					{
						ListingID:    101,
						Quantity:     1,
						VariationIDs: []string{"red", "m"},
					},
				},
				InvoiceAddress: &objects.AddressDetails{
					Name:         "Jane Doe",
					Address1:     "123 Main St",
					City:         "Anytown",
					PostalCode:   "12345",
					Country:      "US",
					EmailAddress: "jane@example.com",
				},
				ShippingAddress: &objects.AddressDetails{
					Name:         "Jane Doe",
					Address1:     "555 Other St",
					City:         "Difftown",
					PostalCode:   "333555",
					Country:      "US",
					EmailAddress: "jane@example.com",
				},
			}),
			validate: func(t *testing.T, s objects.Shop) {
				order, found := s.Orders.Get(5001)
				assert.True(t, found)
				assert.Equal(t, objects.OrderStateOpen, order.State)
			},
		},

		// Step 6: Order gets committed with payment details
		{
			name: "CommitOrder",
			patch: createPatch(t, ReplaceOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"State"},
			}, objects.OrderStateCommitted),
		},

		{
			name: "ChoosePaymentChannel1",
			patch: createPatch(t, AddOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"ChosenPayee"},
			}, objects.Payee{
				Address:        testAddr,
				CallAsContract: false,
			}),
		},
		{
			name: "ChoosePaymentChannel2",
			patch: createPatch(t, AddOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"ChosenCurrency"},
			}, testUsdc),
		},
		{
			name: "ChoosePaymentChannel3",
			patch: createPatch(t, ReplaceOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"State"},
			}, objects.OrderStatePaymentChosen),
		},

		// Step 7: Add payment details
		{
			name: "AddPaymentDetails",
			patch: createPatch(t, ReplaceOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"PaymentDetails"},
			}, objects.PaymentDetails{
				PaymentID: objects.Hash{0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37, 0x13, 0x37},
				ListingHashes: [][]byte{
					{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
				},
				Total: *big.NewInt(1999),
				TTL:   3600,
			}),
		},
		{
			name: "ReadyToPay",
			patch: createPatch(t, ReplaceOp, Path{
				Type:     ObjectTypeOrder,
				ObjectID: testhelper.Uint64ptr(5001),
				Fields:   []any{"State"},
			}, objects.OrderStateUnpaid),
		},

		// Step 8: Update inventory after order
		// {
		// 	name: "DecrementInventory",
		// 	patch: createPatch(t, DecrementOp, Path{
		// 		Type:     ObjectTypeInventory,
		// 		ObjectID: testhelper.Uint64ptr(101),
		// 		Fields:   []any{"red", "m"},
		// 	}, uint64(1)),
		// 	validate: func(t *testing.T, s objects.Shop) {
		// 		count, has := s.Inventory.Get(101, []string{"red", "m"})
		// 		assert.True(t, has)
		// 		assert.Equal(t, uint64(49), count)
		// 	},
		// },

		// // Step 9: Merchant marks order as shipped
		// {
		// 	name: "MarkOrderShipped",
		// 	patch: createPatch(t, ReplaceOp, Path{
		// 		Type:     ObjectTypeOrder,
		// 		ObjectID: testhelper.Uint64ptr(5001),
		// 		Fields:   []any{"State"},
		// 	}, objects.OrderStateCommitted),
		// },

		// // Step 10: Add tracking information
		// {
		// 	name: "AddTrackingInfo",
		// 	patch: createPatch(t, ReplaceOp, Path{
		// 		Type:     ObjectTypeOrder,
		// 		ObjectID: testhelper.Uint64ptr(5001),
		// 		Fields:   []any{"TrackingInfo"},
		// 	}, objects.TrackingInfo{
		// 		Carrier:      "USPS",
		// 		TrackingCode: "1Z999AA10123456784",
		// 		ShippedAt:    time.Now(),
		// 	}),
		// },

		// // Step 11: Customer marks order as received
		// {
		// 	name: "MarkOrderReceived",
		// 	patch: createPatch(t, ReplaceOp, Path{
		// 		Type:     ObjectTypeOrder,
		// 		ObjectID: testhelper.Uint64ptr(5001),
		// 		Fields:   []any{"State"},
		// 	}, objects.OrderStateReceived),
		// },
	}

	// Execute all steps in the story
	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			// Capture before state
			beforeState := clone.Clone(shop)
			beforeEncoded := mustEncode(t, beforeState)
			before := vectorSnapshot{
				Value:   beforeState,
				Encoded: beforeEncoded,
				Hash:    mustHashState(t, beforeState),
			}

			// Apply patch
			patcher := NewPatcher(validate, &shop)
			err := patcher.ApplyPatch(step.patch)
			assert.Nil(t, err)

			// Optional validation
			if step.validate != nil {
				step.validate(t, shop)
			}

			// Capture after state
			afterState := clone.Clone(shop)
			afterEncoded := mustEncode(t, afterState)
			after := vectorSnapshot{
				Value:   afterState,
				Encoded: afterEncoded,
				Hash:    mustHashState(t, afterState),
			}

			// Add to vectors
			vectors.Snapshots = append(vectors.Snapshots, vectorEntryOkay{
				Name:   step.name,
				Before: before,
				After:  after,
			})
			vectors.PatchSet.Patches = append(vectors.PatchSet.Patches, step.patch)
		})
	}

	// Sign the patch set
	kp.TestSignPatchSet(t, &vectors.PatchSet)
}

func orderCancellationStory(t *testing.T, vectors *vectorFileOkay) {
	// Flow showing order creation and then cancellation
	// - Create order
	// - Add payment
	// - Customer requests cancellation
	// - Shop approves cancellation
	// - Refund issued
	// ...implementation similar to above
}
