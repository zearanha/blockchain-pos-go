package main

import (
	"testing"

	"github.com/zearanha/blockchain-pos-go/internal/chain"
	"github.com/zearanha/blockchain-pos-go/internal/wallet"
)

func TestReceiveBlockAppliesSlashingOnDoubleSigning(t *testing.T) {
	validator, err := wallet.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	bc := chain.NewBlockchain()
	bc.Validators.AddStake(validator.Address(), 10)

	app := &NodeApp{
		bc:           bc,
		wallet:       validator,
		seenBlocks:   make(map[string]chain.Block),
		seenEvidence: make(map[string]bool),
	}

	prevHash := bc.LastBlock().Hash
	blockA := signedTestBlock(t, validator, 1, prevHash, "a")
	blockB := signedTestBlock(t, validator, 1, prevHash, "b")

	if err := app.ReceiveBlock(blockA); err != nil {
		t.Fatal(err)
	}
	if err := app.ReceiveBlock(blockB); err != nil {
		t.Fatal(err)
	}

	stake, exists := bc.Validators.StakeOf(validator.Address())
	if !exists {
		t.Fatal("expected validator to remain registered")
	}
	if stake != 0 {
		t.Fatalf("expected validator stake to be slashed to 0, got %.2f", stake)
	}
}

func signedTestBlock(t *testing.T, validator *wallet.Wallet, index int, prevHash string, marker string) chain.Block {
	t.Helper()

	block := chain.NewBlock(index, []chain.Transaction{{From: marker, To: marker, Amount: 1}}, prevHash, validator.Address())
	if err := chain.SignBlock(block, validator); err != nil {
		t.Fatal(err)
	}

	return *block
}
