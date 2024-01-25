package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/airchains-network/evm-sequencer-node/common"
	"github.com/airchains-network/evm-sequencer-node/common/logs"
	settlement_client "github.com/airchains-network/evm-sequencer-node/handlers/settlement-client"
	"github.com/airchains-network/evm-sequencer-node/prover"
	"github.com/airchains-network/evm-sequencer-node/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
)

// ConvertEtherToWei converts a string representation of an Ether amount to a string representation of Wei.
func ConvertEtherToWei(etherStr string) (string, error) {
	// Parse the Ether amount
	ether := new(big.Float)
	_, ok := ether.SetString(etherStr)
	if !ok {
		return "", fmt.Errorf("invalid ether amount")
	}

	// Define the conversion factor (10^18)
	weiMultiplier := new(big.Float).SetInt(big.NewInt(1e18))

	// Multiply the Ether amount by the conversion factor to get Wei
	wei := new(big.Float).Mul(ether, weiMultiplier)

	// Convert the Wei amount to an integer
	weiInt, _ := wei.Int(nil)

	return weiInt.String(), nil
}

func BatchGeneration(client *ethclient.Client, ctx context.Context, lds *leveldb.DB, ldt *leveldb.DB, ldbatch *leveldb.DB, ldda *leveldb.DB, batchStartIndex []byte) {
	// defer wg.Done()

	limit, err := lds.Get([]byte("batchCount"), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting batchCount from static db : %s", err.Error()))
		os.Exit(0)
	}
	limitInt, _ := strconv.Atoi(strings.TrimSpace(string(limit)))
	batchStartIndexInt, _ := strconv.Atoi(strings.TrimSpace(string(batchStartIndex)))

	var batch types.BatchStruct

	var From []string
	var To []string
	var Amounts []string
	var TransactionHash []string
	var SenderBalances []string
	var ReceiverBalances []string
	var Messages []string
	var TransactionNonces []string
	var AccountNonces []string

	for i := batchStartIndexInt; i < (common.BatchSize * (limitInt + 1)); i++ {
		findKey := fmt.Sprintf("txns-%d", i+1)
		txData, err := ldt.Get([]byte(findKey), nil)
		if err != nil {
			i--
			time.Sleep(1 * time.Second)
			continue
		}
		var tx types.TransactionStruct
		err = json.Unmarshal(txData, &tx)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in unmarshalling tx data : %s", err.Error()))
			os.Exit(0)
		}

		senderBalanceInEtherCheck, err := common.GetBalance(tx.From, (tx.BlockNumber - 1))
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting sender balance : %s", err.Error()))
			os.Exit(0)
		}
		// convert senderbalance from ether to wei
		senderBalancesCheck, err := ConvertEtherToWei(senderBalanceInEtherCheck)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in converting sender balance : %s", err.Error()))
			os.Exit(0)
		}

		receiverBalancesEtherCheck, err := common.GetBalance(tx.To, (tx.BlockNumber - 1))
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting reciver balance : %s", err.Error()))
			os.Exit(0)
		}
		// convert receiverbalance from ether to wei
		receiverBalancesCheck, err := ConvertEtherToWei(receiverBalancesEtherCheck)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in converting receiver balance : %s", err.Error()))
			os.Exit(0)
		}

		accountNouceCheck, err := common.GetAccountNonce(ctx, tx.Hash, tx.BlockNumber)
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in getting account nonce : %s", err.Error()))
			os.Exit(0)
		}

		From = append(From, tx.From)
		To = append(To, tx.To)
		TransactionHash = append(TransactionHash, tx.Hash)

		Amounts = append(Amounts, tx.Value)
		SenderBalances = append(SenderBalances, senderBalancesCheck)
		ReceiverBalances = append(ReceiverBalances, receiverBalancesCheck)

		fmt.Println(senderBalancesCheck)
		fmt.Println(receiverBalancesCheck)

		Messages = append(Messages, tx.Input)
		TransactionNonces = append(TransactionNonces, tx.Nonce)
		AccountNonces = append(AccountNonces, accountNouceCheck)
	}

	batch.From = From
	batch.To = To
	batch.Amounts = Amounts
	batch.TransactionHash = TransactionHash
	batch.SenderBalances = SenderBalances
	batch.ReceiverBalances = ReceiverBalances
	batch.Messages = Messages
	batch.TransactionNonces = TransactionNonces
	batch.AccountNonces = AccountNonces

	witnessVector, currentStatusHash, proofByte, pkErr := prover.GenerateProof(batch, limitInt+1)
	if pkErr != nil {
		logs.Log.Error(fmt.Sprintf("Error in generating proof : %s", pkErr.Error()))
		os.Exit(0)
	}

	daKeyHash, err := DaCall(batch.TransactionHash, client, ctx, currentStatusHash, limitInt+1, ldda)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in adding Da client : %s", err.Error()))
		logs.Log.Warn("Trying again...")
		time.Sleep(3 * time.Second)
		BatchGeneration(client, ctx, lds, ldt, ldbatch, ldda, []byte(strconv.Itoa(common.BatchSize*(limitInt+1))))
	}

	logs.Log.Warn(fmt.Sprintf("Successfully added Da client for Batch %s in the latest phase", daKeyHash))

	currentTime := uint64(time.Now().Unix())
	addBatchRes := settlement_client.AddBatch(witnessVector, limitInt+1, currentStatusHash, currentTime, lds, ldda)
	if addBatchRes == "nil" {
		logs.Log.Error(fmt.Sprintf("Error in adding batch to settlement client : %s", addBatchRes))
		os.Exit(0)
	}

	status := settlement_client.VerifyBatch(limitInt+1, proofByte, ldda, lds)
	if !status {
		logs.Log.Error(fmt.Sprintf("Error in verifying batch to settlement client : "))
		os.Exit(0)
	}

	logs.Log.Warn(fmt.Sprintf("Successfully generated proof for Batch %s in the latest phase", strconv.Itoa(limitInt+1)))

	batchJSON, err := json.Marshal(batch)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling batch data : %s", err.Error()))
		os.Exit(0)
	}

	batchKey := fmt.Sprintf("batch-%d", limitInt+1)
	err = ldbatch.Put([]byte(batchKey), batchJSON, nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in writing batch data to file : %s", err.Error()))
		os.Exit(0)
	}

	err = lds.Put([]byte("batchStartIndex"), []byte(strconv.Itoa(common.BatchSize*(limitInt+1))), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating batchStartIndex in static db : %s", err.Error()))
		os.Exit(0)
	}

	err = lds.Put([]byte("batchCount"), []byte(strconv.Itoa(limitInt+1)), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in updating batchCount in static db : %s", err.Error()))
		os.Exit(0)
	}

	err = os.WriteFile("data/batchCount.txt", []byte(strconv.Itoa(limitInt+1)), 0666)
	if err != nil {
		panic("Failed to update batch number: " + err.Error())
	}

	logs.Log.Warn(fmt.Sprintf("Successfully saved Batch %s in the latest phase", strconv.Itoa(limitInt+1)))

	BatchGeneration(client, ctx, lds, ldt, ldbatch, ldda, []byte(strconv.Itoa(common.BatchSize*(limitInt+1))))
}
