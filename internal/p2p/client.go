package p2p

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/zearanha/blockchain-pos-go/internal/chain"
)

func BroadcastTransaction(peers []string, tx chain.Transaction) {
	msg := TransactionMessage{Transaction: tx}
	broadcast(peers, "/transaction", msg)
}

func BroadcastBlock(peers []string, block chain.Block) {
	msg := BlockMessage{Block: block}
	broadcast(peers, "/block", msg)
}

func broadcast(peers []string, path string, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	client := http.Client{Timeout: 2 * time.Second}

	for _, peer := range peers {
		go func(peerAddr string) {
			url := fmt.Sprintf("http://%s%s", peerAddr, path)
			client.Post(url, "application/json", bytes.NewReader(data))
		}(peer)
	}
}
