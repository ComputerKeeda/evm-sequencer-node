package handlers

import (
	"bufio"
	"context"
	"github.com/airchains-network/evm-sequencer-node/common"
	"github.com/airchains-network/evm-sequencer-node/common/logs"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

func BlockCheck( ctx context.Context, client *ethclient.Client, ldb *leveldb.DB, ldt *leveldb.DB) {
	// defer wg.Done()

	fileOpen, err := os.Open("data/blockCount.txt")
	if err != nil {
		logs.Log.Error("Failed to read file: %s" + err.Error())
		os.Exit(0)
	}
	defer fileOpen.Close()

	scanner := bufio.NewScanner(fileOpen)

	blockNumberBytes := ""

	for scanner.Scan() {
		blockNumberBytes = scanner.Text()
	}

	blockNumber, blockNumberErr := strconv.Atoi(strings.TrimSpace(string(blockNumberBytes)))
	if blockNumberErr != nil {
		logs.Log.Error("Invalid block number : %s" + blockNumberErr.Error())
		os.Exit(0)
	}

	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	latestBlock := int(header.Number.Int64())

	if latestBlock == blockNumber {
		time.Sleep(common.BlockDelay * time.Second)
		logs.Log.Info("Block number is same as latest block number : " + strconv.Itoa(blockNumber))
		logs.Log.Info("Waiting for " + strconv.Itoa(common.BlockDelay) + " seconds")
		BlockCheck(ctx, client, ldb, ldt)
	} else {
		BlockSave(client, ctx, blockNumber, ldb, ldt)
	}
}
