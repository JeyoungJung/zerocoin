package db

import (
	"fmt"
	"os"

	"github.com/jeyoungjung/zerocoin/utils"
	bolt "go.etcd.io/bbolt"
)

const (
	dbName           = "blockchain"
	checkpointBucket = "checkpoints"
	blocksBucket     = "blocks"
	checkpoint       = "checkpoint"
)

var db *bolt.DB

func getDbName() string {
	port := os.Args[2][6:]
	return fmt.Sprintf("%s_%s.db", dbName, port)
}

func DB() *bolt.DB { // initializes the database
	if db == nil {
		dbPointer, err := bolt.Open(getDbName(), 0600, nil) // 0600 is the code required for read and write permissions?
		db = dbPointer                                      // now that the database is made, we can assign it to the original db
		utils.HandleErr(err)
		err = db.Update(func(t *bolt.Tx) error { // makes buckets, this is the syntax shown in the actual github for bolt (just copy paste)
			_, err := t.CreateBucketIfNotExists([]byte(checkpointBucket)) // creates a bucket named "checkpoints", this bucket only holds the data for the checkpoints
			utils.HandleErr(err)
			_, err = t.CreateBucketIfNotExists([]byte(blocksBucket)) // creates a bucket named "blocks"
			utils.HandleErr(err)
			return err
		})
		utils.HandleErr(err)
	}
	return db
}

func SaveBlock(hash string, data []byte) { // this function saves the actual data for the block
	// [hash : data] key value pair
	err := DB().Update(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(blocksBucket))
		err := bucket.Put([]byte(hash), data)
		return err
	})
	utils.HandleErr(err)
}

func SaveCheckpoint(data []byte) { // this function saves the data for the checkpoint only
	// ["checkpoint": data]
	err := DB().Update(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(checkpointBucket))
		err := bucket.Put([]byte(checkpoint), data) // since the "key" has to be unique,
		// the data is replaced everytime there is a new block as the "key" remains to be "checkpoint" all the time
		return err
	})
	utils.HandleErr(err)
}

func GetCheckpointData() []byte { // this function retrieves the data for the checkpoint
	var data []byte // has to be declared to return the data
	DB().View(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(checkpointBucket))
		data = bucket.Get([]byte(checkpoint)) // gets the data from the checkpoint
		return nil
	})
	return data
}

func GetBlockData(hash string) []byte {
	var data []byte // has to be declared to return the data
	DB().View(func(t *bolt.Tx) error {
		bucket := t.Bucket([]byte(blocksBucket)) // goes to the blockbucket see if theres the hash that was given
		data = bucket.Get([]byte(hash))          // if there is, get the data for that block and put it into data
		return nil
	})
	return data
}

func CloseDatabase() {
	DB().Close()
}

func EmptyBlocks() { // deletes the current blocksbucket and creates a new one
	DB().Update(func(t *bolt.Tx) error {
		utils.HandleErr(t.DeleteBucket([]byte(blocksBucket)))
		_, err := t.CreateBucket([]byte(blocksBucket))
		utils.HandleErr(err)
		return nil
	})
}
