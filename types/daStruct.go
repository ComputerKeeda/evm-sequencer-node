package types

type ProofStruct struct {
	Ar struct {
		X string `json:"X"`
		Y string `json:"Y"`
	} `json:"Ar"`
	Krs struct {
		X string `json:"X"`
		Y string `json:"Y"`
	} `json:"Krs"`
	Bs struct {
		X struct {
			A0 string `json:"A0"`
			A1 string `json:"A1"`
		} `json:"X"`
		Y struct {
			A0 string `json:"A0"`
			A1 string `json:"A1"`
		} `json:"Y"`
	} `json:"Bs"`
	Commitments   []interface{} `json:"Commitments"`
	CommitmentPok struct {
		X int `json:"X"`
		Y int `json:"Y"`
	} `json:"CommitmentPok"`
}

type DAUploadStruct struct {
	Proof             ProofStruct `json:"proof"`
	TxnHashes         []string    `json:"txnHashes"`
	CurrentStateHash  string      `json:"currentStateHash"`
	PreviousStateHash string      `json:"previousStateHash"`
	MetaData          struct {
		ChainID     string `json:"chainID"`
		BatchNumber int    `json:"batchNumber"`
	} `json:"metaData"`
}

type DAResponseStruct struct {
	Status    int    `json:"status"`
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	DaKeyHash string `json:"daKeyHash"`
}
