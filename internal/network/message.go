package network

import (
	"encoding/gob"
	"fmt"
	"time"

	"mirochain/internal/blockchain"
)

// MessageType определяет тип сообщения
type MessageType string

const (
	// Handshake сообщения
	MessageTypeHandshake    MessageType = "handshake"
	MessageTypeHandshakeAck MessageType = "handshake_ack"

	// Блоки
	MessageTypeNewBlock      MessageType = "new_block"
	MessageTypeBlockRequest  MessageType = "block_request"
	MessageTypeBlockResponse MessageType = "block_response"

	// Транзакции
	MessageTypeNewTransaction MessageType = "new_transaction"
	MessageTypeTxRequest      MessageType = "tx_request"
	MessageTypeTxResponse     MessageType = "tx_response"

	// Синхронизация
	MessageTypeSyncRequest  MessageType = "sync_request"
	MessageTypeSyncResponse MessageType = "sync_response"

	// Peer discovery
	MessageTypePeerList    MessageType = "peer_list"
	MessageTypePeerRequest MessageType = "peer_request"

	// Общие
	MessageTypePing  MessageType = "ping"
	MessageTypePong  MessageType = "pong"
	MessageTypeError MessageType = "error"
)

// Message представляет сообщение в сети
type Message struct {
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	From      string      `json:"from"`
	To        string      `json:"to"`
	ID        string      `json:"id"`
}

// NewMessage создает новое сообщение
func NewMessage(msgType MessageType, data interface{}, from, to string) *Message {
	return &Message{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now().Unix(),
		From:      from,
		To:        to,
		ID:        generateMessageID(),
	}
}

// generateMessageID генерирует уникальный ID сообщения
func generateMessageID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// HandshakeData содержит данные для handshake
type HandshakeData struct {
	Version     string `json:"version"`
	NodeID      string `json:"node_id"`
	BestHeight  int64  `json:"best_height"`
	GenesisHash []byte `json:"genesis_hash"`
}

// BlockData содержит данные блока
type BlockData struct {
	Block *blockchain.Block `json:"block"`
}

// TransactionData содержит данные транзакции
type TransactionData struct {
	Transaction *blockchain.Transaction `json:"transaction"`
}

// SyncData содержит данные для синхронизации
type SyncData struct {
	StartHeight int64    `json:"start_height"`
	EndHeight   int64    `json:"end_height"`
	BlockHashes [][]byte `json:"block_hashes"`
}

// PeerListData содержит список peer'ов
type PeerListData struct {
	Peers []string `json:"peers"`
}

// ErrorData содержит данные об ошибке
type ErrorData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// RegisterTypes регистрирует типы для gob
func RegisterTypes() {
	gob.Register(&Message{})
	gob.Register(&HandshakeData{})
	gob.Register(&BlockData{})
	gob.Register(&TransactionData{})
	gob.Register(&SyncData{})
	gob.Register(&PeerListData{})
	gob.Register(&ErrorData{})
	gob.Register(&blockchain.Block{})
	gob.Register(&blockchain.Transaction{})
	gob.Register(&blockchain.TransactionInput{})
	gob.Register(&blockchain.TransactionOutput{})
}

// String возвращает строковое представление сообщения
func (m *Message) String() string {
	return fmt.Sprintf("Message{Type: %s, From: %s, To: %s, ID: %s, Timestamp: %d}",
		m.Type, m.From, m.To, m.ID, m.Timestamp)
}
