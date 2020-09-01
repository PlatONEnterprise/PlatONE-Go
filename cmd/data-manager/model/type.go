package model

type Block struct {
	Hash       string `json:"hash" bson:"hash"`
	Height     uint64 `json:"height" bson:"height"`
	Timestamp  int64  `json:"timestamp" bson:"timestamp"`
	TxAmount   uint64 `json:"tx_amount" bson:"tx_amount"`
	Proposer   string `json:"proposer" bson:"proposer"`
	GasUsed    uint64 `json:"gas_used" bson:"gas_used"`
	GasLimit   uint64 `json:"gas_limit" bson:"gas_limit"`
	ParentHash string `json:"parent_hash" bson:"parent_hash"`
	ExtraData  string `json:"extra_data" bson:"extra_data"`
}

type Tx struct {
	Hash      string   `json:"tx_hash" bson:"tx_hash"`
	Height    uint64   `json:"block_height" bson:"block_height"`
	Timestamp int64    `json:"timestamp" bson:"timestamp"`
	From      string   `json:"from" bson:"from"`
	To        string   `json:"to" bson:"to"`
	GasLimit  uint64   `json:"gas_limit" bson:"gas_limit"`
	GasPrice  uint64   `json:"gas_price" bson:"gas_price"`
	Nonce     string   `json:"nonce" bson:"nonce"`
	Input     string   `json:"input" bson:"input"`
	Typ       uint64   `json:"tx_type" bson:"tx_type"`
	Value     uint64   `json:"value" bson:"value"`
	Receipt   *Receipt `json:"receipt" bson:"receipt"`
}

type Receipt struct {
	ContractAddress string `json:"contract_address" bson:"contract_address"`
	Status          uint64 `json:"status" bson:"status"`
	Event           string `json:"event" bson:"event"`
	GasUsed         uint64 `json:"gas_used" bson:"gas_used"`
}

type Node struct {
	Name       string `json:"name"`
	PubKey     string `json:"pub_key"`
	Desc       string `json:"desc"`
	IsAlive    bool   `json:"is_alive"`
	InternalIP string `json:"internal_ip"`
	ExternalIP string `json:"external_ip"`
	RPCPort    int    `json:"rpc_port"`
	P2PPort    int    `json:"p2p_port"`
	Typ        int    `json:"type"`
}

type Stats struct {
	LatestBlock   uint64 `json:"latest_block"`
	TotalTx       int64  `json:"total_tx"`
	TotalContract int64  `json:"total_contract"`
	TotalNode     uint64 `json:"total_node"`
}
