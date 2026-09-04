package main

import (
	"fmt"

	"github.com/zearanha/blockchain-pos-go/internal/chain"
)
func main() {
	genesis := chain.NewGenesisBlock()
	fmt.Printf("Bloco gênese criado:\n  Index: %d\n  Hash: %s\n  PrevHash: %q\n",
		genesis.Index, genesis.Hash, genesis.PrevHash)

	tx := chain.Transaction{From: "alice", To: "bob", Amount: 10}
	block1 := chain.NewBlock(1, []chain.Transaction{tx}, genesis.Hash, "validador-teste")

	fmt.Printf("\nBloco 1 criado:\n  Index: %d\n  Hash: %s\n  PrevHash: %s\n  Válido: %v\n",
		block1.Index, block1.Hash, block1.PrevHash, block1.IsHashValid())

	block1.Transactions[0].Amount = 999
	fmt.Printf("\nApós adulterar a transação:\n  Hash ainda bate? %v\n", block1.IsHashValid())
}