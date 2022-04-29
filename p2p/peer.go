package p2p

import (
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type peers struct { // made a struct so we can use mutex
	v map[string]*peer
	m sync.Mutex
}

var Peers peers = peers{ // "Peers" have the map
	v: make(map[string]*peer),
}

type peer struct {
	port    string
	address string
	key     string
	conn    *websocket.Conn
	inbox   chan []byte
}

func initPeer(conn *websocket.Conn, address, port string) *peer {
	Peers.m.Lock()
	defer Peers.m.Unlock()
	key := fmt.Sprintf("%s:%s", address, port)
	p := &peer{
		port:    port,
		address: address,
		key:     key,
		conn:    conn,
		inbox:   make(chan []byte),
	}
	go p.read() // this function runs concurrently, so the initPeer() will go on, but p.read() will still be running
	go p.write()
	Peers.v[key] = p // adds this peer to the map. ex) "127.0.0.1:4000" : peer
	return p
}

func (p *peer) read() { // this function reads messages
	defer p.close() // closes the peer at the end of the function (means that the connection has been closed)
	for {
		m := Message{}
		err := p.conn.ReadJSON(&m)
		if err != nil {
			break
		}
		handleMsg(&m, p)
	}
}

func (p *peer) write() { // this function writes messages
	defer p.close() // closes the peer at the end of the function (means that the connection has been closed)
	for {
		m, ok := <-p.inbox // waits for something to be received from the channel
		if !ok {
			break
		}
		p.conn.WriteMessage(websocket.TextMessage, m) // write the received message to the ws
	}
}

func (p *peer) close() {
	Peers.m.Lock()         // locks it so that no other functions would be able to read it while this function is running
	defer Peers.m.Unlock() // is unlocked when the function is done running
	p.conn.Close()
	delete(Peers.v, p.key) // delete the peer that disconnected from the map of peers
}

func GetPeers(p *peers) []string {
	p.m.Lock()
	defer p.m.Unlock()
	var keys []string
	for key := range p.v {
		keys = append(keys, key)
	}
	return keys
}
