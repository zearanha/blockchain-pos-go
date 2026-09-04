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
	PublicKey *ecdsa.PublicKey
}

func NewWallet() (*Wallet, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		PrivateKey: privKey,
		PublicKey: &privKey.PublicKey,
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

	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

func VerifySignature(pub *ecdsa.PublicKey, data, signature []byte) bool {
	hash := sha256.Sum256(data)

	half := len(signature) / 2
	r := new(big.Int).SetBytes(signature[:half])
	s := new(big.Int).SetBytes(signature[half:])

	return ecdsa.Verify(pub, hash[:], r, s)

}