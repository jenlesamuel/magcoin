package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jenlesamuel/magcoin/cryptography"
	"github.com/jenlesamuel/magcoin/share"
)

func init() {
	gob.Register(TrxInput{})
	gob.Register(TrxOutput{})
	gob.Register(Transaction{})
}

type TrxInput struct {
	OutpointHash  [32]byte // the hash of the referenced transaction
	OutpointIndex [4]byte  //the index of the referenced transaction output
	SigOrData     []byte
	PublicKey     []byte
}

type TrxOutput struct {
	Amount        [8]byte //amount of maglia
	PublicKeyHash [20]byte
}

type Transaction struct {
	Input  []*TrxInput
	Output []*TrxOutput
}

// TODO: implement validate
func (trx *Transaction) Validate() bool {
	return false
}

func (trx *Transaction) Hash() ([32]byte, error) {
	// TODO: separate signature from hash to prevent transaction malleability
	trxBytes, err := trx.Serialize()
	if err != nil {
		return [32]byte{}, err
	}

	return share.DoubleSha256(trxBytes), nil
}

func (trx *Transaction) Serialize() ([]byte, error) {
	buff := new(bytes.Buffer)
	empty := make([]byte, 0)

	for idx, input := range trx.Input {
		if _, err := buff.Write(input.OutpointHash[:]); err != nil {
			return empty, err
		}

		if _, err := buff.Write(input.OutpointIndex[:]); err != nil {
			return empty, err
		}

		if _, err := buff.Write(input.PublicKey); err != nil {
			return empty, err
		}

		if _, err := buff.Write(input.SigOrData); err != nil {
			return empty, err
		}

		output := trx.Output[idx]

		if _, err := buff.Write(output.Amount[:]); err != nil {
			return empty, err
		}

		if _, err := buff.Write(output.PublicKeyHash[:]); err != nil {
			return empty, err
		}
	}

	return buff.Bytes(), nil
}

type TransactionManager struct {
	mempool    *MemPool
	keymanager *cryptography.KeyManager
}

func NewTransactionManager(mempool *MemPool, keymanager *cryptography.KeyManager) *TransactionManager {
	return &TransactionManager{
		mempool:    mempool,
		keymanager: keymanager,
	}
}

func (tm *TransactionManager) CreateStdTransaction(inputs []*TrxInput, outputs []*TrxOutput) (*Transaction, error) {
	if len(inputs) != len(outputs) {
		return nil, errors.New("transaction input and output must be of the same length")
	}

	trx := &Transaction{
		Input:  inputs,
		Output: outputs,
	}

	buff := new(bytes.Buffer)

	for idx, input := range trx.Input {
		if _, err := buff.Write(input.OutpointHash[:]); err != nil {
			return nil, err
		}

		if _, err := buff.Write(input.OutpointIndex[:]); err != nil {
			return nil, err
		}

		pkBytes, err := share.GetPublicKeyBytes(tm.keymanager.PublicKey)
		if err != nil {
			return nil, err
		}
		input.PublicKey = pkBytes

		if _, err := buff.Write(pkBytes); err != nil {
			return nil, err
		}

		// parse trx output
		output := trx.Output[idx]
		if _, err := buff.Write(output.Amount[:]); err != nil {
			return nil, err
		}

		if _, err := buff.Write(output.PublicKeyHash[:]); err != nil {
			return nil, err
		}

		trxSign, err := tm.keymanager.Sign(buff.Bytes())
		if err != nil {
			return nil, err
		}

		input.SigOrData = trxSign.Bytes()
	}

	hash, err := trx.Hash()
	if err != nil {
		return nil, err
	}

	hashHex := hex.EncodeToString(hash[:])

	tm.mempool.AddTransaction(hashHex, trx)

	return trx, nil
}

func (tm *TransactionManager) CreateCoinbaseTransaction(data string) (*Transaction, error) {
	dataBytes := []byte(data)
	if len(dataBytes) < 2 || len(dataBytes) > 100 {
		return nil, fmt.Errorf("coinbase data size of %d exceeds limit of 2 - 100 bytes", len(dataBytes))
	}

	input := &TrxInput{
		OutpointHash:  [32]byte{},
		OutpointIndex: [4]byte{0x01, 0x00, 0x00, 0x00}, //big endian
		SigOrData:     dataBytes,
		PublicKey:     []byte{},
	}

	pkHash, err := share.GetPublicKeyHashFromPublicKey(tm.keymanager.PublicKey)
	if err != nil {
		return nil, err
	}

	output := &TrxOutput{
		Amount:        share.Int64ToByte8(5_000_000_000), // 5,000,000,000 maglia, equivalent of 1 magcoin
		PublicKeyHash: pkHash,
	}

	return &Transaction{Input: []*TrxInput{input}, Output: []*TrxOutput{output}}, nil
}
