package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"mirochain/internal/wallet"
)

func main() {
	// Параметры командной строки
	var (
		create  = flag.Bool("create", false, "Create a new wallet")
		list    = flag.Bool("list", false, "List all wallets")
		info    = flag.String("info", "", "Show wallet info for address")
		balance = flag.String("balance", "", "Show wallet balance for address")
		sign    = flag.String("sign", "", "Sign a message with wallet address")
		verify  = flag.String("verify", "", "Verify a signature with wallet address")
		message = flag.String("message", "", "Message to sign/verify")
	)
	flag.Parse()

	// Настраиваем логгер
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Создаем менеджер кошельков
	walletManager := wallet.NewWalletManager()

	// Загружаем существующие кошельки
	if err := walletManager.LoadWallets(); err != nil {
		slog.Warn("Failed to load wallets", "error", err)
	}

	if *create {
		// Создаем новый кошелек
		newWallet, err := walletManager.CreateWallet()
		if err != nil {
			slog.Error("Failed to create wallet", "error", err)
			os.Exit(1)
		}

		// Сохраняем кошелек
		if err := walletManager.SaveWallet(newWallet); err != nil {
			slog.Error("Failed to save wallet", "error", err)
		}

		// Сохраняем общий список кошельков
		if err := walletManager.SaveWallets(); err != nil {
			slog.Error("Failed to save wallets list", "error", err)
		}

		fmt.Printf("New wallet created:\n")
		fmt.Printf("Address: %s\n", newWallet.GetAddress())
		fmt.Printf("Public Key: %x\n", newWallet.GetPublicKeyBytes())
		fmt.Printf("Private Key: %x\n", newWallet.GetPrivateKeyBytes())

	} else if *list {
		// Список всех кошельков
		wallets := walletManager.GetWallets()
		if len(wallets) == 0 {
			fmt.Println("No wallets found")
			return
		}

		fmt.Printf("Found %d wallets:\n", len(wallets))
		for address, w := range wallets {
			fmt.Printf("Address: %s\n", address)
			fmt.Printf("Public Key: %x\n", w.GetPublicKeyBytes())
			fmt.Println("---")
		}

	} else if *info != "" {
		// Информация о конкретном кошельке
		w, exists := walletManager.GetWallet(*info)
		if !exists {
			fmt.Printf("Wallet not found: %s\n", *info)
			os.Exit(1)
		}

		fmt.Printf("Wallet Info:\n")
		fmt.Printf("Address: %s\n", w.GetAddress())
		fmt.Printf("Public Key: %x\n", w.GetPublicKeyBytes())
		fmt.Printf("Private Key: %x\n", w.GetPrivateKeyBytes())

	} else if *balance != "" {
		// Показываем баланс кошелька
		w, exists := walletManager.GetWallet(*balance)
		if !exists {
			fmt.Printf("Wallet not found: %s\n", *balance)
			os.Exit(1)
		}

		balance, err := w.GetBalance(nil) // nil - заглушка для блокчейна
		if err != nil {
			fmt.Printf("Failed to get balance: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Wallet Balance:\n")
		fmt.Printf("Address: %s\n", w.GetAddress())
		fmt.Printf("Balance: %d\n", balance)

	} else if *sign != "" {
		// Подписываем сообщение
		if *message == "" {
			fmt.Println("Error: -message is required for signing")
			os.Exit(1)
		}

		w, exists := walletManager.GetWallet(*sign)
		if !exists {
			fmt.Printf("Wallet not found: %s\n", *sign)
			os.Exit(1)
		}

		signature, err := w.SignMessage([]byte(*message))
		if err != nil {
			fmt.Printf("Failed to sign message: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Message signed:\n")
		fmt.Printf("Address: %s\n", w.GetAddress())
		fmt.Printf("Message: %s\n", *message)
		fmt.Printf("Signature: %x\n", signature)

	} else if *verify != "" {
		// Проверяем подпись
		if *message == "" {
			fmt.Println("Error: -message is required for verification")
			os.Exit(1)
		}

		w, exists := walletManager.GetWallet(*verify)
		if !exists {
			fmt.Printf("Wallet not found: %s\n", *verify)
			os.Exit(1)
		}

		// Для демонстрации создаем подпись и проверяем её
		signature, err := w.SignMessage([]byte(*message))
		if err != nil {
			fmt.Printf("Failed to create signature for verification: %v\n", err)
			os.Exit(1)
		}

		isValid := w.VerifySignature([]byte(*message), signature)
		fmt.Printf("Signature verification:\n")
		fmt.Printf("Address: %s\n", w.GetAddress())
		fmt.Printf("Message: %s\n", *message)
		fmt.Printf("Signature: %x\n", signature)
		fmt.Printf("Valid: %t\n", isValid)

	} else {
		// Показываем справку
		fmt.Println("MiroChain Wallet CLI")
		fmt.Println("Usage:")
		fmt.Println("  -create                    Create a new wallet")
		fmt.Println("  -list                      List all wallets")
		fmt.Println("  -info <address>            Show wallet info")
		fmt.Println("  -balance <address>         Show wallet balance")
		fmt.Println("  -sign <address> -message <text>  Sign a message")
		fmt.Println("  -verify <address> -message <text>  Verify a signature")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/wallet/main.go -create")
		fmt.Println("  go run cmd/wallet/main.go -list")
		fmt.Println("  go run cmd/wallet/main.go -info <address>")
		fmt.Println("  go run cmd/wallet/main.go -balance <address>")
		fmt.Println("  go run cmd/wallet/main.go -sign <address> -message \"Hello World\"")
		fmt.Println("  go run cmd/wallet/main.go -verify <address> -message \"Hello World\"")
	}
}
