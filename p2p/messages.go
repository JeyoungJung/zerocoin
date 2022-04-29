package p2p

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jeyoungjung/zerocoin/blockchain"
	"github.com/jeyoungjung/zerocoin/utils"
)

type MessageKind int // created a MessageKind type to have better readability

const (
	MessageNewestBlock MessageKind = iota // the "iota" numbers the constants starting from 0,
	// hover on MessageAllBlocksRequest and MessageAllBlocksResponse to see their values
	MessageAllBlocksRequest
	MessageAllBlocksResponse
	MessageNewBlockNotify
	MessageNewTxNotify
	MessageNewPeerNotify
)

type Message struct {
	Kind    MessageKind
	Payload []byte
}

func makeMessage(kind MessageKind, payload interface{}) []byte {
	m := Message{
		Kind:    kind,
		Payload: utils.MarshalToJSON(payload), // **using utils.EncodeToBytes also works**
		// but marshaling to json is easier to unmarshal different types of payloads
	}
	return utils.MarshalToJSON(m) // we have to change to json once in the payload, and once in the message because
	// the message should be turned into json, and the payload also has to be changed to json before.
}

func sendNewestBlock(p *peer) {
	fmt.Printf("Sending newest block to %s\n", p.key)
	b, err := blockchain.FindBlock(blockchain.Blockchain().NewestHash) // find the block with the newest hash
	utils.HandleErr(err)
	m := makeMessage(MessageNewestBlock, b)
	p.inbox <- m // send the message to the channel, next funtion would be write()
}

func handleMsg(m *Message, p *peer) {
	switch m.Kind {
	case MessageNewestBlock:
		fmt.Printf("Received the newest block from %s\n", p.key)
		var payload blockchain.Block
		utils.HandleErr(json.Unmarshal(m.Payload, &payload))               // payload holds the newest block for port 4000 (the sender's blockchain)
		b, err := blockchain.FindBlock(blockchain.Blockchain().NewestHash) // this function gets the height of the current port's blockchain
		utils.HandleErr(err)
		if payload.Height >= b.Height { // if the height of the sender's blockchain is higher than the current port's blockchain
			// it means that the sender has an updated blockchain that the current port is not up to date with
			// in that case, request every block from the sender, to update the current port's blockchain
			fmt.Printf("Requesting all blocks from %s\n", p.key)
			requestAllBlocks(p)
		} else { // else, it means that the current port has more up to date blockchain, so give the sender
			// the current newest blockchain.
			// then that port would run this if statement again to find out that the height is different,
			// and would requestAllBlocks form this port.
			sendNewestBlock(p)
		}
	case MessageAllBlocksRequest: // if that port requested for all blocks, send all blocks
		fmt.Printf("%s wants all the blocks\n", p.key)
		sendAllBlocks(p)
	case MessageAllBlocksResponse: // if that port sent all blocks, unmarshal the blocks, and put it into payload
		fmt.Printf("Received all the blocks from %s\n", p.key)
		var payload []*blockchain.Block
		utils.HandleErr(json.Unmarshal(m.Payload, &payload))
		blockchain.Blockchain().Replace(payload)
	case MessageNewBlockNotify:
		var payload *blockchain.Block
		utils.HandleErr(json.Unmarshal(m.Payload, &payload))
		blockchain.Blockchain().AddPeerBlock(payload)
	case MessageNewTxNotify:
		var payload *blockchain.Tx
		utils.HandleErr(json.Unmarshal(m.Payload, &payload))
		blockchain.Mempool().AddPeerTx(payload)
	case MessageNewPeerNotify:
		var payload string
		utils.HandleErr(json.Unmarshal(m.Payload, &payload))
		parts := strings.Split(payload, ":")
		AddPeer(parts[0], parts[1], parts[2], false) // the broadcast part of the AddPeer() is false because the 
		// peer is beign broadcasted right now, doesnt have to be broadcasted again
	}
}

func requestAllBlocks(p *peer) { // this function requests for the entire blockchain to be sent
	m := makeMessage(MessageAllBlocksRequest, nil)
	p.inbox <- m
}

func sendAllBlocks(p *peer) { // this function sends the response and the entire blockchain
	fmt.Printf("Sent all blocks to %s\n", p.key)
	m := makeMessage(MessageAllBlocksResponse, blockchain.GetBlockchain(blockchain.Blockchain()))
	p.inbox <- m
}

func notifyNewBlock(b *blockchain.Block, p *peer) { // sends the MessageKind with the Block to the peer
	m := makeMessage(MessageNewBlockNotify, b)
	p.inbox <- m
}

func notifyNewTx(b *blockchain.Tx, p *peer) { // sends the MessageKind with the tx to the peer
	m := makeMessage(MessageNewTxNotify, b)
	p.inbox <- m
}

func notifyNewPeer(b string, p *peer) { // sends the MessageKind with the tx to the peer
	m := makeMessage(MessageNewPeerNotify, b)
	p.inbox <- m
}
