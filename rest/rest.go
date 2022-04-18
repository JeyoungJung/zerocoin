package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jeyoungjung/zerocoin/blockchain"
	"github.com/jeyoungjung/zerocoin/utils"
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
			URL:         url("/blocks/{hash}"),
			Method:      "GET",
			Description: "See A Block",
		},
	}
	json.NewEncoder(rw).Encode(data)
}

type addBlockBody struct {
	Message string
}

// You may ask, why struct?
// It has to be a struct decode something.
// "error: cannot unmarshal object into Go value of type string" if it was `var addBlockBody string` instead of a struct

func blocks(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		json.NewEncoder(rw).Encode(blockchain.Blockchain().GetBlockchain())
		//converts the blockchain data to json
	case "POST":
		var addBlockBody addBlockBody
		utils.HandleErr(json.NewDecoder(r.Body).Decode(&addBlockBody)) // this function returns an error, hence the utils.HandleErr()
		// Explanation: new decoder is made, the the r.body (consisting of data like "second block") is decoded into the actual addBlockBody
		// https://stackoverflow.com/questions/21197239/decoding-json-using-json-unmarshal-vs-json-newdecoder-decode
		blockchain.Blockchain().AddBlock(addBlockBody.Message)
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

func Start(startPort int) {
	port = fmt.Sprintf(":%d", startPort)
	router := mux.NewRouter()
	router.Use(jsonContentTypeMiddleware)
	router.HandleFunc("/", documentation)
	router.HandleFunc("/blocks", blocks)
	router.HandleFunc("/blocks/{hash:[a-f0-9]+}", block) // means that the hash can have values from a-f and 0-9 (hexadecimal)
	fmt.Printf("Listening on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
