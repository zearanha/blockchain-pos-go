package main

import (
	"fmt"
	"log"

	"github.com/zearanha/blockchain-pos-go/internal/chain"
	"github.com/zearanha/blockchain-pos-go/internal/wallet"
)

func main() {
	aliceWallet, err := wallet.NewWallet()
	if err != nil {
		log.Fatal(err)
	}
	bobWallet, err := wallet.NewWallet()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Endereço da Alice:", aliceWallet.Address())
	fmt.Println("Endereço do Bob:  ", bobWallet.Address())

	tx := chain.Transaction{
		From:   aliceWallet.Address(),
		To:     bobWallet.Address(),
		Amount: 10,
	}

	if err := chain.SignTransaction(&tx, aliceWallet); err != nil {
		log.Fatal("erro ao assinar:", err)
	}

	valid, err := chain.VerifyTransaction(&tx)
	fmt.Printf("\nAssinatura válida? %v (err: %v)\n", valid, err)

	bc := chain.NewBlockchain()
	_, err = bc.AddBlock([]chain.Transaction{tx}, "validador-1")
	if err != nil {
		log.Fatal("erro ao adicionar bloco:", err)
	}
	fmt.Println("\nBloco com transação assinada adicionado com sucesso!")

	// simula uma transação forjada: Bob tentando assinar como se fosse a Alice
	fakeTx := chain.Transaction{
		From:   aliceWallet.Address(), // finge ser a Alice
		To:     bobWallet.Address(),
		Amount: 1000,
	}
	err = chain.SignTransaction(&fakeTx, bobWallet) // mas assina com a chave do Bob
	fmt.Printf("\nTentativa de forjar transação: %v\n", err)
}