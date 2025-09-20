package parallel

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync"
	"time"

	"mirochain/internal/blockchain"
)

// TransactionProcessor представляет процессор для параллельной обработки транзакций
type TransactionProcessor struct {
	workerCount int
	queue       chan *TransactionTask
	workers     []*Worker
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex
	running     bool
}

// TransactionTask представляет задачу обработки транзакции
type TransactionTask struct {
	Transaction *blockchain.Transaction
	UTXOSet     *blockchain.UTXOSet
	Result      chan *TransactionResult
	Priority    int // Чем выше, тем приоритетнее
}

// TransactionResult представляет результат обработки транзакции
type TransactionResult struct {
	Transaction *blockchain.Transaction
	Valid       bool
	Error       error
	ProcessTime time.Duration
}

// Worker представляет воркера для обработки транзакций
type Worker struct {
	ID      int
	Queue   chan *TransactionTask
	UTXOSet *blockchain.UTXOSet
	ctx     context.Context
	cancel  context.CancelFunc
	wg      *sync.WaitGroup
	mu      *sync.RWMutex
	stats   *WorkerStats
}

// WorkerStats представляет статистику воркера
type WorkerStats struct {
	ProcessedCount int64
	ErrorCount     int64
	TotalTime      time.Duration
	LastProcessed  time.Time
}

// NewTransactionProcessor создает новый процессор транзакций
func NewTransactionProcessor(workerCount int, queueSize int) *TransactionProcessor {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}

	if queueSize <= 0 {
		queueSize = 1000
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TransactionProcessor{
		workerCount: workerCount,
		queue:       make(chan *TransactionTask, queueSize),
		ctx:         ctx,
		cancel:      cancel,
		running:     false,
	}
}

// Start запускает процессор
func (tp *TransactionProcessor) Start(utxoSet *blockchain.UTXOSet) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if tp.running {
		return fmt.Errorf("processor is already running")
	}

	// Создаем воркеров
	tp.workers = make([]*Worker, tp.workerCount)
	for i := 0; i < tp.workerCount; i++ {
		workerCtx, workerCancel := context.WithCancel(tp.ctx)
		tp.workers[i] = &Worker{
			ID:      i,
			Queue:   tp.queue,
			UTXOSet: utxoSet,
			ctx:     workerCtx,
			cancel:  workerCancel,
			wg:      &tp.wg,
			mu:      &tp.mu,
			stats:   &WorkerStats{},
		}
	}

	// Запускаем воркеров
	for _, worker := range tp.workers {
		tp.wg.Add(1)
		go worker.start()
	}

	tp.running = true
	slog.Info("Transaction processor started", "workers", tp.workerCount, "queue_size", cap(tp.queue))

	return nil
}

// Stop останавливает процессор
func (tp *TransactionProcessor) Stop() error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if !tp.running {
		return fmt.Errorf("processor is not running")
	}

	// Отменяем контекст
	tp.cancel()

	// Закрываем очередь
	close(tp.queue)

	// Ждем завершения всех воркеров
	tp.wg.Wait()

	tp.running = false
	slog.Info("Transaction processor stopped")

	return nil
}

// ProcessTransaction обрабатывает транзакцию параллельно
func (tp *TransactionProcessor) ProcessTransaction(transaction *blockchain.Transaction, utxoSet *blockchain.UTXOSet, priority int) (*TransactionResult, error) {
	tp.mu.RLock()
	if !tp.running {
		tp.mu.RUnlock()
		return nil, fmt.Errorf("processor is not running")
	}
	tp.mu.RUnlock()

	// Создаем задачу
	task := &TransactionTask{
		Transaction: transaction,
		UTXOSet:     utxoSet,
		Result:      make(chan *TransactionResult, 1),
		Priority:    priority,
	}

	// Отправляем задачу в очередь
	select {
	case tp.queue <- task:
		// Задача отправлена
	case <-tp.ctx.Done():
		return nil, fmt.Errorf("processor is stopping")
	default:
		return nil, fmt.Errorf("queue is full")
	}

	// Ждем результат
	select {
	case result := <-task.Result:
		return result, nil
	case <-tp.ctx.Done():
		return nil, fmt.Errorf("processor is stopping")
	}
}

// ProcessTransactionsBatch обрабатывает несколько транзакций параллельно
func (tp *TransactionProcessor) ProcessTransactionsBatch(transactions []*blockchain.Transaction, utxoSet *blockchain.UTXOSet, priority int) ([]*TransactionResult, error) {
	if len(transactions) == 0 {
		return []*TransactionResult{}, nil
	}

	// Создаем задачи
	tasks := make([]*TransactionTask, len(transactions))
	for i, tx := range transactions {
		tasks[i] = &TransactionTask{
			Transaction: tx,
			UTXOSet:     utxoSet,
			Result:      make(chan *TransactionResult, 1),
			Priority:    priority,
		}
	}

	// Отправляем все задачи
	for _, task := range tasks {
		select {
		case tp.queue <- task:
			// Задача отправлена
		case <-tp.ctx.Done():
			return nil, fmt.Errorf("processor is stopping")
		default:
			return nil, fmt.Errorf("queue is full")
		}
	}

	// Собираем результаты
	results := make([]*TransactionResult, len(tasks))
	for i, task := range tasks {
		select {
		case result := <-task.Result:
			results[i] = result
		case <-tp.ctx.Done():
			return nil, fmt.Errorf("processor is stopping")
		}
	}

	return results, nil
}

// GetStats возвращает статистику процессора
func (tp *TransactionProcessor) GetStats() map[string]interface{} {
	tp.mu.RLock()
	defer tp.mu.RUnlock()

	stats := map[string]interface{}{
		"running":      tp.running,
		"worker_count": tp.workerCount,
		"queue_size":   cap(tp.queue),
		"queue_length": len(tp.queue),
		"workers":      make([]map[string]interface{}, len(tp.workers)),
	}

	// Собираем статистику воркеров
	for i, worker := range tp.workers {
		workerStats := map[string]interface{}{
			"id":              worker.ID,
			"processed_count": worker.stats.ProcessedCount,
			"error_count":     worker.stats.ErrorCount,
			"total_time":      worker.stats.TotalTime,
			"last_processed":  worker.stats.LastProcessed,
		}
		stats["workers"].([]map[string]interface{})[i] = workerStats
	}

	return stats
}

// start запускает воркера
func (w *Worker) start() {
	defer w.wg.Done()

	slog.Info("Worker started", "id", w.ID)

	for {
		select {
		case task := <-w.Queue:
			if task == nil {
				// Канал закрыт
				slog.Info("Worker stopping", "id", w.ID)
				return
			}

			// Обрабатываем задачу
			result := w.processTask(task)

			// Отправляем результат
			select {
			case task.Result <- result:
				// Результат отправлен
			case <-w.ctx.Done():
				// Контекст отменен
				return
			}

		case <-w.ctx.Done():
			// Контекст отменен
			slog.Info("Worker stopping", "id", w.ID)
			return
		}
	}
}

// processTask обрабатывает задачу
func (w *Worker) processTask(task *TransactionTask) *TransactionResult {
	start := time.Now()

	// Валидируем транзакцию
	valid, err := w.validateTransaction(task.Transaction, task.UTXOSet)

	processTime := time.Since(start)

	// Обновляем статистику
	w.mu.Lock()
	w.stats.ProcessedCount++
	if err != nil {
		w.stats.ErrorCount++
	}
	w.stats.TotalTime += processTime
	w.stats.LastProcessed = time.Now()
	w.mu.Unlock()

	return &TransactionResult{
		Transaction: task.Transaction,
		Valid:       valid,
		Error:       err,
		ProcessTime: processTime,
	}
}

// validateTransaction валидирует транзакцию
func (w *Worker) validateTransaction(transaction *blockchain.Transaction, utxoSet *blockchain.UTXOSet) (bool, error) {
	// Проверяем базовую валидность транзакции
	if !transaction.IsValid() {
		return false, fmt.Errorf("transaction is not valid")
	}

	// Для coinbase транзакций дополнительная валидация не нужна
	if transaction.IsCoinbase() {
		return true, nil
	}

	// Проверяем, что все входы существуют в UTXO
	for _, input := range transaction.Inputs {
		_, exists := utxoSet.GetUTXO(input.TransactionID, input.OutputIndex)
		if !exists {
			return false, fmt.Errorf("UTXO not found for input %x:%d", input.TransactionID, input.OutputIndex)
		}
	}

	return true, nil
}

// GetWorkerStats возвращает статистику воркера
func (w *Worker) GetWorkerStats() *WorkerStats {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Возвращаем копию статистики
	return &WorkerStats{
		ProcessedCount: w.stats.ProcessedCount,
		ErrorCount:     w.stats.ErrorCount,
		TotalTime:      w.stats.TotalTime,
		LastProcessed:  w.stats.LastProcessed,
	}
}
