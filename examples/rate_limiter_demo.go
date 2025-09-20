//go:build rate_limiter_demo
// +build rate_limiter_demo

package main

import (
	"fmt"
	"log"
	"time"

	"mirochain/internal/blockchain"
	"mirochain/internal/network"
	"mirochain/internal/persistent"
)

func main() {
	fmt.Println("=== Rate Limiter Demo ===")

	// Создаем блокчейн
	bc, err := persistent.NewCachedPersistentBlockchain("data/rate_limiter_demo", "test_address", []byte("test_public_key"), 1)
	if err != nil {
		log.Fatalf("Failed to create blockchain: %v", err)
	}
	defer bc.Close()

	// Создаем P2P сервер
	server := network.NewServer("localhost", 8080, &blockchain.Blockchain{})

	// Запускаем сервер
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Ждем инициализации
	time.Sleep(2 * time.Second)

	// Тестируем API rate limiter
	fmt.Println("\nTesting API Rate Limiter (Token Bucket):")
	fmt.Println("Max requests: 100/minute, Burst: 20, Refill rate: 10/second")

	clientID := "test_client_001"
	allowedCount := 0
	blockedCount := 0

	// Быстро отправляем запросы
	for i := 0; i < 50; i++ {
		if server.CheckRateLimit("api", clientID) {
			allowedCount++
			fmt.Printf("Request %d: ALLOWED\n", i+1)
		} else {
			blockedCount++
			fmt.Printf("Request %d: BLOCKED\n", i+1)
		}

		// Небольшая задержка между запросами
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Printf("\nAPI Rate Limiter Results:\n")
	fmt.Printf("Allowed: %d\n", allowedCount)
	fmt.Printf("Blocked: %d\n", blockedCount)

	// Тестируем P2P rate limiter
	fmt.Println("\nTesting P2P Rate Limiter (Sliding Window):")
	fmt.Println("Max requests: 50/minute")

	allowedCount = 0
	blockedCount = 0

	// Быстро отправляем запросы
	for i := 0; i < 60; i++ {
		if server.CheckRateLimit("p2p", clientID) {
			allowedCount++
			fmt.Printf("P2P Request %d: ALLOWED\n", i+1)
		} else {
			blockedCount++
			fmt.Printf("P2P Request %d: BLOCKED\n", i+1)
		}

		// Небольшая задержка между запросами
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Printf("\nP2P Rate Limiter Results:\n")
	fmt.Printf("Allowed: %d\n", allowedCount)
	fmt.Printf("Blocked: %d\n", blockedCount)

	// Тестируем с разными клиентами
	fmt.Println("\nTesting with different clients:")

	clients := []string{"client_001", "client_002", "client_003"}

	for _, client := range clients {
		allowed := 0
		blocked := 0

		for i := 0; i < 10; i++ {
			if server.CheckRateLimit("api", client) {
				allowed++
			} else {
				blocked++
			}
			time.Sleep(50 * time.Millisecond)
		}

		fmt.Printf("Client %s: Allowed=%d, Blocked=%d\n", client, allowed, blocked)
	}

	// Получаем статистику rate limiter'ов
	fmt.Println("\nRate Limiter Statistics:")
	stats := server.GetRateLimiterStats()

	for limiterName, limiterStats := range stats {
		fmt.Printf("\n%s:\n", limiterName)
		if statsMap, ok := limiterStats.(map[string]interface{}); ok {
			for key, value := range statsMap {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	}

	// Тестируем восстановление после блокировки
	fmt.Println("\nTesting recovery after blocking:")
	fmt.Println("Waiting for rate limiter to recover...")

	time.Sleep(5 * time.Second)

	recoveryAllowed := 0
	for i := 0; i < 10; i++ {
		if server.CheckRateLimit("api", clientID) {
			recoveryAllowed++
			fmt.Printf("Recovery request %d: ALLOWED\n", i+1)
		} else {
			fmt.Printf("Recovery request %d: BLOCKED\n", i+1)
		}
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Printf("Recovery allowed: %d/10 requests\n", recoveryAllowed)

	// Симулируем реальную нагрузку
	fmt.Println("\nSimulating real-world load:")

	// Создаем несколько клиентов с разной активностью
	activeClients := []string{"active_001", "active_002", "active_003"}
	normalClients := []string{"normal_001", "normal_002"}

	// Активные клиенты отправляют много запросов
	for _, client := range activeClients {
		go func(c string) {
			for i := 0; i < 30; i++ {
				server.CheckRateLimit("api", c)
				time.Sleep(100 * time.Millisecond)
			}
		}(client)
	}

	// Обычные клиенты отправляют умеренное количество запросов
	for _, client := range normalClients {
		go func(c string) {
			for i := 0; i < 10; i++ {
				server.CheckRateLimit("api", c)
				time.Sleep(500 * time.Millisecond)
			}
		}(client)
	}

	// Ждем завершения симуляции
	time.Sleep(10 * time.Second)

	// Финальная статистика
	fmt.Println("\nFinal Rate Limiter Statistics:")
	finalStats := server.GetRateLimiterStats()

	for limiterName, limiterStats := range finalStats {
		fmt.Printf("\n%s:\n", limiterName)
		if statsMap, ok := limiterStats.(map[string]interface{}); ok {
			for key, value := range statsMap {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	}

	fmt.Println("\nRate Limiter Demo completed!")
}
