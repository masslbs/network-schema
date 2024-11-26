package main

import (
	"fmt"
	"encoding/hex"
	
	"github.com/fxamacker/cbor/v2"
)

func main() {
	var t Tag
	t.Name = "foo"
	t.ListingIds = []uint64{1,2,3}
	fmt.Printf("Simple Tag: %+v\n", t)

	data, err := cbor.Marshal(t)
	check(err)
	
	fmt.Println("CBOR:")
	fmt.Println(hex.Dump(data))
}

func check(err error) {
	if err!=nil{
		panic(err)
	}
}
