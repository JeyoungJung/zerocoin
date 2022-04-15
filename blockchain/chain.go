package blockchain

// Order of operation :

// if the blockchain is initiated for the first time:
// Blockchain() -> db.GetCheckpointData() -> AddBlock() -> createBlock() -> persist() (saves block) ->
// db.SaveBlock(ToBytes()) -> persist() (saves checkpoint) -> db.SaveCheckpoint(ToBytes())

// if the blockchain already exists:
// Blockchain() -> db.GetCheckpointData() -> restore()

import (
	"fmt"
	"sync"

	"github.com/jeyoungjung/zerocoin/db"
	"github.com/jeyoungjung/zerocoin/utils"
)

type blockchain struct {
	NewestHash string `json:"newestHash"`
	Height     int    `json:"height"`
}

var b *blockchain
var once sync.Once

func (b *blockchain) AddBlock(data string) {
	block := createBlock(data, b.NewestHash, b.Height+1)
	b.NewestHash = block.Hash
	b.Height = block.Height
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
	hashCursor := b.NewestHash	// start from the newest hash
	for {
		block, _ := FindBlock(hashCursor) // find the block with the hashcursor
		blocks = append(blocks, block)	// add it to the blocks
		if block.PrevHash != "" {  // if there is prevhash, make the hashcursor be the prevhash
			hashCursor = block.PrevHash		
		} else {	// if there's no prevhash, meaning, the genesis block
			break
		}
	}
	return blocks
}

func Blockchain() *blockchain {
	if b == nil {
		once.Do(func() { // https://medium.com/easyread/just-call-your-code-only-once-256f69ed39a8
			b = &blockchain{"", 0}
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
