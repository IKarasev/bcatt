package agregator

import (
	"fmt"

	"github.com/IKarasev/bcatt/internal/blockchain"
)

type ChainFork struct {
	First *ChainForkItem
	Items map[string]*ChainForkItem
}

type ChainForkBlock struct {
	Height int    `json:"height"`
	Hash   string `json:"hash"`
	Prev   string `json:"prev"`
	Node   string `json:"node"`
	Tick   int    `json:"tick"`
	Evil   bool   `json:"evil"`
}

type ChainForkItem struct {
	Block ChainForkBlock
	Next  []*ChainForkItem
}

func NewChainFork() *ChainFork {
	bf := &ChainFork{
		First: nil,
		Items: make(map[string]*ChainForkItem),
	}
	return bf
}

func NewChainForkBlock(b *blockchain.Block, miner string, tick int, evil bool) ChainForkBlock {
	bf := ChainForkBlock{
		Height: b.Header.Height,
		Hash:   b.HashString(),
		Prev:   b.PrevString(),
		Node:   miner,
		Tick:   tick,
		Evil:   evil,
	}
	return bf
}

func NewChainForkItem(b *blockchain.Block, miner string, tick int, evil bool) *ChainForkItem {
	return &ChainForkItem{
		Block: NewChainForkBlock(b, miner, tick, evil),
		Next:  []*ChainForkItem{},
	}
}

func (cf *ChainFork) AddBlock(b *blockchain.Block, miner string, tick int, evil bool) (*ChainForkItem, error) {
	ci := NewChainForkItem(b, miner, tick, evil)
	if cf.First == nil || len(cf.Items) == 0 {
		if b.Header.Height != 0 {
			return nil, fmt.Errorf("Add block: Empty blockchain and block height is not 0")
		}
		ci.Block.Prev = ""
		cf.First = ci
		cf.Items[ci.Block.Hash] = ci
		return ci, nil
	}

	if _, ok := cf.Items[ci.Block.Hash]; ok {
		return nil, fmt.Errorf("Add block: block alreadi in chain")
	}

	prev, ok := cf.Items[ci.Block.Prev]
	if !ok {
		return nil, fmt.Errorf("Add block: Previous block not found")
	}
	prev.Next = append(prev.Next, ci)
	cf.Items[ci.Block.Hash] = ci
	return ci, nil
}

func (cf *ChainFork) GetBlocksList() []ChainForkBlock {
	itemsLen := len(cf.Items)
	if itemsLen == 0 {
		return []ChainForkBlock{}
	}
	blocks := make([]ChainForkBlock, 0, itemsLen)
	for _, item := range cf.Items {
		blocks = append(blocks, item.Block)
	}
	return blocks
}
