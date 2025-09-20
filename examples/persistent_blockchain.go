//go:build persistent_blockchain
// +build persistent_blockchain

package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/persistent"
	"mirochain/internal/wallet"
)

func main() {
	// Настройка логирования
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	fmt.Println("=== Пример использования персистентного блокчейна ===")

	// Определяем директорию для данных
	dataDir := "./data/blockchain"

	// Создаем директорию, если она не существует
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Не удалось создать директорию данных: %v", err)
	}

	fmt.Printf("Директория данных: %s\n", dataDir)

	// Создаем кошелек для узла
	walletManager := wallet.NewWalletManager()
	nodeWallet, err := walletManager.CreateWallet()
	if err != nil {
		log.Fatalf("Не удалось создать кошелек: %v", err)
	}

	fmt.Printf("Адрес узла: %s\n", nodeWallet.GetAddress())

	// Создаем персистентный блокчейн
	pbc, err := persistent.NewPersistentBlockchain(dataDir, nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 2)
	if err != nil {
		log.Fatalf("Не удалось создать персистентный блокчейн: %v", err)
	}
	defer pbc.Close()

	// Получаем начальную статистику
	stats := pbc.GetStats()
	fmt.Printf("\n=== Начальная статистика ===\n")
	fmt.Printf("Высота: %d\n", stats["height"])
	fmt.Printf("Сложность: %d\n", stats["difficulty"])
	fmt.Printf("Количество UTXO: %d\n", stats["utxo_count"])

	// Получаем начальный баланс
	balance := pbc.GetBalance(nodeWallet.GetAddress())
	fmt.Printf("Начальный баланс: %d\n", balance)

	// Добавляем несколько блоков
	fmt.Printf("\n=== Добавление блоков ===\n")
	for i := 0; i < 3; i++ {
		// Создаем coinbase транзакцию
		coinbaseTx := blockchain.NewCoinbaseTransaction(nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 50)

		// Получаем предыдущий блок
		height, _ := pbc.GetHeight()
		previousBlock, err := pbc.GetBlockByHeight(height)
		if err != nil {
			log.Fatalf("Не удалось получить предыдущий блок: %v", err)
		}

		// Создаем новый блок
		newBlock := blockchain.NewBlock([]*blockchain.Transaction{coinbaseTx}, previousBlock.Hash, height+1, 2)

		// Добавляем блок в блокчейн
		err = pbc.AddBlock(newBlock)
		if err != nil {
			log.Fatalf("Не удалось добавить блок: %v", err)
		}

		fmt.Printf("Добавлен блок #%d с хешем %x\n", newBlock.Height, newBlock.Hash)
	}

	// Получаем финальную статистику
	stats = pbc.GetStats()
	fmt.Printf("\n=== Финальная статистика ===\n")
	fmt.Printf("Высота: %d\n", stats["height"])
	fmt.Printf("Сложность: %d\n", stats["difficulty"])
	fmt.Printf("Количество UTXO: %d\n", stats["utxo_count"])

	// Получаем финальный баланс
	balance = pbc.GetBalance(nodeWallet.GetAddress())
	fmt.Printf("Финальный баланс: %d\n", balance)

	// Показываем UTXO
	utxos := pbc.GetUTXOs(nodeWallet.GetAddress())
	fmt.Printf("\n=== UTXO ===\n")
	for i, utxo := range utxos {
		fmt.Printf("UTXO #%d: Значение=%d, TxID=%x, OutputIndex=%d\n",
			i+1, utxo.Value, utxo.TransactionID, utxo.OutputIndex)
	}

	// Демонстрируем персистентность
	fmt.Printf("\n=== Тест персистентности ===\n")
	fmt.Printf("Закрываем блокчейн...\n")
	pbc.Close()

	// Небольшая пауза для демонстрации
	time.Sleep(1 * time.Second)

	fmt.Printf("Открываем блокчейн заново...\n")

	// Открываем блокчейн заново
	pbc2, err := persistent.NewPersistentBlockchain(dataDir, nodeWallet.GetAddress(), nodeWallet.GetPublicKeyBytes(), 2)
	if err != nil {
		log.Fatalf("Не удалось открыть блокчейн заново: %v", err)
	}
	defer pbc2.Close()

	// Проверяем, что данные сохранились
	stats2 := pbc2.GetStats()
	balance2 := pbc2.GetBalance(nodeWallet.GetAddress())

	fmt.Printf("Высота после перезапуска: %d\n", stats2["height"])
	fmt.Printf("Баланс после перезапуска: %d\n", balance2)

	if stats2["height"] == stats["height"] && balance2 == balance {
		fmt.Printf("✅ Персистентность работает корректно!\n")
	} else {
		fmt.Printf("❌ Проблема с персистентностью!\n")
	}

	// Показываем информацию о файлах
	fmt.Printf("\n=== Информация о хранилище ===\n")
	err = filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fmt.Printf("Файл: %s, Размер: %d байт\n", path, info.Size())
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Ошибка при чтении директории: %v\n", err)
	}

	fmt.Printf("\n=== Пример завершен успешно ===\n")
}
