package network

import (
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"
)

// Peer представляет подключенный узел
type Peer struct {
	ID       string    `json:"id"`
	Address  string    `json:"address"`
	Conn     net.Conn  `json:"-"`
	LastSeen time.Time `json:"last_seen"`
	IsActive bool      `json:"is_active"`
	mutex    sync.RWMutex
}

// NewPeer создает новый peer
func NewPeer(id, address string, conn net.Conn) *Peer {
	return &Peer{
		ID:       id,
		Address:  address,
		Conn:     conn,
		LastSeen: time.Now(),
		IsActive: true,
	}
}

// SendMessage отправляет сообщение peer'у
func (p *Peer) SendMessage(msg *Message) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.IsActive || p.Conn == nil {
		return fmt.Errorf("peer is not active")
	}

	// Устанавливаем таймаут для записи
	p.Conn.SetWriteDeadline(time.Now().Add(30 * time.Second))

	// Кодируем и отправляем сообщение
	encoder := gob.NewEncoder(p.Conn)
	err := encoder.Encode(msg)
	if err != nil {
		p.IsActive = false
		return fmt.Errorf("failed to send message: %w", err)
	}

	p.LastSeen = time.Now()
	return nil
}

// ReadMessage читает сообщение от peer'а
func (p *Peer) ReadMessage() (*Message, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.IsActive || p.Conn == nil {
		return nil, fmt.Errorf("peer is not active")
	}

	// Устанавливаем таймаут для чтения
	p.Conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Декодируем сообщение
	decoder := gob.NewDecoder(p.Conn)
	var msg Message
	err := decoder.Decode(&msg)
	if err != nil {
		p.IsActive = false
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	p.LastSeen = time.Now()
	return &msg, nil
}

// Close закрывает соединение с peer'ом
func (p *Peer) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.IsActive = false
	if p.Conn != nil {
		return p.Conn.Close()
	}
	return nil
}

// IsConnected проверяет, активно ли соединение
func (p *Peer) IsConnected() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.IsActive && p.Conn != nil
}

// UpdateLastSeen обновляет время последнего контакта
func (p *Peer) UpdateLastSeen() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.LastSeen = time.Now()
}

// String возвращает строковое представление peer'а
func (p *Peer) String() string {
	return fmt.Sprintf("Peer{ID: %s, Address: %s, Active: %t, LastSeen: %s}",
		p.ID, p.Address, p.IsActive, p.LastSeen.Format(time.RFC3339))
}
