package blockchain

import (
	"bytes"
	"fmt"
)

type Transaction struct {
	InputUtxo  UtxoList
	OutputUtxo UtxoList
	Sign       []byte
	Pk         []byte
}

func InitTransaction() Transaction {
	return Transaction{InputUtxo: make(UtxoList), OutputUtxo: make(UtxoList)}
}

func NewTransaction() *Transaction {
	return &Transaction{InputUtxo: make(UtxoList), OutputUtxo: make(UtxoList)}
}

func (t *Transaction) SetInputUtxo(u UtxoList) *Transaction {
	t.InputUtxo = u
	return t
}

func (t *Transaction) SetOutputUtxo(u UtxoList) *Transaction {
	t.OutputUtxo = u
	return t
}

func (t *Transaction) SetSign(s, p []byte) *Transaction {
	t.Sign = s
	t.Pk = p
	return t
}

// Filters transaction utxos by address. Returns (InputUtxo, OutputUtxo)
func (t *Transaction) FilterUtxoByWallet(addr string) (UtxoList, UtxoList) {
	inputUtxo := t.InputUtxo.FilterAddress(addr)
	outputUtxo := t.OutputUtxo.FilterAddress(addr)
	return inputUtxo, outputUtxo
}

func (t *Transaction) SignString() string {
	return BytesToString(t.Sign)
}

func (t *Transaction) PkString() string {
	return BytesToString(t.Pk)
}

func (t *Transaction) Bytes() []byte {
	bf := new(bytes.Buffer)
	bf.Write(t.InputUtxo.Bytes())
	bf.Write(t.OutputUtxo.Bytes())
	return bf.Bytes()
}

func (t *Transaction) Clone() Transaction {
	tt := Transaction{
		InputUtxo:  t.InputUtxo.Clone(),
		OutputUtxo: t.OutputUtxo.Clone(),
		Sign:       bytes.Clone(t.Sign),
		Pk:         bytes.Clone(t.Pk),
	}
	return tt
}

func (t *Transaction) UpdateInputUtxo(uid string, amount int, addr string) error {
	if u, ok := t.InputUtxo[uid]; ok {
		u.Amount = amount
		u.Addr = addr
		t.InputUtxo[uid] = u
		return nil
	}
	return fmt.Errorf("Input utxo with id [%s] not found", uid)
}

func (t *Transaction) UpdateOutputUtxo(uid string, amount int, addr string) error {
	if u, ok := t.OutputUtxo[uid]; ok {
		u.Amount = amount
		u.Addr = addr
		t.OutputUtxo[uid] = u
		return nil
	}
	return fmt.Errorf("Output utxo with id [%s] not found", uid)
}

func (t *Transaction) DeleteInputUtxo(uid string) error {
	if _, ok := t.InputUtxo[uid]; !ok {
		return fmt.Errorf("Input utxo [%s] not found", uid)
	}
	delete(t.InputUtxo, uid)
	return nil
}

func (t *Transaction) DeleteOutputUtxo(uid string) error {
	if _, ok := t.OutputUtxo[uid]; !ok {
		return fmt.Errorf("Output utxo [%s] not found", uid)
	}
	delete(t.OutputUtxo, uid)
	return nil
}

func (t *Transaction) InputBalance() int {
	b := 0
	for _, u := range t.InputUtxo {
		b += u.Amount
	}
	return b
}

func (t *Transaction) OutputBalance() int {
	b := 0
	for _, u := range t.OutputUtxo {
		b += u.Amount
	}
	return b
}
