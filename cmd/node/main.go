package main

import (
	"fmt"
	"log"

	"github.com/zearanha/blockchain-pos-go/internal/chain"
	"github.com/zearanha/blockchain-pos-go/internal/wallet"
)

func main() {
	aliceWallet, _ := wallet.NewWallet()
	bobWallet, _ := wallet.NewWallet()
	carolWallet, _ := wallet.NewWallet()

	bc := chain.NewBlockchain()

	// registra stakes: Carol tem bem mais stake que os outros
	bc.Validators.AddStake(aliceWallet.Address(), 10)
	bc.Validators.AddStake(bobWallet.Address(), 15)
	bc.Validators.AddStake(carolWallet.Address(), 75)

	fmt.Println("Stake total:", bc.Validators.TotalStake())

	// adiciona vários blocos e observa quem é escolhido validador em cada um
	validatorCounts := map[string]int{}

	for i := 0; i < 20; i++ {
		tx := chain.Transaction{
			From:   aliceWallet.Address(),
			To:     bobWallet.Address(),
			Amount: 1,
		}
		if err := chain.SignTransaction(&tx, aliceWallet); err != nil {
			log.Fatal(err)
		}

		block, err := bc.AddBlock([]chain.Transaction{tx})
		if err != nil {
			log.Fatal(err)
		}

		validatorCounts[block.Validator]++
	}

	fmt.Println("\nDistribuição de blocos validados em 20 rodadas:")
	fmt.Printf("Alice (stake 10): %d blocos\n", validatorCounts[aliceWallet.Address()])
	fmt.Printf("Bob   (stake 15): %d blocos\n", validatorCounts[bobWallet.Address()])
	fmt.Printf("Carol (stake 75): %d blocos\n", validatorCounts[carolWallet.Address()])

	if err := bc.IsValid(); err != nil {
		fmt.Println("\nChain inválida:", err)
	} else {
		fmt.Println("\nChain válida!")
	}
}