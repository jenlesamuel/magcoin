package blockchain

import "sync"

type MemPool struct {
	transactions map[string]*StdTransaction
	mu           sync.RWMutex
}

func NewMemPool() *MemPool {
	return &MemPool{
		transactions: make(map[string]*StdTransaction),
	}
}

func (mp *MemPool) AddTransaction(idx string, trx *StdTransaction) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.transactions[idx] = trx
}

func (mp *MemPool) DeleteTransaction(idx string) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	delete(mp.transactions, idx)
}

func (mp *MemPool) GetTransaction(idx string) *StdTransaction {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	return mp.transactions[idx]
}
