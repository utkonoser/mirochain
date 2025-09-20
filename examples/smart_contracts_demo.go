//go:build contracts_demo

package main

import (
	"fmt"
	"log"
	"math/big"

	"mirochain/internal/vm"
)

func main() {
	fmt.Println("=== Smart Contracts Demo ===")
	fmt.Println()

	// Создаем виртуальную машину
	vm := vm.NewVM(1000000) // 1M газа
	fmt.Printf("VM created with gas limit: %d\n", vm.GetGasRemaining())
	fmt.Println()

	// 1. Демонстрация простого счетчика
	fmt.Println("1. Simple Counter Contract Demo:")
	demonstrateCounterContract(vm)
	fmt.Println()

	// 2. Демонстрация токена
	fmt.Println("2. Token Contract Demo:")
	demonstrateTokenContract(vm)
	fmt.Println()

	// 3. Демонстрация калькулятора
	fmt.Println("3. Calculator Contract Demo:")
	demonstrateCalculatorContract(vm)
	fmt.Println()

	// 4. Демонстрация системы газа
	fmt.Println("4. Gas System Demo:")
	demonstrateGasSystem(vm)
	fmt.Println()

	// 5. Демонстрация шаблонов контрактов
	fmt.Println("5. Contract Templates Demo:")
	demonstrateContractTemplates(vm)
	fmt.Println()

	// 6. Демонстрация оптимизации газа
	fmt.Println("6. Gas Optimization Demo:")
	demonstrateGasOptimization(vm)
	fmt.Println()
}

func demonstrateCounterContract(vmInstance *vm.VM) {
	// Создаем простой счетчик
	counterCode := `// Simple Counter Contract
PUSH 0
STORE "counter"

// Функция: увеличить счетчик
PUSH 1
LOAD "counter"
ADD
STORE "counter"
RETURN

// Функция: получить значение счетчика
LOAD "counter"
RETURN`

	// Компилируем код
	compiler := vm.NewCompiler()
	instructions, err := compiler.Compile(counterCode)
	if err != nil {
		log.Printf("Compilation error: %v", err)
		return
	}

	// Развертываем контракт
	contract, err := vmInstance.DeployContract(instructions, "user1", big.NewInt(1000))
	if err != nil {
		log.Printf("Deployment error: %v", err)
		return
	}

	fmt.Printf("Counter contract deployed at: %s\n", contract.Address)
	fmt.Printf("Owner: %s\n", contract.Owner)
	fmt.Printf("Initial balance: %s\n", contract.Balance.String())

	// Выполняем контракт несколько раз
	for i := 0; i < 5; i++ {
		result, err := vmInstance.ExecuteContract(contract.Address, []byte("increment"))
		if err != nil {
			log.Printf("Execution error: %v", err)
			continue
		}

		fmt.Printf("Execution %d: Success=%t, Gas used=%d\n", i+1, result.Success, result.GasUsed)
	}

	// Получаем финальное значение
	result, err := vmInstance.ExecuteContract(contract.Address, []byte("get_value"))
	if err != nil {
		log.Printf("Execution error: %v", err)
		return
	}

	fmt.Printf("Final counter value: %s\n", string(result.ReturnData))
}

func demonstrateTokenContract(vm *vm.VM) {
	// Создаем простой токен
	tokenCode := `// Token Contract
PUSH 1000
STORE "total_supply"

PUSH 1000
STORE "balance"

// Функция: получить баланс
LOAD "balance"
RETURN

// Функция: передать токены
LOAD "amount"
LOAD "balance"
SUB
STORE "balance"
RETURN`

	// Компилируем код
	compiler := vm.NewCompiler()
	instructions, err := compiler.Compile(tokenCode)
	if err != nil {
		log.Printf("Compilation error: %v", err)
		return
	}

	// Развертываем контракт
	contract, err := vm.DeployContract(instructions, "user2", big.NewInt(2000))
	if err != nil {
		log.Printf("Deployment error: %v", err)
		return
	}

	fmt.Printf("Token contract deployed at: %s\n", contract.Address)
	fmt.Printf("Owner: %s\n", contract.Owner)

	// Получаем баланс
	result, err := vmInstance.ExecuteContract(contract.Address, []byte("get_balance"))
	if err != nil {
		log.Printf("Execution error: %v", err)
		return
	}

	fmt.Printf("Token balance: %s\n", string(result.ReturnData))
}

func demonstrateCalculatorContract(vm *vm.VM) {
	// Создаем калькулятор
	calculatorCode := `// Calculator Contract
// Функция: сложение
PUSH 10
STORE "a"
PUSH 20
STORE "b"

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
RETURN`

	// Компилируем код
	compiler := vm.NewCompiler()
	instructions, err := compiler.Compile(calculatorCode)
	if err != nil {
		log.Printf("Compilation error: %v", err)
		return
	}

	// Развертываем контракт
	contract, err := vm.DeployContract(instructions, "user3", big.NewInt(500))
	if err != nil {
		log.Printf("Deployment error: %v", err)
		return
	}

	fmt.Printf("Calculator contract deployed at: %s\n", contract.Address)

	// Выполняем сложение
	result, err := vmInstance.ExecuteContract(contract.Address, []byte("add"))
	if err != nil {
		log.Printf("Execution error: %v", err)
		return
	}

	fmt.Printf("Addition result: %s\n", string(result.ReturnData))

	// Выполняем умножение
	result, err = vmInstance.ExecuteContract(contract.Address, []byte("multiply"))
	if err != nil {
		log.Printf("Execution error: %v", err)
		return
	}

	fmt.Printf("Multiplication result: %s\n", string(result.ReturnData))
}

func demonstrateGasSystem(vm *vm.VM) {
	// Создаем контракт для демонстрации газа
	gasCode := `// Gas Demo Contract
PUSH 1
STORE "value"

LOAD "value"
ADD
STORE "value"
RETURN`

	// Компилируем код
	compiler := vm.NewCompiler()
	instructions, err := compiler.Compile(gasCode)
	if err != nil {
		log.Printf("Compilation error: %v", err)
		return
	}

	// Развертываем контракт
	contract, err := vm.DeployContract(instructions, "user4", big.NewInt(100))
	if err != nil {
		log.Printf("Deployment error: %v", err)
		return
	}

	fmt.Printf("Gas demo contract deployed at: %s\n", contract.Address)

	// Создаем трекер газа
	gasTracker := vm.NewGasTracker(1000, 1) // 1000 газа, цена 1

	// Выполняем контракт
	result, err := vmInstance.ExecuteContract(contract.Address, []byte("execute"))
	if err != nil {
		log.Printf("Execution error: %v", err)
		return
	}

	fmt.Printf("Execution result: Success=%t, Gas used=%d\n", result.Success, result.GasUsed)
	fmt.Printf("Gas tracker: Used=%d, Remaining=%d, Total cost=%d\n", 
		gasTracker.GetGasUsed(), gasTracker.GetGasRemaining(), gasTracker.GetTotalCost())

	// Демонстрируем оценку газа
	estimator := vm.NewGasEstimator()
	gasEstimate, err := estimator.EstimateContractGas(gasCode)
	if err != nil {
		log.Printf("Gas estimation error: %v", err)
		return
	}

	fmt.Printf("Gas estimate: %d\n", gasEstimate)
}

func demonstrateContractTemplates(vm *vm.VM) {
	// Получаем шаблоны контрактов
	templates := vm.GetContractTemplates()
	
	fmt.Printf("Available contract templates (%d):\n", len(templates))
	for i, template := range templates {
		fmt.Printf("%d. %s\n", i+1, template.Name)
		fmt.Printf("   Description: %s\n", template.Description)
		fmt.Printf("   Gas estimate: %d\n", template.GasEstimate)
		fmt.Println()
	}

	// Компилируем шаблон "Simple Counter"
	instructions, err := vm.CompileTemplate("Simple Counter")
	if err != nil {
		log.Printf("Template compilation error: %v", err)
		return
	}

	fmt.Printf("Simple Counter template compiled: %d instructions\n", len(instructions))

	// Развертываем контракт из шаблона
	contract, err := vm.DeployContract(instructions, "template_user", big.NewInt(500))
	if err != nil {
		log.Printf("Template deployment error: %v", err)
		return
	}

	fmt.Printf("Template contract deployed at: %s\n", contract.Address)
}

func demonstrateGasOptimization(vm *vm.VM) {
	// Создаем неоптимизированный код
	unoptimizedCode := `// Unoptimized code
PUSH 0
PUSH 1
ADD
PUSH 1
MUL
STORE "result"
RETURN`

	// Компилируем код
	compiler := vm.NewCompiler()
	instructions, err := compiler.Compile(unoptimizedCode)
	if err != nil {
		log.Printf("Compilation error: %v", err)
		return
	}

	fmt.Printf("Unoptimized code: %d instructions\n", len(instructions))

	// Оцениваем газ
	estimator := vm.NewGasEstimator()
	unoptimizedGas := estimator.EstimateGas(instructions)
	fmt.Printf("Unoptimized gas estimate: %d\n", unoptimizedGas)

	// Оптимизируем код
	optimizer := vm.NewGasOptimizer()
	optimizedInstructions := optimizer.OptimizeInstructions(instructions)
	
	fmt.Printf("Optimized code: %d instructions\n", len(optimizedInstructions))

	// Оцениваем оптимизированный газ
	optimizedGas := estimator.EstimateGas(optimizedInstructions)
	fmt.Printf("Optimized gas estimate: %d\n", optimizedGas)

	// Показываем экономию
	gasSaved := unoptimizedGas - optimizedGas
	percentSaved := float64(gasSaved) / float64(unoptimizedGas) * 100
	
	fmt.Printf("Gas saved: %d (%.2f%%)\n", gasSaved, percentSaved)
}

// Дополнительные функции для демонстрации
func demonstrateContractInteraction(vm *vm.VM) {
	fmt.Println("=== Contract Interaction Demo ===")
	
	// Создаем контракт A
	compiler := vm.NewCompiler()
	instructionsA, _ := compiler.Compile("PUSH 1\nRETURN")
	contractA, err := vm.DeployContract(instructionsA, "user1", big.NewInt(1000))
	if err != nil {
		log.Printf("Error deploying contract A: %v", err)
		return
	}
	
	// Создаем контракт B
	instructionsB, _ := compiler.Compile("PUSH 2\nRETURN")
	contractB, err := vm.DeployContract(instructionsB, "user2", big.NewInt(2000))
	if err != nil {
		log.Printf("Error deploying contract B: %v", err)
		return
	}
	
	fmt.Printf("Contract A: %s\n", contractA.Address)
	fmt.Printf("Contract B: %s\n", contractB.Address)
	
	// Демонстрируем вызов между контрактами
	result, err := vm.CallContract(contractA.Address, "call_contract_b", []interface{}{contractB.Address})
	if err != nil {
		log.Printf("Contract call error: %v", err)
		return
	}
	
	fmt.Printf("Contract interaction result: %t\n", result.Success)
}

func demonstrateContractStorage(vm *vm.VM) {
	fmt.Println("=== Contract Storage Demo ===")
	
	// Создаем контракт с хранилищем
	storageCode := `// Storage Contract
PUSH 42
SSTORE "important_data"

SLOAD "important_data"
RETURN`

	// Компилируем код
	compiler := vm.NewCompiler()
	instructions, err := compiler.Compile(storageCode)
	if err != nil {
		log.Printf("Compilation error: %v", err)
		return
	}

	// Развертываем контракт
	contract, err := vm.DeployContract(instructions, "storage_user", big.NewInt(1000))
	if err != nil {
		log.Printf("Deployment error: %v", err)
		return
	}

	fmt.Printf("Storage contract deployed at: %s\n", contract.Address)

	// Выполняем контракт
	result, err := vmInstance.ExecuteContract(contract.Address, []byte("store_and_retrieve"))
	if err != nil {
		log.Printf("Execution error: %v", err)
		return
	}

	fmt.Printf("Storage result: %s\n", string(result.ReturnData))
	fmt.Printf("Contract storage: %v\n", contract.Storage)
}
