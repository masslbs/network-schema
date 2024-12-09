package main

import (
	"fmt"

	ipldCbor "github.com/ipfs/go-ipld-cbor"
	mh "github.com/multiformats/go-multihash"

	schema "github.com/masslbs/network-schema/go"
)

func main() {
	var order schema.Order
	order.InvoiceAddress = &schema.AddressDetails{
		Name: "John Doe",
	}
	order.Items = []schema.OrderedItem{
		{
			ListingID: 1,
			Quantity:  2,
		},
		{
			ListingID:    2,
			VariationIDs: []string{"red", "blue"},
			Quantity:     3,
		},
	}
	// also doesnt work
	// ipldCbor.RegisterCborType(schema.Order{})

	wrapped, err := ipldCbor.WrapObject(&order, mh.BLAKE2B_MIN, -1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Links: %+v\n", wrapped.Links())
	fmt.Printf("Raw: %x\n", wrapped.RawData())
}
