package cr

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

type ChainReader struct {
	Blocks map[uint64]*types.Block
}

func (cr *ChainReader) Config() *params.ChainConfig {
	panic("throw 1")
}

// CurrentHeader retrieves the current header from the local chain.
func (cr *ChainReader) CurrentHeader() *types.Header {
	panic("throw 2")
}

// GetHeader retrieves a block header from the database by hash and number.
func (cr *ChainReader) GetHeader(hash common.Hash, number uint64) *types.Header {
	return cr.Blocks[number].Header()
}

// GetHeaderByNumber retrieves a block header from the database by number.
func (cr *ChainReader) GetHeaderByNumber(number uint64) *types.Header {
	return cr.Blocks[number].Header()
}

// GetHeaderByHash retrieves a block header from the database by its hash.
func (cr *ChainReader) GetHeaderByHash(hash common.Hash) *types.Header {
	panic("throw 3")
}

func (cr *ChainReader) PutBlock(blk *types.Block) {
	if len(cr.Blocks) >= 10 {
		delete(cr.Blocks, blk.NumberU64()-10)
	}
	cr.Blocks[blk.NumberU64()] = blk
}
