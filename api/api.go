package api

import (
	"github.com/jenlesamuel/magcoin/blockchain"
	"github.com/jenlesamuel/magcoin/transaction"
	"github.com/jenlesamuel/magcoin/wallet"
)

type API struct {
	blockIterator *blockchain.BlockIterator
	walletManager *wallet.WalletManager
}

func NewAPI(bi *blockchain.BlockIterator, wm *wallet.WalletManager) *API {
	return &API{
		blockIterator: bi,
		walletManager: wm,
	}
}

func (api *API) GetIterator() *blockchain.BlockIterator {
	return api.blockIterator
}

func (api *API) CreateTransaction(amount uint64, receiverAddress string) (*transaction.Transaction, error) {
	return api.walletManager.CreateTransaction(amount, receiverAddress)
}
