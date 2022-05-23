package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jeyoungjung/zerocoin/blockchain"
	"github.com/jeyoungjung/zerocoin/p2p"
	"github.com/jeyoungjung/zerocoin/utils"
	"github.com/jeyoungjung/zerocoin/wallet"
)

var port string

type url string

func (u url) MarshalText() ([]byte, error) {
	url := fmt.Sprintf("http://localhost%s%s", port, u)
	return []byte(url), nil
}

type urlDescription struct {
	URL         url    `json:"url"`
	Method      string `json:"method"`
	Description string `json:"description"`
	Payload     string `json:"payload,omitempty"`
}

func documentation(rw http.ResponseWriter, r *http.Request) {
	data := []urlDescription{
		{
			URL:         url("/"),
			Method:      "GET",
			Description: "See description",
		}, {
			URL:         url("/blocks"),
			Method:      "POST",
			Description: "Add A Block",
			Payload:     "data:string",
		}, {
			URL:         url("/status"),
			Method:      "GET",
			Description: "See the Status of the Blockchain",
		}, {
			URL:         url("/blocks/{hash}"),
			Method:      "GET",
			Description: "See A Block",
		}, {
			URL:         url("/balance/{address}"),
			Method:      "GET",
			Description: "Get TxOuts for an Address",
		}, {
			URL:         url("/ws"),
			Method:      "GET",
			Description: "Upgrade to WebSockets",
		},
	}
	json.NewEncoder(rw).Encode(data)
}

// type addBlockBody struct {
// 	Message string
// }

// You may ask, why struct?
// It has to be a struct decode something.
// "error: cannot unmarshal object into Go value of type string" if it was `var addBlockBody string` instead of a struct

func blocks(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		json.NewEncoder(rw).Encode(blockchain.GetBlockchain(blockchain.Blockchain())) //converts the blockchain data to json
	case "POST":
		//var addBlockBody addBlockBody
		//utils.HandleErr(json.NewDecoder(r.Body).Decode(&addBlockBody)) // this function returns an error, hence the utils.HandleErr()
		// Explanation: new decoder is made, the the r.body (consisting of data like "second block") is decoded into the actual addBlockBody
		// https://stackoverflow.com/questions/21197239/decoding-json-using-json-unmarshal-vs-json-newdecoder-decode
		newBlock := blockchain.Blockchain().AddBlock()
		p2p.BroadcastNewBlock(newBlock)
		rw.WriteHeader(http.StatusCreated)
	}
}

type errorResponse struct {
	ErrorMessage string `json:"errorMessage"`
}

func block(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r) // from the documentation for gorilla mux
	hash := vars["hash"]
	block, err := blockchain.FindBlock(hash)
	encoder := json.NewEncoder(rw)
	if err == blockchain.ErrBlockNotFound { // if error is same as the error we made in blockchain.go
		encoder.Encode(errorResponse{fmt.Sprint(err)}) // put the error inside the errorResponse and encode it as json
	} else {
		encoder.Encode(block)
	}
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler { // this function is a middleware that sets the content-type to be json,
	// this let's the browser know that the file we are passing is json.
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(rw, r) // this just goes to the next handler
	})
}

func status(rw http.ResponseWriter, r *http.Request) {
	blockchain.Status(blockchain.Blockchain(), rw)
}

type balanceResponse struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

func loggerMiddleware(next http.Handler) http.Handler { // just logs which url it's on
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		next.ServeHTTP(rw, r)
	})
}

func balance(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	address := vars["address"]
	total := r.URL.Query().Get("total")
	switch total {
	case "true": // if "http://localhost:4000/balance/zero?total=true" return the total balance
		amount := blockchain.TotalBalanceByAddress(address, blockchain.Blockchain())
		json.NewEncoder(rw).Encode(balanceResponse{address, amount})
	default: // if "http://localhost:4000/balance/zero" return receipts for each transactions
		utils.HandleErr(json.NewEncoder(rw).Encode(blockchain.UTxOutsByAddress(address, blockchain.Blockchain())))
	}
}

func mempool(rw http.ResponseWriter, r *http.Request) {
	utils.HandleErr(json.NewEncoder(rw).Encode(blockchain.Mempool()))
}

type addTxPayload struct {
	To     string
	Amount int
}

func transactions(rw http.ResponseWriter, r *http.Request) { // this is a POST only function
	// the payload consists of "To" and "Amount" which will send that much amount to that someone.
	// if there is an error, it means that theres not enough money.
	var payload addTxPayload
	utils.HandleErr(json.NewDecoder(r.Body).Decode(&payload))
	newTx, err := blockchain.Mempool().AddTx(payload.To, payload.Amount)
	if err != nil {
		json.NewEncoder(rw).Encode(errorResponse{"not enough funds"})
		return
	}
	p2p.BroadcastNewTx(newTx) // sends this transaction to other peers
}

type myWalletResponse struct {
	Address string `json:"address"`
}

func myWallet(rw http.ResponseWriter, r *http.Request) {
	address := wallet.Wallet().Address
	json.NewEncoder(rw).Encode(myWalletResponse{Address: address})
	// json.NewEncoder(rw).Encode(struct {
	// 	Address string `json:"address"`
	// }{Address: address})
	// the commented out way is called a anonymous struct
}

type addPeerPayload struct {
	Address, Port string
}

func peers(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		var payload addPeerPayload
		json.NewDecoder(r.Body).Decode(&payload)
		p2p.AddPeer(payload.Address, payload.Port, port[1:], true)
		rw.WriteHeader(http.StatusOK)
	case "GET":
		json.NewEncoder(rw).Encode(p2p.GetPeers(&p2p.Peers))
	}
}

func Start(startPort int) {
	port = fmt.Sprintf(":%d", startPort)
	router := mux.NewRouter()
	router.Use(jsonContentTypeMiddleware, loggerMiddleware)
	router.HandleFunc("/", documentation).Methods("GET")
	router.HandleFunc("/status", status).Methods("GET")
	router.HandleFunc("/blocks", blocks).Methods("GET", "POST")
	router.HandleFunc("/balance/{address}", balance).Methods("GET")
	router.HandleFunc("/mempool", mempool).Methods("GET")
	router.HandleFunc("/transactions", transactions).Methods("POST")
	router.HandleFunc("/wallet", myWallet).Methods("GET")
	router.HandleFunc("/ws", p2p.Upgrade).Methods("GET")
	router.HandleFunc("/peers", peers).Methods("GET", "POST")
	router.HandleFunc("/blocks/{hash:[a-f0-9]+}", block).Methods("GET") // means that the hash can have values from a-f and 0-9 (hexadecimal)
	fmt.Printf("Listening on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
