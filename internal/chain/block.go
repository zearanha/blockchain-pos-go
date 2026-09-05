package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

type Block struct {
	Index        int
	Timestamp    int64
	Transactions []Transaction
	PrevHash     string
	Hash         string
	Validator    string
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
