package blockchain

import (
	"bytes"
	"encoding/gob"
	"errors"
	"math/big"
	"time"

	"github.com/jenlesamuel/magcoin/share"
	"github.com/jenlesamuel/magcoin/transaction"
)

const MaxBlockSize = 5

var ErrMaxBlockSizeExceeded = errors.New("maximum block size exceeded")

type Block struct {
	Version      []byte //4 bytes
	PreviousHash []byte //32 bytes
	MerkleRoot   []byte // 32 bytes
	Nonce        []byte // 4 bytes
	Target       []byte
	Timestamp    []byte // 8 bytes
	Transactions []*transaction.Transaction
}

func DecodeToBlock(data []byte) (*Block, error) {
	r := bytes.NewReader(data)

	decoder := gob.NewDecoder(r)
	block := new(Block)
	if err := decoder.Decode(block); err != nil {
		return nil, err
	}

	return block, nil
}

func (block *Block) HeaderHashWithNonce(nonce []byte) []byte {
	concat := bytes.Join([][]byte{
		block.Version[:],
		block.PreviousHash[:],
		block.MerkleRoot[:],
		nonce[:],
		block.Timestamp[:],
	}, []byte{})

	hash := share.DoubleSha256(concat)
	return hash[:]
}

func (block *Block) HeaderHash() []byte {
	return block.HeaderHashWithNonce(block.Nonce)
}

func (block *Block) Encode() ([]byte, error) {
	buff := new(bytes.Buffer)

	encoder := gob.NewEncoder(buff)
	if err := encoder.Encode(block); err != nil {
		return make([]byte, 0), err
	}

	return buff.Bytes(), nil
}

func (block *Block) AddTransaction(trx *transaction.Transaction) error {
	if len(block.Transactions) > MaxBlockSize {
		return ErrMaxBlockSizeExceeded
	}

	// TODO: validate transaction before adding to block
	block.Transactions = append(block.Transactions, trx)
	return nil
}

func (block *Block) Mine() bool {
	pow := NewProofOfWork(block)
	return pow.Run()
}

func (block *Block) Validate() error {
	//TODO: validate other consensus rukes for block
	if !block.validatePOW() {
		return errors.New("proof of work validation failed")
	}

	return nil
}

func (block *Block) validatePOW() bool {
	blockHash := block.HeaderHash()
	blockHashInt := new(big.Int).SetBytes(blockHash[:])

	targetInt := new(big.Int).SetBytes(block.Target[:])

	return blockHashInt.Cmp(targetInt) == -1
}

func (block *Block) IsGenesis() bool {
	return bytes.Equal(block.PreviousHash[:], make([]byte, 32))
}

type BlockManager struct {
	transactionManager *transaction.TransactionManager
}

func NewBlockManager(tm *transaction.TransactionManager) *BlockManager {
	return &BlockManager{
		transactionManager: tm,
	}
}

func (bm *BlockManager) CreateBlock(previousHash []byte, coinbaseData string) (*Block, error) {
	coinbase, err := bm.transactionManager.CreateCoinbaseTransaction(coinbaseData)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()

	return &Block{
		Version:      share.IntToBytes(1),
		PreviousHash: previousHash,
		MerkleRoot:   make([]byte, 32),
		Timestamp:    share.Int64ToBytes(timestamp),
		Transactions: []*transaction.Transaction{coinbase},
	}, nil
}

// Creates the first block in the blockchain
func (bm *BlockManager) GenesisBlock() (*Block, error) {
	coinbase, err := bm.transactionManager.CreateCoinbaseTransaction("MagCoin: Bitcoin Parody 0x1F923")
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()

	block := &Block{
		Version:      share.IntToBytes(1),
		PreviousHash: make([]byte, 32),
		MerkleRoot:   make([]byte, 32),
		Timestamp:    share.Int64ToBytes(timestamp),
		Transactions: []*transaction.Transaction{coinbase},
	}

	return block, nil
}
