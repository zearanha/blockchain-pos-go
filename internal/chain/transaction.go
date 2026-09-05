package chain

import (
	"encoding/json"
	"errors"

	"github.com/zearanha/blockchain-pos-go/internal/wallet"
)

type Transaction struct {
	From      string
	To        string
	Amount    float64
	Signature []byte `json:"signature"`
}

func (t *Transaction) dataToSign() []byte {
	record := struct {
		From   string
		To     string
		Amount float64
	}{t.From, t.To, t.Amount}

	data, _ := json.Marshal(record)
	return data
}

func SignTransaction(tx *Transaction, w *wallet.Wallet) error {
	if tx.From != w.Address() {
		return errors.New("a carteira não corresponde ao remetente da transação")
	}

	signature, err := w.Sign(tx.dataToSign())
	if err != nil {
		return err
	}

	tx.Signature = signature
	return nil
}

func VerifyTransaction(tx *Transaction) (bool, error) {
	if len(tx.Signature) == 0 {
		return false, errors.New("transação sem assinatura")
	}

	pub, err := wallet.AddressToPublicKey(tx.From)
	if err != nil {
		return false, err
	}

	return wallet.VerifySignature(pub, tx.dataToSign(), tx.Signature), nil
}
