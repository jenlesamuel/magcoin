package blockchain

import (
	"bytes"
	"encoding/gob"
	"errors"
	"math/big"
	"time"

	"github.com/jenlesamuel/magcoin/share"
)

const MaxBlockSize = 5

var ErrMaxBlockSizeExceeded = errors.New("maximum block size exceeded")

type Block struct {
	Version      [4]byte
	PreviousHash [32]byte
	MerkleRoot   [32]byte
	Nonce        [4]byte
	Target       [32]byte
	Timestamp    [8]byte
	Coinbase     *CoinbaseTransaction
	// had issue with encoding block, so had to separate coinbase from standard transaction
	Transactions []*StdTransaction
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

func (block *Block) HeaderHashWithNonce(nonce [4]byte) [32]byte {
	concat := bytes.Join([][]byte{
		block.Version[:],
		block.PreviousHash[:],
		block.MerkleRoot[:],
		nonce[:],
		block.Timestamp[:],
	}, []byte{})

	return share.DoubleSha256(concat)
}

func (block *Block) HeaderHash() [32]byte {
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

func (block *Block) AddTransaction(trx *StdTransaction) error {
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

type BlockManager struct {
	transactionManager *TransactionManager
}

func NewBlockManager(tm *TransactionManager) *BlockManager {
	return &BlockManager{
		transactionManager: tm,
	}
}

func (bm *BlockManager) CreateBlock(previousHash [32]byte, coinbaseData string) (*Block, error) {
	coinbase, err := bm.transactionManager.CreateCoinbaseTransaction(coinbaseData)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	block := &Block{
		Version:      share.Uint32ToByte4(uint32(1)),
		PreviousHash: previousHash,
		MerkleRoot:   [32]byte{},
		Timestamp:    share.Int64ToByte8(timestamp),
		Coinbase:     coinbase,
		Transactions: make([]*StdTransaction, 0),
	}

	counter := 0
	for _, trx := range bm.transactionManager.mempool.transactions {
		if counter == MaxBlockSize {
			break
		}

		block.AddTransaction(trx)
		counter += 1
	}

	return block, nil
}
