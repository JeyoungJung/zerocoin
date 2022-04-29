package blockchain

import (
	"errors"
	"strings"
	"time"

	"github.com/jeyoungjung/zerocoin/db"
	"github.com/jeyoungjung/zerocoin/utils"
)

type Block struct {
	Hash         string `json:"hash"`
	PrevHash     string `json:"prevHash,omitempty"`
	Height       int    `json:"height"`
	Difficulty   int    `json:"difficulty"`
	Nonce        int    `json:"nonce"`
	Timestamp    int    `json:"timestamp"`
	Transactions []*Tx  `json:"transactions"`
}

func createBlock(prevHash string, height int, diff int) *Block {
	block := Block{
		Hash:       "",
		PrevHash:   prevHash,
		Height:     height,
		Difficulty: diff,
		Nonce:      0,
	}
	block.mine()
	block.Transactions = Mempool().TxToConfirm() // transactions are only confirmed after blocks are mined
	persistBlock(&block)
	return &block
}

func persistBlock(b *Block) {
	db.SaveBlock(b.Hash, utils.EncodeToBytes(b)) // saves the data, hash, prevhash and height in bytes, to the db
}

// mine is the function where you have to "solve" the "puzzle"
func (b *Block) mine() {
	target := strings.Repeat("0", b.Difficulty) // amount of zeros required for the hash; repeated b.difficulty amount of times
	for {
		b.Timestamp = int(time.Now().Unix()) // current time since jan 1st 1970, in seconds
		hash := utils.Hash(b)                // sets the hash value for the block
		if strings.HasPrefix(hash, target) { // if the hash has the amount of zeros required
			b.Hash = hash // set the hash and break
			break
		} else {
			b.Nonce++ // increase the Nonce, since Nonce is the only thing the miner can change
		}
	}
}

func (b *Block) restore(data []byte) {
	utils.DecodeFromBytesToStruct(data, b) // send data to be decoded into b (block pointer)
}

var ErrBlockNotFound = errors.New("this block is not in the blockchain")

func FindBlock(hash string) (*Block, error) {
	BlockBytes := db.GetBlockData(hash) // returns the block in bytes
	if BlockBytes == nil {
		return nil, ErrBlockNotFound
	}
	block := &Block{}
	block.restore(BlockBytes)
	return block, nil
}
