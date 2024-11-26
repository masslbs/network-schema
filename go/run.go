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

	dump(t)

	// some simple checks
	var sig Signature
	var short = []byte{'x', 'x','x'}
	err:=cbor.Unmarshal( short, &sig)
	if err==nil {
		panic("expected error")
	}
	fmt.Println("got expected error:", err)

	copy(sig[:], []byte("foooo"))
	dump(sig)
	
}

func check(err error) {
	if err!=nil{
		panic(err)
	}
}

func dump(val interface{}) {
	data, err := cbor.Marshal(val)
	check(err)
	
	fmt.Println("CBOR:")
	fmt.Println(hex.Dump(data))
}
