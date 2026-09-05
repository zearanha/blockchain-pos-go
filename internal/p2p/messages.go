package p2p

import "github.com/zearanha/blockchain-pos-go/internal/chain"

type TransactionMessage struct {
	Transaction chain.Transaction `json:"transaction"`
}

type BlockMessage struct {
	Block chain.Block `json:"block"`
}
