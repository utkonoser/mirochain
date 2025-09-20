package vm

import (
	"fmt"
	"math/big"
)

// GasCalculator рассчитывает стоимость газа для операций
type GasCalculator struct {
	baseCosts map[OpCode]uint64
}

// NewGasCalculator создает новый калькулятор газа
func NewGasCalculator() *GasCalculator {
	return &GasCalculator{
		baseCosts: map[OpCode]uint64{
			// Арифметические операции
			OP_ADD: 3,
			OP_SUB: 3,
			OP_MUL: 5,
			OP_DIV: 5,
			OP_MOD: 5,
			OP_POW: 10,
			
			// Логические операции
			OP_EQ:  3,
			OP_NE:  3,
			OP_LT:  3,
			OP_LE:  3,
			OP_GT:  3,
			OP_GE:  3,
			OP_AND: 3,
			OP_OR:  3,
			OP_NOT: 3,
			
			// Операции со стеком
			OP_PUSH: 3,
			OP_POP:  2,
			OP_DUP:  3,
			OP_SWAP: 3,
			
			// Операции с памятью
			OP_LOAD:  3,
			OP_STORE: 3,
			
			// Операции с хранилищем контракта
			OP_SLOAD:  200,
			OP_SSTORE: 20000,
			
			// Управление потоком
			OP_JUMP:   8,
			OP_JUMPI:  10,
			OP_CALL:   100,
			OP_RETURN: 0,
			OP_STOP:   0,
			
			// Операции с газом
			OP_GAS:      2,
			OP_GASLIMIT: 2,
			
			// Операции с балансом
			OP_BALANCE: 400,
			OP_SEND:    2300,
			
			// Операции с блоками
			OP_BLOCKHASH:   20,
			OP_BLOCKNUMBER: 2,
			OP_TIMESTAMP:   2,
			
			// Операции с транзакциями
			OP_TXORIGIN: 2,
			OP_TXVALUE:  2,
			OP_TXDATA:   3,
		},
	}
}

// CalculateGasCost рассчитывает стоимость газа для инструкции
func (gc *GasCalculator) CalculateGasCost(instruction Instruction) uint64 {
	baseCost := gc.baseCosts[instruction.OpCode]
	
	// Дополнительные расчеты для специфических операций
	switch instruction.OpCode {
	case OP_SSTORE:
		// Стоимость SSTORE зависит от того, устанавливаем ли мы новое значение
		// или изменяем существующее
		return baseCost
		
	case OP_PUSH:
		// Стоимость PUSH зависит от размера операнда
		if operand, ok := instruction.Operand.(*big.Int); ok {
			// Больше байт = больше газа
			byteSize := len(operand.Bytes())
			return baseCost + uint64(byteSize)
		}
		return baseCost
		
	case OP_CALL:
		// Стоимость CALL зависит от сложности вызова
		return baseCost + 100 // Базовая стоимость + дополнительная
		
	default:
		return baseCost
	}
}

// GasLimits определяет лимиты газа для разных типов операций
type GasLimits struct {
	MaxGasPerBlock    uint64 `json:"max_gas_per_block"`
	MaxGasPerTx       uint64 `json:"max_gas_per_tx"`
	MaxGasPerContract uint64 `json:"max_gas_per_contract"`
	MinGasPrice       uint64 `json:"min_gas_price"`
}

// DefaultGasLimits возвращает лимиты газа по умолчанию
func DefaultGasLimits() *GasLimits {
	return &GasLimits{
		MaxGasPerBlock:    10000000, // 10M газа на блок
		MaxGasPerTx:       1000000,  // 1M газа на транзакцию
		MaxGasPerContract: 500000,   // 500K газа на контракт
		MinGasPrice:       1,        // Минимальная цена газа
	}
}

// GasTracker отслеживает использование газа
type GasTracker struct {
	used     uint64
	limit    uint64
	price    uint64
	calculator *GasCalculator
}

// NewGasTracker создает новый трекер газа
func NewGasTracker(limit uint64, price uint64) *GasTracker {
	return &GasTracker{
		used:       0,
		limit:      limit,
		price:      price,
		calculator: NewGasCalculator(),
	}
}

// ConsumeGas потребляет газ
func (gt *GasTracker) ConsumeGas(amount uint64) error {
	if gt.used+amount > gt.limit {
		return fmt.Errorf("gas limit exceeded: used %d, limit %d", gt.used+amount, gt.limit)
	}
	
	gt.used += amount
	return nil
}

// ConsumeInstructionGas потребляет газ для инструкции
func (gt *GasTracker) ConsumeInstructionGas(instruction Instruction) error {
	cost := gt.calculator.CalculateGasCost(instruction)
	return gt.ConsumeGas(cost)
}

// GetGasUsed возвращает использованный газ
func (gt *GasTracker) GetGasUsed() uint64 {
	return gt.used
}

// GetGasRemaining возвращает оставшийся газ
func (gt *GasTracker) GetGasRemaining() uint64 {
	if gt.used >= gt.limit {
		return 0
	}
	return gt.limit - gt.used
}

// GetGasPrice возвращает цену газа
func (gt *GasTracker) GetGasPrice() uint64 {
	return gt.price
}

// GetTotalCost возвращает общую стоимость в токенах
func (gt *GasTracker) GetTotalCost() uint64 {
	return gt.used * gt.price
}

// GasOptimizer оптимизирует использование газа
type GasOptimizer struct {
	calculator *GasCalculator
}

// NewGasOptimizer создает новый оптимизатор газа
func NewGasOptimizer() *GasOptimizer {
	return &GasOptimizer{
		calculator: NewGasCalculator(),
	}
}

// OptimizeInstructions оптимизирует инструкции для экономии газа
func (g *GasOptimizer) OptimizeInstructions(instructions []Instruction) []Instruction {
	optimized := make([]Instruction, 0, len(instructions))
	
	for i, instruction := range instructions {
		// Простые оптимизации
		if g.canOptimize(instruction, i, instructions) {
			optimized = append(optimized, g.optimizeInstruction(instruction))
		} else {
			optimized = append(optimized, instruction)
		}
	}
	
	return optimized
}

// canOptimize проверяет, можно ли оптимизировать инструкцию
func (g *GasOptimizer) canOptimize(instruction Instruction, index int, instructions []Instruction) bool {
	// Оптимизация: замена PUSH 0 + ADD на исходное значение
	if instruction.OpCode == OP_PUSH && instruction.Operand == 0 {
		if index+1 < len(instructions) && instructions[index+1].OpCode == OP_ADD {
			return true
		}
	}
	
	// Оптимизация: замена PUSH 1 + MUL на исходное значение
	if instruction.OpCode == OP_PUSH && instruction.Operand == 1 {
		if index+1 < len(instructions) && instructions[index+1].OpCode == OP_MUL {
			return true
		}
	}
	
	return false
}

// optimizeInstruction оптимизирует инструкцию
func (g *GasOptimizer) optimizeInstruction(instruction Instruction) Instruction {
	// Простая оптимизация: уменьшаем стоимость газа
	optimized := instruction
	optimized.GasCost = optimized.GasCost / 2 // Уменьшаем стоимость вдвое
	return optimized
}

// GasEstimator оценивает стоимость газа для контракта
type GasEstimator struct {
	calculator *GasCalculator
}

// NewGasEstimator создает новый оценщик газа
func NewGasEstimator() *GasEstimator {
	return &GasEstimator{
		calculator: NewGasCalculator(),
	}
}

// EstimateGas оценивает стоимость газа для инструкций
func (ge *GasEstimator) EstimateGas(instructions []Instruction) uint64 {
	total := uint64(0)
	
	for _, instruction := range instructions {
		cost := ge.calculator.CalculateGasCost(instruction)
		total += cost
	}
	
	// Добавляем 20% буфер
	return total + (total * 20 / 100)
}

// EstimateContractGas оценивает стоимость газа для контракта
func (ge *GasEstimator) EstimateContractGas(source string) (uint64, error) {
	compiler := NewCompiler()
	instructions, err := compiler.Compile(source)
	if err != nil {
		return 0, err
	}
	
	return ge.EstimateGas(instructions), nil
}

// GasReport содержит отчет об использовании газа
type GasReport struct {
	TotalGasUsed    uint64            `json:"total_gas_used"`
	TotalCost       uint64            `json:"total_cost"`
	GasPrice        uint64            `json:"gas_price"`
	Instructions    []InstructionGas  `json:"instructions"`
	Optimizations   []Optimization    `json:"optimizations"`
}

// InstructionGas содержит информацию о газе для инструкции
type InstructionGas struct {
	OpCode    OpCode `json:"opcode"`
	GasCost   uint64 `json:"gas_cost"`
	GasPrice  uint64 `json:"gas_price"`
	TotalCost uint64 `json:"total_cost"`
}

// Optimization содержит информацию об оптимизации
type Optimization struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	GasSaved    uint64 `json:"gas_saved"`
}

// GenerateGasReport генерирует отчет об использовании газа
func (ge *GasEstimator) GenerateGasReport(instructions []Instruction, gasPrice uint64) *GasReport {
	report := &GasReport{
		TotalGasUsed:  0,
		TotalCost:     0,
		GasPrice:      gasPrice,
		Instructions:  make([]InstructionGas, 0, len(instructions)),
		Optimizations: make([]Optimization, 0),
	}
	
	for _, instruction := range instructions {
		cost := ge.calculator.CalculateGasCost(instruction)
		totalCost := cost * gasPrice
		
		report.TotalGasUsed += cost
		report.TotalCost += totalCost
		
		report.Instructions = append(report.Instructions, InstructionGas{
			OpCode:    instruction.OpCode,
			GasCost:   cost,
			GasPrice:  gasPrice,
			TotalCost: totalCost,
		})
	}
	
	return report
}
