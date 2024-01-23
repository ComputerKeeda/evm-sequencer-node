package settlement_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/airchains-network/evm-sequencer-node/common"
	"github.com/airchains-network/evm-sequencer-node/common/logs"
	"github.com/airchains-network/evm-sequencer-node/types"
	"github.com/syndtr/goleveldb/leveldb"
	"io"
	"net/http"
	"time"
)

type VerifyBatchPostStruct struct {
	BatchNumber    uint64 `json:"batch_number"`
	ChainID        string `json:"chain_id"`
	MerkleRootHash string `json:"merkle_root_hash"`
	PrevMerkleRoot string `json:"prev_merkle_root"`
	ZkProof        []byte `json:"zk_proof"`
}

func VerifyBatch(batchNumber int, proofByte []byte, ldda *leveldb.DB, lds *leveldb.DB) bool {
	logs.Log.Warn("Verifying the batch ")
	settlementChainInfoByte, err := lds.Get([]byte("settlementChainInfo"), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting settlementChainInfo from static db : %s", err.Error()))
		return false
	}

	var settlementChainInfo types.SettlementLayerChainInfoStruct
	err = json.Unmarshal(settlementChainInfoByte, &settlementChainInfo)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in unmarshalling settlementChainInfo : %s", err.Error()))
		return false
	}
	chainID := settlementChainInfo.ChainId

	batchKey := fmt.Sprintf("batch_%d", batchNumber)
	batchDetailsByte, err := ldda.Get([]byte(batchKey), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting batch from db : %s", err.Error()))
		return false
	}

	var batchDetails types.DAStruct
	err = json.Unmarshal(batchDetailsByte, &batchDetails)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in unmarshalling batchDetails : %s", err.Error()))
		return false
	}

	postVerifyBatchStruct := VerifyBatchPostStruct{
		BatchNumber:    uint64(batchNumber),
		ChainID:        chainID,
		MerkleRootHash: batchDetails.CurrentStateHash,
		PrevMerkleRoot: batchDetails.PreviousStateHash,
		ZkProof:        proofByte,
	}

	jsonData, err := json.Marshal(postVerifyBatchStruct)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling postVerifyBatchStruct : %s", err.Error()))
		return false
	}

	rpcUrl := fmt.Sprintf("%s/verify_batch", common.SettlementClientRPC)

	req, err := http.NewRequest("POST", rpcUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return false
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logs.Log.Error(fmt.Sprintf("Error in closing response body : %s", err.Error()))
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return false
	}

	var response types.SettlementClientResponseStruct
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return false
	}

	if !response.Status {
		logs.Log.Error(fmt.Sprintf("Error in verifying batch : %s", response.Description))
		logs.Log.Warn("Trying again... in 5 seconds")
		time.Sleep(5 * time.Second)
		VerifyBatch(batchNumber, proofByte, ldda, lds)
	}
	return response.Status
}
