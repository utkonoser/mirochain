package vm

import (
	"fmt"
	"strconv"
	"strings"
)

// Compiler компилирует код контракта в инструкции VM
type Compiler struct {
	instructions []Instruction
	labels       map[string]int
}

// NewCompiler создает новый компилятор
func NewCompiler() *Compiler {
	return &Compiler{
		instructions: make([]Instruction, 0),
		labels:       make(map[string]int),
	}
}

// Compile компилирует исходный код контракта
func (c *Compiler) Compile(source string) ([]Instruction, error) {
	c.instructions = make([]Instruction, 0)
	c.labels = make(map[string]int)
	
	lines := strings.Split(source, "\n")
	
	// Первый проход: собираем метки
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		
		if strings.HasSuffix(line, ":") {
			label := strings.TrimSuffix(line, ":")
			c.labels[label] = i
		}
	}
	
	// Второй проход: компилируем инструкции
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		
		if strings.HasSuffix(line, ":") {
			continue // Пропускаем метки
		}
		
		instruction, err := c.compileLine(line)
		if err != nil {
			return nil, fmt.Errorf("compilation error: %v", err)
		}
		
		c.instructions = append(c.instructions, instruction)
	}
	
	return c.instructions, nil
}

// compileLine компилирует одну строку кода
func (c *Compiler) compileLine(line string) (Instruction, error) {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return Instruction{}, fmt.Errorf("empty line")
	}
	
	opcode := parts[0]
	
	switch strings.ToUpper(opcode) {
	case "PUSH":
		if len(parts) < 2 {
			return Instruction{}, fmt.Errorf("PUSH requires operand")
		}
		operand, err := c.parseOperand(parts[1])
		if err != nil {
			return Instruction{}, err
		}
		return Instruction{OpCode: OP_PUSH, Operand: operand, GasCost: 3}, nil
		
	case "POP":
		return Instruction{OpCode: OP_POP, GasCost: 2}, nil
		
	case "ADD":
		return Instruction{OpCode: OP_ADD, GasCost: 3}, nil
		
	case "SUB":
		return Instruction{OpCode: OP_SUB, GasCost: 3}, nil
		
	case "MUL":
		return Instruction{OpCode: OP_MUL, GasCost: 5}, nil
		
	case "DIV":
		return Instruction{OpCode: OP_DIV, GasCost: 5}, nil
		
	case "LOAD":
		if len(parts) < 2 {
			return Instruction{}, fmt.Errorf("LOAD requires operand")
		}
		return Instruction{OpCode: OP_LOAD, Operand: parts[1], GasCost: 3}, nil
		
	case "STORE":
		if len(parts) < 2 {
			return Instruction{}, fmt.Errorf("STORE requires operand")
		}
		return Instruction{OpCode: OP_STORE, Operand: parts[1], GasCost: 3}, nil
		
	case "SLOAD":
		if len(parts) < 2 {
			return Instruction{}, fmt.Errorf("SLOAD requires operand")
		}
		return Instruction{OpCode: OP_SLOAD, Operand: parts[1], GasCost: 200}, nil
		
	case "SSTORE":
		if len(parts) < 2 {
			return Instruction{}, fmt.Errorf("SSTORE requires operand")
		}
		return Instruction{OpCode: OP_SSTORE, Operand: parts[1], GasCost: 20000}, nil
		
	case "JUMP":
		if len(parts) < 2 {
			return Instruction{}, fmt.Errorf("JUMP requires operand")
		}
		address, err := c.resolveLabel(parts[1])
		if err != nil {
			return Instruction{}, err
		}
		return Instruction{OpCode: OP_JUMP, Operand: address, GasCost: 8}, nil
		
	case "JUMPI":
		if len(parts) < 2 {
			return Instruction{}, fmt.Errorf("JUMPI requires operand")
		}
		address, err := c.resolveLabel(parts[1])
		if err != nil {
			return Instruction{}, err
		}
		return Instruction{OpCode: OP_JUMPI, Operand: address, GasCost: 10}, nil
		
	case "RETURN":
		return Instruction{OpCode: OP_RETURN, GasCost: 0}, nil
		
	case "STOP":
		return Instruction{OpCode: OP_STOP, GasCost: 0}, nil
		
	default:
		return Instruction{}, fmt.Errorf("unknown opcode: %s", opcode)
	}
}

// parseOperand парсит операнд инструкции
func (c *Compiler) parseOperand(operand string) (interface{}, error) {
	// Пробуем парсить как число
	if num, err := strconv.ParseInt(operand, 10, 64); err == nil {
		return num, nil
	}
	
	// Пробуем парсить как строку
	if strings.HasPrefix(operand, "\"") && strings.HasSuffix(operand, "\"") {
		return strings.Trim(operand, "\""), nil
	}
	
	// Возвращаем как строку
	return operand, nil
}

// resolveLabel разрешает метку в адрес
func (c *Compiler) resolveLabel(label string) (int, error) {
	address, exists := c.labels[label]
	if !exists {
		return 0, fmt.Errorf("undefined label: %s", label)
	}
	return address, nil
}

// ContractTemplate представляет шаблон контракта
type ContractTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	GasEstimate uint64 `json:"gas_estimate"`
}

// GetContractTemplates возвращает шаблоны контрактов
func GetContractTemplates() []ContractTemplate {
	return []ContractTemplate{
		{
			Name:        "Simple Counter",
			Description: "Простой счетчик с функциями увеличения и получения значения",
			Source: `// Simple Counter Contract
// Функция: увеличить счетчик
PUSH 1
LOAD "counter"
ADD
STORE "counter"
RETURN

// Функция: получить значение счетчика
LOAD "counter"
RETURN`,
			GasEstimate: 1000,
		},
		{
			Name:        "Token Balance",
			Description: "Простой токен с балансом",
			Source: `// Token Balance Contract
// Функция: установить баланс
PUSH 1000
STORE "balance"
RETURN

// Функция: получить баланс
LOAD "balance"
RETURN`,
			GasEstimate: 1500,
		},
		{
			Name:        "Math Calculator",
			Description: "Простой калькулятор для базовых операций",
			Source: `// Math Calculator Contract
// Функция: сложение
LOAD "a"
LOAD "b"
ADD
STORE "result"
RETURN

// Функция: умножение
LOAD "a"
LOAD "b"
MUL
STORE "result"
RETURN`,
			GasEstimate: 2000,
		},
		{
			Name:        "Storage Contract",
			Description: "Контракт для хранения данных в блокчейне",
			Source: `// Storage Contract
// Функция: сохранить значение
PUSH 42
SSTORE "data"
RETURN

// Функция: получить значение
SLOAD "data"
RETURN`,
			GasEstimate: 25000,
		},
	}
}

// CompileTemplate компилирует шаблон контракта
func CompileTemplate(templateName string) ([]Instruction, error) {
	templates := GetContractTemplates()
	
	for _, template := range templates {
		if template.Name == templateName {
			compiler := NewCompiler()
			return compiler.Compile(template.Source)
		}
	}
	
	return nil, fmt.Errorf("template not found: %s", templateName)
}
