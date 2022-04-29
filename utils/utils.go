package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func HandleErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func EncodeToBytes(i interface{}) []byte { // https://stackoverflow.com/questions/16330490/in-go-how-can-i-convert-a-struct-to-a-byte-array
	var Result bytes.Buffer                      // A buffer is just a container or holding tank to read data from or write data to
	HandleErr(gob.NewEncoder(&Result).Encode(i)) // **important** this code will be replicated a lot, just know
	return Result.Bytes()
}

func DecodeFromBytesToStruct(data []byte, i interface{}) { // here i is the block struct, and the data is still in bytes
	// this function decodes the data in bytes, to the actual block in the struct form
	// since i is the pointer to the actual blockchain, you dont have to return anything, the change is made on the actual thing
	decoder := gob.NewDecoder(bytes.NewReader(data))
	decoder.Decode(i) // the data is now decoded into i
}

func Hash(i interface{}) string {
	s := fmt.Sprintf("%v", i)
	hash := sha256.Sum256([]byte(s)) // hashes the sum of everything inside "i"
	return fmt.Sprintf("%x", hash)   // returns the value in hexadecimal characters but in string format
}

func StringSplitter(s string, sep string, i int) string {
	r := strings.Split(s, sep)
	if len(r)-1 < i {
		return ""
	}
	return r[i]
}

func MarshalToJSON(i interface{}) []byte {
	r, err := json.Marshal(i)
	HandleErr(err)
	return r
}
