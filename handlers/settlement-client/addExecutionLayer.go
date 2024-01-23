package settlement_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	air "github.com/airchains-network/evm-sequencer-node/airdb/air-leveldb"
	"github.com/airchains-network/evm-sequencer-node/common"
	"github.com/airchains-network/evm-sequencer-node/common/logs"
	"github.com/airchains-network/evm-sequencer-node/types"
	"io"
	"net/http"
	"os"
	"time"
)

type PostAddExecutionLayerStruct struct {
	VerificationKey []byte `json:"verification_key"`
	ChainInfo       string `json:"chain_info"`
}

func AddExecutionLayer() string {

	logs.Log.Info("Adding execution layer")

	if _, err := os.Stat("verificationKey.json"); os.IsNotExist(err) {
		logs.Log.Error("Verification key not found. Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
		AddExecutionLayer()
	}

	verificationKeyContents, err := os.ReadFile("verificationKey.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return "nil"
	}

	//verificationKeyContentsAsString := string(verificationKeyContents)

	chainInfoFile, err := os.ReadFile("config/chainInfo.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		logs.Log.Error("Error reading chainInfo.json file")
		os.Exit(0)
	}

	var chainInfo types.ChainInfoStruct

	err = json.Unmarshal(chainInfoFile, &chainInfo)
	if err != nil {
		fmt.Println("Error reading file:", err)
		logs.Log.Error("Error reading chainInfo.json file")
		os.Exit(0)
	}

	chainInfoAsString, err := json.Marshal(chainInfo.ChainInfo)
	if err != nil {
		fmt.Println("Error marshalling chain info:", err)
		return "nil"
	}

	postAddExecutionLayerStruct := PostAddExecutionLayerStruct{
		//VerificationKey: verificationKeyContentsAsString,
		VerificationKey: verificationKeyContents,
		ChainInfo:       string(chainInfoAsString),
	}

	jsonData, err := json.Marshal(postAddExecutionLayerStruct)
	if err != nil {
		fmt.Println("Error marshalling postAddExecutionLayerStruct:", err)
		return "nil"
	}
	rpcUrl := fmt.Sprintf("%s/add-station", common.SettlementClientRPC)
	req, err := http.NewRequest("POST", rpcUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "nil"
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "nil"
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing body:", err)
		}
	}(resp.Body)

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
	
	fmt.Println(response.Status)
	fmt.Println(response.Data)
	fmt.Println(response.Description)

	if response.Data != "nil" && response.Data != "exist" {
		var settlementChainInfo = types.SettlementLayerChainInfoStruct{
			ChainId:   response.Data,
			ChainName: chainInfo.ChainInfo.Moniker,
		}

		settlementChainInfoBytes, err := json.Marshal(settlementChainInfo)
		if err != nil {
			fmt.Println("Error marshalling settlementChainInfo:", err)
			return "nil"
		}

		err = air.GetStaticDbInstance().Put([]byte("settlementChainInfo"), settlementChainInfoBytes, nil)
		if err != nil {
			fmt.Println("Error putting settlementChainInfo:", err)
			return "nil"
		}
	}

	return response.Data
}