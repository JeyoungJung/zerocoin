package main

import (
	"crypto/sha256"
	"fmt"
)

type block struct {
	data     string
	hash     string
	prevHash string
}

type blockchain struct {
	blocks []block
}

func (b *blockchain) getPrevHash () string {
	if (len(b.blocks) > 0) {
		return b.blocks[len(b.blocks)-1].hash
	}
	return ""
}

func (b *blockchain) addBlock (data string) {
	newBlock := block{data, "", b.getPrevHash()}
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(newBlock.data + newBlock.prevHash)))
	newBlock.hash = hash
	b.blocks = append(b.blocks, newBlock)
}

func (b blockchain) listBlockchain () {
	for _, block := range b.blocks {
		fmt.Printf("data: %s\n", block.data)
		fmt.Printf("hash: %s\n", block.hash)
		fmt.Printf("prev data: %s\n", block.prevHash)
	}
}

func main() {
	blockchain := blockchain{};
	blockchain.addBlock("Genesis Block")
	blockchain.addBlock("Second Block")
	blockchain.addBlock("Third Block")
	blockchain.listBlockchain()
}
