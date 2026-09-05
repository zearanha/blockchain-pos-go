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
	mu           sync.Mutex
	bc           *chain.Blockchain
	mempool      []chain.Transaction // transações pendentes, ainda não incluídas em um bloco
	peers        []string
	wallet       *wallet.Wallet
	seenBlocks   map[string]chain.Block
	seenEvidence map[string]bool
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
	valid, err := chain.VerifyBlock(&block)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("assinatura do bloco inválida")
	}

	var evidence *chain.SlashingEvidence
	var shouldBroadcastEvidence bool

	app.mu.Lock()

	evidence = app.recordBlockLocked(block)
	if evidence != nil {
		applied, err := app.applySlashingEvidenceLocked(*evidence)
		if err != nil {
			app.mu.Unlock()
			return err
		}
		shouldBroadcastEvidence = applied
	}

	last := app.bc.LastBlock()

	// só aceita se encadear corretamente com o que já temos
	if block.PrevHash != last.Hash || block.Index != last.Index+1 {
		app.mu.Unlock()
		if shouldBroadcastEvidence {
			p2p.BroadcastSlashingEvidence(app.peers, *evidence)
		}
		if evidence != nil {
			return nil
		}
		return fmt.Errorf("bloco não encadeia com a chain local (possível divergência)")
	}

	app.bc.Blocks = append(app.bc.Blocks, &block)
	app.mu.Unlock()

	log.Printf("Bloco %d recebido e adicionado via P2P (validador: %s)", block.Index, block.Validator)
	if shouldBroadcastEvidence {
		p2p.BroadcastSlashingEvidence(app.peers, *evidence)
	}
	return nil
}

func (app *NodeApp) ReceiveSlashingEvidence(evidence chain.SlashingEvidence) error {
	if err := evidence.Validate(); err != nil {
		return err
	}

	app.mu.Lock()
	applied, err := app.applySlashingEvidenceLocked(evidence)
	if err != nil {
		app.mu.Unlock()
		return err
	}

	app.recordBlockLocked(evidence.BlockA)
	app.recordBlockLocked(evidence.BlockB)
	app.mu.Unlock()

	if applied {
		log.Printf("Slashing recebido via P2P e aplicado contra validador %s", evidence.Validator)
		p2p.BroadcastSlashingEvidence(app.peers, evidence)
	}

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

		txs := append([]chain.Transaction(nil), app.mempool...)

		block, err := app.bc.AddBlock(txs, app.wallet)
		if err == nil {
			app.mempool = nil
			app.recordBlockLocked(*block)
		}
		app.mu.Unlock()

		if err != nil {
			log.Println("erro ao produzir bloco:", err)
			continue
		}

		log.Printf("Bloco %d produzido localmente (validador: %s)", block.Index, block.Validator)
		p2p.BroadcastBlock(app.peers, *block)
	}
}

func (app *NodeApp) recordBlockLocked(block chain.Block) *chain.SlashingEvidence {
	if block.Index == 0 {
		return nil
	}
	if app.seenBlocks == nil {
		app.seenBlocks = make(map[string]chain.Block)
	}

	key := blockConflictKey(block)
	previous, exists := app.seenBlocks[key]
	if !exists {
		app.seenBlocks[key] = block
		return nil
	}
	if previous.Hash == block.Hash {
		return nil
	}

	evidence := chain.NewSlashingEvidence(previous, block)
	if err := evidence.Validate(); err != nil {
		log.Printf("Evidência de slashing ignorada: %v", err)
		return nil
	}
	return &evidence
}

func (app *NodeApp) applySlashingEvidenceLocked(evidence chain.SlashingEvidence) (bool, error) {
	if app.seenEvidence == nil {
		app.seenEvidence = make(map[string]bool)
	}

	evidenceID := evidence.ID()
	if app.seenEvidence[evidenceID] {
		return false, nil
	}
	if err := app.bc.ApplySlashingEvidence(evidence); err != nil {
		return false, err
	}

	app.seenEvidence[evidenceID] = true
	log.Printf("Slashing aplicado: stake do validador %s foi zerado", evidence.Validator)
	return true, nil
}

func blockConflictKey(block chain.Block) string {
	return fmt.Sprintf("%d:%s:%s", block.Index, block.PrevHash, block.Validator)
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
		bc:           bc,
		peers:        peers,
		wallet:       w,
		seenBlocks:   make(map[string]chain.Block),
		seenEvidence: make(map[string]bool),
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
