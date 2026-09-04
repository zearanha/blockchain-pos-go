package main

import (
	"fmt"

	"github.com/zearanha/blockchain-pos-go/internal/chain"
)

func main() {
	bc := chain.NewBlockchain()
	fmt.Printf("Chain iniciada com bloco gênese (hash: %s)\n\n", bc.LastBlock().Hash)

	bc.AddBlock([]chain.Transaction{
		{From: "alice", To: "bob", Amount: 10},
	}, "validador-1")

	bc.AddBlock([]chain.Transaction{
		{From: "bob", To: "carol", Amount: 5},
	}, "validador-2")

	for _, b := range bc.Blocks {
		fmt.Printf("Bloco %d | Hash: %s | PrevHash: %s\n", b.Index, b.Hash, b.PrevHash)
	}

	if err := bc.IsValid(); err != nil {
		fmt.Println("\nChain inválida:", err)
	} else {
		fmt.Println("\nChain válida!")
	}

	// simula adulteração num bloco do meio da chain
	fmt.Println("\nAdulterando o bloco 1...")
	bc.Blocks[1].Transactions[0].Amount = 999999

	if err := bc.IsValid(); err != nil {
		fmt.Println("Chain inválida (como esperado):", err)
	} else {
		fmt.Println("Chain válida (isso seria um bug!)")
	}
}