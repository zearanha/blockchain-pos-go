package chain

import "errors"

type Blockchain struct {
	Blocks []*Block
}

func NewBlockchain() *Blockchain {
	return &Blockchain {
		Blocks: []*Block{NewGenesisBlock()},
	}
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.Blocks[len(bc.Blocks)-1]
}

func (bc *Blockchain) AddBlock(transactions []Transaction, validator string) *Block {
	last := bc.LastBlock()
	newBlock := NewBlock(last.Index+1, transactions, last.Hash, validator)
	bc.Blocks = append(bc.Blocks, newBlock)
	return newBlock
}

func (bc *Blockchain) IsValid() error {
	for i := 1; i < len(bc.Blocks); i++ {
		current := bc.Blocks[i]
		previous := bc.Blocks[i-1]

		if !current.IsHashValid() {
			return errors.New("hash invalido no bloco " + itoa(current.Index))
		}

		if current.PrevHash != previous.Hash {
			return errors.New("encadeamento quebrado no bloco " + itoa(current.Index))
		}

		if current.Index != previous.Index+1 {
			return errors.New("indice fora de ordem no bloco " + itoa(current.Index))
		}
	}
	return nil
}


func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	neg := i < 0
	if neg {
		i = -i
	}

	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}