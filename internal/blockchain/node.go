package blockchain

import (
	"bytes"
	"fmt"
	"math/big"
	"slices"
	"time"
)

type Node struct {
	Name                 string
	Id                   string
	Utxo                 UtxoList
	Wallet               *Wallet
	BlockChain           []*Block
	BlockCandidate       *Block
	Neighbours           map[string]*Node
	BlockVerifiers       []func(*Node, *Block) error
	TransactionVerifiers []func(*Node, *Transaction) error
}

func NewNode(name string) (*Node, error) {
	n := &Node{
		Name:                 name,
		Id:                   GenUniqueIdString(),
		Utxo:                 NewUtxoList(),
		Neighbours:           make(map[string]*Node),
		BlockCandidate:       nil,
		BlockVerifiers:       make([]func(*Node, *Block) error, 0),
		TransactionVerifiers: make([]func(*Node, *Transaction) error, 0),
	}
	n.Utxo.Put(COINBASE_ADDR, COINBASE_ADDR, COINBASE_START_AMOUNT)
	w, err := NewWallet(name)
	if err != nil {
		return nil, n.Error("NewNode", fmt.Sprintf("Failed to create wallet:\n%s", err))
	}
	n.Wallet = w
	return n, nil
}

func (n *Node) Error(f, msg string) error {
	return fmt.Errorf("Node %s: %s: %s", n.Name, f, msg)
}

func (n *Node) BlockVerificationError(msg string) error {
	return n.Error("Block verifiaction failed", msg)
}

func (n *Node) TransactionVerificatoinError(msg string) error {
	return n.Error("Transaction verification failed", msg)
}

func (n *Node) AssignWallet(w *Wallet) *Node {
	n.Wallet = w
	return n
}

func (n *Node) CoinbaseUtxoAmount() int {
	if u, ok := n.Utxo[COINBASE_ADDR]; ok {
		return u.Amount
	}
	return 0
}

func (n *Node) AddNeighbour(m *Node) *Node {
	if m.Id == n.Id {
		return n
	}
	n.Neighbours[m.Id] = m
	return n
}

func (n *Node) AddNeighbourList(nlist []*Node) {
	for _, m := range nlist {
		n.AddNeighbour(m)
	}
}

func (n *Node) AddBlockVerifier(bf func(*Node, *Block) error) *Node {
	n.BlockVerifiers = append(n.BlockVerifiers, bf)
	return n
}

func (n *Node) GetLastBlock() *Block {
	l := len(n.BlockChain)
	if l == 0 {
		return nil
	}
	return n.BlockChain[l-1]
}

// Initialises the Genesis Block and sets it as Block candidate
func (n *Node) InitGenesisBlock() (*Block, error) {
	if len(n.BlockChain) > 0 {
		return nil, n.Error("CreateGenesisBlock", "Block chain is not empty")
	}
	n.BlockCandidate = NewGenesisBlock()
	return n.BlockCandidate, nil
}

// Creates Genesis block and mines it
func (n *Node) CreateGenesisBlock() (*Block, error) {
	if len(n.BlockChain) > 0 {
		return nil, n.Error("CreateGenesisBlock", "Block chain is not empty")
	}
	n.BlockCandidate = NewGenesisBlock()
	b, err := n.Mine()
	n.BlockCandidate = nil
	if err != nil {
		return nil, n.Error("CreateGenesisBlock", fmt.Sprintf("Failed to mine block:\n%s", err))
	}
	return b, nil
}

func (n *Node) AddVerifyBlock(b *Block) error {
	if err := n.VerifyBlock(b); err != nil {
		return err
	}
	n.AddBlock(b)
	return nil
}

func (n *Node) NewBlockCandidate() *Block {
	n.BlockCandidate = nil
	b := NewBlock()
	b.Header.Height = len(n.BlockChain)
	lb := n.GetLastBlock()
	if lb != nil {
		b.Header.Prev = bytes.Clone(lb.Header.Hash)
		b.Body.Coinbase = lb.Body.Coinbase
	}
	n.BlockCandidate = b
	return b
}

func (n *Node) AddTransaction(t Transaction) {
	if n.BlockCandidate == nil {
		n.NewBlockCandidate()
	}
	n.BlockCandidate.AddTransaction(t)
}

func (n *Node) VerifyTransaction(t Transaction) error {
	var empty interface{}
	inAmount := 0
	outAmount := 0
	utxoHash := make(map[string]interface{})
	uids := n.candidateTransactionUtxoIds()

	// Check unique utxo ids
	for id, u := range t.InputUtxo {
		if _, ok := utxoHash[id]; ok {
			return n.TransactionVerificatoinError("Input Utxo used more then once")
		} else {
			utxoHash[id] = empty
		}
		if _, ok := uids[id]; ok {
			return n.TransactionVerificatoinError("Input Utxo already in block candidate transactions")
		}
		inAmount += u.Amount
	}

	for id, u := range t.OutputUtxo {
		if _, ok := utxoHash[id]; ok {
			return n.TransactionVerificatoinError("Output Utxo used more then once")
		} else {
			utxoHash[id] = empty
		}
		if _, ok := uids[id]; ok {
			return n.TransactionVerificatoinError("Output Utxo already in block candidate transactions")
		}
		outAmount += u.Amount
	}

	if inAmount != outAmount {
		return n.TransactionVerificatoinError("InputUtxo and OutputUtxo sums are not equal")
	}

	return nil
}

func (n *Node) AddVerifyTransaction(t Transaction) error {
	if err := n.VerifyTransaction(t); err != nil {
		return err
	}
	n.AddTransaction(t)
	return nil
}

func (n *Node) Mine() (*Block, error) {

	err := n.mineBlockCandidate()
	if err != nil {
		return nil, err
	}
	b := n.BlockCandidate
	n.BlockCandidate = nil
	// n.addBlock(b)
	if err := n.AddVerifyBlock(b); err != nil {
		return b, err
	}
	return b, nil
}

func (n *Node) MineUnsafe() (*Block, error) {
	err := n.mineBlockCandidate()
	if err != nil {
		return nil, err
	}
	b := n.BlockCandidate
	n.BlockCandidate = nil
	n.AddBlock(b)
	return b, nil
}

func (n *Node) mineBlockCandidate() error {
	if n.BlockCandidate == nil {
		return n.Error("Mine", "No block candidate")
	}

	if err := n.AddRewardTransaction(n.BlockCandidate); err != nil {
		return n.Error("Mine", fmt.Sprintf("Failed to add reward Transaction:\n%s", err))
	}

	nonce, h, err := MineBlock(n.BlockCandidate)
	if err != nil {
		return n.Error("Mine", fmt.Sprintf("Failed to mine:\n\t%s", err))
	}
	n.BlockCandidate.Header.Nonce = nonce
	n.BlockCandidate.Header.Hash = h
	n.BlockCandidate.Body.Coinbase = n.BlockCandidate.Body.Coinbase - REWARD_AMOUNT
	return nil
}

func (n *Node) AddRewardTransaction(b *Block) error {
	cb := n.Utxo[COINBASE_ADDR].Amount
	if cb < REWARD_AMOUNT {
		return n.Error("AddRewardTransaction", "Not enough coinbase")
	}
	rt := InitTransaction()
	rt.Sign = []byte{}
	rt.Pk = []byte{}
	rt.InputUtxo.Put(COINBASE_ADDR, COINBASE_ADDR, cb)
	rt.OutputUtxo.NewRecord(n.Wallet.Addr, REWARD_AMOUNT)
	rt.OutputUtxo.Put(COINBASE_ADDR, COINBASE_ADDR, cb-REWARD_AMOUNT)
	b.Body.Transactions = slices.Insert(b.Body.Transactions, 0, rt)
	return nil
}

func (n *Node) VerifyBlock2(b *Block) error {
	for _, bf := range n.BlockVerifiers {
		if err := bf(n, b); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) VerifyBlock(b *Block) error {
	// lb := len(n.BlockChain)
	// if lb == 0 {
	// 	return n.VerificationError("Genesis block is not set")
	// }

	// 1. Check merkle
	root, err := b.CalcMerkleRoot()
	if err != nil {
		return n.BlockVerificationError(fmt.Sprintf("failed to calculate Merkle Root\n\t: %s", err))
	}
	if bytes.Compare(root, b.Header.Root) != 0 {
		return n.BlockVerificationError("Merkle root check failed")
	}

	// 2. Check nonce
	if !checkNonce(b) {
		return n.BlockVerificationError("Nonce check failed")
	}

	// 3. Check block height
	lb := len(n.BlockChain)
	if lb != b.Header.Height {
		return n.BlockVerificationError("Height check failed")
	}

	// Check reward transaction
	if err = n.checkRewardTransaction(b); err != nil {
		return err
	}

	if b.Header.Height == 0 {
		if err = n.verifyGenesisBlock(b); err != nil {
			return err
		}
	} else {
		if err = n.verifyCommonBlock(b); err != nil {
			return err
		}
	}
	return nil
}

// Verifications for genesis block
func (n *Node) verifyGenesisBlock(b *Block) error {
	if len(n.BlockChain) != 0 {
		return n.BlockVerificationError("Genesis block: BlockChain is not empty")
	}
	// Check height
	if b.Header.Height != 0 {
		return n.BlockVerificationError("Genesis block height is not 0")
	}
	// Check Prev hash
	if !(len(b.Header.Prev) == 1 && b.Header.Prev[0] == byte(GENESIS_BLOCK_PREV)) {
		return n.BlockVerificationError("Genesis block invalid Prev hash")
	}
	// Check time
	if time.Now().Before(b.Header.Time) {
		return n.BlockVerificationError("Genesis block: invalid time")
	}
	// Check coinbase
	if b.Body.Coinbase != n.Utxo[COINBASE_ADDR].Amount-REWARD_AMOUNT {
		return n.BlockVerificationError("Genesis block: invalid block coinbase")
	}
	return nil
}

// Verification for non-genesis bock
func (n *Node) verifyCommonBlock(b *Block) error {
	lastBlock := n.GetLastBlock()
	// 4. Check prev Hash
	if bytes.Compare(lastBlock.Header.Hash, b.Header.Prev) != 0 {
		return n.BlockVerificationError("Prev hash check failed")
	}

	// 5. Check time
	if lastBlock.Header.Time.After(b.Header.Time) {
		return n.BlockVerificationError("Time check failed")
	}

	// 6. coinbase check
	if b.Body.Coinbase < 0 || lastBlock.Body.Coinbase-REWARD_AMOUNT != b.Body.Coinbase {
		return n.BlockVerificationError("Coinbase check failed")
	}

	// 7,8,9
	for _, t := range b.Body.Transactions[1:] {
		// 7 and 8. Input Utxo check
		if !n.Utxo.Contains(t.InputUtxo) {
			return n.BlockVerificationError("InputUtxo check failed")
		}
		// 9. Transaction sign check
		if !CheckSign(t.Bytes(), t.Sign, t.Pk) {
			return n.BlockVerificationError("Transaction check failed")
		}
	}
	return nil
}

func (n *Node) checkRewardTransaction(b *Block) error {
	funcName := "CheckRewardTransaction"
	if len(b.Body.Transactions) == 0 {
		return n.Error(funcName, "No transactions in block")
	}
	rt := b.Body.Transactions[0]
	if len(rt.InputUtxo) != 1 {
		return n.Error(funcName, "InputUtxo wrong length")
	}
	if len(rt.OutputUtxo) != 2 {
		return n.Error(funcName, "OutputUtxo wrong lenth")
	}
	cbUtxo := n.Utxo[COINBASE_ADDR]
	if c, ok := rt.InputUtxo[COINBASE_ADDR]; ok {
		if cbUtxo.Amount != c.Amount {
			return n.Error(funcName, "InputUtxo amount not equal to utxo coinbase amount")
		}
	} else {
		return n.Error(funcName, "No coinbase InputUtxo")
	}
	if _, ok := rt.OutputUtxo[COINBASE_ADDR]; !ok {
		return n.Error(funcName, "No coinbase OutputUtxo")
	}
	for k, v := range rt.OutputUtxo {
		if k == COINBASE_ADDR {
			if v.Amount != cbUtxo.Amount-REWARD_AMOUNT {
				return n.Error(funcName, "OutputUtxo - coinbase - amount is wrong")
			}
		} else if v.Amount != REWARD_AMOUNT {
			return n.Error(funcName, "OutputUtxo - miner - reward value is wrong")
		}
	}
	return nil
}

func (n *Node) AddBlock(b *Block) {
	bAdd := b.Clone()
	n.BlockChain = append(n.BlockChain, bAdd)
	for _, t := range b.Body.Transactions {
		n.Utxo.RemoveRecords(t.InputUtxo)
		n.Utxo.AddRecords(t.OutputUtxo)
	}
	if n.Utxo[COINBASE_ADDR].Amount != b.Body.Coinbase {
		n.Utxo[COINBASE_ADDR] = Utxo{Addr: COINBASE_ADDR, Amount: b.Body.Coinbase}
	}
}

func (n *Node) candidateTransactionUtxoIds() map[string]interface{} {
	var empty interface{}
	r := make(map[string]interface{})
	if n.BlockCandidate == nil {
		return r
	}
	for _, t := range n.BlockCandidate.Body.Transactions {
		for id := range t.InputUtxo {
			r[id] = empty
		}
		for id := range t.OutputUtxo {
			r[id] = empty
		}
	}
	return r
}

func MineBlock(b *Block) (int, []byte, error) {
	target, err := getMineTarget()
	if err != nil {
		return 0, nil, fmt.Errorf("MineBlock: %s", err)
	}
	bf := blockBaseBytes(b)

	i := 0
	h := []byte{}
	for ; i < NONCE_MAX; i++ {
		msg := append(bf, IntToBytes(i)...)
		h, err = GetHashGost3411(msg)
		if err != nil {
			continue
		}
		if new(big.Int).SetBytes(h).Cmp(target) == -1 {
			break
		}
	}
	return i, h, nil
}

func blockBaseBytes(b *Block) []byte {
	bf := new(bytes.Buffer)
	bf.Write(IntToBytes(b.Header.Height))
	bf.Write(IntToBytes(b.Header.Time.Unix()))
	bf.Write(b.GetMerkleRoot())
	bf.Write(b.Header.Prev)
	return bf.Bytes()
}

func checkNonce(b *Block) bool {
	h, err := GetHashGost3411(append(blockBaseBytes(b), IntToBytes(b.Header.Nonce)...))
	if err != nil {
		return false
	}
	target, err := getMineTarget()
	if err != nil {
		return false
	}

	return new(big.Int).SetBytes(h).Cmp(target) == -1
}

func getMineTarget() (*big.Int, error) {
	bs, ok := new(big.Int).SetString(MINE_BASE, 10)
	if !ok {
		return nil, fmt.Errorf("Failed to read MINE_BASE")
	}
	diff, ok := new(big.Int).SetString(MINE_DIFF, 10)
	if !ok {
		return nil, fmt.Errorf("Failed to read MINE_BASE")
	}
	return bs.Div(bs, diff), nil
}
