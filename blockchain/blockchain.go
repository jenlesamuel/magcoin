package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger"
	"github.com/jenlesamuel/magcoin/share"
)

const (
	LastBlockHeaderHash = "last_block_header_hash"
)

type Blockchain struct {
	DB                  *badger.DB
	LastBlockHeaderHash [32]byte
}

func InitBlockchain(db *badger.DB) (*Blockchain, error) {

	var lastBlockHeaderHash []byte

	err := db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(LastBlockHeaderHash))
		if err != nil {
			if err == badger.ErrKeyNotFound { // blockchain not yet persisted in db
				genesisBlock := NewBlock(
					share.Uint32ToByte4(uint32(1)),
					share.Int32ToByte32(0),
					share.SliceToByte32([]byte("Genesis")),
				)
				genesisBlockHeaderHash := genesisBlock.HeaderHash()
				lastBlockHeaderHash = genesisBlockHeaderHash[:]
				genesisBlockBytes, err := genesisBlock.Serialize()
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

	blockChain := &Blockchain{DB: db, LastBlockHeaderHash: share.SliceToByte32(lastBlockHeaderHash)}

	return blockChain, nil
}

func (bc *Blockchain) AddBlock(data string) error {

	lastBlockHeaderHash := bc.LastBlockHeaderHash
	version := share.Uint32ToByte4(1)
	dataByte32 := share.SliceToByte32([]byte(data))

	block := NewBlock(version, lastBlockHeaderHash, dataByte32)
	blockHeaderHash := block.HeaderHash()
	blockBytes, err := block.Serialize()
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

func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{
		DB:          bc.DB,
		CurrentHash: bc.LastBlockHeaderHash,
	}
}

type BlockchainIterator struct {
	DB          *badger.DB
	CurrentHash [32]byte
}

func (iterator *BlockchainIterator) Next() (*Block, error) {
	var next *Block

	err := iterator.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iterator.CurrentHash[:])
		if err != nil {
			return err
		}

		err = item.Value(func(value []byte) error {
			next, err = FromBytes(value)
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
