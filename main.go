package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	air "github.com/airchains-network/evm-sequencer-node/airdb/air-leveldb"
	"github.com/airchains-network/evm-sequencer-node/common"
	"github.com/airchains-network/evm-sequencer-node/common/logs"
	"github.com/airchains-network/evm-sequencer-node/handlers"
	settlement_client "github.com/airchains-network/evm-sequencer-node/handlers/settlement-client"
	"github.com/airchains-network/evm-sequencer-node/prover"
	"github.com/airchains-network/evm-sequencer-node/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

func main() {
	logs.Log.Info("Starting EVM Sequencer")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	daClientRPC := os.Getenv("DA_CLIENT_RPC")

	if daClientRPC == "" {
		logs.Log.Error("Please set all the environment variables")
		os.Exit(0)
	}

	ctx := context.Background()
	_ = ctx
	dbStatus := air.InitDb()
	if !dbStatus {
		logs.Log.Error("Error in initializing db")
		os.Exit(0)
	}

	prover.CreateVkPk()
	chainId := settlement_client.AddExecutionLayer()
	if chainId == "nil" {
		logs.Log.Error("Something went wrong while adding execution layer")
		logs.Log.Warn("Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
		_ = settlement_client.AddExecutionLayer()
	} else if chainId == "exist" {
		logs.Log.Info("Chain already exist")
	}

	ldt := air.GetTxDbInstance()
	ldb := air.GetBlockDbInstance()
	lds := air.GetStaticDbInstance()
	ldbatch := air.GetBatchesDbInstance()
	ldda := air.GetDaDbInstance()

	da := types.DAStruct{
		DAKey:             "0",
		DAClientName:      "0",
		BatchNumber:       "0",
		PreviousStateHash: "0",
		CurrentStateHash:  "0",
	}

	daBytes, err := json.Marshal(da)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling da : %s", err.Error()))
		os.Exit(0)
	}

	_, err = ldda.Get([]byte("batch_0"), nil)
	if err != nil {
		err = ldda.Put([]byte("batch_0"), daBytes, nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving da in static db : %s", err.Error()))
			os.Exit(0)
		}
	}

	batchStartIndex, err := lds.Get([]byte("batchStartIndex"), nil)
	if err != nil {
		err = lds.Put([]byte("batchStartIndex"), []byte("0"), nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving batchStartIndex in static db : %s", err.Error()))
			os.Exit(0)
		}
	}

	client, err := ethclient.Dial(common.ExecutionClientRPC)
	if err != nil {
		log.Fatal("Failed to connect to the Ethereum client:", err)
	}

	_, err = lds.Get([]byte("batchCount"), nil)
	if err != nil {
		err = lds.Put([]byte("batchCount"), []byte("0"), nil)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in saving batchCount in static db : %s", err.Error()))
			os.Exit(0)
		}
	}

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		handlers.BlockCheck(&wg, ctx, client, ldb, ldt)
	}()
	go func() {
		defer wg.Done()
		handlers.BatchGeneration(&wg, client, ctx, lds, ldt, ldbatch, ldda, batchStartIndex)
	}()
	wg.Wait()
}
