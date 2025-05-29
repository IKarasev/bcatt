package nodemanager

import (
	"fmt"
	"math/rand"

	"github.com/IKarasev/bcatt/internal/blockchain"
)

type NodeManager struct {
	Nodes     map[string]*blockchain.Node
	Wallets   map[string]*blockchain.Wallet
	mainNode  *blockchain.Node
	EvilBlock *blockchain.Block
	Tick      int
	TrGen     *TrGenerator
	NetMap    map[string]map[string]struct{}
}

func NewNodeManager() *NodeManager {
	return &NodeManager{
		Nodes:    make(map[string]*blockchain.Node),
		Wallets:  make(map[string]*blockchain.Wallet),
		mainNode: nil,
		TrGen:    NewTrGenerator(),
		NetMap:   make(map[string]map[string]struct{}),
	}
}

func DefaultNodeManager() *NodeManager {
	nm := NewNodeManager()

	for i := 1; i <= blockchain.NODE_NUM; i++ {
		nm.NewNode(fmt.Sprintf("Node%d", i))
	}
	for i := 1; i <= blockchain.WALLET_NUM; i++ {
		nm.NewWallet(fmt.Sprintf("User%d", i))
	}
	return nm
}

func NewNodeManagerSet(nodes, wallets int) *NodeManager {
	nm := NewNodeManager()
	for i := 1; i <= nodes; i++ {
		nm.NewNode(fmt.Sprintf("Node%d", i))
	}
	for i := 1; i <= wallets; i++ {
		nm.NewNode(fmt.Sprintf("Wallet%d", i))
	}
	return nm
}

func (nm *NodeManager) WithNetmap(netmap map[string][]string) *NodeManager {

	if nm.NetMap == nil {
		nm.NetMap = make(map[string]map[string]struct{})
	}

	// all connected to all if map not provided
	if netmap == nil || len(netmap) == 0 {
		l := len(nm.Nodes)
		nodes := make([]string, 0, l)
		for id := range nm.Nodes {
			nodes = append(nodes, id)
		}
		for i := 0; i < l-1; i++ {
			for j := i + 1; j < l; j++ {
				nm.addToNetMap(nodes[i], nodes[j])
			}
		}
		return nm
	}

	nameToId := make(map[string]string)
	for id, n := range nm.Nodes {
		nameToId[n.Name] = id
	}

	for from, toList := range netmap {
		if toList == nil || len(toList) == 0 {
			continue
		}
		fromId, ok := nameToId["Node"+from]
		if !ok {
			continue
		}
		for _, to := range toList {
			if from == to {
				continue
			}
			toId, ok := nameToId["Node"+to]
			if !ok {
				continue
			}
			nm.addToNetMap(fromId, toId)
		}
	}

	return nm
}

func (nm *NodeManager) SelectMainNode() *blockchain.Node {
	if nm.mainNode != nil {
		nm.mainNode.BlockCandidate = nil
	}
	l := len(nm.Nodes)
	ids := make([]string, 0, l)
	for k := range nm.Nodes {
		ids = append(ids, k)
	}
	nm.mainNode = nm.Nodes[ids[rand.Intn(l)]]
	nm.mainNode.NewBlockCandidate()
	nm.TrGen.Clear()
	return nm.mainNode
}

func (nm *NodeManager) SetMainNode(id string) (*blockchain.Node, error) {
	if id == "" {
		return nil, fmt.Errorf("Node manager: Set main node: empty id")
	}
	var oldMain *blockchain.Node

	if nm.mainNode != nil {
		if nm.mainNode.Id == id {
			return nm.mainNode, nil
		}
		oldMain = nm.mainNode
	}

	n, ok := nm.Nodes[id]

	if !ok {
		return nil, fmt.Errorf("Node manager: Set main node: Node %s not found", id)
	}

	b := n.NewBlockCandidate()

	if oldMain != nil && oldMain.BlockCandidate != nil && len(oldMain.BlockCandidate.Body.Transactions) > 0 {
		oldTr := oldMain.BlockCandidate.Body.Transactions
		b.Body.Transactions = make([]blockchain.Transaction, 0, len(oldTr))
		for t := range oldTr {
			b.Body.Transactions = append(b.Body.Transactions, oldTr[t].Clone())
		}
		oldMain.BlockCandidate = nil
	} else {
		b.Body.Transactions = make([]blockchain.Transaction, 0)
	}
	nm.mainNode = n
	return nm.mainNode, nil
}

func (nm *NodeManager) GetSetMainNode() *blockchain.Node {
	if nm.mainNode != nil {
		return nm.mainNode
	}
	return nm.SelectMainNode()
}

func (nm *NodeManager) MainNode() *blockchain.Node {
	return nm.mainNode
}

func (nm *NodeManager) NodeNames() []string {
	names := make([]string, len(nm.Nodes))
	i := 0
	for _, n := range nm.Nodes {
		names[i] = n.Name
		i++
	}
	return names
}

func (nm *NodeManager) NewWallet(name string) (*blockchain.Wallet, error) {
	w, err := blockchain.NewWallet(name)
	if err != nil {
		return nil, err
	}
	nm.Wallets[w.Addr] = w
	return w, nil
}

func (nm *NodeManager) AddWallet(w *blockchain.Wallet) {
	nm.Wallets[w.Addr] = w
}

func (nm *NodeManager) NewNode(name string) (*blockchain.Node, error) {
	n, err := blockchain.NewNode(name)
	if err != nil {
		return nil, err
	}
	nm.Nodes[n.Id] = n
	nm.AddWallet(n.Wallet)
	return n, err
}

func (nm *NodeManager) Mine() (*blockchain.Block, error) {
	b, err := nm.GetSetMainNode().Mine()
	if err != nil {
		return nil, err
	}
	nm.UpdateWalletsUtxo(b)
	return b, nil
}

func (nm *NodeManager) EvryNode(f func(n *blockchain.Node) error) error {
	for _, n := range nm.Nodes {
		if err := f(n); err != nil {
			return err
		}
	}
	return nil
}

func (nm *NodeManager) EvryNonMainNode(f func(n *blockchain.Node) error) error {
	if nm.mainNode == nil {
		return fmt.Errorf("Emulator Server: main node not set")
	}
	for k, n := range nm.Nodes {
		if k == nm.mainNode.Id {
			continue
		}
		if err := f(n); err != nil {
			return err
		}
	}
	return nil
}

func (nm *NodeManager) GetNodeBlock(nodeId string, height int) (*blockchain.Block, error) {
	n, ok := nm.Nodes[nodeId]
	if !ok {
		return nil, fmt.Errorf("NodeManager: Node [%s] not found", nodeId)
	}
	if height == -1 {
		if n.BlockCandidate == nil {
			return nil, fmt.Errorf("NodeManager: Node [%s]: no block candidate", nodeId)
		}
		return n.BlockCandidate, nil
	} else if height < 0 || height >= len(n.BlockChain) {
		return nil, fmt.Errorf("NodeManager: Node [%s]: invalid block height", nodeId)
	}
	return n.BlockChain[height], nil
}

func (nm *NodeManager) UpdateWalletsUtxo(b *blockchain.Block) {
	for _, t := range b.Body.Transactions {
		for id, u := range t.InputUtxo {
			if w, ok := nm.Wallets[u.Addr]; ok {
				w.RemoveUtxo(id)
			}
		}
		for id, u := range t.OutputUtxo {
			if w, ok := nm.Wallets[u.Addr]; ok {
				w.AddUtxo(id, u.Addr, u.Amount)
			}
		}
	}
}

func (nm *NodeManager) NetLinkCount() int {
	if nm.NetMap == nil {
		return len(nm.Nodes)
	}
	count := 0
	seen := make(map[string]struct{})
	for from, toList := range nm.NetMap {
		for to := range toList {
			conn := from + "->" + to
			if _, ok := seen[conn]; ok {
				continue
			}
			count++
			seen[conn] = struct{}{}
			seen[to+"->"+from] = struct{}{}
		}
	}
	return count
}

func (nm *NodeManager) NetLinkListWithNames() [][2]string {
	l := len(nm.Nodes)
	linkList := make([][2]string, 0, l*2)
	if nm.NetMap == nil || len(nm.NetMap) == 0 {
		nodeList := make([]string, 0, l)
		for _, n := range nm.Nodes {
			nodeList = append(nodeList, n.Name)
		}
		for i := 0; i < l-1; i++ {
			for j := i + 1; j < l; j++ {
				linkList = append(linkList, [2]string{nodeList[i], nodeList[j]})
			}
		}
		return linkList
	}

	seen := make(map[string]struct{})
	for fromId, toList := range nm.NetMap {
		for toId := range toList {
			if _, ok := seen[fromId+"->"+toId]; ok {
				continue
			}
			linkList = append(linkList, [2]string{nm.Nodes[fromId].Name, nm.Nodes[toId].Name})
			seen[fromId+"->"+toId] = struct{}{}
			seen[toId+"->"+fromId] = struct{}{}
		}
	}

	return linkList
}

func (nm *NodeManager) GenerateUtxoForAll(nMin, nMax, vMin, vMax int) {
	utxoList := blockchain.NewUtxoList()
	for wid, w := range nm.Wallets {
		if ul, err := nm.TrGen.GenerateUtxos(wid, nMin, nMax, vMin, vMax); err == nil {
			w.Utxo.AddRecords(ul)
			utxoList.AddRecords(ul)
		} else {
			fmt.Println(err)
		}
	}
	for _, n := range nm.Nodes {
		n.Utxo.AddRecords(utxoList)
	}
}

func (nm *NodeManager) GenerateUtxoForWallet(walletName []string, nMin, nMax, vMin, vMax int) {
	utxoList := blockchain.NewUtxoList()
	walletSet := make(map[string]struct{})
	for _, n := range walletName {
		walletSet[n] = struct{}{}
	}
	for wid, wallet := range nm.Wallets {
		if _, ok := walletSet[wallet.Name]; !ok {
			continue
		}
		if ul, err := nm.TrGen.GenerateUtxos(wid, nMin, nMax, vMin, vMax); err == nil {
			wallet.Utxo.AddRecords(ul)
			utxoList.AddRecords(ul)
		} else {
			fmt.Println(err)
		}
	}
	for _, n := range nm.Nodes {
		n.Utxo.AddRecords(utxoList)
	}
}

func (nm *NodeManager) GenerateTransactions(tr, utxo int) error {
	return nm.TrGen.GenTransactionN(nm, tr, utxo)
}

func (nm *NodeManager) consensusCheck(b *blockchain.Block) (bool, []error) {
	nl := len(nm.Nodes)
	errs := make([]error, 0, nl)
	for _, n := range nm.Nodes {
		if err := n.VerifyBlock(b); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > nl/2 {
		return false, errs
	}
	return true, errs
}

func (nm *NodeManager) addToNetMap(from, to string) {
	fromConn, ok := nm.NetMap[from]
	if !ok {
		nm.NetMap[from] = make(map[string]struct{})
		fromConn = nm.NetMap[from]
	}
	toConn, ok := nm.NetMap[to]
	if !ok {
		nm.NetMap[to] = make(map[string]struct{})
		toConn = nm.NetMap[to]
	}
	fromConn[to] = struct{}{}
	toConn[from] = struct{}{}

}
