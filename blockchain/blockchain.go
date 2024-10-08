package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger"
)

const (
	LastBlockHeaderHash = "last_block_header_hash"
)

type Blockchain struct {
	DB                  *badger.DB
	LastBlockHeaderHash []byte
}

func LoadBlockchain(db *badger.DB, genesisBlock *Block) (*Blockchain, error) {

	var lastBlockHeaderHash []byte

	err := db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(LastBlockHeaderHash))
		if err != nil {
			if err == badger.ErrKeyNotFound { // blockchain not yet persisted in db

				if !genesisBlock.Mine() {
					return fmt.Errorf("could not mine Genesis block: %s", err)
				}

				headerHash := genesisBlock.HeaderHash()
				lastBlockHeaderHash = headerHash[:]
				genesisBlockBytes, err := genesisBlock.Encode()
				if err != nil {
					return fmt.Errorf("error serializing genesis block: %s", err)
				}

				if err = txn.Set([]byte(LastBlockHeaderHash), lastBlockHeaderHash); err != nil {
					return fmt.Errorf("could not persist last block hash to db: %s", err)
				}

				if err = txn.Set(lastBlockHeaderHash, genesisBlockBytes[:]); err != nil {
					return fmt.Errorf("could not persist block to db: %s", err)
				}

			} else {
				return fmt.Errorf("could not fetch last block header hash from db: %s", err)
			}
		} else {
			err = item.Value(func(value []byte) error {
				lastBlockHeaderHash = value
				return nil
			})

			if err != nil {
				return fmt.Errorf("error parsing last block header hash: %s", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not initialize blockchain from db: %s", err)
	}

	blockChain := &Blockchain{
		DB:                  db,
		LastBlockHeaderHash: lastBlockHeaderHash,
	}

	return blockChain, nil
}

func (bc *Blockchain) AddBlock(block *Block) error {

	var err error

	if err = block.Validate(); err != nil {
		return err
	}

	blockHeaderHash := block.HeaderHash()
	blockBytes, err := block.Encode()
	if err != nil {
		return fmt.Errorf("could not serialize block: %s", err)
	}

	err = bc.DB.Update(func(txn *badger.Txn) error {
		if err = txn.Set(blockHeaderHash[:], blockBytes); err != nil {
			return err
		}

		if err = txn.Set([]byte(LastBlockHeaderHash), blockHeaderHash[:]); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("could not persist block to db: %s", err)
	}

	bc.LastBlockHeaderHash = blockHeaderHash

	return nil
}

func (bc *Blockchain) Iterator() *BlockIterator {
	return &BlockIterator{
		DB:          bc.DB,
		CurrentHash: bc.LastBlockHeaderHash,
	}
}

type BlockIterator struct {
	DB          *badger.DB
	CurrentHash []byte
}

func (iterator *BlockIterator) Next() (*Block, error) {
	var next *Block

	err := iterator.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iterator.CurrentHash[:])

		if err != nil {
			return err
		}

		err = item.Value(func(value []byte) error {
			next, err = DecodeToBlock(value)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	iterator.CurrentHash = next.PreviousHash

	return next, nil
}
