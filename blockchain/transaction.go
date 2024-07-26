package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/jenlesamuel/magcoin/cryptography"
	"github.com/jenlesamuel/magcoin/share"
)

func init() {
	gob.Register(TrxInput{})
	gob.Register(TrxOutput{})
	gob.Register(CoinbaseInput{})
	gob.Register(cryptography.Signature{})
	gob.Register(StdTransaction{})
	gob.Register(CoinbaseTransaction{})
}

type TrxInput struct {
	OutpointHash  [32]byte // the hash of the referenced transaction
	OutpointIndex [4]byte  //the index of the referenced transaction output
	Signature     *cryptography.Signature
	PublicKey     []byte
}

type TrxOutput struct {
	Amount        uint64 //amount of maglia
	PublicKeyHash [20]byte
}

type CoinbaseInput struct {
	Data []byte
}

type CoinbaseTransaction struct {
	Input  *CoinbaseInput
	Output []*TrxOutput
}

type StdTransaction struct {
	Input  []*TrxInput
	Output []*TrxOutput
}

// TODO: implement validate
func (st *StdTransaction) Validate() bool {
	return false
}

func (st *StdTransaction) Hash() ([32]byte, error) {
	// TODO: separate signature from hash to prevent transaction malleability
	trxBytes, err := st.Serialize()
	if err != nil {
		return [32]byte{}, err
	}

	return share.DoubleSha256(trxBytes), nil
}

func (st *StdTransaction) Serialize() ([]byte, error) {
	buff := new(bytes.Buffer)
	empty := make([]byte, 0)

	for idx, input := range st.Input {
		if _, err := buff.Write(input.OutpointHash[:]); err != nil {
			return empty, err
		}

		if _, err := buff.Write(input.OutpointIndex[:]); err != nil {
			return empty, err
		}

		if _, err := buff.Write(input.PublicKey); err != nil {
			return empty, err
		}

		rBytes := new(big.Int).Set(input.Signature.R).Bytes()
		sBytes := new(big.Int).Set(input.Signature.S).Bytes()

		if _, err := buff.Write(rBytes[:]); err != nil {
			return empty, err
		}

		if _, err := buff.Write(sBytes[:]); err != nil {
			return empty, err
		}

		output := st.Output[idx]

		amountBytes := share.Int64ToByte8(int64(output.Amount))
		if _, err := buff.Write(amountBytes[:]); err != nil {
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

func (tm *TransactionManager) CreateStdTransaction(inputs []*TrxInput, outputs []*TrxOutput) (*StdTransaction, error) {
	if len(inputs) != len(outputs) {
		return nil, errors.New("transaction input and output must be of the same length")
	}

	trx := &StdTransaction{
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
		amountBytes8 := share.Int64ToByte8(int64(output.Amount))
		if _, err := buff.Write(amountBytes8[:]); err != nil {
			return nil, err
		}

		if _, err := buff.Write(output.PublicKeyHash[:]); err != nil {
			return nil, err
		}

		trxSign, err := tm.keymanager.Sign(buff.Bytes())
		if err != nil {
			return nil, err
		}

		input.Signature = trxSign
	}

	hash, err := trx.Hash()
	if err != nil {
		return nil, err
	}

	hashHex := hex.EncodeToString(hash[:])

	tm.mempool.AddTransaction(hashHex, trx)

	return trx, nil
}

func (tm *TransactionManager) CreateCoinbaseTransaction(data string) (*CoinbaseTransaction, error) {
	input := &CoinbaseInput{
		Data: []byte(data),
	}

	pkHash, err := share.GetPublicKeyHashFromPublicKey(tm.keymanager.PublicKey)
	if err != nil {
		return nil, err
	}

	output := &TrxOutput{
		Amount:        5_000_000_000, // 5,000,000,000 maglia, equivalent of 1 magcoin
		PublicKeyHash: pkHash,
	}

	return &CoinbaseTransaction{Input: input, Output: []*TrxOutput{output}}, nil
}
