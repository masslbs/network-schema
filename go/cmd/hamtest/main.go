package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	hamt "github.com/ipld/go-ipld-adl-hamt"
	"github.com/ipld/go-ipld-prime"
	cidLink "github.com/ipld/go-ipld-prime/linking/cid"
	ipldCodec "github.com/ipld/go-ipld-prime/multicodec"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/storage/fsstore"
	"github.com/multiformats/go-multicodec"

	"github.com/ipld/go-ipld-prime/codec/dagcbor"
	_ "github.com/ipld/go-ipld-prime/codec/dagcbor"
)

var Linkproto = cidLink.LinkPrototype{
	Prefix: cid.Prefix{
		Version:  1,
		Codec:    uint64(multicodec.DagCbor),
		MhType:   uint64(multicodec.Blake3),
		MhLength: -1,
	},
}

var lsys = cidLink.DefaultLinkSystem()
var store fsstore.Store

func main() {
	err := store.InitDefaults("/tmp/hamt")
	check(err)

	lsys.SetWriteStorage(&store)

	if len(os.Args) < 2 {
		fmt.Println("usage: hamtest write")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "write":
		write()
	case "read":
		read()
	default:
		fmt.Println("usage: hamtest write|read")
		os.Exit(1)
	}
}

func read() {
	hamtCid, err := cid.Parse(os.Args[2])
	check(err)

	data, err := store.Get(context.Background(), hamtCid.KeyString())
	check(err)

	io.Copy(os.Stdout, bytes.NewReader(data))

	// data, err := lsys.LoadRaw(ipld.LinkContext{}, hamtCid)
	// check(err)
	// fmt.Printf("data: %x\n", data)
}

func write() {
	hb := hamt.NewBuilder(hamt.Prototype{
		BitWidth:   3,
		BucketSize: 32,
	}.WithHashAlg(multicodec.Murmur3X64_64)).WithLinking(lsys, Linkproto)

	hmap, err := hb.BeginMap(0)
	check(err)

	for i := 0; i < 66; i++ {
		hmap.AssembleKey().AssignString(fmt.Sprintf("key-%d", i))
		var randomBytes [6 * 1024]byte
		rand.Read(randomBytes[:])
		hmap.AssembleValue().AssignBytes(randomBytes[:])
	}

	err = hmap.Finish()
	check(err)

	hamtRoot, ok := hb.Build().(ipld.ADL)
	if !ok {
		panic("not an ADL node")
	}

	lnk, err := lsys.Store(ipld.LinkContext{}, Linkproto, hamtRoot.Substrate())
	check(err)

	fmt.Printf("link: %s\n", lnk)

	var buf bytes.Buffer
	err = dagcbor.Encode(hamtRoot, &buf)
	check(err)
	os.WriteFile("debug.cbor", buf.Bytes(), 0644)

}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type Ingester struct {
	dsTmp datastore.Batching
	lsys  ipld.LinkSystem
}

func (ing *Ingester) loadHamt(c cid.Cid) (*hamt.Node, error) {
	node, err := ing.loadNode(c, hamt.HashMapRootPrototype)
	if err != nil {
		return nil, err
	}
	root := bindnode.Unwrap(node).(*hamt.HashMapRoot)
	if root == nil {
		return nil, errors.New("cannot unwrap node as hamt.HashMapRoot")
	}
	hn := hamt.Node{
		HashMapRoot: *root,
	}.WithLinking(ing.lsys, Linkproto)
	return hn, nil
}

func (ing *Ingester) loadNode(c cid.Cid, prototype ipld.NodePrototype) (ipld.Node, error) {
	val, err := ing.dsTmp.Get(context.Background(), datastore.NewKey(c.String()))
	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			return nil, errors.New("node not present on indexer")
		}
		return nil, fmt.Errorf("cannot fetch the node from datastore: %w", err)
	}
	node, err := decodeIPLDNode(c.Prefix().Codec, bytes.NewBuffer(val), prototype)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ipldNode: %w", err)
	}
	return node, nil

}

// decodeIPLDNode decodes an ipld.Node from bytes read from an io.Reader.
func decodeIPLDNode(codec uint64, r io.Reader, prototype ipld.NodePrototype) (ipld.Node, error) {
	// NOTE: Considering using the schema prototypes. This was failing, using a
	// map gives flexibility. Maybe is worth revisiting this again in the
	// future.
	nb := prototype.NewBuilder()
	decoder, err := ipldCodec.LookupDecoder(codec)
	if err != nil {
		return nil, err
	}
	err = decoder(nb, r)
	if err != nil {
		return nil, err
	}
	return nb.Build(), nil
}
