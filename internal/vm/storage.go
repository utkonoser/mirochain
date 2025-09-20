package vm

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/dgraph-io/badger/v4"
)

// ContractStorageManager управляет хранением состояния контрактов
type ContractStorageManager struct {
	db *badger.DB
}

// NewContractStorageManager создает новый менеджер хранения контрактов
func NewContractStorageManager(db *badger.DB) *ContractStorageManager {
	return &ContractStorageManager{
		db: db,
	}
}

// SaveContract сохраняет контракт в базу данных
func (csm *ContractStorageManager) SaveContract(contract *Contract) error {
	return csm.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("contract:%s", contract.Address))
		data, err := json.Marshal(contract)
		if err != nil {
			return fmt.Errorf("failed to marshal contract: %w", err)
		}

		return txn.Set(key, data)
	})
}

// GetContract получает контракт из базы данных
func (csm *ContractStorageManager) GetContract(address string) (*Contract, error) {
	var contract *Contract
	err := csm.db.View(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("contract:%s", address))
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			contract = &Contract{}
			return json.Unmarshal(val, contract)
		})
	})

	if err != nil {
		return nil, err
	}

	return contract, nil
}

// UpdateContractStorage обновляет хранилище контракта
func (csm *ContractStorageManager) UpdateContractStorage(address string, storage map[string]*big.Int) error {
	contract, err := csm.GetContract(address)
	if err != nil {
		return fmt.Errorf("failed to get contract: %w", err)
	}

	if contract == nil {
		return fmt.Errorf("contract not found: %s", address)
	}

	contract.Storage = storage
	contract.UpdatedAt = time.Now()

	return csm.SaveContract(contract)
}

// GetContractStorage получает хранилище контракта
func (csm *ContractStorageManager) GetContractStorage(address string) (map[string]*big.Int, error) {
	contract, err := csm.GetContract(address)
	if err != nil {
		return nil, err
	}

	if contract == nil {
		return nil, fmt.Errorf("contract not found: %s", address)
	}

	return contract.Storage, nil
}

// SetStorageValue устанавливает значение в хранилище контракта
func (csm *ContractStorageManager) SetStorageValue(address, key string, value *big.Int) error {
	contract, err := csm.GetContract(address)
	if err != nil {
		return fmt.Errorf("failed to get contract: %w", err)
	}

	if contract == nil {
		return fmt.Errorf("contract not found: %s", address)
	}

	if contract.Storage == nil {
		contract.Storage = make(map[string]*big.Int)
	}

	contract.Storage[key] = value
	contract.UpdatedAt = time.Now()

	return csm.SaveContract(contract)
}

// GetStorageValue получает значение из хранилища контракта
func (csm *ContractStorageManager) GetStorageValue(address, key string) (*big.Int, error) {
	contract, err := csm.GetContract(address)
	if err != nil {
		return nil, err
	}

	if contract == nil {
		return nil, fmt.Errorf("contract not found: %s", address)
	}

	if contract.Storage == nil {
		return big.NewInt(0), nil
	}

	value, exists := contract.Storage[key]
	if !exists {
		return big.NewInt(0), nil
	}

	return value, nil
}

// ListContracts возвращает список всех контрактов
func (csm *ContractStorageManager) ListContracts() ([]*Contract, error) {
	var contracts []*Contract

	err := csm.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("contract:")

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var contract Contract
				if err := json.Unmarshal(val, &contract); err != nil {
					return err
				}
				contracts = append(contracts, &contract)
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return contracts, err
}

// DeleteContract удаляет контракт из базы данных
func (csm *ContractStorageManager) DeleteContract(address string) error {
	return csm.db.Update(func(txn *badger.Txn) error {
		key := []byte(fmt.Sprintf("contract:%s", address))
		return txn.Delete(key)
	})
}

// GetContractCount возвращает количество контрактов
func (csm *ContractStorageManager) GetContractCount() (int, error) {
	count := 0

	err := csm.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte("contract:")

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}

		return nil
	})

	return count, err
}

// GetStorageStats возвращает статистику хранилища контрактов
func (csm *ContractStorageManager) GetStorageStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	contracts, err := csm.ListContracts()
	if err != nil {
		return nil, err
	}

	stats["total_contracts"] = len(contracts)

	totalStorageKeys := 0
	for _, contract := range contracts {
		if contract.Storage != nil {
			totalStorageKeys += len(contract.Storage)
		}
	}

	stats["total_storage_keys"] = totalStorageKeys
	stats["average_storage_keys_per_contract"] = float64(totalStorageKeys) / float64(len(contracts))

	return stats, nil
}

// Close закрывает соединение с базой данных
func (csm *ContractStorageManager) Close() error {
	return csm.db.Close()
}
