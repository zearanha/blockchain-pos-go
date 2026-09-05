package chain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
)

type SlashingEvidence struct {
	Validator string `json:"validator"`
	BlockA    Block  `json:"block_a"`
	BlockB    Block  `json:"block_b"`
}

func NewSlashingEvidence(blockA, blockB Block) SlashingEvidence {
	return SlashingEvidence{
		Validator: blockA.Validator,
		BlockA:    blockA,
		BlockB:    blockB,
	}
}

func (e SlashingEvidence) ID() string {
	hashes := []string{e.BlockA.Hash, e.BlockB.Hash}
	sort.Strings(hashes)

	data := fmt.Sprintf("%s|%d|%s|%s|%s", e.Validator, e.BlockA.Index, e.BlockA.PrevHash, hashes[0], hashes[1])
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func (e SlashingEvidence) Validate() error {
	if e.Validator == "" {
		return errors.New("evidência sem validador")
	}
	if e.BlockA.Index == 0 || e.BlockB.Index == 0 {
		return errors.New("bloco genesis não pode gerar slashing")
	}
	if e.BlockA.Validator != e.Validator || e.BlockB.Validator != e.Validator {
		return errors.New("blocos não pertencem ao mesmo validador acusado")
	}
	if e.BlockA.Index != e.BlockB.Index {
		return errors.New("blocos têm índices diferentes")
	}
	if e.BlockA.PrevHash != e.BlockB.PrevHash {
		return errors.New("blocos não competem pelo mesmo bloco anterior")
	}
	if e.BlockA.Hash == e.BlockB.Hash {
		return errors.New("blocos não são conflitantes")
	}

	validA, err := VerifyBlock(&e.BlockA)
	if err != nil {
		return fmt.Errorf("primeiro bloco inválido: %w", err)
	}
	if !validA {
		return errors.New("assinatura do primeiro bloco inválida")
	}

	validB, err := VerifyBlock(&e.BlockB)
	if err != nil {
		return fmt.Errorf("segundo bloco inválido: %w", err)
	}
	if !validB {
		return errors.New("assinatura do segundo bloco inválida")
	}

	return nil
}

func (bc *Blockchain) ApplySlashingEvidence(evidence SlashingEvidence) error {
	if err := evidence.Validate(); err != nil {
		return err
	}

	bc.Validators.Slash(evidence.Validator)
	return nil
}
