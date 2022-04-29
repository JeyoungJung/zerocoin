package blockchain

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/jeyoungjung/zerocoin/db"
	"github.com/jeyoungjung/zerocoin/utils"
)

const (
	defaultDifficulty              int = 2 // set to have 2 leading zeros
	blocksRequiredForRecalculation int = 5 // checked every 5 blocks
	expectedMineTime               int = 2 // should take around 2 minutes per block
	allowedRange                   int = 2 // allowed +- 2 minutes (8 to 12 minutes every 5 blocks)
)

type blockchain struct {
	NewestHash        string `json:"newestHash"`
	Height            int    `json:"height"`
	CurrentDifficulty int    `json:"currentdifficulty"`
	m                 sync.Mutex
}

var b *blockchain
var once sync.Once

// Blockchain makes a new blockchain or restores an existing blockchain when ran for the first time
func Blockchain() *blockchain {
	// or just returns the blockchain in the databse if its not the first time running
	once.Do(func() { // https://medium.com/easyread/just-call-your-code-only-once-256f69ed39a8
		// Do function will not end until the function inside the Do function ends. Meaning it will cause a deadlock
		b = &blockchain{
			Height: 0,
		}
		checkpoint := db.GetCheckpointData()
		if checkpoint == nil { // if there is no checkpoint make a genesis block
			b.AddBlock()
		} else { // if there is checkpoint, restore that block
			b.restore(checkpoint)
		}
	})
	return b
}

func Status(b *blockchain, rw http.ResponseWriter) {
	b.m.Lock()
	defer b.m.Unlock()
	utils.HandleErr(json.NewEncoder(rw).Encode(b))
}

func (b *blockchain) AddBlock() *Block {
	block := createBlock(b.NewestHash, b.Height+1, getDifficulty(b))
	b.NewestHash = block.Hash // newesthash, height and currentdifficulty is updated every time a new block is created
	b.Height = block.Height
	b.CurrentDifficulty = block.Difficulty
	persistBlockchain(b)
	return block
}

func (b *blockchain) restore(data []byte) {
	utils.DecodeFromBytesToStruct(data, b) // the data is decoded into b, so now the checkpoint data is in b
}

func persistBlockchain(b *blockchain) {
	db.SaveCheckpoint(utils.EncodeToBytes(b)) // Sends the newest hash and height in bytes
}

// GetBlockchain gets every block from the blockchain
func GetBlockchain(b *blockchain) []*Block {
	b.m.Lock()
	defer b.m.Unlock()
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

func getDifficulty(b *blockchain) int {
	if b.Height == 0 { // if there is no block, just set the difficulty to default difficulty
		return defaultDifficulty
	} else if b.Height%blocksRequiredForRecalculation == 0 { // we are recalculating every 5 blocks
		// (it's just what we set it to, bitcoin has it set to 2016, since they hopefully want to check every 2 weeks, and they want 1 block every 10 min, 1*6*24*14 = 2016)
		// , recalculate the difficulty
		return recalculateDifficulty(b)
	} else { // anything else, just return the current difficulty
		return b.CurrentDifficulty
	}
}

func recalculateDifficulty(b *blockchain) int {
	allBlocks := GetBlockchain(b)                                                // gets every block
	newestBlock := allBlocks[0]                                                  // newest block added
	lastRecalculatedBlock := allBlocks[blocksRequiredForRecalculation-1]         // gets the 5th newest block
	actualTime := (newestBlock.Timestamp - lastRecalculatedBlock.Timestamp) / 60 // actual time took to mine 5 blocks
	expectedTime := blocksRequiredForRecalculation * expectedMineTime            // time that should've taken to mine 5 blocks
	if actualTime <= (expectedTime - allowedRange) {                             // if it took less than 8 minutes
		return b.CurrentDifficulty + 1 // increase difficulty because it took shorter than expected
	} else if actualTime >= (expectedTime + allowedRange) { // if it took more than 12 minutes
		return b.CurrentDifficulty - 1 // decrease difficulty because it took longer than expected
	}
	return b.CurrentDifficulty // if its in range, return the current difficulty
}

// Replace empties the current blockchain and replaces it with the new blockchain
func (b *blockchain) Replace(newBlockchain []*Block) {
	b.m.Lock()
	defer b.m.Unlock()
	b.CurrentDifficulty = newBlockchain[0].Difficulty
	b.Height = len(newBlockchain)
	b.NewestHash = newBlockchain[0].Hash
	db.EmptyBlocksBucket()
	persistBlockchain(b)
	for _, block := range newBlockchain {
		persistBlock(block)
	}
}

// AddPeerBlock adds the new block from the peer to the current blockchain
// AddPeerBlock is called everytime a new block is made by someone
func (b *blockchain) AddPeerBlock(newBlock *Block) {

	b.m.Lock()
	m.m.Lock()
	defer b.m.Unlock()
	defer m.m.Unlock()

	b.CurrentDifficulty = newBlock.Difficulty
	b.Height += 1
	b.NewestHash = newBlock.Hash

	persistBlockchain(b)
	persistBlock(newBlock)

	for _, tx := range newBlock.Transactions { // if the transaction inside the current mempool is resolved by the
		// new incoming block, delete the transaction from the current mempool
		_, ok := m.Txs[tx.ID]
		if ok {
			delete(m.Txs, tx.ID)
		}
	}
}
