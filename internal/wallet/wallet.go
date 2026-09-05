package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
)

type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

const signaturePartSize = 32

func NewWallet() (*Wallet, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		PrivateKey: privKey,
		PublicKey:  &privKey.PublicKey,
	}, nil
}

func (w *Wallet) Address() string {
	return PublicKeyToAddress(w.PublicKey)
}

func PublicKeyToAddress(pub *ecdsa.PublicKey) string {
	bytes := elliptic.Marshal(pub.Curve, pub.X, pub.Y)
	return hex.EncodeToString(bytes)
}

func AddressToPublicKey(address string) (*ecdsa.PublicKey, error) {
	bytes, err := hex.DecodeString(address)

	if err != nil {
		return nil, err
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), bytes)
	if x == nil {
		return nil, errors.New("endereço invalido")
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

func (w *Wallet) Sign(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)

	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash[:])
	if err != nil {
		return nil, err
	}

	signature := make([]byte, signaturePartSize*2)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	copy(signature[signaturePartSize-len(rBytes):signaturePartSize], rBytes)
	copy(signature[len(signature)-len(sBytes):], sBytes)
	return signature, nil
}

func VerifySignature(pub *ecdsa.PublicKey, data, signature []byte) bool {
	if len(signature) != signaturePartSize*2 {
		return false
	}

	hash := sha256.Sum256(data)

	r := new(big.Int).SetBytes(signature[:signaturePartSize])
	s := new(big.Int).SetBytes(signature[signaturePartSize:])

	return ecdsa.Verify(pub, hash[:], r, s)

}
