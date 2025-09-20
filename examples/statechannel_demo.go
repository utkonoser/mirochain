//go:build statechannel_demo

package main

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"mirochain/internal/statechannel"
)

func main() {
	fmt.Println("=== State Channel Demo ===")
	fmt.Println()

	// Создаем менеджер state channels
	manager := statechannel.NewStateChannelManager()

	// 1. Создание каналов
	fmt.Println("1. Creating State Channels:")
	demonstrateChannelCreation(manager)
	fmt.Println()

	// 2. Депозиты в каналы
	fmt.Println("2. Making Deposits:")
	demonstrateDeposits(manager)
	fmt.Println()

	// 3. Транзакции в каналах
	fmt.Println("3. Channel Transactions:")
	demonstrateTransactions(manager)
	fmt.Println()

	// 4. Выводы из каналов
	fmt.Println("4. Withdrawals:")
	demonstrateWithdrawals(manager)
	fmt.Println()

	// 5. Споры и урегулирование
	fmt.Println("5. Disputes and Settlement:")
	demonstrateDisputes(manager)
	fmt.Println()

	// 6. Статистика и управление
	fmt.Println("6. Statistics and Management:")
	demonstrateStatistics(manager)
	fmt.Println()

	fmt.Println("State Channel demo completed!")
}

func demonstrateChannelCreation(manager *statechannel.StateChannelManager) {
	// Создаем различные типы каналов
	channels := []struct {
		participants  []string
		channelType   statechannel.ChannelType
		disputePeriod time.Duration
		metadata      map[string]interface{}
	}{
		{
			participants:  []string{"alice", "bob"},
			channelType:   statechannel.TypePayment,
			disputePeriod: 24 * time.Hour,
			metadata: map[string]interface{}{
				"description": "Payment channel between Alice and Bob",
				"max_amount":  "1000000",
			},
		},
		{
			participants:  []string{"alice", "charlie", "dave"},
			channelType:   statechannel.TypeMicropayment,
			disputePeriod: 1 * time.Hour,
			metadata: map[string]interface{}{
				"description": "Micropayment channel for gaming",
				"max_amount":  "10000",
			},
		},
		{
			participants:  []string{"bob", "eve"},
			channelType:   statechannel.TypeGaming,
			disputePeriod: 2 * time.Hour,
			metadata: map[string]interface{}{
				"description": "Gaming channel for tournaments",
				"max_amount":  "500000",
			},
		},
		{
			participants:  []string{"alice", "bob", "charlie", "dave", "eve"},
			channelType:   statechannel.TypePrediction,
			disputePeriod: 7 * 24 * time.Hour,
			metadata: map[string]interface{}{
				"description": "Prediction market channel",
				"max_amount":  "2000000",
			},
		},
	}

	for i, channelData := range channels {
		channel, err := manager.CreateChannel(channelData.participants, channelData.channelType, channelData.disputePeriod, channelData.metadata)
		if err != nil {
			log.Printf("Error creating channel %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Channel %d created:\n", i+1)
		fmt.Printf("  ID: %s\n", channel.ID)
		fmt.Printf("  Participants: %v\n", channel.Participants)
		fmt.Printf("  Type: %s\n", channel.ChannelType)
		fmt.Printf("  State: %s\n", channel.State)
		fmt.Printf("  Dispute Period: %v\n", channel.DisputePeriod)
		fmt.Printf("  Created At: %s\n", channel.CreatedAt.Format(time.RFC3339))
		fmt.Printf("  Expires At: %s\n", channel.ExpiresAt.Format(time.RFC3339))
		fmt.Println()
	}
}

func demonstrateDeposits(manager *statechannel.StateChannelManager) {
	// Получаем первый канал
	channels := manager.ListChannels()
	if len(channels) == 0 {
		fmt.Println("No channels available for deposit demo")
		return
	}

	channel := channels[0]
	fmt.Printf("Using channel: %s (%s)\n", channel.ID, channel.ChannelType)

	// Создаем депозиты для каждого участника
	deposits := []struct {
		participant string
		amount      *big.Int
		txHash      string
	}{
		{
			participant: "alice",
			amount:      big.NewInt(100000),
			txHash:      "0x1234567890abcdef",
		},
		{
			participant: "bob",
			amount:      big.NewInt(150000),
			txHash:      "0xabcdef1234567890",
		},
	}

	for i, depositData := range deposits {
		deposit, err := manager.Deposit(channel.ID, depositData.participant, depositData.amount, depositData.txHash)
		if err != nil {
			log.Printf("Error creating deposit %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Deposit %d created:\n", i+1)
		fmt.Printf("  ID: %s\n", deposit.ID)
		fmt.Printf("  Channel ID: %s\n", deposit.ChannelID)
		fmt.Printf("  Participant: %s\n", deposit.Participant)
		fmt.Printf("  Amount: %s\n", deposit.Amount.String())
		fmt.Printf("  Tx Hash: %s\n", deposit.TxHash)
		fmt.Printf("  Status: %s\n", deposit.Status)
		fmt.Printf("  Timestamp: %s\n", deposit.Timestamp.Format(time.RFC3339))
		fmt.Println()

		// Получаем обновленный канал
		updatedChannel, _ := manager.GetChannel(channel.ID)
		fmt.Printf("Updated channel state: %s\n", updatedChannel.State)
		fmt.Printf("Total deposit: %s\n", updatedChannel.TotalDeposit.String())
		fmt.Printf("Alice balance: %s\n", updatedChannel.Balances["alice"].String())
		fmt.Printf("Bob balance: %s\n", updatedChannel.Balances["bob"].String())
		fmt.Println()
	}
}

func demonstrateTransactions(manager *statechannel.StateChannelManager) {
	// Получаем первый канал
	channels := manager.ListChannels()
	if len(channels) == 0 {
		fmt.Println("No channels available for transaction demo")
		return
	}

	channel := channels[0]
	fmt.Printf("Using channel: %s (%s)\n", channel.ID, channel.ChannelType)

	// Создаем транзакции в канале
	transactions := []struct {
		from      string
		to        string
		amount    *big.Int
		data      map[string]interface{}
		signature string
	}{
		{
			from:   "alice",
			to:     "bob",
			amount: big.NewInt(5000),
			data: map[string]interface{}{
				"description": "Payment for services",
				"category":    "payment",
			},
			signature: "alice_signature_1",
		},
		{
			from:   "bob",
			to:     "alice",
			amount: big.NewInt(2000),
			data: map[string]interface{}{
				"description": "Refund for overpayment",
				"category":    "refund",
			},
			signature: "bob_signature_1",
		},
		{
			from:   "alice",
			to:     "bob",
			amount: big.NewInt(10000),
			data: map[string]interface{}{
				"description": "Large payment",
				"category":    "payment",
			},
			signature: "alice_signature_2",
		},
	}

	for i, txData := range transactions {
		transaction, err := manager.CreateTransaction(channel.ID, txData.from, txData.to, txData.amount, txData.data, txData.signature)
		if err != nil {
			log.Printf("Error creating transaction %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Transaction %d created:\n", i+1)
		fmt.Printf("  ID: %s\n", transaction.ID)
		fmt.Printf("  Channel ID: %s\n", transaction.ChannelID)
		fmt.Printf("  From: %s\n", transaction.From)
		fmt.Printf("  To: %s\n", transaction.To)
		fmt.Printf("  Amount: %s\n", transaction.Amount.String())
		fmt.Printf("  Nonce: %d\n", transaction.Nonce)
		fmt.Printf("  Status: %s\n", transaction.Status)
		fmt.Printf("  Gas Used: %d\n", transaction.GasUsed)
		fmt.Printf("  Gas Price: %s\n", transaction.GasPrice.String())
		fmt.Printf("  Timestamp: %s\n", transaction.Timestamp.Format(time.RFC3339))
		fmt.Println()

		// Получаем обновленный канал
		updatedChannel, _ := manager.GetChannel(channel.ID)
		fmt.Printf("Updated balances:\n")
		fmt.Printf("  Alice: %s\n", updatedChannel.Balances["alice"].String())
		fmt.Printf("  Bob: %s\n", updatedChannel.Balances["bob"].String())
		fmt.Printf("  Nonce: %d\n", updatedChannel.Nonce)
		fmt.Println()
	}
}

func demonstrateWithdrawals(manager *statechannel.StateChannelManager) {
	// Получаем первый канал
	channels := manager.ListChannels()
	if len(channels) == 0 {
		fmt.Println("No channels available for withdrawal demo")
		return
	}

	channel := channels[0]
	fmt.Printf("Using channel: %s (%s)\n", channel.ID, channel.ChannelType)

	// Создаем выводы
	withdrawals := []struct {
		participant string
		amount      *big.Int
	}{
		{
			participant: "alice",
			amount:      big.NewInt(20000),
		},
		{
			participant: "bob",
			amount:      big.NewInt(15000),
		},
	}

	for i, withdrawalData := range withdrawals {
		withdrawal, err := manager.Withdraw(channel.ID, withdrawalData.participant, withdrawalData.amount)
		if err != nil {
			log.Printf("Error creating withdrawal %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Withdrawal %d created:\n", i+1)
		fmt.Printf("  ID: %s\n", withdrawal.ID)
		fmt.Printf("  Channel ID: %s\n", withdrawal.ChannelID)
		fmt.Printf("  Participant: %s\n", withdrawal.Participant)
		fmt.Printf("  Amount: %s\n", withdrawal.Amount.String())
		fmt.Printf("  Status: %s\n", withdrawal.Status)
		fmt.Printf("  Timestamp: %s\n", withdrawal.Timestamp.Format(time.RFC3339))
		fmt.Println()

		// Обрабатываем вывод
		txHash := fmt.Sprintf("withdrawal_tx_%d", i+1)
		err = manager.ProcessWithdrawal(withdrawal.ID, txHash)
		if err != nil {
			log.Printf("Error processing withdrawal %d: %v", i+1, err)
			continue
		}

		fmt.Printf("Withdrawal %d processed with tx hash: %s\n", i+1, txHash)
		fmt.Println()

		// Получаем обновленный канал
		updatedChannel, _ := manager.GetChannel(channel.ID)
		fmt.Printf("Updated balances after withdrawal:\n")
		fmt.Printf("  Alice: %s\n", updatedChannel.Balances["alice"].String())
		fmt.Printf("  Bob: %s\n", updatedChannel.Balances["bob"].String())
		fmt.Printf("  Total Deposit: %s\n", updatedChannel.TotalDeposit.String())
		fmt.Println()
	}
}

func demonstrateDisputes(manager *statechannel.StateChannelManager) {
	// Получаем первый канал
	channels := manager.ListChannels()
	if len(channels) == 0 {
		fmt.Println("No channels available for dispute demo")
		return
	}

	channel := channels[0]
	fmt.Printf("Using channel: %s (%s)\n", channel.ID, channel.ChannelType)

	// Создаем обновление состояния для спора
	disputedState := &statechannel.StateUpdate{
		ChannelID:    channel.ID,
		Nonce:        channel.Nonce + 1,
		Balances: map[string]*big.Int{
			"alice": big.NewInt(50000),
			"bob":   big.NewInt(200000),
		},
		Participants: channel.Participants,
		Data: map[string]interface{}{
			"dispute_reason": "Incorrect balance calculation",
		},
		Signature:  "disputed_signature",
		Timestamp:  time.Now(),
		UpdateType: statechannel.UpdateDispute,
	}

	// Инициируем спор
	dispute, err := manager.InitiateDispute(channel.ID, "alice", "Balance calculation error", "Evidence of incorrect calculation", disputedState)
	if err != nil {
		log.Printf("Error creating dispute: %v", err)
		return
	}

	fmt.Printf("Dispute created:\n")
	fmt.Printf("  ID: %s\n", dispute.ID)
	fmt.Printf("  Channel ID: %s\n", dispute.ChannelID)
	fmt.Printf("  Initiator: %s\n", dispute.Initiator)
	fmt.Printf("  Reason: %s\n", dispute.Reason)
	fmt.Printf("  Evidence: %s\n", dispute.Evidence)
	fmt.Printf("  Status: %s\n", dispute.Status)
	fmt.Printf("  Timestamp: %s\n", dispute.Timestamp.Format(time.RFC3339))
	fmt.Println()

	// Получаем обновленный канал
	updatedChannel, _ := manager.GetChannel(channel.ID)
	fmt.Printf("Channel state after dispute: %s\n", updatedChannel.State)
	fmt.Println()

	// Разрешаем спор
	resolution := "Dispute resolved in favor of Alice. Balance corrected."
	err = manager.ResolveDispute(dispute.ID, resolution)
	if err != nil {
		log.Printf("Error resolving dispute: %v", err)
		return
	}

	fmt.Printf("Dispute resolved:\n")
	fmt.Printf("  Resolution: %s\n", resolution)
	fmt.Println()

	// Получаем обновленный канал
	updatedChannel, _ = manager.GetChannel(channel.ID)
	fmt.Printf("Channel state after resolution: %s\n", updatedChannel.State)
	fmt.Println()

	// Закрываем канал
	err = manager.CloseChannel(channel.ID)
	if err != nil {
		log.Printf("Error closing channel: %v", err)
		return
	}

	fmt.Printf("Channel closed successfully\n")
	fmt.Println()

	// Урегулируем канал
	finalState := &statechannel.StateUpdate{
		ChannelID:    channel.ID,
		Nonce:        channel.Nonce + 1,
		Balances:     channel.Balances,
		Participants: channel.Participants,
		Data: map[string]interface{}{
			"settlement": "Final settlement",
		},
		Signature:  "final_signature",
		Timestamp:  time.Now(),
		UpdateType: statechannel.UpdateSettlement,
	}

	settlement, err := manager.SettleChannel(channel.ID, finalState, "settlement_tx_hash")
	if err != nil {
		log.Printf("Error settling channel: %v", err)
		return
	}

	fmt.Printf("Channel settled:\n")
	fmt.Printf("  Settlement ID: %s\n", settlement.ID)
	fmt.Printf("  Channel ID: %s\n", settlement.ChannelID)
	fmt.Printf("  Tx Hash: %s\n", settlement.TxHash)
	fmt.Printf("  Gas Used: %d\n", settlement.GasUsed)
	fmt.Printf("  Gas Price: %s\n", settlement.GasPrice.String())
	fmt.Printf("  Timestamp: %s\n", settlement.Timestamp.Format(time.RFC3339))
	fmt.Println()
}

func demonstrateStatistics(manager *statechannel.StateChannelManager) {
	channels := manager.ListChannels()
	if len(channels) == 0 {
		fmt.Println("No channels available for statistics demo")
		return
	}

	// Показываем статистику для каждого канала
	for i, channel := range channels {
		fmt.Printf("Channel %d: %s (%s)\n", i+1, channel.ID, channel.ChannelType)
		
		stats, err := manager.GetChannelStats(channel.ID)
		if err != nil {
			log.Printf("Failed to get stats for channel %d: %v", i+1, err)
			continue
		}

		fmt.Printf("  State: %s\n", stats["state"])
		fmt.Printf("  Participants: %v\n", stats["participants"])
		fmt.Printf("  Nonce: %d\n", stats["nonce"])
		fmt.Printf("  Total Deposit: %s\n", stats["total_deposit"])
		fmt.Printf("  Channel Type: %s\n", stats["channel_type"])
		fmt.Printf("  Created At: %s\n", stats["created_at"])
		fmt.Printf("  Updated At: %s\n", stats["updated_at"])
		fmt.Printf("  Expires At: %s\n", stats["expires_at"])
		fmt.Printf("  Total Transactions: %d\n", stats["total_transactions"])
		fmt.Printf("  Total Deposits: %d\n", stats["total_deposits"])
		fmt.Printf("  Total Withdrawals: %d\n", stats["total_withdrawals"])
		fmt.Printf("  Active Disputes: %d\n", stats["active_disputes"])
		fmt.Println()

		// Показываем балансы
		balances := stats["balances"].(map[string]*big.Int)
		fmt.Printf("  Balances:\n")
		for participant, balance := range balances {
			fmt.Printf("    %s: %s\n", participant, balance.String())
		}
		fmt.Println()

		// Показываем транзакции
		transactions, err := manager.GetChannelTransactions(channel.ID)
		if err != nil {
			log.Printf("Failed to get transactions for channel %d: %v", i+1, err)
			continue
		}

		fmt.Printf("  Transactions (%d):\n", len(transactions))
		for j, tx := range transactions {
			if j >= 3 { // Показываем только первые 3
				fmt.Printf("    ... and %d more\n", len(transactions)-3)
				break
			}
			fmt.Printf("    %d. %s -> %s: %s (%s)\n", j+1, tx.From, tx.To, tx.Amount.String(), tx.Status)
		}
		fmt.Println()
	}
}
