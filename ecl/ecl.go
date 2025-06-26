package ecl

import (
	"context"
	"fmt"
	"time"

	"github.com/berachain/beacon-kit/execution/client/ethclient"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/wcgcyx/ecl/cliq"
	"github.com/wcgcyx/ecl/cr"
	"github.com/wcgcyx/ecl/engine"
)

// CHANGE(immutable): EmbeddedClique is used to act as an embedded consensus layer
// that drives the execution layer.
type EmbeddedClique struct {
	reader    *cr.ChainReader
	engineAPI *ethclient.Client
	clique    *cliq.Clique

	// feeRecipient is only applicable when in mining mode
	feeRecipient common.Address
}

func NewEmbeddedClique(reader *cr.ChainReader, engineAPI *ethclient.Client, clique *cliq.Clique, feeRecipient common.Address) *EmbeddedClique {
	return &EmbeddedClique{
		reader:       reader,
		engineAPI:    engineAPI,
		clique:       clique,
		feeRecipient: feeRecipient,
	}
}

// CHANGE(immutable): miningLoop is the main mining loop.
func (ecl *EmbeddedClique) MiningLoop(ctx context.Context, prvBlkHash common.Hash, prvBlkNumber uint64) {
	nextBlkNumber := prvBlkNumber + 1

	// Keep building block
	for {
		select {
		case <-ctx.Done():
			log.Warn("Stop block building process due to context being cancelled", "reason", ctx.Err())
			return
		default:
			// Start next block
			fmt.Printf("Start building block %v...\n", nextBlkNumber)
			resp, err := engine.ForkchoiceUpdatedV3(ecl.engineAPI, prvBlkHash)
			if err != nil {
				panic(err)
			}
			if resp.PayloadStatus.Status != "VALID" {
				panic("fail to fcu")
			}

			fmt.Println("Wait 2 seconds...")
			time.Sleep(2 * time.Second)

			ev, err := engine.GetPayloadV4(ecl.engineAPI, resp.PayloadID)
			if err != nil {
				panic(err)
			}

			br := common.Hash{}
			// TODO: Fix difficulty, diff can be calculated alongside with block nonce with prepare.
			diff := ecl.clique.CalcDifficulty(ecl.reader, ev.Timestamp, ecl.reader.GetHeaderByNumber(nextBlkNumber-1))
			blk, err := engine.ExecutableDataToBlockWithDifficulty(*ev, diff, []common.Hash{}, &br, nil)
			if err != nil {
				panic(err)
			}

			// Seal block
			resultCh := make(chan *types.Block, 128)
			stopCh := make(chan struct{})
			err = ecl.clique.SealECL(ecl.reader, blk, resultCh, stopCh)
			if err != nil {
				panic(err)
			}
			sealedBlk := <-resultCh
			ecl.reader.PutBlock(sealedBlk)

			// Now pass to EL.
			fmt.Printf("Ask EL to import block %v...\n", nextBlkNumber)
			resp2, err := engine.NewPayloadV3(ecl.engineAPI, sealedBlk)
			if err != nil {
				panic(err)
			}
			if resp2.Status != "VALID" {
				panic("fail to import payload")
			}

			// Broadcast block
			fmt.Println("TODO: Broadcast block...")

			prvBlkHash = sealedBlk.Hash()
			nextBlkNumber++
		}
	}
}
