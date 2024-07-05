package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
	"time"

	"github.com/jenlesamuel/magcoin/share"
)

type Block struct {
	Version      [4]byte
	PreviousHash [32]byte
	Data         [32]byte
	Nonce        [4]byte
	Target       [32]byte
	Timestamp    [8]byte
}

func NewBlock(version [4]byte, previousHash, data [32]byte) *Block {

	timestamp := time.Now().Unix()
	block := &Block{
		Version:      version,
		PreviousHash: previousHash,
		Data:         data,
		Timestamp:    share.Int64ToByte8(timestamp),
	}

	pow := NewProofOfWork(block)
	pow.Run()

	return block
}

func FromBytes(data []byte) (*Block, error) {
	buff := bytes.NewBuffer(data)
	b := new(Block)

	err := binary.Read(buff, binary.BigEndian, b)
	if err != nil {
		return nil, fmt.Errorf("error creating block from bytes: %s", err)
	}

	return b, nil
}

func (block *Block) HeaderHashFromNonce(nonce [4]byte) [32]byte {
	concat := bytes.Join([][]byte{
		block.Version[:],
		block.PreviousHash[:],
		block.Data[:],
		nonce[:],
		block.Timestamp[:],
	}, []byte{})

	firstHash := sha256.Sum256(concat)

	return sha256.Sum256(firstHash[:])
}

func (block *Block) HeaderHash() [32]byte {
	return block.HeaderHashFromNonce(block.Nonce)
}

func (block *Block) Validate() bool {
	blockHash := block.HeaderHash()
	blockHashInt := new(big.Int).SetBytes(blockHash[:])

	targetInt := new(big.Int).SetBytes(block.Target[:])

	return blockHashInt.Cmp(targetInt) == -1
}

func (block *Block) Serialize() ([]byte, error) {
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.BigEndian, block)
	if err != nil {
		return []byte{}, fmt.Errorf("error serializing block: %s", err)
	}

	return buff.Bytes(), nil
}
