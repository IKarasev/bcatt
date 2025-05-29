package blockchain

import (
	"bytes"
	"fmt"
	"iter"
	"maps"
	"slices"
)

type Utxo struct {
	Addr   string
	Amount int
}

type UtxoList map[string]Utxo

func NewUtxoList() UtxoList {
	return make(UtxoList)
}

func (ul UtxoList) Error(i, f, msg string) error {
	return fmt.Errorf("Utxo id %s: %s: %s", i, f, msg)
}

func (ul UtxoList) CheckId(i string) bool {
	_, ok := ul[i]
	return ok
}

func (ul UtxoList) Get(i string) *Utxo {
	if u, ok := ul[i]; ok {
		return &u
	}
	return nil
}

// Creates new record. Returns it's id
func (ul UtxoList) NewRecord(addr string, amount int) string {
	u := Utxo{Addr: addr, Amount: amount}
	id := GenUniqueIdString()
	ul[id] = u
	return id
}

func (ul UtxoList) Put(id string, addr string, amount int) {
	ul[id] = Utxo{Addr: addr, Amount: amount}
}

func (ul UtxoList) RemoveId(i string) {
	delete(ul, i)
}

func (ul UtxoList) RemoveRecords(ul2 UtxoList) {
	for i := range ul2 {
		ul.RemoveId(i)
	}
}

func (ul UtxoList) AddRecords(ul2 UtxoList) {
	for i, v := range ul2 {
		ul[i] = v
	}
}

func (ul UtxoList) Sum() int {
	s := int(0)
	for _, u := range ul {
		s += u.Amount
	}
	return s
}

// Concationation of all records in its bytes representation
func (ul UtxoList) Bytes() []byte {
	bf := new(bytes.Buffer)
	for _, u := range ul.SortedItems() {
		bf.Write([]byte(u.Addr))
		bf.Write(IntToBytes(u.Amount))
	}
	return bf.Bytes()

}

func (ul UtxoList) Clone() UtxoList {
	return maps.Clone(ul)
}

// Searches for utxo records with given ammount. Returns new UtxoList with searh results
func (ul UtxoList) FilterAmmount(a int) UtxoList {
	res := make(UtxoList)
	for i, u := range ul {
		if u.Amount == a {
			res[i] = u
		}
	}
	return res
}

// Searches for utxo recods with given address. Reutrns new UtxoList with search results
func (ul UtxoList) FilterAddress(addr string) UtxoList {
	res := make(UtxoList)
	for i, u := range ul {
		if u.Addr == addr {
			res[i] = u
		}
	}
	return res
}

// Applies f filter function to every utxo record in list, if its true - adds record to result.
// Return new UtxoList with filtering results
func (ul UtxoList) FilterFunction(f func(i, addr string, ammount int) bool) UtxoList {
	res := make(UtxoList)
	for i, u := range ul {
		if f(i, u.Addr, u.Amount) {
			res[i] = u
		}
	}
	return res
}

// Checks if current list contains all items from given UtxoList. Checks keys and values
func (ul UtxoList) Contains(ul2 UtxoList) bool {
	if len(ul) < len(ul2) {
		return false
	}
	for k, u2 := range ul2 {
		if u, ok := ul[k]; ok {
			if u != u2 {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// Checks if utxo record with given id has it's amount < given amount
func (ul UtxoList) ValidateAmmount(id string, ammount int) error {
	u, ok := ul[id]
	if !ok {
		return ul.Error(id, "ValidateAmmount", "utxo record not found")
	}
	if u.Amount < ammount {
		return ul.Error(id, "ValidateAmmount", "not enough amount")
	}
	return nil
}

// Iterator with sorted keys
func (ul UtxoList) SortedItems() iter.Seq2[string, Utxo] {
	keys := make([]string, 0, len(ul))
	for k := range ul {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return func(yield func(string, Utxo) bool) {
		for _, k := range keys {
			if !yield(k, ul[k]) {
				return
			}
		}
	}
}
