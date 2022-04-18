package blockchain

import (
	"fmt"
	"sync"

	"github.com/jeyoungjung/zerocoin/db"
	"github.com/jeyoungjung/zerocoin/utils"
)

const (
	defaultDifficulty         int = 2 // set to have 2 leading zeros
	blocksRequiredForChecking int = 5 // checked every 5 blocks
	expectedMineTime          int = 2 // should take around 2 minutes per block
	allowedRange              int = 2 // allowed +- 2 minutes (8 to 12 minutes every 5 blocks)
)

type blockchain struct {
	NewestHash        string `json:"newestHash"`
	Height            int    `json:"height"`
	CurrentDifficulty int    `json:"currentdifficulty"`
}

var b *blockchain
var once sync.Once

func (b *blockchain) AddBlock(data string) {
	block := createBlock(data, b.NewestHash, b.Height+1)
	b.NewestHash = block.Hash
	b.Height = block.Height
	b.CurrentDifficulty = block.Difficulty
	b.persist()
	fmt.Printf("NewestHash: %s\nHeight:%d\n", b.NewestHash, b.Height)
}

func (b *blockchain) restore(data []byte) {
	utils.DecodeFromBytesToStruct(data, b) // the data is decoded into b, so now the checkpoint data is in b
}

func (b *blockchain) persist() {
	db.SaveCheckpoint(utils.EncodeFromStructToBytes(b)) // Sends the newest hash and height in bytes
}

func (b *blockchain) GetBlockchain() []*Block { // this function gets every block
	var blocks []*Block
	hashCursor := b.NewestHash // start from the newest hash
	for {
		block, _ := FindBlock(hashCursor) // find the block with the hashcursor
		blocks = append(blocks, block)    // add it to the blocks
		if block.PrevHash != "" {         // if there is prevhash, make the hashcursor be the prevhash
			hashCursor = block.PrevHash
		} else { // if there's no prevhash, meaning, the genesis block
			break
		}
	}
	return blocks
}

func (b *blockchain) calculateDifficulty() int {
	allBlocks := b.GetBlockchain() // gets every block
	newestBlock := allBlocks[0]    // newest block is added to the beginning of the array
	lastRecalculatedBlock := allBlocks[blocksRequiredForChecking-1]
	actualTime := (newestBlock.Timestamp / 60) - (lastRecalculatedBlock.Timestamp / 60)
	expectedTime := blocksRequiredForChecking * expectedMineTime
	if actualTime <= (expectedTime - allowedRange) {
		return b.CurrentDifficulty + 1
	} else if actualTime >= (expectedTime + allowedRange) {
		return b.CurrentDifficulty - 1
	}
	return b.CurrentDifficulty
}

func Blockchain() *blockchain {
	if b == nil {
		once.Do(func() { // https://medium.com/easyread/just-call-your-code-only-once-256f69ed39a8
			b = &blockchain{
				Height: 0,
			}
			fmt.Printf("NewestHash: %s\nHeight:%d\n", b.NewestHash, b.Height)
			checkpoint := db.GetCheckpointData()
			if checkpoint == nil { // if there is no checkpoint make a genesis block
				b.AddBlock("Genesis")
			} else { // if there is checkpoint, restore that block
				fmt.Println("Restoring...")
				b.restore(checkpoint)
			}
		})
	}
	fmt.Printf("NewestHash: %s\nHeight:%d\n", b.NewestHash, b.Height)
	return b
}
