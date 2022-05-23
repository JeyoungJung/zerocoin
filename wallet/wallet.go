package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"

	"github.com/jeyoungjung/zerocoin/utils"
)

const (
	fileName string = "zerocoin.wallet"
)

var w *wallet

type wallet struct {
	privateKey *ecdsa.PrivateKey
	Address    string
}

func hasWalletFile() bool {
	_, err := os.Stat(fileName)
	return os.IsExist(err) // if the file exists return true
}

func Wallet() *wallet {
	if w == nil {
		w = &wallet{}
		if hasWalletFile() {
			w.privateKey = restoreKey()
		} else {
			key := createPrivKey() // if wallet doesnt exist, create new priv key
			persistKey(key)        // keep the priv key on a new file
			w.privateKey = key
		}

		w.Address = addressFromKey(w.privateKey)
	}
	return w
}

func createPrivKey() *ecdsa.PrivateKey {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	utils.HandleErr(err)
	return privKey
}

func persistKey(key *ecdsa.PrivateKey) {
	bytes, err := x509.MarshalECPrivateKey(key) // takes privkey, returns as bytes
	utils.HandleErr(err)
	err = os.WriteFile(fileName, bytes, 0644) // make a new file consisting the priv key
	utils.HandleErr(err)
}

func restoreKey() *ecdsa.PrivateKey {
	keyAsBytes, err := os.ReadFile(fileName)
	utils.HandleErr(err)
	key, err := x509.ParseECPrivateKey(keyAsBytes)
	utils.HandleErr(err)
	return key
}

// addressFromKey gets the public key (aka address)
func addressFromKey(key *ecdsa.PrivateKey) string {
	// this makes the publicKey to be together as a string
	// if you do address := key.PublicKey it will include the eliptic curve as an element, and the X and Y would be divided into 2
	// (Look at the key.PublicKey struct to be reminded)
	// by appending the X and Y we have 1 string that can be used as a publicKey.
	return encodeBigInts(key.X.Bytes(), key.Y.Bytes())
}

func Sign(payload string, w *wallet) string { // for signature you will need: the data you want to sign (payload) + privateKey
	payloadBytes := decodeString(payload)
	r, s, err := ecdsa.Sign(rand.Reader, w.privateKey, payloadBytes)
	utils.HandleErr(err)
	return encodeBigInts(r.Bytes(), s.Bytes())
}

func encodeBigInts(a, b []byte) string {
	z := append(a, b...)
	return fmt.Sprintf("%x", z)
}

// restoreBigInts gets string, return them in big Ints
func restoreBigInts(payload string) (*big.Int, *big.Int, error) {
	bytes, err := hex.DecodeString(payload) // decodes string into bytes
	if err != nil {
		return nil, nil, err
	}
	firstHalfBytes := bytes[:len(bytes)/2] // divides the slice of bytes into 2 equal parts
	secondHalfBytes := bytes[len(bytes)/2:]
	bigA, bigB := big.Int{}, big.Int{}
	bigA.SetBytes(firstHalfBytes)
	bigB.SetBytes(secondHalfBytes)
	return &bigA, &bigB, nil
}

func decodeString(payload string) []byte {
	payloadBytes, err := hex.DecodeString(payload)
	utils.HandleErr(err)
	return payloadBytes
}

// Verify checks if the transaction is the owner's
// for verification you will need, the signature, payload and publicKey (address)
func Verify(signature, payload, address string) bool {
	r, s, err := restoreBigInts(signature) // changed the string into bigInts
	utils.HandleErr(err)
	x, y, err := restoreBigInts(address) // changed the string into bigInts
	utils.HandleErr(err)
	publicKey := ecdsa.PublicKey{
		Curve: elliptic.P256(), // we are using p256 curve
		X:     x,
		Y:     y,
	}
	payloadBytes := decodeString(payload)
	ok := ecdsa.Verify(&publicKey, payloadBytes, r, s)
	return ok
}
