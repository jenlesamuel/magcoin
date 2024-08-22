package transaction

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jenlesamuel/magcoin/share"
)

type TrxInput struct {
	OutpointHash  []byte // the hash of the referenced transaction (32 bytes
	OutpointIndex []byte //the index of the referenced transaction output (4 bytes)
	SigOrData     []byte
	PublicKey     []byte
}

type TrxOutput struct {
	Amount        []byte //amount of maglia (8 bytes)
	PublicKeyHash []byte // 20 bytes
}

type Transaction struct {
	ID     []byte // aka Transaction Hash, 32 bytes
	Input  []*TrxInput
	Output []*TrxOutput
}

type UTXO struct {
	TransactionHash []byte // 32 bytes
	OutpointIndex   []byte // 4 bytes
	Amount          []byte // 8 bytes
}

// Returns a new non-coinbase transaction
func NewTransaction(inputs []*TrxInput, outputs []*TrxOutput) (*Transaction, error) {
	trx := &Transaction{Input: inputs, Output: outputs}
	hash, err := trx.Hash(false)
	if err != nil {
		return nil, err
	}

	trx.ID = hash[:]

	return trx, nil
}

// Returns a new coinbase transaction
// A coinbase transaction is the first transaction added to a block.
// It references no previous transaction and its output is paid to the miner of the block.
func NewCoinbaseTransaction(inputs []*TrxInput, outputs []*TrxOutput) (*Transaction, error) {
	trx := &Transaction{Input: inputs, Output: outputs}
	hash, err := trx.Hash(true)
	if err != nil {
		return nil, err
	}

	trx.ID = hash[:]

	return trx, nil
}

func (trx *Transaction) IsCoinbase() bool {
	if len(trx.Input) != 1 {
		return false
	}

	input := trx.Input[0]

	return bytes.Equal(input.OutpointHash, make([]byte, 32)) &&
		bytes.Equal(input.OutpointIndex, []byte{0xFF, 0xFF, 0xFF, 0xFF})
}

// TODO: implement validate
func (trx *Transaction) Validate() bool {
	return false
}

func (trx *Transaction) Hash(withSigOrData bool) ([32]byte, error) {
	trxBytes, err := trx.Serialize(withSigOrData)
	if err != nil {
		return [32]byte{}, err
	}

	return share.DoubleSha256(trxBytes), nil
}

func (trx *Transaction) Serialize(withSigOrData bool) ([]byte, error) {
	buff := new(bytes.Buffer)
	empty := make([]byte, 0)

	for _, input := range trx.Input {
		if _, err := buff.Write(input.OutpointHash[:]); err != nil {
			return empty, err
		}

		if _, err := buff.Write(input.OutpointIndex[:]); err != nil {
			return empty, err
		}

		if _, err := buff.Write(input.PublicKey); err != nil {
			return empty, err
		}

		if withSigOrData {
			if _, err := buff.Write(input.SigOrData); err != nil {
				return empty, err
			}
		}
	}

	for _, output := range trx.Output {
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
	keyManager *share.KeyManager
}

func NewTransactionManager(km *share.KeyManager) *TransactionManager {
	return &TransactionManager{keyManager: km}
}

func (tm *TransactionManager) CreateCoinbaseTransaction(data string) (*Transaction, error) {
	dataBytes := []byte(data)
	if len(dataBytes) < 2 || len(dataBytes) > 100 {
		return nil, fmt.Errorf("coinbase data size of %d exceeds limit of 2 - 100 bytes", len(dataBytes))
	}
	timestampBytes := share.Int64ToBytes(time.Now().UnixMilli())
	dataBytes = append(dataBytes, timestampBytes...)

	input := &TrxInput{
		OutpointHash:  make([]byte, 32),
		OutpointIndex: []byte{0xFF, 0xFF, 0xFF, 0xFF},
		SigOrData:     dataBytes,
		PublicKey:     make([]byte, 0),
	}

	pkHash, err := share.GetPublicKeyHashFromPublicKey(tm.keyManager.PublicKey)
	if err != nil {
		return nil, err
	}

	output := &TrxOutput{
		Amount:        share.Int64ToBytes(5_000_000_000), // 5,000,000,000 maglia, equivalent of 50 magcoin
		PublicKeyHash: pkHash[:],
	}

	return NewCoinbaseTransaction(
		[]*TrxInput{input},
		[]*TrxOutput{output},
	)
}
