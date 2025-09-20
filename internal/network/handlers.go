package network

import (
	"fmt"
	"log/slog"

	"mirochain/internal/blockchain"
)

// handleHandshake обрабатывает handshake сообщение
func (s *Server) handleHandshake(msg *Message) {
	data, ok := msg.Data.(*HandshakeData)
	if !ok {
		slog.Error("Invalid handshake data")
		return
	}

	slog.Info("Received handshake", "from", msg.From, "version", data.Version, "height", data.BestHeight)

	// Отправляем подтверждение handshake
	responseData := &HandshakeData{
		Version:     "1.0.0",
		NodeID:      s.ID,
		BestHeight:  s.Blockchain.GetHeight(),
		GenesisHash: s.Blockchain.GetGenesisHash(),
	}

	response := NewMessage(MessageTypeHandshakeAck, responseData, s.ID, msg.From)
	s.broadcastToPeer(msg.From, response)
}

// handleHandshakeAck обрабатывает подтверждение handshake
func (s *Server) handleHandshakeAck(msg *Message) {
	data, ok := msg.Data.(*HandshakeData)
	if !ok {
		slog.Error("Invalid handshake ack data")
		return
	}

	slog.Info("Handshake acknowledged", "from", msg.From, "version", data.Version, "height", data.BestHeight)
}

// handleNewBlock обрабатывает новый блок
func (s *Server) handleNewBlock(msg *Message) {
	data, ok := msg.Data.(*BlockData)
	if !ok {
		slog.Error("Invalid block data")
		return
	}

	block := data.Block
	slog.Info("Received new block", "from", msg.From, "height", block.Height, "hash", fmt.Sprintf("%x", block.Hash))

	// Проверяем валидность блока
	if !block.IsValid(s.Blockchain.GetLastBlock()) {
		slog.Error("Invalid block received", "from", msg.From, "height", block.Height)
		return
	}

	// Добавляем блок в блокчейн
	err := s.Blockchain.AddBlock(block)
	if err != nil {
		slog.Error("Failed to add block", "error", err)
		return
	}

	slog.Info("Block added to blockchain", "height", block.Height, "hash", fmt.Sprintf("%x", block.Hash))

	// Распространяем блок другим peer'ам
	s.broadcastBlock(block, msg.From)
}

// handleNewTransaction обрабатывает новую транзакцию
func (s *Server) handleNewTransaction(msg *Message) {
	data, ok := msg.Data.(*TransactionData)
	if !ok {
		slog.Error("Invalid transaction data")
		return
	}

	tx := data.Transaction
	slog.Info("Received new transaction", "from", msg.From, "tx_id", fmt.Sprintf("%x", tx.ID))

	// Проверяем валидность транзакции
	if !tx.IsValid() {
		slog.Error("Invalid transaction received", "from", msg.From, "tx_id", fmt.Sprintf("%x", tx.ID))
		return
	}

	// TODO: Добавить транзакцию в mempool
	slog.Info("Transaction validated", "tx_id", fmt.Sprintf("%x", tx.ID))

	// Распространяем транзакцию другим peer'ам
	s.broadcastTransaction(tx, msg.From)
}

// handleSyncRequest обрабатывает запрос синхронизации
func (s *Server) handleSyncRequest(msg *Message) {
	data, ok := msg.Data.(*SyncData)
	if !ok {
		slog.Error("Invalid sync data")
		return
	}

	slog.Info("Received sync request", "from", msg.From, "start", data.StartHeight, "end", data.EndHeight)

	// Отправляем блоки для синхронизации
	s.sendBlocksForSync(msg.From, data.StartHeight, data.EndHeight)
}

// handlePing обрабатывает ping сообщение
func (s *Server) handlePing(msg *Message) {
	slog.Debug("Received ping", "from", msg.From)

	// Отправляем pong
	pong := NewMessage(MessageTypePong, nil, s.ID, msg.From)
	s.broadcastToPeer(msg.From, pong)
}

// broadcastBlock распространяет блок всем peer'ам кроме отправителя
func (s *Server) broadcastBlock(block *blockchain.Block, excludeFrom string) {
	data := &BlockData{Block: block}
	msg := NewMessage(MessageTypeNewBlock, data, s.ID, "")

	for _, peer := range s.GetPeers() {
		if peer.ID != excludeFrom {
			err := peer.SendMessage(msg)
			if err != nil {
				slog.Error("Failed to broadcast block to peer", "peer", peer.Address, "error", err)
			}
		}
	}
}

// broadcastTransaction распространяет транзакцию всем peer'ам кроме отправителя
func (s *Server) broadcastTransaction(tx *blockchain.Transaction, excludeFrom string) {
	data := &TransactionData{Transaction: tx}
	msg := NewMessage(MessageTypeNewTransaction, data, s.ID, "")

	for _, peer := range s.GetPeers() {
		if peer.ID != excludeFrom {
			err := peer.SendMessage(msg)
			if err != nil {
				slog.Error("Failed to broadcast transaction to peer", "peer", peer.Address, "error", err)
			}
		}
	}
}

// broadcastToPeer отправляет сообщение конкретному peer'у
func (s *Server) broadcastToPeer(peerID string, msg *Message) {
	s.mutex.RLock()
	peer, exists := s.Peers[peerID]
	s.mutex.RUnlock()

	if !exists {
		slog.Error("Peer not found", "peer_id", peerID)
		return
	}

	err := peer.SendMessage(msg)
	if err != nil {
		slog.Error("Failed to send message to peer", "peer", peer.Address, "error", err)
	}
}

// sendBlocksForSync отправляет блоки для синхронизации
func (s *Server) sendBlocksForSync(peerID string, startHeight, endHeight int64) {
	for height := startHeight; height <= endHeight; height++ {
		block := s.Blockchain.GetBlockByHeight(height)
		if block == nil {
			slog.Error("Block not found", "height", height)
			continue
		}

		data := &BlockData{Block: block}
		msg := NewMessage(MessageTypeBlockResponse, data, s.ID, peerID)
		s.broadcastToPeer(peerID, msg)
	}
}

// BroadcastNewBlock распространяет новый блок по сети
func (s *Server) BroadcastNewBlock(block *blockchain.Block) {
	s.broadcastBlock(block, "")
}

// BroadcastNewTransaction распространяет новую транзакцию по сети
func (s *Server) BroadcastNewTransaction(tx *blockchain.Transaction) {
	s.broadcastTransaction(tx, "")
}
