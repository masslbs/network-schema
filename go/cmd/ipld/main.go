package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	ipfsFiles "github.com/ipfs/boxo/files"
	ipfsBlocks "github.com/ipfs/go-block-format"
	"github.com/ipfs/go-cid"
	ipldFormat "github.com/ipfs/go-ipld-format"
	ipldLegacy "github.com/ipfs/go-ipld-legacy"
	ipfsRpc "github.com/ipfs/kubo/client/rpc"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/linking"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	ipldSchema "github.com/ipld/go-ipld-prime/schema"
	"github.com/ipld/go-ipld-prime/storage/memstore"
	"github.com/multiformats/go-multiaddr"
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
	case "schema": // ipld schema
		assert(len(os.Args) == 3, "schema requires a file path")
		ipldSchemaRead()
	case "write":
		ipldWrite()
	case "direct":
		directCbor()
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

func directCbor() {
	var (
		rd  io.Reader
		err error
	)
	assert(len(os.Args) == 3, "decode requires a file path")
	if os.Args[2] == "-" {
		rd = os.Stdin
	} else {
		var file *os.File
		file, err = os.Open(os.Args[2])
		assert(err == nil, "failed to open file")
		defer file.Close()
		rd = file
	}
	var order schema.Order
	dec := schema.DefaultDecoder(rd)
	err = dec.Decode(&order)
	check(err)
	fmt.Printf("%+v\n", order)
}

func ipldSchemaRead() {
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
	nodeTipe, ok := node.Prototype().(ipldSchema.Type)
	if !ok {
		fmt.Printf("node prototype is not a type: %T\n", node.Prototype())
	} else {
		fmt.Printf("Type: %s\n", nodeTipe.Name())
	}
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
				fmt.Printf("\n===\nKey: %s (empty)\n", keyStr)
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
				fmt.Printf("\n===\nKey: %s\n", keyStr)
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
		printValue(node)
	}
}

// TODO: find a way to store in an ipfs node
var store = memstore.Store{}

func ipldStore() {

	lsys := cidlink.DefaultLinkSystem()

	// We want to store the serialized data somewhere.
	//  We'll use an in-memory store for this.  (It's a package scoped variable.)
	//  You can use any kind of storage system here;
	//   or if you need even more control, you could also write a function that conforms to the linking.BlockWriteOpener interface.
	lsys.SetWriteStorage(&store)

	lp := cidlink.LinkPrototype{Prefix: cid.Prefix{
		Version: 1, // Usually '1'.
		// See the multicodecs table: https://github.com/multiformats/multicodec/
		Codec:    0x71, // "dag-cbor"
		MhType:   0x1e, //  "blake3"
		MhLength: -1,
	}}

	order := testOrder()

	schemaType := ts.TypeByName("Order")
	node := bindnode.Wrap(&order, schemaType)

	// Now: time to apply the LinkSystem, and do the actual store operation!
	lnk, err := lsys.Store(
		linking.LinkContext{}, // The zero value is fine.  Configure it it you want cancellability or other features.
		lp,                    // The LinkPrototype says what codec and hashing to use.
		node.Representation(), // And here's our data.
	)
	check(err)
	// That's it!  We got a link.
	fmt.Printf("link: %s\n", lnk)
	fmt.Printf("concrete type: `%T`\n", lnk)
}

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

const SoftBlockLimit = 1024 * 1024 // https://github.com/ipfs/kubo/issues/7421#issuecomment-910833499

func ipldWrite() {
	order := testOrder()

	schemaType := ts.TypeByName("Order")
	node := bindnode.Wrap(&order, schemaType)

	var buf bytes.Buffer
	err := dagcbor.Encode(node, &buf)
	check(err)
	fmt.Printf("Hex data: %x\n", buf.Bytes())

	ctx := context.Background()
	client, err := getIpfsClient(ctx, 0, nil)
	check(err)
	var adder ipldFormat.NodeAdder = client.Dag()
	if false {
		adder = client.Dag().Pinning()
	}
	b := ipldFormat.NewBatch(ctx, adder)

	linkPrefix := cid.Prefix{
		Version:  1,    // Usually '1'.
		Codec:    0x71, // 0x71 means "dag-cbor" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhType:   0x1e, // 0x1e means "blake3" -- See the multicodecs table: https://github.com/multiformats/multicodec/
		MhLength: -1,   // blake3 hash has a variable length.
	}

	blockCid, err := linkPrefix.Sum(buf.Bytes())
	check(err)
	blk, err := ipfsBlocks.NewBlockWithCid(buf.Bytes(), blockCid)
	check(err)
	ln := ipldLegacy.LegacyNode{
		Block: blk,
		Node:  node,
	}
	if buf.Len() > SoftBlockLimit {
		check(fmt.Errorf("produced block is over 1MiB: big blocks can't be exchanged with other peers. consider using UnixFS for automatic chunking of bigger files, or pass --allow-big-block to override"))
	}

	err = b.Add(ctx, &ln)
	check(err)

	cid := ln.Cid()
	fmt.Printf("cid: %s\n", cid)
	err = b.Commit()
	check(err)
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

func testCID(i uint) cid.Cid {
	h, err := mh.Sum([]byte(fmt.Sprintf("TEST-%d", i)), mh.SHA3, 8)
	check(err)
	// TODO: check what the codec number should be
	return cid.NewCidV1(666, h)
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
	} else if val.Kind() == datamodel.Kind_Map {
		fmt.Printf("Map: %d\n", val.Length())
		it := val.MapIterator()
		for !it.Done() {
			key, val, err := it.Next()
			check(err)
			keyStr, err := key.AsString()
			check(err)
			fmt.Printf("%s: ", keyStr)
			printValue(val)
		}
	} else if val.Kind() == datamodel.Kind_List {
		fmt.Printf("List: %d\n", val.Length())
		it := val.ListIterator()
		for !it.Done() {
			idx, val, err := it.Next()
			check(err)
			fmt.Printf("%d: ", idx)
			printValue(val)
		}
	} else if val.Kind() == datamodel.Kind_Null {
		fmt.Printf("Null\n")
	} else {
		fmt.Printf("unknown type: %s\n", val.Kind())
	}
}

// IPFS integration
const ipfsMaxConnectTries = 3

// getIpfsClient recursivly calls itself until it was able to connect or until ipfsMaxConnectTries is reached.
func getIpfsClient(ctx context.Context, errCount int, lastErr error) (*ipfsRpc.HttpApi, error) {
	if errCount >= ipfsMaxConnectTries {
		return nil, fmt.Errorf("getIpfsClient: tried %d times.. last error: %w", errCount, lastErr)
	}
	if errCount > 0 {
		fmt.Printf("getIpfsClient.retrying lastErr=%s\n", lastErr)
		// TODO: exp backoff
		time.Sleep(1 * time.Second)
	}
	ipfsAPIAddr, err := multiaddr.NewMultiaddr(os.Getenv("IPFS_API_PATH"))
	if err != nil {
		// TODO: check type of error
		return getIpfsClient(ctx, errCount+1, fmt.Errorf("getIpfsClient: multiaddr.NewMultiaddr failed with %w", err))
	}
	ipfsClient, err := ipfsRpc.NewApi(ipfsAPIAddr)
	if err != nil {
		// TODO: check type of error
		return getIpfsClient(ctx, errCount+1, fmt.Errorf("getIpfsClient: ipfsRpc.NewApi failed with %w", err))
	}
	// check connectivity
	if os.Getenv("ENV") == "dev" {
		_, err := ipfsClient.Unixfs().Add(ctx, ipfsFiles.NewBytesFile([]byte("test")))
		if err != nil {
			return getIpfsClient(ctx, errCount+1, fmt.Errorf("getIpfsClient: (dev env) add 'test' failed %w", err))
		}
		//	d.Get(ctx, ipfsPath.NewPath("/ipfs/QmQg162Z6YZ4KvtVx6yLrdw3452RFSLYQXZLKQd2Y4wX5m"))

	} else {
		peers, err := ipfsClient.Swarm().Peers(ctx)
		if err != nil {
			// TODO: check type of error
			return getIpfsClient(ctx, errCount+1, fmt.Errorf("getIpfsClient: ipfsClient.Swarm.Peers failed with %w", err))
		}
		if len(peers) == 0 {
			// TODO: dial another peer
			// return getIpfsClient(ctx, errCount+1, fmt.Errorf("ipfs node has no peers"))
			fmt.Printf("getIpfsClient.warning: no peers\n")
		}
	}
	return ipfsClient, nil
}
