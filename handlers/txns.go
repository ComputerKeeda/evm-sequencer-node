package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	evmcommon "github.com/airchains-network/evm-sequencer-node/common"
	"github.com/airchains-network/evm-sequencer-node/common/logs"
	evmtypes "github.com/airchains-network/evm-sequencer-node/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
)

func insertTxn(db *leveldb.DB, txns evmtypes.TransactionStruct, transactionNumber int) error {

	println("insertTxn")
	data, err := json.Marshal(txns)
	if err != nil {
		return err
	}

	txnsKey := fmt.Sprintf("txns-%d", transactionNumber+1)
	err = db.Put([]byte(txnsKey), data, nil)
	if err != nil {
		return err
	}
	err = os.WriteFile("data/transactionCount.txt", []byte(strconv.Itoa(transactionNumber+1)), 0666)
	if err != nil {
		return err
	}

	return nil
}

func SaveTxns(client *ethclient.Client, ctx context.Context, ldt *leveldb.DB, transactionHash string, blockNumber int, blockHash string) {
	blockNumberUint64, err := strconv.ParseUint(strconv.Itoa(blockNumber), 10, 64)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error parsing block number to uint64:"))
		time.Sleep(2 * time.Second)
		logs.Log.Info("Retrying in 2s...")
		SaveTxns(client, ctx, ldt, transactionHash, blockNumber, blockHash)
	}

	txHash := common.HexToHash(transactionHash)
	tx, isPending, err := client.TransactionByHash(ctx, txHash)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to get transaction by hash: %s", err))
		os.Exit(0)
	}

	if isPending {
		logs.Log.Warn("Transaction is pending")
		logs.Log.Info(fmt.Sprintf("Transaction type: %d\n", tx.Type()))
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to get the network ID: %v", err))
		os.Exit(0)
	}
	msg, err := types.Sender(types.NewLondonSigner(chainID), tx)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to derive the sender address: %v", err))
		os.Exit(0)
	}

	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to fetch the transaction receipt: %v", err))
		os.Exit(0)
	}
	// fmt.Println(tx)
	// v, r, s :=
	// v *big.Int, r *big.Int, s *big.Int

	fmt.Println("Error testing")
	var v, r, s = tx.RawSignatureValues()
	fmt.Println("Error testing")

	// if v == nil || r == nil || s == nil {
	// 	log.Fatal("Transaction has an invalid signature")
	// }

	if tx.To() == nil {
		// Contract creation
		fmt.Printf("Contract created at %s\n", msg.Hex())
	} else {
		fmt.Printf("Transaction to: %s\n", tx.To().Hex())
	}

	// fmt.Printf("BlockHash: %s\n", blockHash)
	// fmt.Printf("BlockNumber: %d\n", blockNumberUint64)
	// fmt.Printf("From: %s\n", msg.Hex())
	// fmt.Printf("Gas: %s\n", evmcommon.ToString(tx.Gas()))
	// fmt.Printf("GasPrice: %s\n", tx.GasPrice().String())
	// fmt.Printf("Hash: %s\n", tx.Hash().Hex())
	// fmt.Printf("Input: %s\n", string(tx.Data()))
	// fmt.Printf("Nonce: %s\n", evmcommon.ToString(tx.Nonce()))
	// fmt.Printf("R: %s\n", r.String())
	// fmt.Printf("S: %s\n", s.String())
	// fmt.Printf("To: %s\n", tx.To().Hex())
	// fmt.Printf("TransactionIndex: %s\n", evmcommon.ToString(receipt.TransactionIndex))
	// fmt.Printf("Type: %d\n", tx.Type())
	// fmt.Printf("V: %s\n", v.String())
	// fmt.Printf("Value: %s\n", tx.Value().String())

	var toValue string

	if tx.To() == nil {
		// Contract creation
		toValue = msg.Hex()
	} else {
		toValue = tx.To().Hex()
	}

	txData := evmtypes.TransactionStruct{
		BlockHash:        blockHash,
		BlockNumber:      blockNumberUint64,
		From:             msg.Hex(),
		Gas:              evmcommon.ToString(tx.Gas()),
		GasPrice:         tx.GasPrice().String(),
		Hash:             tx.Hash().Hex(),
		Input:            string(tx.Data()),
		Nonce:            evmcommon.ToString(tx.Nonce()),
		R:                r.String(),
		S:                s.String(),
		To:               toValue,
		TransactionIndex: evmcommon.ToString(receipt.TransactionIndex),
		Type:             fmt.Sprintf("%d", tx.Type()),
		V:                v.String(),
		Value:            tx.Value().String(),
	}

	fileOpen, err := os.Open("data/transactionCount.txt")
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to read file: %s" + err.Error()))
		os.Exit(0)
	}
	defer fileOpen.Close()

	scanner := bufio.NewScanner(fileOpen)

	transactionNumberBytes := ""

	for scanner.Scan() {
		transactionNumberBytes = scanner.Text()
	}

	transactionNumber, err := strconv.Atoi(strings.TrimSpace(string(transactionNumberBytes)))
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Invalid transaction number : %s" + err.Error()))
		os.Exit(0)
	}

	insetTxnErr := insertTxn(ldt, txData, transactionNumber)
	if insetTxnErr != nil {
		logs.Log.Error(fmt.Sprintf("Failed to insert transaction: %s" + insetTxnErr.Error()))
		os.Exit(0)
	}

	logs.Log.Debug(fmt.Sprintf("Successfully saved Transation %s in the latest phase", txHash))

}
