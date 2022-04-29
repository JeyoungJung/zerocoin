package blockchain

import (
	"errors"
	"sync"
	"time"

	"github.com/jeyoungjung/zerocoin/utils"
	"github.com/jeyoungjung/zerocoin/wallet"
)

const (
	minerReward int = 50
)

type Tx struct {
	ID        string   `json:"id"`
	Timestamp int      `json:"timestamp"`
	TxIns     []*TxIn  `json:"txIns"`
	TxOuts    []*TxOut `json:"txOuts"`
}

func (t *Tx) hashId() {
	t.ID = utils.Hash(t)
}

type TxIn struct {
	TxID      string `json:"txid"`
	Index     int    `json:"index"`
	Signature string `json:"signature"`
}

type TxOut struct {
	Address string `json:"address"`
	Amount  int    `json:"amount"`
}

type UTxOut struct {
	TxID   string
	Index  int
	Amount int
}

type mempool struct { // mempool is where transactions are held before verification, it just stays in the memory
	// no need to go to the database
	Txs map[string]*Tx // "txID" : tx
	m   sync.Mutex
}

var m *mempool
var memOnce sync.Once

func Mempool() *mempool {
	memOnce.Do(func() {
		m = &mempool{
			Txs: make(map[string]*Tx),
		}
	})
	return m
}

func makeCoinbaseTx(address string) *Tx { // coinbase transaction is the first transaction in a block,
	// where the reward is given to the miner, added immediately when a block in added to the blockchain
	txIns := []*TxIn{
		{"", -1, "COINBASE"},
	}
	txOuts := []*TxOut{
		{address, minerReward},
	}
	tx := Tx{
		ID:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.hashId()
	return &tx
}

func (tx *Tx) sign() { // signs every txIn
	for _, txIn := range tx.TxIns {
		txIn.Signature = wallet.Sign(tx.ID, wallet.Wallet())
	}
}

func validate(tx *Tx) bool { // validates the ownership of the money
	valid := true
	for _, txIn := range tx.TxIns {
		prevTx := FindTx(Blockchain(), txIn.TxID) // find a tx with the same ID as the txIn
		if prevTx == nil {                        // if there is none, it means that there was no tx with the same ID as txIn (no such money)
			valid = false
			break
		}
		address := prevTx.TxOuts[txIn.Index].Address // if there is such tx with the same ID as the txIn,
		// get the txOut.address of that tx, and verify it with the signature of the txIn
		// if that txOut was owned by the owner of this txIn, it would be verifed, if not, it won't be verified
		valid = wallet.Verify(txIn.Signature, tx.ID, address)
		if !valid {
			break
		}
	}
	return valid
}

func isOnMempool(uTxOut *UTxOut) bool { // this functions checks if the uTxOut already exists on the mempool
	// if it already exists, you shouldn't be able to use it again since, it has already been used.
	exists := false
Outer: // this is called a "label"
	for _, tx := range Mempool().Txs {
		for _, input := range tx.TxIns {
			if input.Index == uTxOut.Index && input.TxID == uTxOut.TxID {
				exists = true
				break Outer
			}
		}
	}
	return exists
}

var ErrorNoMoney = errors.New("not enough funds")
var ErrorNotValid = errors.New("Tx Invalid")

func makeTx(from, to string, amount int) (*Tx, error) { // this function makes the transactions
	if TotalBalanceByAddress(from, Blockchain()) < amount {
		return nil, ErrorNoMoney
	}
	var txOuts []*TxOut
	var txIns []*TxIn
	total := 0
	uTxOuts := UTxOutsByAddress(from, Blockchain()) // gets the unspent transaction output for "from"
	for _, uTxOut := range uTxOuts {
		if total >= amount {
			break
		}
		txIn := &TxIn{uTxOut.TxID, uTxOut.Index, from}
		txIns = append(txIns, txIn)
		total += uTxOut.Amount
	}
	if change := total - amount; change != 0 {
		changeTxOut := &TxOut{from, change}
		txOuts = append(txOuts, changeTxOut)
	}
	txOut := &TxOut{to, amount}
	txOuts = append(txOuts, txOut)
	tx := &Tx{
		ID:        "",
		Timestamp: int(time.Now().Unix()),
		TxIns:     txIns,
		TxOuts:    txOuts,
	}
	tx.hashId()
	tx.sign()
	valid := validate(tx)
	if !valid {
		return nil, ErrorNotValid
	}
	return tx, nil
}

func (m *mempool) AddTx(to string, amount int) (*Tx, error) {
	tx, err := makeTx(wallet.Wallet().Address, to, amount)
	if err != nil {
		return nil, err
	}
	m.Txs[tx.ID] = tx // add the transaction to the mempool
	return tx, nil
}

func (m *mempool) TxToConfirm() []*Tx {
	coinbase := makeCoinbaseTx(wallet.Wallet().Address) // adds the coinbase transaction right away
	var txs []*Tx
	for _, tx := range m.Txs { // goes through all the transactions inside the mempool
		txs = append(txs, tx)
	}
	txs = append([]*Tx{coinbase}, txs...) // coinbase transaction is prepended (appended to the front)
	// https://medium.com/@tzuni_eh/go-append-prepend-item-into-slice-a4bf167eb7af
	m.Txs = make(map[string]*Tx) // empty mempool
	return txs
}

// hardest function to understand
// Inside a transaction inside a block, it always starts with a coinbase transaction (reward given when a block is mined)
// and the transaction ITSELF has an "id".
// If a transaction was to be made between 2 people, and there is sufficient fund in that transaction's txOut,
// the TxID of that transaction's txIn (between 2 people) would have the same TxID as the "id" of the actual transaction.
// Notice how the TxID of every coinbase transaction is just empty, that part should be the same as the actual id
// of the transaction.

// Conclusion: ID is the same if you used money from that transaction

func UTxOutsByAddress(address string, b *blockchain) []*UTxOut { // this function finds all of the TxOuts that haven't been used by an input yet
	// so basically finding the unused money, aka remaining balance
	var uTxOuts []*UTxOut
	creatorTxs := make(map[string]bool)
	for _, block := range GetBlockchain(b) { // inside block
		for _, tx := range block.Transactions { // inside transactions
			for _, input := range tx.TxIns { // inside transaction inputs
				if input.Signature == "COINBASE" {
					break
				}
				if FindTx(b, input.TxID).TxOuts[input.Index].Address == address {
					// if the ID of the txOut is same as the txIn, that txOut has been used by that txIn,
					// so add that ID to the map
					creatorTxs[input.TxID] = true // here input.tx id is same as tx.id
				}
			}
			for index, output := range tx.TxOuts {
				if output.Address == address {
					if _, ok := creatorTxs[tx.ID]; !ok {
						uTxOut := &UTxOut{tx.ID, index, output.Amount}
						if !isOnMempool(uTxOut) {
							uTxOuts = append(uTxOuts, uTxOut)
						}
					}
				}
			}
		}
	}
	return uTxOuts
}

func TotalBalanceByAddress(address string, b *blockchain) int { // this function finds the total balance for a specific address
	txOuts := UTxOutsByAddress(address, b) // Gathered txOuts for that address
	var amount int
	for _, txOut := range txOuts {
		amount += txOut.Amount
	}
	return amount
}

func GetTxs(b *blockchain) []*Tx { // returns every transaction inside the blockchain
	var txs []*Tx
	for _, block := range GetBlockchain(b) {
		txs = append(txs, block.Transactions...)
	}
	return txs
}

func FindTx(b *blockchain, targetID string) *Tx { // returns a transaction with the targetID,
	for _, tx := range GetTxs(b) {
		if tx.ID == targetID {
			return tx
		}
	}
	return nil
}

func (m *mempool) AddPeerTx(tx *Tx) { // this function adds the new transaction from the peer to the current mempool
	// this function is called everytime a new transaction is made by someone
	m.m.Lock()
	defer m.m.Unlock()
	m.Txs[tx.ID] = tx
}
