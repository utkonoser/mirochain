//go:build simple_contracts_demo

package main

import (
	"fmt"
	"log"
	"math/big"

	"mirochain/internal/vm"
)

func main() {
	fmt.Println("=== Simple Smart Contracts Demo ===")
	fmt.Println()

	// Создаем виртуальную машину
	vmInstance := vm.NewVM(1000000) // 1M газа
	fmt.Printf("VM created with gas limit: %d\n", vmInstance.GetGasRemaining())
	fmt.Println()

	// Демонстрация простого контракта
	demonstrateSimpleContract(vmInstance)
}

func demonstrateSimpleContract(vmInstance *vm.VM) {
	// Создаем простой контракт
	simpleCode := `// Simple Contract
PUSH 42
STORE "value"
LOAD "value"
RETURN`

	fmt.Printf("Contract code:\n%s\n\n", simpleCode)

	// Компилируем код
	compiler := vm.NewCompiler()
	instructions, err := compiler.Compile(simpleCode)
	if err != nil {
		log.Printf("Compilation error: %v", err)
		return
	}

	fmt.Printf("Compiled to %d instructions\n", len(instructions))

	// Развертываем контракт
	contract, err := vmInstance.DeployContract(instructions, "user1", big.NewInt(1000))
	if err != nil {
		log.Printf("Deployment error: %v", err)
		return
	}

	fmt.Printf("Contract deployed successfully!\n")
	fmt.Printf("Address: %s\n", contract.Address)
	fmt.Printf("Owner: %s\n", contract.Owner)
	fmt.Printf("Balance: %s\n", contract.Balance.String())
	fmt.Println()

	// Выполняем контракт
	result, err := vmInstance.ExecuteContract(contract.Address, []byte("execute"))
	if err != nil {
		log.Printf("Execution error: %v", err)
		return
	}

	fmt.Printf("Contract execution completed!\n")
	fmt.Printf("Success: %t\n", result.Success)
	fmt.Printf("Gas used: %d\n", result.GasUsed)
	fmt.Printf("Return data: %s\n", string(result.ReturnData))
	
	if result.Error != "" {
		fmt.Printf("Error: %s\n", result.Error)
	}
	
	fmt.Printf("Contract storage: %v\n", result.Storage)
	fmt.Println()

	// Демонстрация газа
	fmt.Println("=== Gas System Demo ===")
	gasTracker := vm.NewGasTracker(1000, 1) // 1000 газа, цена 1
	fmt.Printf("Gas tracker created: Limit=%d, Price=%d\n", 1000, 1)
	
	// Оценка газа
	estimator := vm.NewGasEstimator()
	gasEstimate := estimator.EstimateGas(instructions)
	fmt.Printf("Gas estimate for contract: %d\n", gasEstimate)
	fmt.Printf("Gas tracker remaining: %d\n", gasTracker.GetGasRemaining())
	
	// Демонстрация шаблонов
	fmt.Println()
	fmt.Println("=== Contract Templates Demo ===")
	templates := vm.GetContractTemplates()
	fmt.Printf("Available templates: %d\n", len(templates))
	
	for i, template := range templates {
		fmt.Printf("%d. %s - %s (Gas: %d)\n", i+1, template.Name, template.Description, template.GasEstimate)
	}
	
	// Компилируем шаблон
	if len(templates) > 0 {
		fmt.Printf("\nCompiling template: %s\n", templates[0].Name)
		templateInstructions, err := vm.CompileTemplate(templates[0].Name)
		if err != nil {
			log.Printf("Template compilation error: %v", err)
		} else {
			fmt.Printf("Template compiled to %d instructions\n", len(templateInstructions))
		}
	}
	
	fmt.Println()
	fmt.Println("Smart contracts demo completed!")
}
