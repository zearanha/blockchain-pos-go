package chain

import (
	"encoding/json"
	"testing"

	"github.com/zearanha/blockchain-pos-go/internal/wallet"
)

func TestSignedTransactionSurvivesJSONRoundTrip(t *testing.T) {
	w, err := wallet.NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	tx := Transaction{From: w.Address(), To: w.Address(), Amount: 1}
	if err := SignTransaction(&tx, w); err != nil {
		t.Fatal(err)
	}

	data, err := json.Marshal(tx)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Transaction
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	valid, err := VerifyTransaction(&decoded)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Fatal("expected decoded transaction signature to be valid")
	}
}
