package share

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"
)

const PrivateKeyFilename = "mag_ecdsa_private_key.pem"
const AddressFilename = "mag_wallet_address.txt"

type Signature struct {
	R *big.Int
	S *big.Int
}

func (s *Signature) Bytes() []byte {
	rBytes := s.R.Bytes()
	sBytes := s.S.Bytes()

	return bytes.Join([][]byte{rBytes, sBytes}, []byte{})
}

// TODO: implement
func SignatureFromBytes(b []byte) *Signature {
	return nil
}

type KeyManager struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

func LoadKeyManager(dest string) (*KeyManager, error) {
	if dest == "" {
		dest = "/tmp"
	}

	var privateKey *ecdsa.PrivateKey

	privateKey, err := loadPrivateKeyFromFile(fmt.Sprintf("%s/%s", dest, PrivateKeyFilename))
	if err != nil {
		privateKey, err = generatePrivateKey()
		if err != nil {
			return nil, err
		}

		if err = writePrivateKeyToFile(privateKey, dest); err != nil {
			return nil, err
		}

		address, err := AddressFromPublicKey(&privateKey.PublicKey)
		if err != nil {
			return nil, err
		}

		path := fmt.Sprintf("%s/%s", dest, AddressFilename)
		if err = writeAddressToFile(address, path); err != nil {
			return nil, err
		}
	}

	return &KeyManager{PrivateKey: privateKey, PublicKey: &privateKey.PublicKey}, nil
}

func (km *KeyManager) Sign(hash []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, km.PrivateKey, hash)
	if err != nil {
		return nil, err
	}

	signature := new(Signature)
	signature.R = r
	signature.S = s

	return signature, nil
}

func (km *KeyManager) VerifySignature(hash []byte, signature *Signature) bool {
	return ecdsa.Verify(km.PublicKey, hash, signature.R, signature.S)
}

func generatePrivateKey() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

func writePrivateKeyToFile(privateKey *ecdsa.PrivateKey, dest string) error {
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return err
	}

	pemBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	dest = strings.TrimSuffix(dest, "/")
	file, err := os.Create(fmt.Sprintf("%s/%s", dest, PrivateKeyFilename))
	if err != nil {
		return err
	}
	defer file.Close()

	if err := pem.Encode(file, pemBlock); err != nil {
		return err
	}

	return nil
}

func loadPrivateKeyFromFile(file string) (*ecdsa.PrivateKey, error) {
	privateKeyPEM, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	pemBlock, _ := pem.Decode(privateKeyPEM)
	if pemBlock == nil || pemBlock.Type != "EC PRIVATE KEY" {
		return nil, errors.New("could not decode PEM block")
	}

	privateKey, err := x509.ParseECPrivateKey(pemBlock.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func (km *KeyManager) GetPublicKeyHash() ([20]byte, error) {
	return GetPublicKeyHashFromPublicKey(km.PublicKey)
}

func (km *KeyManager) GetAddress() (string, error) {
	return AddressFromPublicKey(km.PublicKey)
}

func writeAddressToFile(address, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(address); err != nil {
		return err
	}

	return nil
}
