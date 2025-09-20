package vm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// VM представляет виртуальную машину для выполнения смарт-контрактов
type VM struct {
	stack     []*big.Int
	memory    map[string]*big.Int
	storage   map[string]*big.Int
	gas       uint64
	gasLimit  uint64
	pc        int // Program counter
	contracts map[string]*Contract
}

// Contract представляет смарт-контракт
type Contract struct {
	Address   string            `json:"address"`
	Code      []byte            `json:"code"`
	Storage   map[string]*big.Int `json:"storage"`
	Balance   *big.Int          `json:"balance"`
	Owner     string            `json:"owner"`
	CreatedAt time.Time         `json:"created_at"`
}

// Instruction представляет инструкцию виртуальной машины
type Instruction struct {
	OpCode   OpCode      `json:"opcode"`
	Operand  interface{} `json:"operand,omitempty"`
	GasCost  uint64      `json:"gas_cost"`
}

// OpCode определяет операции виртуальной машины
type OpCode int

const (
	// Арифметические операции
	OP_ADD OpCode = iota
	OP_SUB
	OP_MUL
	OP_DIV
	OP_MOD
	OP_POW
	
	// Логические операции
	OP_EQ
	OP_NE
	OP_LT
	OP_LE
	OP_GT
	OP_GE
	OP_AND
	OP_OR
	OP_NOT
	
	// Операции со стеком
	OP_PUSH
	OP_POP
	OP_DUP
	OP_SWAP
	
	// Операции с памятью
	OP_LOAD
	OP_STORE
	
	// Операции с хранилищем контракта
	OP_SLOAD
	OP_SSTORE
	
	// Управление потоком
	OP_JUMP
	OP_JUMPI
	OP_CALL
	OP_RETURN
	OP_STOP
	
	// Операции с газом
	OP_GAS
	OP_GASLIMIT
	
	// Операции с балансом
	OP_BALANCE
	OP_SEND
	
	// Операции с блоками
	OP_BLOCKHASH
	OP_BLOCKNUMBER
	OP_TIMESTAMP
	
	// Операции с транзакциями
	OP_TXORIGIN
	OP_TXVALUE
	OP_TXDATA
)

// ExecutionResult содержит результат выполнения контракта
type ExecutionResult struct {
	Success    bool        `json:"success"`
	GasUsed    uint64      `json:"gas_used"`
	ReturnData []byte      `json:"return_data"`
	Error      string      `json:"error,omitempty"`
	Logs       []Log       `json:"logs"`
	Storage    map[string]*big.Int `json:"storage"`
}

// Log представляет лог контракта
type Log struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    []byte   `json:"data"`
}

// NewVM создает новую виртуальную машину
func NewVM(gasLimit uint64) *VM {
	return &VM{
		stack:     make([]*big.Int, 0),
		memory:    make(map[string]*big.Int),
		storage:   make(map[string]*big.Int),
		gas:       gasLimit,
		gasLimit:  gasLimit,
		pc:        0,
		contracts: make(map[string]*Contract),
	}
}

// DeployContract развертывает новый контракт
func (vm *VM) DeployContract(instructions []Instruction, owner string, initialBalance *big.Int) (*Contract, error) {
	// Генерируем адрес контракта
	address := vm.generateContractAddress()
	
	contract := &Contract{
		Address:   address,
		Code:      vm.instructionsToBytes(instructions),
		Storage:   make(map[string]*big.Int),
		Balance:   new(big.Int).Set(initialBalance),
		Owner:     owner,
		CreatedAt: time.Now(),
	}
	
	vm.contracts[address] = contract
	return contract, nil
}

// ExecuteContract выполняет контракт
func (vm *VM) ExecuteContract(contractAddress string, input []byte) (*ExecutionResult, error) {
	contract, exists := vm.contracts[contractAddress]
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", contractAddress)
	}
	
	// Сбрасываем состояние VM
	vm.reset()
	
	// Парсим код контракта в инструкции
	instructions, err := vm.parseCode(contract.Code)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("failed to parse contract code: %v", err),
		}, nil
	}
	
	// Выполняем инструкции
	result := vm.executeInstructions(instructions, contract, input)
	
	return result, nil
}

// CallContract вызывает функцию контракта
func (vm *VM) CallContract(contractAddress string, function string, args []interface{}) (*ExecutionResult, error) {
	// Конвертируем аргументы в байты
	input, err := json.Marshal(map[string]interface{}{
		"function": function,
		"args":     args,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %v", err)
	}
	
	return vm.ExecuteContract(contractAddress, input)
}

// GetContract возвращает контракт по адресу
func (vm *VM) GetContract(address string) (*Contract, bool) {
	contract, exists := vm.contracts[address]
	return contract, exists
}

// ListContracts возвращает список всех контрактов
func (vm *VM) ListContracts() []*Contract {
	contracts := make([]*Contract, 0, len(vm.contracts))
	for _, contract := range vm.contracts {
		contracts = append(contracts, contract)
	}
	return contracts
}

// GetGasRemaining возвращает оставшийся газ
func (vm *VM) GetGasRemaining() uint64 {
	return vm.gas
}

// GetGasUsed возвращает использованный газ
func (vm *VM) GetGasUsed() uint64 {
	return vm.gasLimit - vm.gas
}

// generateContractAddress генерирует адрес контракта
func (vm *VM) generateContractAddress() string {
	// Простая генерация адреса на основе времени и количества контрактов
	timestamp := time.Now().UnixNano()
	count := len(vm.contracts)
	return fmt.Sprintf("contract_%d_%d", timestamp, count)
}

// instructionsToBytes конвертирует инструкции в байты
func (vm *VM) instructionsToBytes(instructions []Instruction) []byte {
	// Простая сериализация инструкций в JSON
	data, err := json.Marshal(instructions)
	if err != nil {
		return []byte{}
	}
	return data
}

// reset сбрасывает состояние VM
func (vm *VM) reset() {
	vm.stack = vm.stack[:0]
	vm.memory = make(map[string]*big.Int)
	vm.pc = 0
	vm.gas = vm.gasLimit
}

// parseCode парсит код контракта в инструкции
func (vm *VM) parseCode(code []byte) ([]Instruction, error) {
	// Простой парсер для демонстрации
	// В реальной реализации здесь должен быть полноценный парсер
	instructions := make([]Instruction, 0)
	
	// Пока что возвращаем пустой список
	// TODO: Реализовать парсинг кода контракта
	return instructions, nil
}

// executeInstructions выполняет инструкции контракта
func (vm *VM) executeInstructions(instructions []Instruction, contract *Contract, input []byte) *ExecutionResult {
	result := &ExecutionResult{
		Success: true,
		Logs:    make([]Log, 0),
		Storage: make(map[string]*big.Int),
	}
	
	// Копируем хранилище контракта
	for k, v := range contract.Storage {
		result.Storage[k] = new(big.Int).Set(v)
	}
	
	// Выполняем инструкции
	for vm.pc < len(instructions) {
		instruction := instructions[vm.pc]
		
		// Проверяем газ
		if vm.gas < instruction.GasCost {
			result.Success = false
			result.Error = "out of gas"
			break
		}
		
		// Выполняем инструкцию
		err := vm.executeInstruction(instruction, contract, result)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			break
		}
		
		// Уменьшаем газ
		vm.gas -= instruction.GasCost
		vm.pc++
	}
	
	result.GasUsed = vm.GetGasUsed()
	return result
}

// executeInstruction выполняет одну инструкцию
func (vm *VM) executeInstruction(instruction Instruction, contract *Contract, result *ExecutionResult) error {
	switch instruction.OpCode {
	case OP_ADD:
		return vm.opAdd()
	case OP_SUB:
		return vm.opSub()
	case OP_MUL:
		return vm.opMul()
	case OP_DIV:
		return vm.opDiv()
	case OP_PUSH:
		return vm.opPush(instruction.Operand)
	case OP_POP:
		return vm.opPop()
	case OP_LOAD:
		return vm.opLoad(instruction.Operand)
	case OP_STORE:
		return vm.opStore(instruction.Operand)
	case OP_SLOAD:
		return vm.opSLoad(instruction.Operand, contract)
	case OP_SSTORE:
		return vm.opSStore(instruction.Operand, contract, result)
	case OP_JUMP:
		return vm.opJump(instruction.Operand)
	case OP_JUMPI:
		return vm.opJumpI(instruction.Operand)
	case OP_RETURN:
		return vm.opReturn(instruction.Operand, result)
	case OP_STOP:
		return vm.opStop()
	default:
		return fmt.Errorf("unknown opcode: %d", instruction.OpCode)
	}
}

// Арифметические операции
func (vm *VM) opAdd() error {
	if len(vm.stack) < 2 {
		return fmt.Errorf("stack underflow")
	}
	
	b := vm.stack[len(vm.stack)-1]
	a := vm.stack[len(vm.stack)-2]
	vm.stack = vm.stack[:len(vm.stack)-2]
	
	result := new(big.Int).Add(a, b)
	vm.stack = append(vm.stack, result)
	return nil
}

func (vm *VM) opSub() error {
	if len(vm.stack) < 2 {
		return fmt.Errorf("stack underflow")
	}
	
	b := vm.stack[len(vm.stack)-1]
	a := vm.stack[len(vm.stack)-2]
	vm.stack = vm.stack[:len(vm.stack)-2]
	
	result := new(big.Int).Sub(a, b)
	vm.stack = append(vm.stack, result)
	return nil
}

func (vm *VM) opMul() error {
	if len(vm.stack) < 2 {
		return fmt.Errorf("stack underflow")
	}
	
	b := vm.stack[len(vm.stack)-1]
	a := vm.stack[len(vm.stack)-2]
	vm.stack = vm.stack[:len(vm.stack)-2]
	
	result := new(big.Int).Mul(a, b)
	vm.stack = append(vm.stack, result)
	return nil
}

func (vm *VM) opDiv() error {
	if len(vm.stack) < 2 {
		return fmt.Errorf("stack underflow")
	}
	
	b := vm.stack[len(vm.stack)-1]
	a := vm.stack[len(vm.stack)-2]
	vm.stack = vm.stack[:len(vm.stack)-2]
	
	if b.Cmp(big.NewInt(0)) == 0 {
		return fmt.Errorf("division by zero")
	}
	
	result := new(big.Int).Div(a, b)
	vm.stack = append(vm.stack, result)
	return nil
}

// Операции со стеком
func (vm *VM) opPush(operand interface{}) error {
	var value *big.Int
	
	switch v := operand.(type) {
	case int:
		value = big.NewInt(int64(v))
	case int64:
		value = big.NewInt(v)
	case *big.Int:
		value = new(big.Int).Set(v)
	case string:
		// Парсим строку как число
		value = new(big.Int)
		value.SetString(v, 10)
	default:
		return fmt.Errorf("invalid operand type for PUSH: %T", operand)
	}
	
	vm.stack = append(vm.stack, value)
	return nil
}

func (vm *VM) opPop() error {
	if len(vm.stack) == 0 {
		return fmt.Errorf("stack underflow")
	}
	
	vm.stack = vm.stack[:len(vm.stack)-1]
	return nil
}

// Операции с памятью
func (vm *VM) opLoad(operand interface{}) error {
	key, ok := operand.(string)
	if !ok {
		return fmt.Errorf("invalid operand type for LOAD: %T", operand)
	}
	
	value, exists := vm.memory[key]
	if !exists {
		value = big.NewInt(0)
	}
	
	vm.stack = append(vm.stack, value)
	return nil
}

func (vm *VM) opStore(operand interface{}) error {
	if len(vm.stack) == 0 {
		return fmt.Errorf("stack underflow")
	}
	
	key, ok := operand.(string)
	if !ok {
		return fmt.Errorf("invalid operand type for STORE: %T", operand)
	}
	
	value := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	
	vm.memory[key] = new(big.Int).Set(value)
	return nil
}

// Операции с хранилищем контракта
func (vm *VM) opSLoad(operand interface{}, contract *Contract) error {
	key, ok := operand.(string)
	if !ok {
		return fmt.Errorf("invalid operand type for SLOAD: %T", operand)
	}
	
	value, exists := contract.Storage[key]
	if !exists {
		value = big.NewInt(0)
	}
	
	vm.stack = append(vm.stack, value)
	return nil
}

func (vm *VM) opSStore(operand interface{}, contract *Contract, result *ExecutionResult) error {
	if len(vm.stack) == 0 {
		return fmt.Errorf("stack underflow")
	}
	
	key, ok := operand.(string)
	if !ok {
		return fmt.Errorf("invalid operand type for SSTORE: %T", operand)
	}
	
	value := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	
	contract.Storage[key] = new(big.Int).Set(value)
	result.Storage[key] = new(big.Int).Set(value)
	return nil
}

// Управление потоком
func (vm *VM) opJump(operand interface{}) error {
	address, ok := operand.(int)
	if !ok {
		return fmt.Errorf("invalid operand type for JUMP: %T", operand)
	}
	
	vm.pc = address
	return nil
}

func (vm *VM) opJumpI(operand interface{}) error {
	if len(vm.stack) == 0 {
		return fmt.Errorf("stack underflow")
	}
	
	condition := vm.stack[len(vm.stack)-1]
	vm.stack = vm.stack[:len(vm.stack)-1]
	
	if condition.Cmp(big.NewInt(0)) != 0 {
		return vm.opJump(operand)
	}
	
	return nil
}

func (vm *VM) opReturn(operand interface{}, result *ExecutionResult) error {
	// Возвращаем данные со стека
	returnData := make([]byte, 0)
	for len(vm.stack) > 0 {
		value := vm.stack[len(vm.stack)-1]
		vm.stack = vm.stack[:len(vm.stack)-1]
		returnData = append(returnData, value.Bytes()...)
	}
	
	result.ReturnData = returnData
	return nil
}

func (vm *VM) opStop() error {
	vm.pc = len(vm.stack) // Завершаем выполнение
	return nil
}
