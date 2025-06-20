package main

import (
	"context"
	"os"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/wcgcyx/ecl/cliq"
	"github.com/wcgcyx/ecl/cr"
	"github.com/wcgcyx/ecl/ecl"
	"github.com/wcgcyx/ecl/engine"
)

func main() {
	db, head, err := cliq.OpenCliqDB()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	reader := cr.ChainReader{
		Blocks: make(map[uint64]*types.Block),
	}
	reader.PutBlock(head)

	// Create clique
	cq := cliq.New(&params.CliqueConfig{
		Period: 2,
	}, db)
	defer cq.Close()

	content, err := os.ReadFile("../geth-experimental/testdir/keystore/UTC--2025-03-27T14-36-48.387632000Z--97b5853e9a4a9e64342c0bfdefef5a5aa6ba6fe2")
	if err != nil {
		panic(err)
	}

	ky, err := keystore.DecryptKey(content, "")
	if err != nil {
		panic(err)
	}
	cq.AuthorizeWithPrivateKey(common.HexToAddress("0x97B5853E9a4a9E64342C0bFDEFef5a5AA6ba6fe2"), func(account accounts.Account, s string, data []byte) ([]byte, error) {
		return crypto.Sign(crypto.Keccak256(data), ky.PrivateKey)
	})

	// Get Engine API
	ec, err := engine.GetEngineAPI()
	if err != nil {
		panic(err)
	}
	defer ec.Close()

	cl := ecl.NewEmbeddedClique(&reader, ec, cq, common.HexToAddress("0x97B5853E9a4a9E64342C0bFDEFef5a5AA6ba6fe2"))
	cl.MiningLoop(context.Background(), head.Hash(), head.NumberU64())
}
