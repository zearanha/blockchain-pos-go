package chain

import (
	"errors"
	"fmt"

	"github.com/zearanha/blockchain-pos-go/internal/wallet"
)

type Blockchain struct {
	Blocks     []*Block
	Validators *ValidatorSet
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks:     []*Block{NewGenesisBlock()},
		Validators: NewValidatorSet(),
	}
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.Blocks[len(bc.Blocks)-1]
}

func (bc *Blockchain) AddBlock(transactions []Transaction, validatorWallet *wallet.Wallet) (*Block, error) {
	for i, tx := range transactions {
		valid, err := VerifyTransaction(&tx)
		if err != nil {
			return nil, fmt.Errorf("transação %d inválida: %w", i, err)
		}
		if !valid {
			return nil, fmt.Errorf("transação %d com assinatura inválida", i)
		}
	}

	last := bc.LastBlock()

	validator, err := bc.Validators.SelectValidator(last.Hash)
	if err != nil {
		return nil, fmt.Errorf("erro ao selecionar validador: %w", err)
	}
	if validatorWallet == nil || validatorWallet.Address() != validator {
		return nil, fmt.Errorf("carteira local não é o validador selecionado: %s", validator)
	}

	newBlock := NewBlock(last.Index+1, transactions, last.Hash, validator)
	if err := SignBlock(newBlock, validatorWallet); err != nil {
		return nil, fmt.Errorf("erro ao assinar bloco: %w", err)
	}

	bc.Blocks = append(bc.Blocks, newBlock)
	return newBlock, nil
}

func (bc *Blockchain) IsValid() error {
	for i := 1; i < len(bc.Blocks); i++ {
		current := bc.Blocks[i]
		previous := bc.Blocks[i-1]

		if !current.IsHashValid() {
			return errors.New("hash invalido no bloco " + itoa(current.Index))
		}
		validSignature, err := VerifyBlock(current)
		if err != nil {
			return fmt.Errorf("assinatura inválida no bloco %s: %w", itoa(current.Index), err)
		}
		if !validSignature {
			return errors.New("assinatura inválida no bloco " + itoa(current.Index))
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
