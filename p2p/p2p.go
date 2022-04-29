package p2p

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/jeyoungjung/zerocoin/blockchain"
	"github.com/jeyoungjung/zerocoin/utils"
)

var upgrader = websocket.Upgrader{}

func Upgrade(rw http.ResponseWriter, r *http.Request) {
	// Port :3000 will upgrade the request from :4000
	openPort := r.URL.Query().Get("openPort")        // gets the port the upgrade was requested from
	ip := utils.StringSplitter(r.RemoteAddr, ":", 0) // r.RemoteAddr gets the address where the request was sent from
	// splits the 127.0.0.1:4000 at the ":" and returns the [0] index, so the 127.0.0.1
	upgrader.CheckOrigin = func(r *http.Request) bool { // this just gets rid of the error,
		// try without this function and where would be an error in the console
		return openPort != "" && ip != ""
	}
	fmt.Printf(":%s wants an upgrade\n", openPort)
	conn, err := upgrader.Upgrade(rw, r, nil) // Upgrades http to ws
	utils.HandleErr(err)
	initPeer(conn, ip, openPort)
}

func AddPeer(address, port, openPort string, broadcast bool) { // this function takes in the its PEER's port
	// Port :4000 is requesting an upgrade from the port :3000
	fmt.Printf("%s wants to connect to port %s\n", openPort, port)
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s:%s/ws?openPort=%s", address, port, openPort), nil) // it is going to call the Upgrade function, which will upgrade that page to Websocket
	utils.HandleErr(err)
	p := initPeer(conn, address, port)
	if broadcast { // if the peer is 100% new to the network, and needs to be broadcasted
		broadcastNewPeer(p)
	}
	sendNewestBlock(p) // when a new peer is added, you always want to send the newest block to keep the new peer updated
	// with the current most up to date blockchain.
}

func BroadcastNewBlock(b *blockchain.Block) {
	Peers.m.Lock()
	defer Peers.m.Unlock()
	for _, p := range Peers.v { // goes through every peer to send the new Block to
		notifyNewBlock(b, p)
	}
}

func BroadcastNewTx(tx *blockchain.Tx) {
	Peers.m.Lock()
	defer Peers.m.Unlock()
	for _, p := range Peers.v { // goes through every peer to send the new tx to
		notifyNewTx(tx, p)
	}
}

func broadcastNewPeer(newPeer *peer) {
	for key, p := range Peers.v {
		if key != newPeer.key { // if the peer is not the newPeer, notify the peer that a new peer was added
			// an example of key is something like "127.0.0.1:3000"
			payload := fmt.Sprintf("%s:%s", newPeer.key, p.port) // send the key of the new peer, and the port of the p its sending to
			// you have to send the port of itself because they won't know what port they are
			// if the new port is 2000 and the sender is 4000, the payload will looks something like:
			// "127.0.0.1:2000:4000"
			notifyNewPeer(payload, p)
		}
	}
}
