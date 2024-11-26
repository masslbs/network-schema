package main

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/fxamacker/cbor/v2"
)

func main() {
	var t Tag
	t.Name = "foo"
	t.ListingIds = []uint64{1, 2, 3}
	fmt.Printf("Simple Tag")
	dump(t)

	// some simple checks
	var sig Signature
	var short = []byte{'x', 'x', 'x'}
	err := cbor.Unmarshal(short, &sig)
	if err == nil {
		panic("expected error")
	}
	fmt.Println("got expected error:", err)

	copy(sig[:], []byte("foooo"))
	dump(sig)

	// try missing metadata

	var lis Listing
	lis.Metadata.Title = "foo"
	dump(lis)

	type FakeListing struct {
		Price Uint256
		//Metadata  ListingMetadata
		ViewState ListingViewState
		Options   []ListingOption
		// one for each combination of variations
		StockStatuses []ListingStockStatus
	}
	var fl FakeListing
	twentythree :=big.NewInt(230000)
	fl.Price.Int = *twentythree
	testData := dump(fl)

	// TODO: shouldnt be unmarshal with missing Metadata
	err = cbor.Unmarshal(testData, &lis)
	if err == nil {
		dump(lis)
		panic("expected error")
	}
	fmt.Println("got expected error:", err)
	
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func dump(val interface{}) []byte {
	data, err := cbor.Marshal(val)
	check(err)

	fmt.Printf("CBOR of: %+v\n", val)
	fmt.Println(hex.Dump(data))
	return data
}
