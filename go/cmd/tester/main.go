package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"

	masscbor "github.com/masslbs/network-schema/go/cbor"
)

func main() {
	var useHex bool
	flag.BoolVar(&useHex, "hex", false, "decode hex")
	flag.Parse()

	data, err := io.ReadAll(os.Stdin)
	check(err)

	if useHex {
		data, err = hex.DecodeString(string(data))
		check(err)
	}

	var obj interface{}
	err = masscbor.Unmarshal(data, &obj)
	check(err)

	out, err := masscbor.Marshal(obj)
	check(err)

	if bytes.Equal(data, out) {
		fmt.Println("round trip ok")
	} else {
		fmt.Println("round trip failed")
		fmt.Println("original:", hex.EncodeToString(data))
		fmt.Println("encoded: ", hex.EncodeToString(out))

		// Find where they diverge
		minLen := len(data)
		if len(out) < minLen {
			minLen = len(out)
		}

		for i := 0; i < minLen; i++ {
			if data[i] != out[i] {
				fmt.Printf("first difference at position %d: original=0x%02x, encoded=0x%02x\n",
					i, data[i], out[i])
				break
			}
		}

		if len(data) != len(out) {
			fmt.Printf("length mismatch: original=%d, encoded=%d\n", len(data), len(out))
		}
		os.Exit(1)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
