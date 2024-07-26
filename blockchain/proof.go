package blockchain

import (
	"math/big"

	"github.com/jenlesamuel/magcoin/share"
)

const Difficulty = 12

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))

	return &ProofOfWork{
		Block:  block,
		Target: target,
	}
}

func (p *ProofOfWork) Run() bool {
	nonce := uint32(0)
	intHash := new(big.Int)

	for nonce < ^uint32(0) {
		nonceByte4 := share.Uint32ToByte4(nonce)

		blockHash := p.Block.HeaderHashWithNonce(nonceByte4)
		intHash.SetBytes(blockHash[:])

		if intHash.Cmp(p.Target) == -1 {
			p.Block.Nonce = nonceByte4
			p.Block.Target = share.SliceToByte32(p.Target.Bytes())

			return true
		}

		nonce += 1
	}

	return false
}
