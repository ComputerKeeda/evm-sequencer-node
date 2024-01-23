package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/airchains-network/evm-sequencer-node/common"
	"github.com/airchains-network/evm-sequencer-node/common/logs"
	"github.com/airchains-network/evm-sequencer-node/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
)

func BlockSave(client *ethclient.Client, ctx context.Context, blockIndex int, ldb *leveldb.DB, ldt *leveldb.DB) {
	blockData, err := client.BlockByNumber(ctx, big.NewInt(int64(blockIndex)))
	if err != nil {
		errMessage := fmt.Sprintf("Failed to get block data for block number %d: %s", blockIndex, err)
		logs.Log.Error(errMessage)
		logs.Log.Info("Waiting for the next block..")
		time.Sleep(common.BlockDelay * time.Second)
		BlockSave(client, ctx, blockIndex, ldb, ldt)
	}

	block := types.BlockStruct{
		BaseFeePerGas:    common.ToString(blockData.Header().BaseFee),
		Difficulty:       common.ToString(blockData.Difficulty().String()),
		ExtraData:        common.ToString(blockData.Extra()),
		GasLimit:         common.ToString(blockData.GasLimit()),
		GasUsed:          common.ToString(blockData.GasUsed()),
		Hash:             common.ToString(blockData.Hash().String()),
		LogsBloom:        common.ToString(blockData.Bloom()),
		Miner:            common.ToString(blockData.Coinbase().String()),
		MixHash:          common.ToString(blockData.MixDigest().String()),
		Nonce:            common.ToString(blockData.Nonce()),
		Number:           common.ToString(blockData.Number().String()),
		ParentHash:       common.ToString(blockData.ParentHash().String()),
		ReceiptsRoot:     common.ToString(blockData.ReceiptHash().String()),
		Sha3Uncles:       common.ToString(blockData.UncleHash()),
		Size:             common.ToString(blockData.Size()),
		StateRoot:        common.ToString(blockData.Root().String()),
		Timestamp:        common.ToString(blockData.Time()),
		TotalDifficulty:  common.ToString(blockData.Difficulty().String()),
		TransactionCount: blockData.Transactions().Len(), // Assuming transactionCount is already defined and holds the appropriate value
		TransactionsRoot: common.ToString(blockData.TxHash().String()),
		Uncles:           common.ToString(blockData.Uncles()),
	}

	data, err := json.Marshal(block)
	if err != nil {
		errMessage := fmt.Sprintf("Error marshalling block data: %s", err)
		logs.Log.Error(errMessage)
	}
	key := fmt.Sprintf("block_%s", block.Number)
	err = ldb.Put([]byte(key), data, nil)
	if err != nil {
		errMessage := fmt.Sprintf("Error inserting block data into database: %s", err)
		logs.Log.Error(errMessage)
	}

	transactions := blockData.Transactions()
	if transactions == nil {
		fmt.Println("No transactions found in block number", blockIndex)
	}

	infoMessage := fmt.Sprintf("Block number %d has %d transactions", blockIndex, transactions.Len())
	logs.Log.Info(infoMessage)

	for i := 0; i < block.TransactionCount; i++ {
		fmt.Println("Transaction number", i+1, "of block number", blockIndex)
		SaveTxns(client, ctx, ldt, transactions[i].Hash().String(), blockIndex, block.Hash)
		fmt.Println("code have stuck above probably")
	}

	err = os.WriteFile("data/blockCount.txt", []byte(strconv.Itoa(blockIndex+1)), 0666)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in saving blockCount in static db : %s", err.Error()))
		os.Exit(0)
	}

	BlockSave(client, ctx, blockIndex+1, ldb, ldt)
}
