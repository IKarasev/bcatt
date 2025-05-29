package emulator

const (
	V_BN_UPDATE_A_SET_MINER = "miner"
	V_BN_UPDATE_A_TOPOLOGY  = "topology"
	V_BN_UPDATE_A_ATTACKED  = "attacked"
	V_NODE_GROUP_MINER      = "miner"
	V_NODE_GROUP_NODE       = "regular"
	V_NODE_GROUP_USER       = "user"
	V_NODE_GROUP_ATT        = "attack"
)

type BlockChainNetUpdate struct {
	Action string              `json:"action"`
	Nodes  []BlockchainNetNode `json:"nodes"`
	Links  []BlockChainNetLink `json:"links"`
}

type BlockchainNet struct {
	Nodes []BlockchainNetNode `json:"nodes"`
	Links []BlockChainNetLink `json:"links"`
}

type BlockchainNetNode struct {
	Id       string `json:"id"`
	Group    string `json:"group"`
	Attacked bool   `json:"attacked"`
}

type BlockChainNetLink struct {
	Source  string `json:"source"`
	Target  string `json:"target"`
	Latency string `json:"latency"`
}
