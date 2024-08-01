package blockchain

import "sync"

type MemPool struct {
	transactions map[string]*Transaction
	mu           sync.RWMutex
}

func NewMemPool() *MemPool {
	return &MemPool{
		transactions: make(map[string]*Transaction),
	}
}

func (mp *MemPool) AddTransaction(idx string, trx *Transaction) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.transactions[idx] = trx
}

func (mp *MemPool) DeleteTransaction(idx string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	delete(mp.transactions, idx)
}

func (mp *MemPool) GetTransaction(idx string) *Transaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.transactions[idx]
}
