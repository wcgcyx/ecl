package cliq

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/ethdb/pebble"
)

func OpenCliqDB() (ethdb.Database, *types.Block, error) {
	new, err := pebble.New("./testdir/geth/chaindata", 512, 5120, "eth/db/chaindata", false)
	if err != nil {
		panic(err)
	}
	newDb := rawdb.NewDatabase(new)

	if rawdb.ReadHeadBlockHash(newDb).Cmp(common.Hash{}) == 0 {
		fmt.Println("Head block not found, read from old db")
		old, err := pebble.New("../geth-experimental/testdir/geth/chaindata", 512, 5120, "eth/db/chaindata/", true)
		if err != nil {
			newDb.Close()
			return nil, nil, err
		}
		oldDb := rawdb.NewDatabase(old)
		defer oldDb.Close()
		headBlk := rawdb.ReadHeadBlock(oldDb)

		rawdb.WriteHeadBlockHash(newDb, headBlk.Hash())
		rawdb.WriteHeadHeaderHash(newDb, headBlk.Header().Hash())
		rawdb.WriteHeaderNumber(newDb, headBlk.Header().Hash(), headBlk.NumberU64())
		rawdb.WriteBlock(newDb, headBlk)

		fmt.Println("Write head block to be ", headBlk.Hash(), headBlk.Header().Hash(), headBlk.NumberU64())

		return newDb, headBlk, nil
	}
	headBlk := rawdb.ReadHeadBlock(newDb)
	fmt.Println("Clique DB initialised...")
	fmt.Println("Head block is ", headBlk.Hash())

	return newDb, headBlk, nil
}
