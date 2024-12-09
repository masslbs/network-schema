package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing/quick"

	"github.com/ipfs/go-cid"
	schema "github.com/masslbs/network-schema/go"
)

func testOrder() schema.Order {
	var order schema.Order
	order.Items = schema.OrderedItems{
		schema.OrderedItem{
			ListingID: 782193,
			Quantity:  1,
		},
		schema.OrderedItem{
			ListingID:    782194,
			Quantity:     2,
			VariationIDs: []string{"red", "blue"},
		},
	}
	order.PaymentDetails = &schema.PaymentDetails{
		TTL:       81238,
		PaymentID: schema.Hash{0x01, 0x02, 0x03},
		ListingHashes: []cid.Cid{
			testCID(1),
			testCID(2),
			testCID(3),
		},
	}
	order.InvoiceAddress = &schema.AddressDetails{
		Name: "John Doe",
	}
	return order
}

func testListing(seed int64) schema.Listing {
	src := rand.New(rand.NewSource(seed))
	listing := schema.Listing{
		ID:    schema.ObjectId(src.Int63()),
		Price: uint64(src.Int63()),
		Metadata: schema.ListingMetadata{
			Title:       fmt.Sprintf("Test Listing %d", src.Int63n(1000000)),
			Description: fmt.Sprintf("Test Description %d", src.Int63n(1000000)),
			Images:      []string{"image1.jpg", "image2.jpg"},
		},
		ViewState: schema.ListingViewState(src.Intn(3)),
		Options: schema.ListingOptions{
			"Color": schema.ListingOption{
				Title: "Color",
				Variations: schema.ListingVariations{
					"red":   testVariation(1),
					"blue":  testVariation(2),
					"green": testVariation(3),
				},
			},
			"Size": schema.ListingOption{
				Title: "Size",
				Variations: schema.ListingVariations{
					"small":  testVariation(5),
					"medium": testVariation(6),
					"large":  testVariation(7),
				},
			},
		},
		StockStatuses: []schema.ListingStockStatus{
			schema.ListingStockStatus{},
		},
	}
	return listing
}

func testVariation(seed int64) schema.ListingVariation {
	src := rand.New(rand.NewSource(seed))
	v, ok := quick.Value(reflect.TypeOf(schema.ListingVariation{}), src)
	if !ok {
		panic("failed to generate random ListingVariation")
	}
	return v.Interface().(schema.ListingVariation)
}

var orderSchema = `
	
# Listing
# =======

type Listing struct {
	ID        Int
	Price     Int
	Metadata  ListingMetadata
	ViewState Int
	Options   {String:ListingOption}
	StockStatuses [ListingStockStatus]
}

type ListingMetadata struct {
	Title       String
	Description String
	Images      [String]
}

type ListingOption struct {
	Title      String
	Variations {String:ListingVariation}
}

type ListingVariation struct {
	VariationInfo ListingMetadata
	PriceModifier optional PriceModifier
	SKU           optional String
}

type ListingStockStatus struct {
	VariationIDs [String]
	InStock      optional Bool
	ExpectedInStockBy optional Int
}

type PriceModifier struct {
	ModificationPrecents optional Int
	ModificationAbsolute optional Int
}


# Order
# =====
type Order struct {
	Items           [OrderedItem]
	State           Int
	InvoiceAddress  optional AddressDetails
	ShippingAddress optional AddressDetails
	CanceledAt      optional Int
	ChosenPayee     optional Payee
	ChosenCurrency  optional ChainAddress
	PaymentDetails  optional PaymentDetails
	TxDetails       optional OrderPaid
}

type OrderedItem struct {
	ListingID    Int
	VariationIDs optional [String]
	Quantity     Int
}

type AddressDetails struct {
	Name         String
	Address1     String
	Address2     optional String
	City         String
	PostalCode   optional String
	Country      String
	EmailAddress String
	PhoneNumber  optional String
}

type PaymentDetails struct {
	PaymentID     Bytes
	Total         Int
	ListingHashes [Link]
	TTL           Int
	ShopSignature Bytes
}

type OrderPaid struct {
	TxHash    optional Bytes
	BlockHash Bytes
}

type ChainAddress struct {
	ChainId   Int
	Address   Bytes
}
type Time struct {
	Location  Int
	BlockNum  Int
	TxIndex  Int
}

type Payee struct {
	Address        ChainAddress
	CallAsContract Bool
}`
