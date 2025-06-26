package engine

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"time"

	types2 "github.com/berachain/beacon-kit/consensus-types/types"
	engineprimitives "github.com/berachain/beacon-kit/engine-primitives/engine-primitives"
	"github.com/berachain/beacon-kit/execution/client/ethclient"
	ethclientrpc "github.com/berachain/beacon-kit/execution/client/ethclient/rpc"
	"github.com/berachain/beacon-kit/primitives/bytes"
	common2 "github.com/berachain/beacon-kit/primitives/common"
	"github.com/berachain/beacon-kit/primitives/math"
	"github.com/berachain/beacon-kit/primitives/net/jwt"
	"github.com/berachain/beacon-kit/primitives/version"
	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
)

func GetEngineAPI() (*ethclient.Client, error) {
	content, err := os.ReadFile("../geth-experimental/testdir/geth/jwtsecret")
	if err != nil {
		return nil, err
	}
	jwtSecret, err := jwt.NewFromHex(string(content))
	if err != nil {
		return nil, err
	}
	tmp := ethclientrpc.NewClient(
		"http://localhost:8551",
		jwtSecret,
		time.Second,
	)
	ec := ethclient.New(tmp)
	go ec.Start(context.Background())
	time.Sleep(1 * time.Second)

	return ec, nil
}

func ForkchoiceUpdatedV3(ec *ethclient.Client, prvBlkHash common.Hash) (*engineprimitives.ForkchoiceResponseV1, error) {
	return ec.ForkchoiceUpdatedV3(context.Background(), &engineprimitives.ForkchoiceStateV1{
		HeadBlockHash:      common2.NewExecutionHashFromHex(prvBlkHash.Hex()),
		SafeBlockHash:      common2.NewExecutionHashFromHex(prvBlkHash.Hex()),
		FinalizedBlockHash: common2.NewExecutionHashFromHex(prvBlkHash.Hex()),
	}, engineprimitives.PayloadAttributes{
		Timestamp:             math.U64(time.Now().Unix()),
		PrevRandao:            common2.Bytes32{},
		SuggestedFeeRecipient: common2.NewExecutionAddressFromHex("0x97B5853E9a4a9E64342C0bFDEFef5a5AA6ba6fe2"),
		Withdrawals:           []*engineprimitives.Withdrawal{},
		ParentBeaconBlockRoot: common2.NewRootFromBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}),
	})
}

func GetPayloadV4(ec *ethclient.Client, payloadID *engineprimitives.PayloadID) (*engine.ExecutableData, error) {
	ev, err := ec.GetPayloadV4(context.Background(), *payloadID, version.Electra())
	if err != nil {
		return nil, err
	}
	ev2 := engine.ExecutableData{
		ParentHash:       common.HexToHash(ev.GetExecutionPayload().ParentHash.Hex()),
		FeeRecipient:     common.HexToAddress(ev.GetExecutionPayload().FeeRecipient.Hex()),
		StateRoot:        common.HexToHash(ev.GetExecutionPayload().StateRoot.String()),
		ReceiptsRoot:     common.HexToHash(ev.GetExecutionPayload().ReceiptsRoot.String()),
		LogsBloom:        ev.GetExecutionPayload().LogsBloom[:],
		Random:           common.HexToHash(ev.GetExecutionPayload().Random.String()),
		Number:           uint64(ev.GetExecutionPayload().Number),
		GasLimit:         ev.GetExecutionPayload().GasLimit.Unwrap(),
		GasUsed:          ev.GetExecutionPayload().GasUsed.Unwrap(),
		Timestamp:        ev.GetExecutionPayload().Timestamp.Unwrap(),
		ExtraData:        ev.GetExecutionPayload().ExtraData,
		BaseFeePerGas:    ev.GetExecutionPayload().BaseFeePerGas.ToBig(),
		BlockHash:        common.HexToHash(ev.GetExecutionPayload().BlockHash.Hex()),
		Transactions:     [][]byte{},            // TODO.
		Withdrawals:      []*types.Withdrawal{}, // TODO.
		BlobGasUsed:      ev.GetExecutionPayload().BlobGasUsed.UnwrapPtr(),
		ExcessBlobGas:    ev.GetExecutionPayload().ExcessBlobGas.UnwrapPtr(),
		ExecutionWitness: nil,
	}
	return &ev2, nil
}

func NewPayloadV3(ec *ethclient.Client, blk *types.Block) (*engineprimitives.PayloadStatusV1, error) {
	// Now pass to EL.
	beacon := common2.NewRootFromBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	return ec.NewPayloadV3(context.Background(), &types2.ExecutionPayload{
		ParentHash:    common2.NewExecutionHashFromHex(blk.ParentHash().Hex()),
		FeeRecipient:  common2.NewExecutionAddressFromHex(blk.Coinbase().Hex()),
		StateRoot:     common2.Bytes32(blk.Root()),
		ReceiptsRoot:  common2.Bytes32(blk.ReceiptHash()),
		LogsBloom:     bytes.B256(blk.Bloom()),
		Random:        common2.Bytes32{},
		Number:        math.U64(blk.NumberU64()),
		GasLimit:      math.U64(blk.GasLimit()),
		GasUsed:       math.U64(blk.GasUsed()),
		Timestamp:     math.U64(blk.Time()),
		ExtraData:     blk.Extra(),
		BaseFeePerGas: math.NewU256(blk.BaseFee().Uint64()),
		BlockHash:     common2.ExecutionHash(blk.Hash()),
		Transactions:  make(engineprimitives.Transactions, 0),  // TODO...
		Withdrawals:   make([]*engineprimitives.Withdrawal, 0), // TODO...
		BlobGasUsed:   math.U64(*blk.BlobGasUsed()),
		ExcessBlobGas: math.U64(*blk.ExcessBlobGas()),
	}, []common2.ExecutionHash{}, &beacon)
}

func ExecutableDataToBlockWithDifficulty(data engine.ExecutableData, difficulty *big.Int, versionedHashes []common.Hash, beaconRoot *common.Hash, requests [][]byte) (*types.Block, error) {
	txs, err := decodeTransactions(data.Transactions)
	if err != nil {
		return nil, err
	}
	if len(data.LogsBloom) != 256 {
		return nil, fmt.Errorf("invalid logsBloom length: %v", len(data.LogsBloom))
	}
	// Check that baseFeePerGas is not negative or too big
	if data.BaseFeePerGas != nil && (data.BaseFeePerGas.Sign() == -1 || data.BaseFeePerGas.BitLen() > 256) {
		return nil, fmt.Errorf("invalid baseFeePerGas: %v", data.BaseFeePerGas)
	}
	var blobHashes = make([]common.Hash, 0, len(txs))
	for _, tx := range txs {
		blobHashes = append(blobHashes, tx.BlobHashes()...)
	}
	if len(blobHashes) != len(versionedHashes) {
		return nil, fmt.Errorf("invalid number of versionedHashes: %v blobHashes: %v", versionedHashes, blobHashes)
	}
	for i := 0; i < len(blobHashes); i++ {
		if blobHashes[i] != versionedHashes[i] {
			return nil, fmt.Errorf("invalid versionedHash at %v: %v blobHashes: %v", i, versionedHashes, blobHashes)
		}
	}
	// Only set withdrawalsRoot if it is non-nil. This allows CLs to use
	// ExecutableData before withdrawals are enabled by marshaling
	// Withdrawals as the json null value.
	var withdrawalsRoot *common.Hash
	if data.Withdrawals != nil {
		h := types.DeriveSha(types.Withdrawals(data.Withdrawals), trie.NewStackTrie(nil))
		withdrawalsRoot = &h
	}

	var requestsHash *common.Hash
	if requests != nil {
		h := types.CalcRequestsHash(requests)
		requestsHash = &h
	}

	header := &types.Header{
		ParentHash:       data.ParentHash,
		UncleHash:        types.EmptyUncleHash,
		Coinbase:         data.FeeRecipient,
		Root:             data.StateRoot,
		TxHash:           types.DeriveSha(types.Transactions(txs), trie.NewStackTrie(nil)),
		ReceiptHash:      data.ReceiptsRoot,
		Bloom:            types.BytesToBloom(data.LogsBloom),
		Difficulty:       big.NewInt(0),
		Number:           new(big.Int).SetUint64(data.Number),
		GasLimit:         data.GasLimit,
		GasUsed:          data.GasUsed,
		Time:             data.Timestamp,
		BaseFee:          data.BaseFeePerGas,
		Extra:            data.ExtraData,
		MixDigest:        data.Random,
		WithdrawalsHash:  withdrawalsRoot,
		ExcessBlobGas:    data.ExcessBlobGas,
		BlobGasUsed:      data.BlobGasUsed,
		ParentBeaconRoot: beaconRoot,
		RequestsHash:     requestsHash,
	}
	temp := types.NewBlockWithHeader(header).
		WithBody(types.Body{Transactions: txs, Uncles: nil, Withdrawals: data.Withdrawals}).
		WithWitness(data.ExecutionWitness)
	if temp.Hash() != data.BlockHash {
		return nil, fmt.Errorf("blockhash mismatch, want %x, got %x", data.BlockHash, temp.Hash())
	}
	header.Difficulty = difficulty
	// TODO: FIX BLOCK NONCE, Need to be calculated by clique.prepare
	header.Nonce = types.BlockNonce(hexutil.MustDecode("0xffffffffffffffff"))
	block := types.NewBlockWithHeader(header).
		WithBody(types.Body{Transactions: txs, Uncles: nil, Withdrawals: data.Withdrawals}).
		WithWitness(data.ExecutionWitness)

	return block, nil
}

func decodeTransactions(enc [][]byte) ([]*types.Transaction, error) {
	var txs = make([]*types.Transaction, len(enc))
	for i, encTx := range enc {
		var tx types.Transaction
		if err := tx.UnmarshalBinary(encTx); err != nil {
			return nil, fmt.Errorf("invalid transaction %d: %v", i, err)
		}
		txs[i] = &tx
	}
	return txs, nil
}
