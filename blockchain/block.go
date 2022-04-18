package blockchain

import (
	"errors"
	"strings"
	"time"

	"github.com/jeyoungjung/zerocoin/db"
	"github.com/jeyoungjung/zerocoin/utils"
)

const difficulty int = 2

type Block struct {
	Data       string `json:"data"`
	Hash       string `json:"hash"`
	PrevHash   string `json:"prevHash,omitempty"`
	Height     int    `json:"height"`
	Difficulty int    `json:"difficulty`
	Nonce      int    `json: nonce`
	Timestamp  int    `json:"timestamp"`
}

func createBlock(data string, prevHash string, height int) *Block {
	block := Block{
		Data:       data,
		Hash:       "",
		PrevHash:   prevHash,
		Height:     height,
		Difficulty: b.calculateDifficulty(),
		Nonce:      0,
	}
	block.mine()
	block.persist()
	return &block
}

func (b *Block) persist() {
	db.SaveBlock(b.Hash, utils.EncodeFromStructToBytes(b)) // saves the data, hash, prevhash and height in bytes, to the db
}

func (b *Block) mine() {
	target := strings.Repeat("0", b.Difficulty) // amount of zeros required for the hash; repeated b.difficulty amount of times
	for {
		b.Timestamp = int(time.Now().Unix())
		hash := utils.Hash(b)
		if strings.HasPrefix(hash, target) { // if the hash has the amount of zeros required
			b.Hash = hash // set the hash and break
			break
		} else {
			b.Nonce++
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
