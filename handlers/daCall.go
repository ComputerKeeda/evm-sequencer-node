package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	air "github.com/airchains-network/evm-sequencer-node/airdb/air-leveldb"
	"github.com/airchains-network/evm-sequencer-node/common"
	"github.com/airchains-network/evm-sequencer-node/common/logs"
	"github.com/airchains-network/evm-sequencer-node/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/syndtr/goleveldb/leveldb"
	"net/http"
	"os"
	"strconv"
	"time"
)

func DaCall(transactions []string, ethClient *ethclient.Client, ctx context.Context, currentStateHash string, batchNumber int, ldda *leveldb.DB) (string, error) {
	logs.Log.Warn("DA Calling")
	proofGet, proofGetErr := air.GetProofDbInstance().Get([]byte(fmt.Sprintf("proof_%d", batchNumber)), nil)
	if proofGetErr != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting proof from db : %s", proofGetErr.Error()))
		os.Exit(0)
	}

	var proofDecode types.ProofStruct

	proofDecodeErr := json.Unmarshal(proofGet, &proofDecode)
	if proofDecodeErr != nil {
		logs.Log.Error(fmt.Sprintf("Error in unmarshalling proof : %s", proofDecodeErr.Error()))
		os.Exit(0)
	}

	daGet, daGetErr := ldda.Get([]byte(fmt.Sprintf("batch_%d", batchNumber-1)), nil)
	if daGetErr != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting da from db : %s", daGetErr.Error()))
		os.Exit(0)
	}

	var daDecode types.DAStruct
	daDecodeErr := json.Unmarshal(daGet, &daDecode)
	if daDecodeErr != nil {
		logs.Log.Error(fmt.Sprintf("Error in unmarshalling da : %s", daDecodeErr.Error()))
		os.Exit(0)
	}

	chainID, err := ethClient.NetworkID(ctx)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Failed to get the network ID: %v", err))
		os.Exit(0)
	}

	DaStruct := types.DAUploadStruct{
		Proof:             proofDecode,
		TxnHashes:         transactions,
		CurrentStateHash:  currentStateHash,
		PreviousStateHash: daDecode.PreviousStateHash,
		MetaData: struct {
			ChainID     string `json:"chainID"`
			BatchNumber int    `json:"batchNumber"`
		}{
			ChainID:     chainID.String(),
			BatchNumber: batchNumber,
		},
	}

	payloadJSON, payloadJSONErr := json.Marshal(DaStruct)

	if payloadJSONErr != nil {
		fmt.Println("Error in marshalling payload")
		return "", payloadJSONErr
		//logs.LogMessage("ERROR:", fmt.Sprintf("Failed to read file: %s"+payloadJSONErr.Error()))
		//os.Exit(0)
	}

	client := &http.Client{}

	req, reqErr := http.NewRequest("POST", common.DaClientRPC, bytes.NewBuffer(payloadJSON))
	if reqErr != nil {
		return "", reqErr
		//logs.LogMessage("ERROR:", fmt.Sprintf("Resquesting in DA RPC : %s"+reqErr.Error()))
		//os.Exit(0)
	}

	req.Header.Set("Content-Type", "application/json")
	res, resErr := client.Do(req)
	if resErr != nil {
		fmt.Println("Error in requesting")
		return "", resErr
		//logs.LogMessage("ERROR:", fmt.Sprintf("Resquesting in DA RPC : %s"+resErr.Error()))
		//os.Exit(0)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		logs.Log.Error(fmt.Sprintf("DA RPC is not responding, retrying in 1 second"))
		time.Sleep(3 * time.Second)
		_, _ = DaCall(transactions, ethClient, ctx, currentStateHash, batchNumber, ldda)
	}

	var response types.DAResponseStruct
	decodeErr := json.NewDecoder(res.Body).Decode(&response)
	if decodeErr != nil {
		fmt.Println("Error in decoding")
		return "", decodeErr
	}

	if response.DaKeyHash == "nil" {
		logs.Log.Error(fmt.Sprintf("DA RPC is not responding, retrying in 1 second"))
		time.Sleep(3 * time.Second)
		_, _ = DaCall(transactions, ethClient, ctx, currentStateHash, batchNumber, ldda)
	}

	da := types.DAStruct{
		DAKey:             response.DaKeyHash,
		DAClientName:      "celestia",
		BatchNumber:       strconv.Itoa(batchNumber),
		PreviousStateHash: daDecode.PreviousStateHash,
		CurrentStateHash:  currentStateHash,
	}

	daBytes, err := json.Marshal(da)

	batchKey := fmt.Sprintf("batch_%d", batchNumber)
	err = ldda.Put([]byte(batchKey), daBytes, nil)
	if err != nil {
		fmt.Println("Error in putting da in db")
		return "", err
	}

	return response.DaKeyHash, nil
}
