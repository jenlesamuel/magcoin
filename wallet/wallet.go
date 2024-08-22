package wallet

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/jenlesamuel/magcoin/blockchain"
	"github.com/jenlesamuel/magcoin/share"
	"github.com/jenlesamuel/magcoin/transaction"
)

const (
	ErrInsufficientBalance = "insufficient balance"
	ErrInvalidAddress      = "invalid address"
)

type Wallet struct {
	Balance uint64
	Address string
}

type WalletManager struct {
	iterator   *blockchain.BlockIterator
	keymanager *share.KeyManager
	mempool    *transaction.MemPool
}

func NewWalletManager(
	iterator *blockchain.BlockIterator,
	keymanager *share.KeyManager,
	mempool *transaction.MemPool,
) *WalletManager {
	return &WalletManager{
		iterator:   iterator,
		keymanager: keymanager,
		mempool:    mempool,
	}
}

// Retrieves all the UTXOs for an address
func (wm *WalletManager) getUTXO(address string) ([]*transaction.UTXO, error) {
	pkHash := share.PublicKeyHashFromAddress(address)
	spendableTransactionOutputs := make(map[string]map[int]int64)
	empty := make([]*transaction.UTXO, 0)

	for {
		block, err := wm.iterator.Next()
		if err != nil {
			return empty, err
		}

		for _, trx := range block.Transactions {
			trxIDHex := hex.EncodeToString(trx.ID)
			for idx, output := range trx.Output {
				if !bytes.Equal(output.PublicKeyHash, pkHash) { // output not meant for address
					continue
				}
				amount, err := share.BytesToInt64(output.Amount)
				if err != nil {
					return empty, err
				}

				if _, exists := spendableTransactionOutputs[trxIDHex]; !exists {
					spendableTransactionOutputs[trxIDHex] = make(map[int]int64)
				}
				spendableTransactionOutputs[trxIDHex][idx] = amount
			}

			if trx.IsCoinbase() {
				// Coinbase input does not reference any transaction output.
				// Ignore it.
				continue
			}

			for _, input := range trx.Input {
				outHashHex := hex.EncodeToString(input.OutpointHash)
				outIdx, err := share.BytesToInt(input.OutpointIndex)

				if err != nil {
					return empty, err
				}

				delete(spendableTransactionOutputs[outHashHex], outIdx)
			}
		}

		if block.IsGenesis() {
			break
		}

	}

	utxos := make([]*transaction.UTXO, 0)
	for trxIDHex, out := range spendableTransactionOutputs {
		trxID, err := hex.DecodeString(trxIDHex)
		if err != nil {
			return empty, err
		}

		for outIdx, amount := range out {
			utxo := &transaction.UTXO{
				TransactionHash: trxID,
				OutpointIndex:   share.IntToBytes(outIdx),
				Amount:          share.Int64ToBytes(amount),
			}

			utxos = append(utxos, utxo)
		}
	}

	return utxos, nil
}

// Get the balance from UTXOs
func (wm *WalletManager) getUTXOAmount(utxos []*transaction.UTXO) (int64, error) {
	var balance int64

	for _, utxo := range utxos {
		amount, err := share.BytesToInt64(utxo.Amount)
		if err != nil {
			return 0, err
		}

		balance += amount
	}

	return balance, nil
}

func (wm *WalletManager) GetAddressBalance(senderAddress string) (int64, error) {
	utxos, err := wm.getUTXO(senderAddress)
	if err != nil {
		return 0, err
	}

	return wm.getUTXOAmount(utxos)
}

// Returns UTXOs, their total amount sum and nil if sender has enough balance to pay amount
// and no error occured. Returns zero values and error instead.
// TODO: optimize method to prevent a scenario where only little denomination (change) utxos exist
func (wm *WalletManager) getUTXOForAmount(amount uint64, senderAddress string) ([]*transaction.UTXO, uint64, error) {
	empty := make([]*transaction.UTXO, 0)

	utxos, err := wm.getUTXO(senderAddress)
	if err != nil {
		return empty, uint64(0), err
	}

	sum := uint64(0)
	res := make([]*transaction.UTXO, 0)
	for _, utxo := range utxos {
		if sum >= amount {
			break
		}

		amount, err := share.BytesToInt64(utxo.Amount)
		if err != nil {
			return empty, uint64(0), err
		}

		res = append(res, utxo)
		sum += uint64(amount)
	}

	if sum < amount {
		return empty, uint64(0), errors.New(ErrInsufficientBalance)
	}

	return res, sum, nil
}

func (wm *WalletManager) CreateTransaction(amount uint64, receiverAddress string) (*transaction.Transaction, error) {
	if !share.ValidateAddress(receiverAddress) {
		return nil, errors.New(ErrInvalidAddress)
	}

	utxos, total, err := wm.getUTXOForAmount(amount, receiverAddress)
	if err != nil {
		return nil, err
	}

	inputs := make([]*transaction.TrxInput, 0)
	outputs := make([]*transaction.TrxOutput, 0)

	pAmountBytes := share.Int64ToBytes(int64(amount))
	pkHash := share.PublicKeyHashFromAddress(receiverAddress)

	// Payment output is the output that represents the amount to be sent to the receiver
	paymentOutput := &transaction.TrxOutput{
		Amount:        pAmountBytes,
		PublicKeyHash: pkHash,
	}
	outputs = append(outputs, paymentOutput)

	outBuf := new(bytes.Buffer)
	outBuf.Write(pAmountBytes)
	outBuf.Write(pkHash)

	if total > amount {
		// Change Output is the output that represents the change paid back to the sender.
		// Imagine you need to pay a fee of $25 but have a $100 bill, you'll pay the $100
		// but get a change of $75
		cAmountBytes := share.Int64ToBytes(int64(total - amount))

		changeOutput := &transaction.TrxOutput{
			Amount:        cAmountBytes,
			PublicKeyHash: pkHash,
		}
		outputs = append(outputs, changeOutput)

		outBuf.Write(cAmountBytes)
		outBuf.Write(pkHash)
	}

	pubKey, err := share.GetPublicKeyBytes(wm.keymanager.PublicKey)
	if err != nil {
		return nil, err
	}

	for _, utxo := range utxos {
		input := &transaction.TrxInput{
			OutpointHash:  utxo.TransactionHash,
			OutpointIndex: utxo.OutpointIndex,
			PublicKey:     pubKey,
		}

		inBuf := new(bytes.Buffer)
		inBuf.Write(utxo.TransactionHash)
		inBuf.Write(utxo.OutpointIndex)
		inBuf.Write(pubKey)

		// Create signature
		combo := bytes.Join([][]byte{inBuf.Bytes(), outBuf.Bytes()}, []byte{})
		hash := share.DoubleSha256(combo)
		signature, err := wm.keymanager.Sign(hash[:])
		if err != nil {
			return nil, err
		}
		input.SigOrData = signature.Bytes()

		inputs = append(inputs, input)
	}

	trx, err := transaction.NewTransaction(inputs, outputs)
	if err != nil {
		return nil, err
	}

	trxHashHex := hex.EncodeToString(trx.ID[:])

	wm.mempool.AddTransaction(trxHashHex, trx)

	return trx, nil
}
