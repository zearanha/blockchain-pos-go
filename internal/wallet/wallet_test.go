package wallet

import "testing"

func TestSignReturnsFixedSizeSignature(t *testing.T) {
	w, err := NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	signature, err := w.Sign([]byte("payload"))
	if err != nil {
		t.Fatal(err)
	}

	if len(signature) != signaturePartSize*2 {
		t.Fatalf("expected signature length %d, got %d", signaturePartSize*2, len(signature))
	}
	if !VerifySignature(w.PublicKey, []byte("payload"), signature) {
		t.Fatal("expected signature to verify")
	}
}

func TestVerifySignatureRejectsInvalidSignatureLength(t *testing.T) {
	w, err := NewWallet()
	if err != nil {
		t.Fatal(err)
	}

	if VerifySignature(w.PublicKey, []byte("payload"), []byte("short")) {
		t.Fatal("expected invalid signature length to be rejected")
	}
}
