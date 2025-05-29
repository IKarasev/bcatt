package agregator

type DataAgregator struct {
	ChainFork    *ChainFork
	NodesAttacks *NodesAttacks
}

func NewDataAgreator() *DataAgregator {
	return &DataAgregator{
		ChainFork:    NewChainFork(),
		NodesAttacks: NewNodesAttacks(),
	}
}

func (da *DataAgregator) GetForkBlocks() []ChainForkBlock {
	return da.ChainFork.GetBlocksList()
}
