package p2p

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/zearanha/blockchain-pos-go/internal/chain"
)

type ChainHandler interface {
	ReceiveTransaction(tx chain.Transaction) error
	ReceiveBlock(block chain.Block) error
	GetChain() []*chain.Block
}

type Server struct {
	handler ChainHandler
}

func NewServer(handler ChainHandler) *Server {
	return &Server{handler: handler}
}

func (s *Server) Start(address string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/transaction", s.handleTransaction)
	mux.HandleFunc("/block", s.handleBlock)
	mux.HandleFunc("/chain", s.handleGetChain)

	log.Printf("Nó P2P escutando em %s", address)
	go func() {
		if err := http.ListenAndServe(address, mux); err != nil {
			log.Fatal("erro no servidor P2P:", err)
		}
	}()
}

func (s *Server) handleTransaction(w http.ResponseWriter, r *http.Request) {
	var msg TransactionMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}

	if err := s.handler.ReceiveTransaction(msg.Transaction); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleBlock(w http.ResponseWriter, r *http.Request) {
	var msg BlockMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "corpo inválido", http.StatusBadRequest)
		return
	}

	if err := s.handler.ReceiveBlock(msg.Block); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGetChain(w http.ResponseWriter, r *http.Request) {
	chainBlocks := s.handler.GetChain()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chainBlocks)
}
