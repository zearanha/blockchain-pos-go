package chain

import "testing"

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
