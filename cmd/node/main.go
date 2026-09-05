package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/zearanha/blockchain-pos-go/internal/chain"
	"github.com/zearanha/blockchain-pos-go/internal/p2p"
	"github.com/zearanha/blockchain-pos-go/internal/wallet"
)

type NodeApp struct {
	mu      sync.Mutex
	bc      *chain.Blockchain
	mempool []chain.Transaction // transações pendentes, ainda não incluídas em um bloco
	peers   []string
	wallet  *wallet.Wallet
}

func (app *NodeApp) ReceiveTransaction(tx chain.Transaction) error {
	valid, err := chain.VerifyTransaction(&tx)
	if err != nil || !valid {
		return fmt.Errorf("transação inválida")
	}

	app.mu.Lock()
	defer app.mu.Unlock()
	app.mempool = append(app.mempool, tx)
	log.Printf("Transação recebida via P2P: %s -> %s (%.2f)", tx.From, tx.To, tx.Amount)
	return nil
}

func (app *NodeApp) ReceiveBlock(block chain.Block) error {
	app.mu.Lock()
	defer app.mu.Unlock()

	last := app.bc.LastBlock()

	// só aceita se encadear corretamente com o que já temos
	if block.PrevHash != last.Hash || block.Index != last.Index+1 {
		return fmt.Errorf("bloco não encadeia com a chain local (possível divergência)")
	}
	if !block.IsHashValid() {
		return fmt.Errorf("hash do bloco inválido")
	}

	app.bc.Blocks = append(app.bc.Blocks, &block)
	log.Printf("Bloco %d recebido e adicionado via P2P (validador: %s)", block.Index, block.Validator)
	return nil
}

func (app *NodeApp) GetChain() []*chain.Block {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.bc.Blocks
}

func (app *NodeApp) produceBlockLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		app.mu.Lock()
		if len(app.mempool) == 0 {
			app.mu.Unlock()
			continue
		}

		txs := app.mempool
		app.mempool = nil

		block, err := app.bc.AddBlock(txs)
		app.mu.Unlock()

		if err != nil {
			log.Println("erro ao produzir bloco:", err)
			continue
		}

		log.Printf("Bloco %d produzido localmente (validador: %s)", block.Index, block.Validator)
		p2p.BroadcastBlock(app.peers, *block)
	}
}

func main() {
	peersEnv := os.Getenv("PEERS")
	var peers []string
	if peersEnv != "" {
		peers = strings.Split(peersEnv, ",")
	}

	address := os.Getenv("ADDRESS")
	if address == "" {
		address = ":8080"
	}

	w, err := wallet.NewWallet()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Endereço deste nó:", w.Address())

	bc := chain.NewBlockchain()
	bc.Validators.AddStake(w.Address(), 10)

	app := &NodeApp{
		bc:     bc,
		peers:  peers,
		wallet: w,
	}

	server := p2p.NewServer(app)
	server.Start(address)

	go app.produceBlockLoop()

	go func() {
		time.Sleep(2 * time.Second)
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			tx := chain.Transaction{From: w.Address(), To: w.Address(), Amount: 1}
			if err := chain.SignTransaction(&tx, w); err != nil {
				continue
			}
			app.ReceiveTransaction(tx)
			p2p.BroadcastTransaction(app.peers, tx)
		}
	}()

	select {}
}
