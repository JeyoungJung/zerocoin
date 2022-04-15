package blockchain

// Order of operation:
// FindBlock() -> db.GetBlockData() --- if no matching data ---> return nil and ErrBlockNotFound 
//									--- if matching data ---> create a new empty block, put the data in, return the block with no err

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/jeyoungjung/zerocoin/db"
	"github.com/jeyoungjung/zerocoin/utils"
)

type Block struct {
	Data     string `json:"data"`
	Hash     string `json:"hash"`
	PrevHash string `json:"prevHash,omitempty"`
	Height   int    `json:"height"`
}

func createBlock(data string, prevHash string, height int) *Block {
	block := Block{
		Data:     data,
		Hash:     "",
		PrevHash: prevHash,
		Height:   height,
	}
	payload := block.Data + block.PrevHash + fmt.Sprint(block.Height)
	block.Hash = fmt.Sprintf("%x", sha256.Sum256([]byte(payload)))
	block.persist()
	return &block
}

func (b *Block) persist() {
	db.SaveBlock(b.Hash, utils.EncodeFromStructToBytes(b)) // saves the data, hash, prevhash and height in bytes, to the db
}

func (b *Block) restore(data []byte) {
	utils.DecodeFromBytesToStruct(data, b) // send data to be decoded into b (block pointer)
}

var ErrBlockNotFound = errors.New("this block is not in the blockchain")

func FindBlock (hash string) (*Block, error) {
	BlockBytes := db.GetBlockData(hash) // returns the block in bytes
	if BlockBytes == nil {
		return nil, ErrBlockNotFound  
	}
	block := &Block{}
	block.restore(BlockBytes)
	return block, nil
}
