package main

import (
	"context"
	"os"

	"github.com/ipfs/go-blockservice"
	"github.com/ipfs/go-datastore"
	dss "github.com/ipfs/go-datastore/sync"
	bstore "github.com/ipfs/go-ipfs-blockstore"
	chunker "github.com/ipfs/go-ipfs-chunker"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	files "github.com/ipfs/go-ipfs-files"
	ipldformat "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-merkledag"
	"github.com/ipfs/go-unixfs/importer/balanced"
	ihelper "github.com/ipfs/go-unixfs/importer/helpers"
	"github.com/ipld/go-car"
	ipldfree "github.com/ipld/go-ipld-prime/impl/free"
	"github.com/ipld/go-ipld-prime/traversal/selector"
	"github.com/ipld/go-ipld-prime/traversal/selector/builder"
)

const UnixfsChunkSize uint64 = 1 << 10
const UnixfsLinksPerLevel = 1024

type CarGenerator struct {
	InputFile  string
	OutputFile string
}

func (cg CarGenerator) Generate() error {
	ctx := context.Background()

	// make a blockstore and dag service
	bs1 := bstore.NewBlockstore(dss.MutexWrap(datastore.NewMapDatastore()))
	dagService1 := merkledag.NewDAGService(blockservice.New(bs1, offline.Exchange(bs1)))

	f, err := os.Open(cg.InputFile)
	if err != nil {
		return err
	}

	file := files.NewReaderFile(f)

	// import to UnixFS
	bufferedDS := ipldformat.NewBufferedDAG(ctx, dagService1)

	params := ihelper.DagBuilderParams{
		Maxlinks:   UnixfsLinksPerLevel,
		RawLeaves:  true,
		CidBuilder: nil,
		Dagserv:    bufferedDS,
	}

	db, err := params.New(chunker.NewSizeSplitter(file, int64(UnixfsChunkSize)))
	if err != nil {
		return err
	}

	nd, err := balanced.Layout(db)
	if err != nil {
		return err
	}

	err = bufferedDS.Commit()
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	ssb := builder.NewSelectorSpecBuilder(ipldfree.NodeBuilder())
	selector := ssb.ExploreRecursive(selector.RecursionLimitNone(), ssb.ExploreAll(ssb.ExploreRecursiveEdge())).Node()
	carGen := car.NewSelectiveCar(ctx, bs1, []car.Dag{
		{
			Root:     nd.Cid(),
			Selector: selector,
		},
	})

	of, err := os.Create(cg.OutputFile)
	if err != nil {
		return err
	}
	err = carGen.Write(of)
	if err != nil {
		return err
	}
	return of.Close()
}
