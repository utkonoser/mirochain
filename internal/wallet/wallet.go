package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcec/v2"
)

// Wallet представляет кошелек пользователя
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey `json:"private_key"`
	PublicKey  *ecdsa.PublicKey  `json:"public_key"`
	Address    string            `json:"address"`
}

// NewWallet создает новый кошелек
func NewWallet() (*Wallet, error) {
	// Генерируем приватный ключ
	privateKey, err := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Получаем публичный ключ
	publicKey := &privateKey.PublicKey

	// Генерируем адрес
	address, err := generateAddress(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate address: %w", err)
	}

	wallet := &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}

	slog.Info("New wallet created", "address", address)
	return wallet, nil
}

// generateAddress генерирует адрес из публичного ключа
func generateAddress(publicKey *ecdsa.PublicKey) (string, error) {
	// Конвертируем публичный ключ в байты
	publicKeyBytes := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)

	// Вычисляем хеш публичного ключа
	hash := sha256.Sum256(publicKeyBytes)

	// Создаем адрес (первые 20 байт хеша)
	address := hex.EncodeToString(hash[:20])

	return address, nil
}

// Sign подписывает данные приватным ключом
func (w *Wallet) Sign(data []byte) ([]byte, error) {
	// Вычисляем хеш данных
	hash := sha256.Sum256(data)

	// Подписываем хеш
	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// Кодируем подпись
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// VerifySignature проверяет подпись
func (w *Wallet) VerifySignature(data []byte, signature []byte) bool {
	if w.PublicKey == nil {
		return false
	}

	// Вычисляем хеш данных
	hash := sha256.Sum256(data)

	// Декодируем подпись
	if len(signature) != 64 {
		return false
	}

	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:])

	// Проверяем подпись
	return ecdsa.Verify(w.PublicKey, hash[:], r, s)
}

// GetPublicKeyBytes возвращает публичный ключ в виде байтов
func (w *Wallet) GetPublicKeyBytes() []byte {
	if w.PublicKey == nil {
		return []byte("public_key_not_available")
	}
	return elliptic.Marshal(w.PublicKey.Curve, w.PublicKey.X, w.PublicKey.Y)
}

// GetPrivateKeyBytes возвращает приватный ключ в виде байтов
func (w *Wallet) GetPrivateKeyBytes() []byte {
	if w.PrivateKey == nil {
		return []byte("private_key_not_available")
	}
	return w.PrivateKey.D.Bytes()
}

// GetAddress возвращает адрес кошелька
func (w *Wallet) GetAddress() string {
	return w.Address
}

// WalletManager управляет множеством кошельков
type WalletManager struct {
	Wallets map[string]*Wallet `json:"wallets"` // Ключ: адрес
	DataDir string             `json:"-"`       // Директория для хранения данных
}

// NewWalletManager создает новый менеджер кошельков
func NewWalletManager() *WalletManager {
	return &WalletManager{
		Wallets: make(map[string]*Wallet),
		DataDir: "./data/wallets",
	}
}

// NewWalletManagerWithDataDir создает менеджер кошельков с указанной директорией
func NewWalletManagerWithDataDir(dataDir string) *WalletManager {
	return &WalletManager{
		Wallets: make(map[string]*Wallet),
		DataDir: dataDir,
	}
}

// CreateWallet создает новый кошелек
func (wm *WalletManager) CreateWallet() (*Wallet, error) {
	wallet, err := NewWallet()
	if err != nil {
		return nil, err
	}

	wm.Wallets[wallet.Address] = wallet
	slog.Info("Wallet added to manager", "address", wallet.Address)
	return wallet, nil
}

// GetWallet возвращает кошелек по адресу
func (wm *WalletManager) GetWallet(address string) (*Wallet, bool) {
	wallet, exists := wm.Wallets[address]
	return wallet, exists
}

// GetWallets возвращает все кошельки
func (wm *WalletManager) GetWallets() map[string]*Wallet {
	return wm.Wallets
}

// RemoveWallet удаляет кошелек
func (wm *WalletManager) RemoveWallet(address string) {
	delete(wm.Wallets, address)
	slog.Info("Wallet removed from manager", "address", address)
}

// GetWalletCount возвращает количество кошельков
func (wm *WalletManager) GetWalletCount() int {
	return len(wm.Wallets)
}

// ListAddresses возвращает список всех адресов
func (wm *WalletManager) ListAddresses() []string {
	var addresses []string
	for address := range wm.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

// SaveWallets сохраняет кошельки в файл
func (wm *WalletManager) SaveWallets() error {
	// Создаем директорию если не существует
	if err := os.MkdirAll(wm.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Создаем структуру для сериализации (без приватных ключей в JSON)
	type WalletData struct {
		Address   string `json:"address"`
		PublicKey string `json:"public_key"`
	}

	walletsData := make(map[string]WalletData)
	for address, wallet := range wm.Wallets {
		walletsData[address] = WalletData{
			Address:   wallet.Address,
			PublicKey: hex.EncodeToString(wallet.GetPublicKeyBytes()),
		}
	}

	// Сохраняем в JSON файл
	filePath := filepath.Join(wm.DataDir, "wallets.json")
	data, err := json.MarshalIndent(walletsData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallets: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write wallets file: %w", err)
	}

	slog.Info("Wallets saved", "file", filePath, "count", len(wm.Wallets))
	return nil
}

// LoadWallets загружает кошельки из файла
func (wm *WalletManager) LoadWallets() error {
	filePath := filepath.Join(wm.DataDir, "wallets.json")

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		slog.Info("No wallets file found, starting with empty wallet manager")
		return nil
	}

	// Читаем файл
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read wallets file: %w", err)
	}

	// Парсим JSON
	var walletsData map[string]struct {
		Address   string `json:"address"`
		PublicKey string `json:"public_key"`
	}

	if err := json.Unmarshal(data, &walletsData); err != nil {
		return fmt.Errorf("failed to unmarshal wallets: %w", err)
	}

	// Загружаем кошельки с полными данными
	wm.Wallets = make(map[string]*Wallet)
	for address, walletData := range walletsData {
		// Пытаемся загрузить полный кошелек из отдельного файла
		wallet, err := wm.loadWalletFromFile(address)
		if err != nil {
			slog.Warn("Failed to load full wallet, creating read-only wallet", "address", address, "error", err)
			// Создаем кошелек только с публичным ключом
			wallet = &Wallet{
				Address:   walletData.Address,
				PublicKey: nil, // Не можем восстановить без приватного ключа
			}
		}

		wm.Wallets[address] = wallet
	}

	slog.Info("Wallets loaded", "file", filePath, "count", len(wm.Wallets))
	return nil
}

// loadWalletFromFile загружает полный кошелек из отдельного файла
func (wm *WalletManager) loadWalletFromFile(address string) (*Wallet, error) {
	fileName := fmt.Sprintf("wallet_%s.json", address)
	filePath := filepath.Join(wm.DataDir, fileName)

	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("wallet file not found")
	}

	// Читаем файл
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	// Парсим JSON
	var walletData struct {
		Address    string `json:"address"`
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
	}

	if err := json.Unmarshal(data, &walletData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet: %w", err)
	}

	// Декодируем ключи
	publicKeyBytes, err := hex.DecodeString(walletData.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	privateKeyBytes, err := hex.DecodeString(walletData.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Восстанавливаем приватный ключ
	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: btcec.S256(),
		},
		D: new(big.Int).SetBytes(privateKeyBytes),
	}

	// Восстанавливаем публичный ключ
	publicKey := &ecdsa.PublicKey{
		Curve: btcec.S256(),
	}
	publicKey.X, publicKey.Y = elliptic.Unmarshal(btcec.S256(), publicKeyBytes)
	if publicKey.X == nil || publicKey.Y == nil {
		return nil, fmt.Errorf("failed to unmarshal public key")
	}

	privateKey.PublicKey = *publicKey

	// Создаем кошелек
	wallet := &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    walletData.Address,
	}

	return wallet, nil
}

// SaveWallet сохраняет отдельный кошелек (с приватным ключом)
func (wm *WalletManager) SaveWallet(wallet *Wallet) error {
	// Создаем директорию если не существует
	if err := os.MkdirAll(wm.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Создаем структуру для сериализации
	walletData := struct {
		Address    string `json:"address"`
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
	}{
		Address:    wallet.Address,
		PublicKey:  hex.EncodeToString(wallet.GetPublicKeyBytes()),
		PrivateKey: hex.EncodeToString(wallet.GetPrivateKeyBytes()),
	}

	// Сохраняем в отдельный файл для каждого кошелька
	fileName := fmt.Sprintf("wallet_%s.json", wallet.Address)
	filePath := filepath.Join(wm.DataDir, fileName)

	data, err := json.MarshalIndent(walletData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal wallet: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil { // 0600 - только владелец может читать/писать
		return fmt.Errorf("failed to write wallet file: %w", err)
	}

	slog.Info("Wallet saved", "file", filePath, "address", wallet.Address)
	return nil
}

// GetBalance возвращает баланс кошелька (требует блокчейн)
func (w *Wallet) GetBalance(blockchain interface{}) (int64, error) {
	// Это заглушка - в реальной реализации нужно интегрироваться с блокчейном
	// Для демонстрации возвращаем случайный баланс
	return 1000000, nil
}

// CreateTransaction создает транзакцию (требует блокчейн)
func (w *Wallet) CreateTransaction(to string, amount int64, blockchain interface{}) (interface{}, error) {
	// Это заглушка - в реальной реализации нужно интегрироваться с блокчейном
	return nil, fmt.Errorf("transaction creation not implemented yet")
}

// SignMessage подписывает сообщение приватным ключом
func (w *Wallet) SignMessage(message []byte) ([]byte, error) {
	if w.PrivateKey == nil {
		return nil, fmt.Errorf("private key not available for signing")
	}

	hash := sha256.Sum256(message)
	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash[:])
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	// Конвертируем подпись в байты
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// String возвращает строковое представление кошелька
func (w *Wallet) String() string {
	return fmt.Sprintf("Wallet{Address: %s, PublicKey: %x}", w.Address, w.GetPublicKeyBytes())
}
