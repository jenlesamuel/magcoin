package main

import (
	"log"

	"github.com/jenlesamuel/magcoin/api"
	"github.com/jenlesamuel/magcoin/blockchain"
	"github.com/jenlesamuel/magcoin/cli"
	"github.com/jenlesamuel/magcoin/share"
	"github.com/jenlesamuel/magcoin/transaction"
	"github.com/jenlesamuel/magcoin/wallet"
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
	keymanager, err := share.LoadKeyManager(keysPath)
	if err != nil {
		log.Panicf("%s\n", err)
	}

	// Init Mempool
	mempool := transaction.NewMemPool()

	// Init TransactionManager
	transactionManager := transaction.NewTransactionManager(keymanager)

	// Init BlockManager
	blockManager := blockchain.NewBlockManager(transactionManager)

	genesisBlock, err := blockManager.GenesisBlock()
	if err != nil {
		log.Panicf("%s\n", err)
	}

	// Init Blockchain
	bc, err := blockchain.LoadBlockchain(db, genesisBlock)
	if err != nil {
		log.Panicf("%s\n", err)
	}

	//Init Wallet Manager
	walletManager := wallet.NewWalletManager(bc.Iterator(), keymanager, mempool)

	// Init API
	api := api.NewAPI(bc.Iterator(), walletManager)

	// Run CLI
	cli := cli.NewCommandLine(api)
	cli.Exec()
}
