package share

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"math/big"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type Signature struct {
	R *big.Int
	S *big.Int
}

func GeneratePrivateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func Sign(hash []byte, privateKey *ecdsa.PrivateKey) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return nil, err
	}

	signature := new(Signature)
	signature.R = r
	signature.S = s

	return signature, nil
}

func VerifySignature(publicKey *ecdsa.PublicKey, hash []byte, signature *Signature) bool {
	return ecdsa.Verify(publicKey, hash, signature.R, signature.S)
}

func GetPublicKeyBytes(publicKey *ecdsa.PublicKey) ([]byte, error) {
	pkBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return make([]byte, 0), nil
	}
	return pkBytes, nil
}

func GetPublicKeyHashFromPublicKey(publicKey *ecdsa.PublicKey) ([20]byte, error) {
	publicKeyBytes, err := GetPublicKeyBytes(publicKey)
	if err != nil {
		return [20]byte{}, err
	}

	hash := sha256.Sum256(publicKeyBytes)

	r160Hasher := ripemd160.New()
	_, err = r160Hasher.Write(hash[:])
	if err != nil {
		return [20]byte{}, err
	}
	b := r160Hasher.Sum(nil)

	b20 := [20]byte{}
	copy(b20[:], b)

	return b20, nil
}

func AddressFromPublicKey(publicKey *ecdsa.PublicKey) (string, error) {
	pkHash, err := GetPublicKeyHashFromPublicKey(publicKey)
	if err != nil {
		return "", err
	}

	doubleHash := DoubleSha256(pkHash[:])
	checksum := doubleHash[0:4]

	addressBytes := pkHash[:]
	addressBytes = append(addressBytes, checksum...)

	return base58.Encode(addressBytes), nil
}

func ValidateAddress(address string) bool {
	addressBytes := base58.Decode(address)
	publicKeyHash := addressBytes[:20]
	checksum := addressBytes[20:]

	doubleHash := DoubleSha256(publicKeyHash)

	return bytes.Equal(doubleHash[:4], checksum)
}

// func HexToPublicKey(pkHex string) (*ecdsa.PublicKey, error) {
// 	pkHexBytes, err := hex.DecodeString(pkHex)
// 	if err != nil {
// 		return nil, err
// 	}

// 	//validate public key bytes.
// 	if len(pkHexBytes) != 65 || pkHexBytes[0] != 0x04 {
// 		return nil, errors.New("invalid public key format")
// 	}

// 	x := new(big.Int).SetBytes(pkHexBytes[1:33])
// 	y := new(big.Int).SetBytes(pkHexBytes[33:])

// 	publicKey := &ecdsa.PublicKey{
// 		Curve: elliptic.P256(),
// 		X:     x,
// 		Y:     y,
// 	}

// 	return publicKey, nil
// }

func DoubleSha256(data []byte) [32]byte {
	singleHash := sha256.Sum256(data)

	return sha256.Sum256(singleHash[:])
}

func GetPublicKeyHashFromAddress(address string) [20]byte {
	addressBytes := base58.Decode(address)

	pkHash := [20]byte{}
	copy(pkHash[:], addressBytes[:20])

	return pkHash
}
