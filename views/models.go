package views

import "fmt"

const NoEnterEvent string = "if(event.keyCode === 13) {return false;}"

type NodeCellInput struct {
	Name      string
	Id        string
	Coinbase  string
	WName     string
	WAddress  string
	WCoins    string
	BHeight   string
	BHash     string
	BCoinbase string
	BNonce    string
	BRoot     string
	Miner     bool
}

type NodeInfoSm struct {
	Name        string
	Id          string
	Coinbase    string
	TotalUtxo   string
	TotalBlocks string
	Miner       bool
}

type BlockInfoSmallItem struct {
	Height   string
	Coinbase string
	Nonce    string
	Hash     string
	Prev     string
	Root     string
	Time     string
	TotalTr  string
}

type BlockTransactionItem struct {
	Id, Sign, Pk string
	InputUtxo    []UtxoItem
	OutputUtxo   []UtxoItem
}

type UtxoItem struct {
	Id, Addr, Amount string
	Spent            bool
}

type EmulationSettingsItem struct {
	NodeNum       string
	WalletNum     string
	CoinbaseAddr  string
	CoinbaseStart string
	RewardAmount  string
	Diff          string
	OpPause       string
}

type SelectListItem struct {
	Id   string
	Name string
}

type WalletBlockTrItem struct {
	Sign       string
	Pk         string
	InputUtxo  []string
	OutputUtxo []string
}

func rssNodeLabel(n string, ev string) string {
	return n + ev
}

func blockIdInput(id string) string {
	return fmt.Sprintf("blockInfoInvId%s", id)
}

func blockIdInputSelector(id string) string {
	return "#" + blockIdInput(id)
}

var (
	WListTest = []SelectListItem{
		{
			Id:   "1",
			Name: "Wallet1",
		},
		{
			Id:   "2",
			Name: "Wallet2",
		},
		{
			Id:   "3",
			Name: "Wallet3",
		},
		{
			Id:   "4",
			Name: "Wallet4",
		},
	}
)
