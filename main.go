package main

import (
	"log"

	"github.com/jenlesamuel/magcoin/blockchain"
	"github.com/jenlesamuel/magcoin/cli"
)

func main() {

	db, err := blockchain.InitDB()
	if err != nil {
		log.Fatalf("could not intitialize DB: %s", err)
	}

	defer db.Close()

	bc, err := blockchain.InitBlockchain(db)
	if err != nil {
		log.Panicf("%s\n", err) //panic so that db can be closed
	}

	cli := cli.NewCommandLine(bc)
	if err = cli.Exec(); err != nil {
		log.Panicf("%s\n", err) //panic so that db can be closed
	}
}
