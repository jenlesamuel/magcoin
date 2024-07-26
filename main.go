package main

import (
	"log"

	"github.com/jenlesamuel/magcoin/blockchain"
	"github.com/jenlesamuel/magcoin/cli"
	"github.com/jenlesamuel/magcoin/cryptography"
)

func main() {

	// Init DB
	db, err := blockchain.InitDB()
	if err != nil {
		log.Fatalf("could not intitialize DB: %s", err)
	}
	defer db.Close()

	// Init KeyManager
	keysPath := "/tmp"
	keymanager, err := cryptography.LoadKeyManager(keysPath)
	if err != nil {
		log.Panicf("%s\n", err)
	}

	// Init Mempool
	mempool := blockchain.NewMemPool()

	// Init TransactionManager
	transactionManager := blockchain.NewTransactionManager(mempool, keymanager)

	// Init BlockManager
	blockManager := blockchain.NewBlockManager(transactionManager)

	// Init Blockchain
	bc, err := blockchain.LoadBlockchain(db, blockManager)
	if err != nil {
		log.Panicf("%s\n", err)
	}

	// Run CLI
	cli := cli.NewCommandLine(bc, transactionManager)
	cli.Exec()
}
