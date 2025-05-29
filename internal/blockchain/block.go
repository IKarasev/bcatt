package blockchain

import (
	"bytes"
	"time"
)

type BlockHeader struct {
	Height int
	Time   time.Time
	Root   []byte
	Prev   []byte
	Nonce  int
	Hash   []byte
}

type BlockBody struct {
	Coinbase     int
	Transactions []Transaction
}

type Block struct {
	Header BlockHeader
	Body   BlockBody
}

func NewBlock() *Block {
	return &Block{
		Header: BlockHeader{
			Time: time.Now(),
			Root: nil,
			Prev: nil,
			Hash: nil,
		},
		Body: BlockBody{
			Transactions: nil,
		},
	}
}

func NewGenesisBlock() *Block {
	b := &Block{
		Header: BlockHeader{
			Height: 0,
			Time:   time.Now(),
			Root:   []byte{},
			Prev:   []byte{byte(GENESIS_BLOCK_PREV)},
			Nonce:  0,
			Hash:   []byte{},
		},
		Body: BlockBody{
			Coinbase:     COINBASE_START_AMOUNT,
			Transactions: make([]Transaction, 1),
		},
	}
	b.Body.Transactions[0].OutputUtxo = NewUtxoList()
	b.Body.Transactions[0].OutputUtxo.Put(COINBASE_ADDR, COINBASE_ADDR, COINBASE_START_AMOUNT)
	return b
}

func (b *Block) Clone() *Block {
	bClone := &Block{
		Header: b.Header.Clone(),
		Body:   b.Body.Clone(),
	}

	return bClone
}

func (h BlockHeader) Clone() BlockHeader {
	h1 := BlockHeader{
		Height: h.Height,
		Time:   h.Time,
		Root:   bytes.Clone(h.Root),
		Prev:   bytes.Clone(h.Prev),
		Nonce:  h.Nonce,
		Hash:   bytes.Clone(h.Hash),
	}
	return h1
}

func (b BlockBody) Clone() BlockBody {
	b1 := BlockBody{
		Coinbase: b.Coinbase,
	}
	b1.Transactions = make([]Transaction, len(b.Transactions))
	for i := range b.Transactions {
		b1.Transactions[i] = b.Transactions[i].Clone()
	}
	return b1
}

func (b *Block) HashString() string {
	return BytesToString(b.Header.Hash)
}

func (b *Block) RootString() string {
	return BytesToString(b.Header.Root)
}

func (b *Block) PrevString() string {
	return BytesToString(b.Header.Prev)
}

// func (b *Block) AddRewardTransaction(addr string) {
// 	rewardTr := InitTransaction()
// 	rewardTr.InputUtxo.Put(COINBASE_ADDR, COINBASE_ADDR, REWARD_AMOUNT)
// 	rewardTr.OutputUtxo.NewRecord(addr, REWARD_AMOUNT)
// 	b.Body.Transactions = slices.Insert(b.Body.Transactions, 0, rewardTr)
// 	panic("Block: AddRewardTransaction: must move to node")
// }

func (b *Block) AddTransaction(t Transaction) int {
	b.Body.Transactions = append(b.Body.Transactions, t.Clone())
	return len(b.Body.Transactions) - 1
}

func (b *Block) GetMerkleRoot() []byte {
	if b.Header.Root == nil || len(b.Header.Root) == 0 {
		r, err := MerkleRoot(b.TransactionBytes())
		if err != nil {
			return nil
		}
		b.Header.Root = r
	}
	root := make([]byte, len(b.Header.Root))
	copy(root, b.Header.Root)
	return root
}

func (b *Block) CalcMerkleRoot() ([]byte, error) {
	return MerkleRoot(b.TransactionBytes())
}

func (b *Block) TransactionBytes() [][]byte {
	tBytes := make([][]byte, len(b.Body.Transactions))
	for i, v := range b.Body.Transactions {
		tBytes[i] = v.Bytes()
	}
	return tBytes
}
