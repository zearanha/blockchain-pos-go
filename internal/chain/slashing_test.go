package chain

import (
	"testing"

	"github.com/zearanha/blockchain-pos-go/internal/wallet"
)

func TestSlashingEvidenceValidation(t *testing.T) {
	w, err := wallet.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	prevHash := NewGenesisBlock().Hash
	blockA := signedBlock(t, w, 1, prevHash, "a")
	blockB := signedBlock(t, w, 1, prevHash, "b")

	evidence := NewSlashingEvidence(blockA, blockB)
	if err := evidence.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestSlashingEvidenceRejectsBlocksFromDifferentValidators(t *testing.T) {
	validatorA, err := wallet.NewWallet()
	if err != nil {
		t.Fatal(err)
	}
	validatorB, err := wallet.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	prevHash := NewGenesisBlock().Hash
	blockA := signedBlock(t, validatorA, 1, prevHash, "a")
	blockB := signedBlock(t, validatorB, 1, prevHash, "b")

	evidence := NewSlashingEvidence(blockA, blockB)
	if err := evidence.Validate(); err == nil {
		t.Fatal("expected evidence with different validators to be rejected")
	}
}

func TestApplySlashingEvidenceZerosValidatorStake(t *testing.T) {
	slashedValidator, err := wallet.NewWallet()
	if err != nil {
		t.Fatal(err)
	}
	remainingValidator, err := wallet.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	bc := NewBlockchain()
	bc.Validators.AddStake(slashedValidator.Address(), 10)
	bc.Validators.AddStake(remainingValidator.Address(), 5)

	prevHash := bc.LastBlock().Hash
	blockA := signedBlock(t, slashedValidator, 1, prevHash, "a")
	blockB := signedBlock(t, slashedValidator, 1, prevHash, "b")

	if err := bc.ApplySlashingEvidence(NewSlashingEvidence(blockA, blockB)); err != nil {
		t.Fatal(err)
	}

	stake, exists := bc.Validators.StakeOf(slashedValidator.Address())
	if !exists {
		t.Fatal("expected slashed validator to remain registered")
	}
	if stake != 0 {
		t.Fatalf("expected slashed validator stake to be 0, got %.2f", stake)
	}

	selected, err := bc.Validators.SelectValidator("seed")
	if err != nil {
		t.Fatal(err)
	}
	if selected != remainingValidator.Address() {
		t.Fatalf("expected remaining validator to be selected, got %s", selected)
	}
}

func signedBlock(t *testing.T, validator *wallet.Wallet, index int, prevHash string, marker string) Block {
	t.Helper()

	block := NewBlock(index, []Transaction{{From: marker, To: marker, Amount: 1}}, prevHash, validator.Address())
	if err := SignBlock(block, validator); err != nil {
		t.Fatal(err)
	}

	return *block
}
