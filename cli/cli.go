package cli

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jenlesamuel/magcoin/api"
	"github.com/jenlesamuel/magcoin/share"
	"github.com/jenlesamuel/magcoin/transaction"
)

const (
	Add     = "add"
	Publish = "publish"
)

type CommandLine struct {
	api *api.API
}

func NewCommandLine(api *api.API) *CommandLine {
	return &CommandLine{api: api}
}

func (cli *CommandLine) printHelp() {
	fmt.Println(`
		magcoin is a demo blockchain for educational purposes.

		Usage:

			magcoin [flags] <command> [arguments]
		
		The commands are:

		publish				print all the blocks in the blockchain
		create-transaction  creates a standard transaction i.e a non-coinbase transaction
	`)
}

func (cli *CommandLine) Exec() {
	args := os.Args

	if len(args) == 1 {
		cli.printHelp()
		return
	}

	switch args[1] {
	case "publish":
		if err := cli.printBlockchain(); err != nil {
			log.Panic(err)
		}
	case "create-transaction":
		var trx *transaction.Transaction
		var err error
		if trx, err = cli.execCreateTransaction(); err != nil {
			log.Panic(err)
		}
		log.Printf("Transaction Created: %+v", trx)
	default:
		cli.printHelp()
	}
}

func (cli *CommandLine) execCreateTransaction() (*transaction.Transaction, error) {
	os.Args = os.Args[1:]
	receiverAddress := flag.String("receiver-address", "", "address of the receiver")
	amount := flag.Uint64("amount", 0, "amount to be sent to the receiver in maglia (100,000,000 maglia = 1 magcoin)")

	flag.Parse()

	*receiverAddress = strings.TrimSpace(*receiverAddress)

	if *receiverAddress == "" {
		return nil, errors.New("receiver address cannot be empty")
	}

	if *amount < 1 || *amount > share.MAX_MAGLIA {
		return nil, fmt.Errorf("transaction amount should be minimum of 1 maglia and less than %d maglias (21 million magcoins)", share.MAX_MAGLIA)
	}

	return cli.api.CreateTransaction(*amount, *receiverAddress)
}

func (cli *CommandLine) printBlockchain() error {
	iterator := cli.api.GetIterator()

	log.Println("/**********Blocks**********/")

	for {
		block, err := iterator.Next()

		if err != nil {
			return fmt.Errorf("error occured while printing blockchain: %s", err)
		}

		timestamp, err := share.BytesToInt64(block.Timestamp)
		if err != nil {
			return fmt.Errorf("could not convert [8]byte to int64: %s", err)
		}

		log.Printf("Hash: %X\t", block.HeaderHash())
		log.Printf("Previous Hash: %X\t", block.PreviousHash)
		log.Printf("Timestamp: %d\n", timestamp)

		if block.IsGenesis() {
			break
		}
	}

	log.Println("/**********Blocks**********/")

	return nil
}
