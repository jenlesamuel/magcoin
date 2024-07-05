package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/jenlesamuel/magcoin/blockchain"
	"github.com/jenlesamuel/magcoin/share"
)

const (
	Add     = "add"
	Publish = "publish"
)

type CommandLine struct {
	Blockchain *blockchain.Blockchain
}

func NewCommandLine(bc *blockchain.Blockchain) *CommandLine {
	return &CommandLine{Blockchain: bc}
}

func (cli *CommandLine) PrintHelp() {
	fmt.Println(`
		magcoin is a demo blockchain for educational purposes.

		Usage:

			magcoin <command> [arguments]
		
		The commands are:

		publish				print all the blocks in the blockchain
		add <data>			adds a block with data as content to the blockchain. <data> is an argument.
		
	`)
}

func (cli *CommandLine) ValidateArgs(args []string) bool {
	if len(args) == 1 {
		return false
	}

	if args[1] == Add && len(args) < 3 {
		return false
	}

	return true
}

func (cli *CommandLine) Exec() error {
	args := os.Args

	if !cli.ValidateArgs(args) {
		cli.PrintHelp()
		return nil
	}

	command := args[1]

	switch command {
	case Publish:
		return cli.PrintBlockchain()
	case Add:
		if err := cli.Blockchain.AddBlock(args[2]); err != nil {
			return err
		}
		log.Println("Block Added")
		return nil

	default:
		cli.PrintHelp()
	}

	return nil
}

func (cli *CommandLine) PrintBlockchain() error {
	iterator := cli.Blockchain.Iterator()

	log.Println("/**********Blocks**********/")

	for {
		block, err := iterator.Next()
		if err != nil {
			return fmt.Errorf("error occured while printing blockchain: %s", err)
		}

		timestamp, err := share.Byte8ToInt64(block.Timestamp)
		if err != nil {
			return fmt.Errorf("could not convert [8]byte to int64: %s", err)
		}

		log.Printf("Hash: %X\t", block.HeaderHash())
		log.Printf("Previous Hash: %X\t", block.PreviousHash)
		log.Printf("Timestamp: %d\n", timestamp)

		if share.IsZeroArray(block.PreviousHash) {
			break
		}
	}

	log.Println("/**********Blocks**********/")

	return nil
}
