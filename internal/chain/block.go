package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/zearanha/blockchain-pos-go/internal/wallet"
)

type Block struct {
	Index        int
	Timestamp    int64
	Transactions []Transaction
	PrevHash     string
	Hash         string
	Validator    string
	Signature    []byte `json:"signature"`
}

func NewBlock(index int, transactions []Transaction, prevHash string, validator string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PrevHash:     prevHash,
		Validator:    validator,
	}
	block.Hash = block.CalculateHash()
	return block
}

func (b *Block) CalculateHash() string {
	record := struct {
		Index        int
		Timestamp    int64
		Transactions []Transaction
		PrevHash     string
		Validator    string
	}{
		Index:        b.Index,
		Timestamp:    b.Timestamp,
		Transactions: b.Transactions,
		PrevHash:     b.PrevHash,
		Validator:    b.Validator,
	}
	data, _ := json.Marshal(record)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (b *Block) IsHashValid() bool {
	return b.Hash == b.CalculateHash()
}

func (b *Block) dataToSign() []byte {
	return []byte(b.Hash)
}

func SignBlock(block *Block, w *wallet.Wallet) error {
	if w == nil {
		return errors.New("carteira do validador é obrigatória")
	}
	if block.Validator != w.Address() {
		return errors.New("a carteira não corresponde ao validador do bloco")
	}

	block.Hash = block.CalculateHash()

	signature, err := w.Sign(block.dataToSign())
	if err != nil {
		return err
	}

	block.Signature = signature
	return nil
}

func VerifyBlock(block *Block) (bool, error) {
	if !block.IsHashValid() {
		return false, errors.New("hash do bloco inválido")
	}
	if len(block.Signature) == 0 {
		return false, errors.New("bloco sem assinatura")
	}

	pub, err := wallet.AddressToPublicKey(block.Validator)
	if err != nil {
		return false, err
	}

	return wallet.VerifySignature(pub, block.dataToSign(), block.Signature), nil
}

func NewGenesisBlock() *Block {
	block := &Block{
		Index:        0,
		Timestamp:    0,
		Transactions: []Transaction{},
		PrevHash:     "",
		Validator:    "genesis",
	}
	block.Hash = block.CalculateHash()
	return block
}
