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

type PostAddBatchStruct struct {
	ChainID     string `json:"chain_id"`
	BatchNumber uint64 `json:"batch_number"`
	Witness     []byte `json:"witness"`
}

func AddBatch(witnessVector any, batchNumber int, lds *leveldb.DB) string {

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

	wvByte, _ := json.Marshal(witnessVector)

	postAddBatchStruct := PostAddBatchStruct{
		ChainID:     chainID,
		BatchNumber: uint64(batchNumber),
		Witness:     wvByte,
	}

	jsonData, err := json.Marshal(postAddBatchStruct)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("Error in marshalling postAddBatchStruct : %s", err.Error()))
		return "nil"
	}
	rpcUrl := fmt.Sprintf("%s/add_batch", common.SettlementClientRPC)
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
		AddBatch(witnessVector, batchNumber, lds)
	}

	return response.Data
}
