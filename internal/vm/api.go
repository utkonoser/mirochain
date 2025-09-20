package vm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

// ContractAPI предоставляет HTTP API для работы с контрактами
type ContractAPI struct {
	vm             *VM
	storageManager *ContractStorageManager
}

// NewContractAPI создает новый API для контрактов
func NewContractAPI(vm *VM) *ContractAPI {
	return &ContractAPI{vm: vm, storageManager: vm.storageManager}
}

// NewContractAPIWithStorage создает новый API для контрактов с системой хранения
func NewContractAPIWithStorage(vm *VM, storageManager *ContractStorageManager) *ContractAPI {
	return &ContractAPI{vm: vm, storageManager: storageManager}
}

// DeployRequest представляет запрос на развертывание контракта
type DeployRequest struct {
	Code           string `json:"code"`
	Owner          string `json:"owner"`
	InitialBalance string `json:"initial_balance"`
	GasLimit       uint64 `json:"gas_limit"`
}

// DeployResponse представляет ответ на развертывание контракта
type DeployResponse struct {
	Success         bool   `json:"success"`
	ContractAddress string `json:"contract_address,omitempty"`
	GasUsed         uint64 `json:"gas_used,omitempty"`
	Error           string `json:"error,omitempty"`
}

// CallRequest представляет запрос на вызов контракта
type CallRequest struct {
	ContractAddress string        `json:"contract_address"`
	Function        string        `json:"function"`
	Args            []interface{} `json:"args"`
	GasLimit        uint64        `json:"gas_limit"`
}

// CallResponse представляет ответ на вызов контракта
type CallResponse struct {
	Success    bool        `json:"success"`
	ReturnData interface{} `json:"return_data,omitempty"`
	GasUsed    uint64      `json:"gas_used,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// ContractInfo представляет информацию о контракте
type ContractInfo struct {
	Address   string            `json:"address"`
	Owner     string            `json:"owner"`
	Balance   string            `json:"balance"`
	Storage   map[string]string `json:"storage"`
	CreatedAt time.Time         `json:"created_at"`
	CodeSize  int               `json:"code_size"`
}

// DeployContract развертывает контракт
func (api *ContractAPI) DeployContract(w http.ResponseWriter, r *http.Request) {
	var req DeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Парсим начальный баланс
	initialBalance, err := strconv.ParseInt(req.InitialBalance, 10, 64)
	if err != nil {
		http.Error(w, "Invalid initial balance", http.StatusBadRequest)
		return
	}

	// Компилируем код контракта
	compiler := NewCompiler()
	instructions, err := compiler.Compile(req.Code)
	if err != nil {
		response := DeployResponse{
			Success: false,
			Error:   fmt.Sprintf("Compilation error: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Развертываем контракт
	contract, err := api.vm.DeployContract(instructions, req.Owner, big.NewInt(initialBalance))
	if err != nil {
		response := DeployResponse{
			Success: false,
			Error:   fmt.Sprintf("Deployment error: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := DeployResponse{
		Success:         true,
		ContractAddress: contract.Address,
		GasUsed:         api.vm.GetGasUsed(),
	}

	json.NewEncoder(w).Encode(response)
}

// CallContract вызывает контракт
func (api *ContractAPI) CallContract(w http.ResponseWriter, r *http.Request) {
	var req CallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Вызываем контракт
	result, err := api.vm.CallContract(req.ContractAddress, req.Function, req.Args)
	if err != nil {
		response := CallResponse{
			Success: false,
			Error:   fmt.Sprintf("Call error: %v", err),
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CallResponse{
		Success:    result.Success,
		ReturnData: result.ReturnData,
		GasUsed:    result.GasUsed,
		Error:      result.Error,
	}

	json.NewEncoder(w).Encode(response)
}

// GetContract возвращает информацию о контракте
func (api *ContractAPI) GetContract(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address parameter required", http.StatusBadRequest)
		return
	}

	contract, exists := api.vm.GetContract(address)
	if !exists {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}

	// Конвертируем storage в строки для JSON
	storage := make(map[string]string)
	for k, v := range contract.Storage {
		storage[k] = v.String()
	}

	info := ContractInfo{
		Address:   contract.Address,
		Owner:     contract.Owner,
		Balance:   contract.Balance.String(),
		Storage:   storage,
		CreatedAt: contract.CreatedAt,
		CodeSize:  len(contract.Code),
	}

	json.NewEncoder(w).Encode(info)
}

// ListContracts возвращает список всех контрактов
func (api *ContractAPI) ListContracts(w http.ResponseWriter, r *http.Request) {
	contracts := api.vm.ListContracts()

	infos := make([]ContractInfo, 0, len(contracts))
	for _, contract := range contracts {
		// Конвертируем storage в строки для JSON
		storage := make(map[string]string)
		for k, v := range contract.Storage {
			storage[k] = v.String()
		}

		info := ContractInfo{
			Address:   contract.Address,
			Owner:     contract.Owner,
			Balance:   contract.Balance.String(),
			Storage:   storage,
			CreatedAt: contract.CreatedAt,
			CodeSize:  len(contract.Code),
		}

		infos = append(infos, info)
	}

	json.NewEncoder(w).Encode(infos)
}

// GetTemplates возвращает шаблоны контрактов
func (api *ContractAPI) GetTemplates(w http.ResponseWriter, r *http.Request) {
	templates := GetContractTemplates()
	json.NewEncoder(w).Encode(templates)
}

// CompileTemplate компилирует шаблон контракта
func (api *ContractAPI) CompileTemplate(w http.ResponseWriter, r *http.Request) {
	templateName := r.URL.Query().Get("template")
	if templateName == "" {
		http.Error(w, "Template parameter required", http.StatusBadRequest)
		return
	}

	instructions, err := CompileTemplate(templateName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template compilation error: %v", err), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(instructions)
}

// EstimateGas оценивает стоимость газа для контракта
func (api *ContractAPI) EstimateGas(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Code string `json:"code"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	estimator := NewGasEstimator()
	gasEstimate, err := estimator.EstimateContractGas(req.Code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Gas estimation error: %v", err), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"gas_estimate": gasEstimate,
		"success":      true,
	}

	json.NewEncoder(w).Encode(response)
}

// GetGasReport возвращает отчет об использовании газа
func (api *ContractAPI) GetGasReport(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Address parameter required", http.StatusBadRequest)
		return
	}

	contract, exists := api.vm.GetContract(address)
	if !exists {
		http.Error(w, "Contract not found", http.StatusNotFound)
		return
	}

	// Компилируем код для анализа
	compiler := NewCompiler()
	instructions, err := compiler.Compile(string(contract.Code))
	if err != nil {
		http.Error(w, fmt.Sprintf("Compilation error: %v", err), http.StatusBadRequest)
		return
	}

	estimator := NewGasEstimator()
	gasPrice := uint64(1) // Базовая цена газа
	report := estimator.GenerateGasReport(instructions, gasPrice)

	json.NewEncoder(w).Encode(report)
}

// RegisterRoutes регистрирует маршруты API
func (api *ContractAPI) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/contracts/deploy", api.DeployContract)
	mux.HandleFunc("/api/contracts/call", api.CallContract)
	mux.HandleFunc("/api/contracts/get", api.GetContract)
	mux.HandleFunc("/api/contracts/list", api.ListContracts)
	mux.HandleFunc("/api/contracts/templates", api.GetTemplates)
	mux.HandleFunc("/api/contracts/compile", api.CompileTemplate)
	mux.HandleFunc("/api/contracts/estimate-gas", api.EstimateGas)
	mux.HandleFunc("/api/contracts/gas-report", api.GetGasReport)

	// Новые endpoints для работы с хранилищем контрактов
	mux.HandleFunc("/api/contracts/storage/", api.GetContractStorage)
	mux.HandleFunc("/api/contracts/storage/set", api.SetStorageValue)
	mux.HandleFunc("/api/contracts/storage/get", api.GetStorageValue)
	mux.HandleFunc("/api/contracts/stats", api.GetContractStats)
}

// GetContractStorage возвращает хранилище контракта
func (api *ContractAPI) GetContractStorage(w http.ResponseWriter, r *http.Request) {
	if api.storageManager == nil {
		http.Error(w, "Storage manager not available", http.StatusServiceUnavailable)
		return
	}

	address := r.URL.Path[len("/api/contracts/storage/"):]
	if address == "" {
		http.Error(w, "Contract address required", http.StatusBadRequest)
		return
	}

	storage, err := api.storageManager.GetContractStorage(address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": address,
		"storage": storage,
	})
}

// SetStorageValue устанавливает значение в хранилище контракта
func (api *ContractAPI) SetStorageValue(w http.ResponseWriter, r *http.Request) {
	if api.storageManager == nil {
		http.Error(w, "Storage manager not available", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		Address string `json:"address"`
		Key     string `json:"key"`
		Value   string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	value, ok := new(big.Int).SetString(req.Value, 10)
	if !ok {
		http.Error(w, "Invalid value format", http.StatusBadRequest)
		return
	}

	err := api.storageManager.SetStorageValue(req.Address, req.Key, value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"address": req.Address,
		"key":     req.Key,
		"value":   req.Value,
	})
}

// GetStorageValue получает значение из хранилища контракта
func (api *ContractAPI) GetStorageValue(w http.ResponseWriter, r *http.Request) {
	if api.storageManager == nil {
		http.Error(w, "Storage manager not available", http.StatusServiceUnavailable)
		return
	}

	address := r.URL.Query().Get("address")
	key := r.URL.Query().Get("key")

	if address == "" || key == "" {
		http.Error(w, "Address and key required", http.StatusBadRequest)
		return
	}

	value, err := api.storageManager.GetStorageValue(address, key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": address,
		"key":     key,
		"value":   value.String(),
	})
}

// GetContractStats возвращает статистику контрактов
func (api *ContractAPI) GetContractStats(w http.ResponseWriter, r *http.Request) {
	if api.storageManager == nil {
		http.Error(w, "Storage manager not available", http.StatusServiceUnavailable)
		return
	}

	stats, err := api.storageManager.GetStorageStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}
