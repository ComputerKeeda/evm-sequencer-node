package settlement_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/airchains-network/evm-sequencer-node/common"
	"github.com/airchains-network/evm-sequencer-node/common/logs"
	"github.com/airchains-network/evm-sequencer-node/types"
	"github.com/syndtr/goleveldb/leveldb"
)

type PostAddBatchStruct struct {
	StationId              string `json:"station_id"`
	PodNumber              uint64 `json:"pod_number"`
	MerkleRootHash         string `json:"merkle_root_hash"`
	PreviousMerkleRootHash string `json:"previous_merkle_root_hash"`
	PublicWitness          []byte `json:"public_witness"`
	Timestamp              uint64 `json:"timestamp"`
}

func AddBatch(witnessVector any, batchNumber int, mrh string, timestamp uint64, lds *leveldb.DB, ldda *leveldb.DB) string {

	fmt.Println("Creating batch: ", batchNumber)

	logs.Log.Warn("Submitting batch to settlement")

	settlementChainInfoByte, err := lds.Get([]byte("settlementChainInfo"), nil)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in getting settlementChainInfo from static db : %s", err.Error()))
		return "nil"
	}

	var settlementChainInfo types.SettlementLayerChainInfoStruct
	err = json.Unmarshal(settlementChainInfoByte, &settlementChainInfo)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in unmarshalling settlementChainInfo : %s", err.Error()))
		return "nil"
	}
	chainID := settlementChainInfo.ChainId

	//var witnessFrVector fr.Vector
	//witnessFrVector = witnessVector.(fr.Vector)
	wvByte, _ := json.Marshal(witnessVector)

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

	var pMrh string

	if batchNumber < 2 {
		pMrh = "0"
	} else {
		pMrh = daDecode.CurrentStateHash
	}

	postAddBatchStruct := PostAddBatchStruct{
		StationId:              chainID,
		PodNumber:              uint64(batchNumber),
		MerkleRootHash:         mrh,
		PreviousMerkleRootHash: pMrh,
		PublicWitness:          wvByte,
		Timestamp:              timestamp,
	}

	jsonData, err := json.Marshal(postAddBatchStruct)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling postAddBatchStruct : %s", err.Error()))
		return "nil"
	}
	rpcUrl := fmt.Sprintf("%s/add-pod", common.SettlementClientRPC)
	req, err := http.NewRequest("POST", rpcUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "nil"
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "nil"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "nil"
	}

	var response types.SettlementClientResponseStruct
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return "nil"
	}

	if !response.Status {
		logs.Log.Error("error in adding batch to settlement")
		logs.Log.Warn("Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
		AddBatch(witnessVector, batchNumber, mrh, timestamp, lds, ldda)
	}

	return response.Data
}
