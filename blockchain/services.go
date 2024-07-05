package blockchain

import (
	"github.com/dgraph-io/badger"
)

const dbPath = "./tmp/blocks"

func InitDB() (*badger.DB, error) {
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	return db, err
}
