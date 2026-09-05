package chain

import (
	"testing"

	"github.com/zearanha/blockchain-pos-go/internal/wallet"
)

func TestNewBlockSetsAValidHash(t *testing.T) {
	block := NewBlock(1, nil, "previous-hash", "validator")

	if block.Hash == "" {
		t.Fatal("expected block hash to be set")
	}
	if !block.IsHashValid() {
		t.Fatal("expected block hash to be valid")
	}
}

func TestGenesisBlockIsDeterministic(t *testing.T) {
	first := NewGenesisBlock()
	second := NewGenesisBlock()

	if first.Hash == "" {
		t.Fatal("expected genesis hash to be set")
	}
	if first.Hash != second.Hash {
		t.Fatalf("expected deterministic genesis hash, got %q and %q", first.Hash, second.Hash)
	}
}

func TestSignBlockAddsVerifiableSignature(t *testing.T) {
	w, err := wallet.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	block := NewBlock(1, nil, NewGenesisBlock().Hash, w.Address())
	if err := SignBlock(block, w); err != nil {
		t.Fatal(err)
	}

	valid, err := VerifyBlock(block)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Fatal("expected block signature to be valid")
	}
}

func TestVerifyBlockRejectsUnsignedBlock(t *testing.T) {
	w, err := wallet.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	block := NewBlock(1, nil, NewGenesisBlock().Hash, w.Address())
	valid, err := VerifyBlock(block)
	if err == nil {
		t.Fatal("expected unsigned block to be rejected")
	}
	if valid {
		t.Fatal("expected unsigned block to be invalid")
	}
}
