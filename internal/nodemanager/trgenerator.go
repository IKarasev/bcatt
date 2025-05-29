package nodemanager

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/IKarasev/bcatt/internal/blockchain"
	bc "github.com/IKarasev/bcatt/internal/blockchain"
)

type TrGenerator struct {
	spentUtxo map[string]WalletSpentUtxo
}

type WalletSpentUtxo map[string]struct{}

type walletUtxos struct {
	addr  string
	utxos []string
}

func NewTrGenerator() *TrGenerator {
	return &TrGenerator{
		spentUtxo: make(map[string]WalletSpentUtxo),
	}
}

func TrGeneratorGenError(m string, a ...any) error {
	format := "Tr Generator ERROR: failed to generate new TR: " + m
	return fmt.Errorf(format, a...)
}

func (trg *TrGenerator) Clear() {
	for k := range trg.spentUtxo {
		trg.spentUtxo[k] = nil
	}
	trg.spentUtxo = nil
	trg.spentUtxo = make(map[string]WalletSpentUtxo)
}

// Generates new Transaction, source wallet utxo's and amounts are random
// Returns new *blockchain.Transaction, total amount sent, error
func (trg *TrGenerator) GenTransaction(n *bc.Node, fromAddr *bc.Wallet, toAddr string) (*bc.Transaction, int, error) {
	wUtxo := n.Utxo.FilterAddress(fromAddr.Addr)

	uLen := len(wUtxo)
	if uLen == 0 {
		return nil, 0, TrGeneratorGenError("no UTXO for source wallet")
	}

	// Filtering out UTXOs that has already been used
	uIds := make(map[string]struct{})
	for id := range wUtxo {
		uIds[id] = struct{}{}
	}

	sUtxo, ok := trg.spentUtxo[fromAddr.Addr]

	if ok {
		for i := range sUtxo {
			delete(uIds, i)
		}
		if len(uIds) == 0 {
			return nil, 0, TrGeneratorGenError("all source wallet UTXOs been used")
		}
	}

	// Selecting random number of UTXO and send amounts as input for transaction
	gen := newRandGenerator()
	uNum := gen.IntN(len(uIds)) + 1
	inputUId := make([]string, uNum)
	inputAmount := make([]int, uNum)
	totalSend := 0

	for k := range uIds {
		inputUId[uNum-1] = k
		inputAmount[uNum-1] = gen.IntN(wUtxo[k].Amount) + 1
		totalSend += inputAmount[uNum-1]
		uNum -= 1
		if uNum == 0 {
			break
		}
	}

	// Creating transaction
	tr, err := fromAddr.NewTransaction(inputUId, inputAmount, toAddr)
	if err != nil {
		return nil, 0, TrGeneratorGenError("failed to create transaction: %s", err)
	}

	err = n.AddVerifyTransaction(*tr)

	if err != nil {
		return nil, 0, TrGeneratorGenError("node %s rejected trunsaction: %s", n.Name, err)
	}

	// saving used utxo
	for i := range inputUId {
		trg.AddSpentUtxo(fromAddr.Addr, inputUId[i])
	}
	return tr, totalSend, nil
}

func (trg *TrGenerator) GenTransactionN(nm *NodeManager, n int, u int) error {
	node := nm.mainNode
	if node == nil {
		return TrGeneratorGenError("Miner not set")
	}

	if len(node.Utxo)-1 < n*u {
		return TrGeneratorGenError("Not enough utxo to generate %d tr with %d utxos", n, u)
	}

	freeWUtxo := trg.getFreeUtxoByWallet(node.Utxo)

	trInputUtxos := make([]walletUtxos, 0, n)
	used := make(map[string]struct{})

	for addr, uids := range freeWUtxo {
		if len(uids) < u {
			continue
		}

		group := walletUtxos{
			addr:  addr,
			utxos: make([]string, 0, u),
		}
		for _, id := range uids {
			if _, ok := used[id]; !ok {
				group.utxos = append(group.utxos, id)
				used[id] = struct{}{}
				if len(group.utxos) == u {
					break
				}
			}
		}

		if len(group.utxos) == u {
			trInputUtxos = append(trInputUtxos, group)
			if len(trInputUtxos) == n {
				break
			}
		}
	}

	if len(trInputUtxos) < n {
		return TrGeneratorGenError("Not enough wallet utxos to generate %d transactions", n)
	}

	for _, trInput := range trInputUtxos {
		w := nm.Wallets[trInput.addr]

		out_amount := make([]int, u)

		for i := 0; i < u; i++ {
			out_amount[i] = 1
		}

		var to_addr string
		for a := range nm.Wallets {
			if a != trInput.addr {
				to_addr = a
				break
			}
		}

		if tr, err := w.NewTransaction(trInput.utxos, out_amount, to_addr); err == nil {
			if err = node.AddVerifyTransaction(*tr); err == nil {
				for _, u := range trInput.utxos {
					trg.AddSpentUtxo(w.Name, u)
				}
			}
		}
	}

	return nil
}

// Adds utxo to spent list
func (trg *TrGenerator) AddSpentUtxo(walletAddr, utxoId string) {
	if _, ok := trg.spentUtxo[walletAddr]; !ok {
		trg.spentUtxo[walletAddr] = WalletSpentUtxo{utxoId: struct{}{}}
	} else {
		// trg.spentUtxo[walletAddr] = append(trg.spentUtxo[walletAddr], utxoId)
		trg.spentUtxo[walletAddr][utxoId] = struct{}{}
	}
}

// Adds input utxo's from transsaction to spent utxos
func (trg *TrGenerator) AddSpentUtxoFromTr(tr *bc.Transaction) {
	for id, u := range tr.InputUtxo {
		trg.AddSpentUtxo(u.Addr, id)
	}
}

// Chesks if wallet's utxo already spent
func (trg *TrGenerator) IsSpentUtxo(walletAddr, utxoId string) bool {
	if l, ok := trg.spentUtxo[walletAddr]; ok {
		return l.Check(utxoId)
	}
	return false
}

func (trg *TrGenerator) GetWallentSpentUtxo(walletAddr string) WalletSpentUtxo {
	w := trg.spentUtxo[walletAddr]
	return w
}

// Generates utxo list for given wallet addres
// nMin, nMax - min and max number of utxo to generate
// vMin, vMax - min and max value for each utxo
func (trg *TrGenerator) GenerateUtxos(addr string, nMin, nMax, vMin, vMax int) (blockchain.UtxoList, error) {
	if nMin > nMax || vMin > vMax {
		return nil, fmt.Errorf("TrGenerator: GenerateUtxos: min value > max value in input")
	}
	utxoList := blockchain.NewUtxoList()
	generator := newRandGenerator()
	num := 0
	if nMin == nMax {
		num = nMin
	} else {
		num = generator.IntN(nMax-nMin+1) + nMin
	}
	for i := 0; i < num; i++ {
		amount := 0
		if vMin == vMax {
			amount = vMin
		} else {
			amount = generator.IntN(vMax-vMin+1) + vMin
		}
		utxoList.NewRecord(addr, amount)
	}
	return utxoList, nil
}

// Checks if utxo is in spent list via Id
func (wsu WalletSpentUtxo) Check(utxoId string) bool {
	_, ok := wsu[utxoId]
	return ok
}

func (trg *TrGenerator) getFreeUtxoByWallet(ul bc.UtxoList) map[string][]string {
	byAddr := make(map[string][]string)

	if ul == nil {
		return byAddr
	}

	freeUtxo := make(bc.UtxoList)
	for uid, utxo := range ul {
		if !trg.IsSpentUtxo(utxo.Addr, uid) {
			freeUtxo.Put(uid, utxo.Addr, utxo.Amount)
		}
	}

	delete(freeUtxo, bc.COINBASE_ADDR)

	for id, utxo := range freeUtxo {
		byAddr[utxo.Addr] = append(byAddr[utxo.Addr], id)
	}

	return byAddr
}

// Creates new random generator
func newRandGenerator() *rand.Rand {
	t := []byte(time.Now().String())
	var seed [32]byte
	copy(seed[:], t)
	g := rand.New(rand.NewChaCha8(seed))
	return g
}
