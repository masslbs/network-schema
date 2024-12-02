package main

import (
	"fmt"
	"os"

	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	ipldSchema "github.com/ipld/go-ipld-prime/schema"
	mh "github.com/multiformats/go-multihash"

	schema "github.com/masslbs/network-schema/go"
)

var ts *ipldSchema.TypeSystem

func main() {
	var err error
	ts, err = ipld.LoadSchemaBytes([]byte(orderSchema))
	if err != nil {
		panic(err)
	}

	switch os.Args[1] {
	case "read":
		assert(len(os.Args) == 3, "read requires a file path")
		read()
	case "write":
		write()
	default:
		panic("unknown command")
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func assert(condition bool, msg string) {
	if !condition {
		panic(msg)
	}
}

func read() {
	schemaType := ts.TypeByName("Order")
	proto := bindnode.Prototype(nil, schemaType)
	builder := proto.NewBuilder()
	var err error
	assert(len(os.Args) == 3, "read requires a file path")
	if os.Args[2] == "-" {
		err = dagcbor.Decode(builder, os.Stdin)
	} else {
		var file *os.File
		file, err = os.Open(os.Args[2])
		assert(err == nil, "failed to open file")
		defer file.Close()
		err = dagcbor.Decode(builder, file)
	}
	check(err)
	node := builder.Build()
	kind := node.Kind()
	fmt.Printf("Kind: %s %T\n", kind, node)
	fmt.Printf("Length: %d\n", node.Length())
	if kind == datamodel.Kind_Map {
		it := node.MapIterator()
		for !it.Done() {
			key, val, err := it.Next()
			if err != nil {
				break
			}
			keyStr, err := key.AsString()
			check(err)
			if val.IsNull() || val.IsAbsent() {
				fmt.Printf("Key: %s (empty)\n", keyStr)
				continue
			}
			if keyStr == "TxDetails" ||
				keyStr == "InvoiceAddress" ||
				keyStr == "ShippingAddress" ||
				keyStr == "PaymentDetails" {
				it := val.MapIterator()
				if it == nil {
					fmt.Printf("%s is not a map (Kind: %s)\n", keyStr, val.Kind())
					continue
				}
				fmt.Printf("Key: %s\n", keyStr)
				for !it.Done() {
					idx, val, err := it.Next()
					check(err)

					idxStr, err := idx.AsString()
					check(err)
					fmt.Printf("%s.%s: ", keyStr, idxStr)
					if keyStr == "PaymentDetails" && idxStr == "listingHashes" {
						fmt.Printf("\n\nlistingHashes: %d\n", val.Length())
						printListingHashes(val)
					} else {
						printValue(val)
					}
				}
			} else {
				printValue(val)
			}
		}
	} else {
		fmt.Printf("Node is not a map: %s\n", kind)
	}
}

func printListingHashes(val datamodel.Node) {
	it := val.ListIterator()
	for !it.Done() {
		idx, val, err := it.Next()
		check(err)
		fmt.Printf("%d: ", idx)
		printValue(val)
	}
}

func printValue(val datamodel.Node) {
	if valStr, err := val.AsString(); err == nil {
		fmt.Printf("%q\n", valStr)
	} else if valBytes, err := val.AsBytes(); err == nil {
		fmt.Printf("%x\n", valBytes)
	} else if valInt, err := val.AsInt(); err == nil {
		fmt.Printf("%d\n", valInt)
	} else if valBool, err := val.AsBool(); err == nil {
		fmt.Printf("%t\n", valBool)
	} else if valLink, err := val.AsLink(); err == nil {
		fmt.Printf("Link: %s\n", valLink)
	} else {
		fmt.Printf("unknown type: %T\n", val)
	}
}

func write() {

	var order schema.Order
	order.Items = schema.OrderedItems{
		schema.OrderedItem{
			ListingID: 782193,
			Quantity:  1,
		},
	}
	order.PaymentDetails = &schema.PaymentDetails{
		TTL:       81238,
		PaymentID: schema.Hash{0x01, 0x02, 0x03},
		ListingHashes: []cid.Cid{
			testHash(1),
			testHash(2),
			testHash(3),
		},
	}
	schemaType := ts.TypeByName("Order")
	node := bindnode.Wrap(&order, schemaType)
	nodeRepr := node.Representation()
	err := dagcbor.Encode(nodeRepr, os.Stdout)
	assert(err == nil, "failed to encode")
}

var orderSchema = `
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
	paymentID     Bytes
	total         Int
	listingHashes [Link]
	ttl           Int
	shopSignature Bytes
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

func testHash(i uint) cid.Cid {
	h, err := mh.Sum([]byte(fmt.Sprintf("TEST-%d", i)), mh.SHA3, 8)
	check(err)
	// TODO: check what the codec number should be
	return cid.NewCidV1(666, h)
}
