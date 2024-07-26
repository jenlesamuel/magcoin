package cli

import (
	"errors"
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
	Blockchain         *blockchain.Blockchain
	TransactionManager *blockchain.TransactionManager
}

func NewCommandLine(bc *blockchain.Blockchain, tm *blockchain.TransactionManager) *CommandLine {
	return &CommandLine{Blockchain: bc, TransactionManager: tm}
}

func (cli *CommandLine) printHelp() {
	fmt.Println(`
		magcoin is a demo blockchain for educational purposes.

		Usage:

			magcoin <command> [arguments]
		
		The commands are:

		publish				print all the blocks in the blockchain
		create-transaction  creates a standard transaction i.e a non-coinbase transaction
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

func (cli *CommandLine) Exec() {
	args := os.Args

	if len(args) == 1 {
		log.Panic(errors.New("no command specified"))
	}

	switch args[1] {
	case "publish":
		if err := cli.printBlockchain(); err != nil {
			log.Panic(err)
		}
	/*case "create-transaction":
	if err := cli.execCreateTransaction(); err != nil {
		log.Panic(err)
	}*/
	default:
		cli.printHelp()
	}
}

/*func (cli *CommandLine) execCreateTransaction() error {
	outpointHashes := flag.String("outpoint-hashes", "", "command delimited list of referenced transaction hashes")
	outpointIndices := flag.String("outpoint-indices", "", "command delimited list of referenced transaction output indices that corresponding"+
		"to the outpoint hashes")
	receiverAddresses := flag.String("receiver-addresses", "", "a comma delimited list of the addresses of the receivers")
	amounts := flag.String("amounts", "", "a comma delimited list of the amounts in Maglia that coresponds to the"+
		"public-key-hashes of the receivers")

	flag.Parse()

	*outpointHashes = strings.Trim(*outpointHashes, " ")
	*outpointIndices = strings.Trim(*outpointIndices, " ")
	*receiverAddresses = strings.Trim(*receiverAddresses, " ")
	*amounts = strings.Trim(*amounts, " ")

	if *outpointHashes == "" {
		return errors.New("outpoint-hashes cannot be empty")
	}

	if *outpointIndices == "" {
		return errors.New("outpoint-indices cannot be empty")
	}

	if *receiverAddresses == "" {
		return errors.New("public-key-hashes cannot be empty")
	}

	if *amounts == "" {
		return errors.New("amounts cannot be empty")
	}

	outpointHashesArr := strings.Split(*outpointHashes, ",")
	outpointIndicesArr := strings.Split(*outpointIndices, ",")
	receiverAddressesArr := strings.Split(*receiverAddresses, ",")
	amountsArr := strings.Split(*amounts, ",")

	minLength := min(len(outpointHashesArr), len(outpointIndicesArr), len(receiverAddressesArr), len(amountsArr))

	var trxInputs []*blockchain.TrxInput
	var trxOutputs []*blockchain.TrxOutput

	for i := 0; i < minLength; i++ {
		outpointHashByte, err := hex.DecodeString(outpointHashesArr[i])
		if err != nil {
			return err
		}
		outpointHashByte32 := share.SliceToByte32(outpointHashByte)

		outpointIndexByte, err := hex.DecodeString(outpointIndicesArr[i])
		if err != nil {
			return err
		}
		outpointIndexByte4 := share.SliceToByte4(outpointIndexByte)

		publicKeyHashByte20 := share.GetPublicKeyHashFromAddress(receiverAddressesArr[i])

		amount, err := strconv.Atoi(amountsArr[i])
		if err != nil {
			return err
		}

		trxInput := &blockchain.TrxInput{
			OutpointHash:  outpointHashByte32,
			OutpointIndex: outpointIndexByte4,
		}

		trxOutput := &blockchain.TrxOutput{
			Amount:        uint64(amount),
			PublicKeyHash: publicKeyHashByte20,
		}

		trxInputs = append(trxInputs, trxInput)
		trxOutputs = append(trxOutputs, trxOutput)

	}

	if _, err := cli.TransactionManager.CreateStdTransaction(trxInputs, trxOutputs); err != nil {
		return err
	}

	return nil
}

func min(nums ...int) int {
	min := nums[0]

	for _, num := range nums {
		if num < min {
			min = num
		}
	}

	return min
}*/

func (cli *CommandLine) printBlockchain() error {
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
