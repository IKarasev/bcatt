package blockchain

import (
	"encoding/hex"
	"fmt"
)

type Wallet struct {
	Name string
	S    *Signer
	Addr string
	Utxo UtxoList
}

func NewWallet(name string) (*Wallet, error) {
	s, err := NewSigner()
	if err != nil {
		return nil, fmt.Errorf("Failed to create signer")
	}
	addr, err := GetHashGost3411(s.prvKey.Raw())
	if err != nil {
		return nil, fmt.Errorf("Failed to build client address")
	}
	w := &Wallet{
		Name: name,
		S:    s,
		Addr: hex.EncodeToString(addr),
		Utxo: NewUtxoList(),
	}
	return w, nil
}

func (w *Wallet) Error(f, msg string) error {
	return fmt.Errorf("Wallet %s: %s: %s", w.Name, f, msg)
}

func (w *Wallet) ErrorUtxmo(f, id, msg string) error {
	return fmt.Errorf("Wallet %s [%s]: %s: utxo %s: %s", w.Name, w.Addr, f, id, msg)
}

func (w *Wallet) Balance() int {
	b := int(0)
	for _, v := range w.Utxo {
		b += v.Amount
	}
	return b
}

func (w *Wallet) NewTransaction(inputIds []string, out_amount []int, addr string) (*Transaction, error) {
	if addr == w.Addr {
		return nil, w.Error("NewTransaction", "Sending crypto to self not allowed")
	}
	if len(inputIds) != len(out_amount) {
		return nil, w.Error("NewTransaction", "number of input UTXO not equal numver of out amounts")
	}
	for i, v := range out_amount {
		if v < 1 {
			return nil, w.Error("NewTransaction", fmt.Sprintf("out ammount %d is less then 1", i))
		}
	}
	if SliceHasDuplicates(inputIds) {
		return nil, w.Error("NewTransaction", "input ids has duplicates")
	}

	input_utxo := NewUtxoList()
	output_utxo := NewUtxoList()

	for i, v := range inputIds {
		u := w.Utxo.Get(v)
		if u == nil {
			return nil, w.ErrorUtxmo("NewTransaction", v, "record not found")
		}
		if u.Amount < out_amount[i] {
			return nil, w.ErrorUtxmo("NewTransaction", v, "not enough coins")
		}
		input_utxo.Put(v, u.Addr, u.Amount)
		output_utxo.NewRecord(addr, out_amount[i])
		if d := u.Amount - out_amount[i]; d > 0 {
			output_utxo.NewRecord(w.Addr, d)
		}
	}

	if input_utxo.Sum() != output_utxo.Sum() {
		return nil, w.Error("NewTransaction", "input and output sums not equal")
	}

	t := NewTransaction().SetInputUtxo(input_utxo).SetOutputUtxo(output_utxo)
	if err := w.SignTransaction(t); err != nil {
		return nil, w.Error("NewTransaction", "failed to sign transaction")
	}
	return t, nil
}

func (w *Wallet) SignTransaction(t *Transaction) error {
	sig, err := w.S.Sign(t.Bytes())
	if err != nil {
		return w.Error("SignTransaction", "Failed to sign transaction")
	}
	t.SetSign(sig, w.S.PubKey.Raw())
	return nil
}

func (w *Wallet) RemoveUtxo(id string) {
	delete(w.Utxo, id)
}

func (w *Wallet) AddUtxo(id, addr string, a int) {
	w.Utxo[id] = Utxo{
		Addr:   addr,
		Amount: a,
	}
}
