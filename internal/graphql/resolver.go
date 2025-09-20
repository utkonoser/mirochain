package graphql

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/graphql-go/graphql"
	"mirochain/internal/consensus"
	"mirochain/internal/crypto"
	"mirochain/internal/nft"
	"mirochain/internal/network"
	"mirochain/internal/sidechain"
	"mirochain/internal/statechannel"
	"mirochain/internal/tokens"
	"mirochain/internal/vm"
	"mirochain/internal/wallet"
)

// GraphQLResolver содержит все зависимости для GraphQL API
type GraphQLResolver struct {
	blockchain      interface{} // Используем interface{} для совместимости
	walletManager   *wallet.WalletManager
	vm              *vm.VM
	tokenManager    *tokens.ERC20Manager
	nftManager      *nft.ERC721Manager
	sidechainManager *sidechain.SidechainManager
	stateChannelManager *statechannel.StateChannelManager
	consensusManager *consensus.ConsensusComparison
	signatureManager *crypto.SignatureManager
	networkManager  *network.Server
}

// NewGraphQLResolver создает новый GraphQL resolver
func NewGraphQLResolver(
	blockchain interface{},
	walletManager *wallet.WalletManager,
	vm *vm.VM,
	tokenManager *tokens.ERC20Manager,
	nftManager *nft.ERC721Manager,
	sidechainManager *sidechain.SidechainManager,
	stateChannelManager *statechannel.StateChannelManager,
	consensusManager *consensus.ConsensusComparison,
	signatureManager *crypto.SignatureManager,
	networkManager *network.Server,
) *GraphQLResolver {
	return &GraphQLResolver{
		blockchain:      blockchain,
		walletManager:   walletManager,
		vm:              vm,
		tokenManager:    tokenManager,
		nftManager:      nftManager,
		sidechainManager: sidechainManager,
		stateChannelManager: stateChannelManager,
		consensusManager: consensusManager,
		signatureManager: signatureManager,
		networkManager:  networkManager,
	}
}

// CreateSchema создает GraphQL схему
func (r *GraphQLResolver) CreateSchema() (graphql.Schema, error) {
	// Определяем типы
	blockType := r.createBlockType()
	transactionType := r.createTransactionType()
	walletType := r.createWalletType()
	contractType := r.createContractType()
	tokenType := r.createTokenType()
	peerType := r.createPeerType()
	networkStatsType := r.createNetworkStatsType()
	consensusInfoType := r.createConsensusInfoType()

	// Создаем корневые типы
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"blockchain": &graphql.Field{
				Type: r.createBlockchainType(),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return r.getBlockchain(p.Context)
				},
			},
			"block": &graphql.Field{
				Type: blockType,
				Args: graphql.FieldConfigArgument{
					"height": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					height := p.Args["height"].(int)
					return r.getBlock(p.Context, height)
				},
			},
			"transaction": &graphql.Field{
				Type: transactionType,
				Args: graphql.FieldConfigArgument{
					"hash": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					hash := p.Args["hash"].(string)
					return r.getTransaction(p.Context, hash)
				},
			},
			"wallet": &graphql.Field{
				Type: walletType,
				Args: graphql.FieldConfigArgument{
					"address": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					address := p.Args["address"].(string)
					return r.getWallet(p.Context, address)
				},
			},
			"contract": &graphql.Field{
				Type: contractType,
				Args: graphql.FieldConfigArgument{
					"address": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					address := p.Args["address"].(string)
					return r.getContract(p.Context, address)
				},
			},
			"token": &graphql.Field{
				Type: tokenType,
				Args: graphql.FieldConfigArgument{
					"address": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					address := p.Args["address"].(string)
					return r.getToken(p.Context, address)
				},
			},
			"peers": &graphql.Field{
				Type: graphql.NewList(peerType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return r.getPeers(p.Context)
				},
			},
			"networkStats": &graphql.Field{
				Type: networkStatsType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return r.getNetworkStats(p.Context)
				},
			},
			"consensus": &graphql.Field{
				Type: consensusInfoType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return r.getConsensusInfo(p.Context)
				},
			},
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"sendTransaction": &graphql.Field{
				Type: r.createTransactionResultType(),
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(r.createSendTransactionInputType()),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input := p.Args["input"].(map[string]interface{})
					return r.sendTransaction(p.Context, input)
				},
			},
			"deployContract": &graphql.Field{
				Type: r.createContractResultType(),
				Args: graphql.FieldConfigArgument{
					"input": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(r.createDeployContractInputType()),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					input := p.Args["input"].(map[string]interface{})
					return r.deployContract(p.Context, input)
				},
			},
		},
	})

	// Создаем схему
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})

	return schema, err
}

// Helper methods для создания типов
func (r *GraphQLResolver) createBlockchainType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Blockchain",
		Fields: graphql.Fields{
			"height": &graphql.Field{Type: graphql.Int},
			"difficulty": &graphql.Field{Type: graphql.Int},
			"totalSupply": &graphql.Field{Type: graphql.String},
			"hashRate": &graphql.Field{Type: graphql.Float},
			"lastBlock": &graphql.Field{Type: r.createBlockType()},
			"stats": &graphql.Field{Type: r.createBlockchainStatsType()},
		},
	})
}

func (r *GraphQLResolver) createBlockType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Block",
		Fields: graphql.Fields{
			"hash": &graphql.Field{Type: graphql.String},
			"height": &graphql.Field{Type: graphql.Int},
			"timestamp": &graphql.Field{Type: graphql.String},
			"previousHash": &graphql.Field{Type: graphql.String},
			"merkleRoot": &graphql.Field{Type: graphql.String},
			"nonce": &graphql.Field{Type: graphql.Int},
			"difficulty": &graphql.Field{Type: graphql.Int},
			"transactions": &graphql.Field{Type: graphql.NewList(r.createTransactionType())},
			"size": &graphql.Field{Type: graphql.Int},
			"gasUsed": &graphql.Field{Type: graphql.Int},
			"gasLimit": &graphql.Field{Type: graphql.Int},
		},
	})
}

func (r *GraphQLResolver) createTransactionType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Transaction",
		Fields: graphql.Fields{
			"hash": &graphql.Field{Type: graphql.String},
			"from": &graphql.Field{Type: graphql.String},
			"to": &graphql.Field{Type: graphql.String},
			"amount": &graphql.Field{Type: graphql.String},
			"fee": &graphql.Field{Type: graphql.String},
			"timestamp": &graphql.Field{Type: graphql.String},
			"blockHeight": &graphql.Field{Type: graphql.Int},
			"status": &graphql.Field{Type: graphql.String},
			"gasUsed": &graphql.Field{Type: graphql.Int},
			"gasPrice": &graphql.Field{Type: graphql.String},
			"data": &graphql.Field{Type: graphql.String},
			"signature": &graphql.Field{Type: graphql.String},
		},
	})
}

func (r *GraphQLResolver) createWalletType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Wallet",
		Fields: graphql.Fields{
			"address": &graphql.Field{Type: graphql.String},
			"balance": &graphql.Field{Type: graphql.String},
			"nonce": &graphql.Field{Type: graphql.Int},
			"transactions": &graphql.Field{Type: graphql.NewList(r.createTransactionType())},
			"createdAt": &graphql.Field{Type: graphql.String},
		},
	})
}

func (r *GraphQLResolver) createContractType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Contract",
		Fields: graphql.Fields{
			"address": &graphql.Field{Type: graphql.String},
			"code": &graphql.Field{Type: graphql.String},
			"owner": &graphql.Field{Type: graphql.String},
			"balance": &graphql.Field{Type: graphql.String},
			"createdAt": &graphql.Field{Type: graphql.String},
			"updatedAt": &graphql.Field{Type: graphql.String},
			"storage": &graphql.Field{Type: r.createContractStorageType()},
		},
	})
}

func (r *GraphQLResolver) createTokenType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Token",
		Fields: graphql.Fields{
			"address": &graphql.Field{Type: graphql.String},
			"name": &graphql.Field{Type: graphql.String},
			"symbol": &graphql.Field{Type: graphql.String},
			"decimals": &graphql.Field{Type: graphql.Int},
			"totalSupply": &graphql.Field{Type: graphql.String},
			"owner": &graphql.Field{Type: graphql.String},
			"createdAt": &graphql.Field{Type: graphql.String},
		},
	})
}

func (r *GraphQLResolver) createNFTType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "NFT",
		Fields: graphql.Fields{
			"contractAddress": &graphql.Field{Type: graphql.String},
			"tokenId": &graphql.Field{Type: graphql.String},
			"owner": &graphql.Field{Type: graphql.String},
			"metadata": &graphql.Field{Type: graphql.String},
			"createdAt": &graphql.Field{Type: graphql.String},
		},
	})
}

func (r *GraphQLResolver) createSidechainType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Sidechain",
		Fields: graphql.Fields{
			"id": &graphql.Field{Type: graphql.String},
			"name": &graphql.Field{Type: graphql.String},
			"consensus": &graphql.Field{Type: graphql.String},
			"validators": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"assets": &graphql.Field{Type: graphql.NewList(r.createSidechainAssetType())},
			"createdAt": &graphql.Field{Type: graphql.String},
		},
	})
}

func (r *GraphQLResolver) createStateChannelType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "StateChannel",
		Fields: graphql.Fields{
			"id": &graphql.Field{Type: graphql.String},
			"participants": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"balance": &graphql.Field{Type: graphql.String},
			"status": &graphql.Field{Type: graphql.String},
			"createdAt": &graphql.Field{Type: graphql.String},
			"updatedAt": &graphql.Field{Type: graphql.String},
		},
	})
}

func (r *GraphQLResolver) createPeerType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "Peer",
		Fields: graphql.Fields{
			"id": &graphql.Field{Type: graphql.String},
			"address": &graphql.Field{Type: graphql.String},
			"port": &graphql.Field{Type: graphql.Int},
			"lastSeen": &graphql.Field{Type: graphql.String},
			"latency": &graphql.Field{Type: graphql.Float},
		},
	})
}

func (r *GraphQLResolver) createNetworkStatsType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "NetworkStats",
		Fields: graphql.Fields{
			"totalPeers": &graphql.Field{Type: graphql.Int},
			"connectedPeers": &graphql.Field{Type: graphql.Int},
			"averageLatency": &graphql.Field{Type: graphql.Float},
			"bandwidth": &graphql.Field{Type: graphql.Float},
		},
	})
}

func (r *GraphQLResolver) createConsensusInfoType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "ConsensusInfo",
		Fields: graphql.Fields{
			"algorithm": &graphql.Field{Type: graphql.String},
			"validators": &graphql.Field{Type: graphql.NewList(graphql.String)},
			"currentRound": &graphql.Field{Type: graphql.Int},
			"nextValidator": &graphql.Field{Type: graphql.String},
		},
	})
}

// Дополнительные типы
func (r *GraphQLResolver) createBlockchainStatsType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "BlockchainStats",
		Fields: graphql.Fields{
			"totalBlocks": &graphql.Field{Type: graphql.Int},
			"totalTransactions": &graphql.Field{Type: graphql.Int},
			"totalAddresses": &graphql.Field{Type: graphql.Int},
			"averageBlockTime": &graphql.Field{Type: graphql.Float},
			"cacheHitRate": &graphql.Field{Type: graphql.Float},
		},
	})
}

func (r *GraphQLResolver) createContractStorageType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "ContractStorage",
		Fields: graphql.Fields{
			"address": &graphql.Field{Type: graphql.String},
			"values": &graphql.Field{Type: graphql.NewList(r.createStorageValueType())},
		},
	})
}

func (r *GraphQLResolver) createStorageValueType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "StorageValue",
		Fields: graphql.Fields{
			"key": &graphql.Field{Type: graphql.String},
			"value": &graphql.Field{Type: graphql.String},
		},
	})
}

func (r *GraphQLResolver) createSidechainAssetType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "SidechainAsset",
		Fields: graphql.Fields{
			"address": &graphql.Field{Type: graphql.String},
			"type": &graphql.Field{Type: graphql.String},
			"balance": &graphql.Field{Type: graphql.String},
		},
	})
}

// Input types
func (r *GraphQLResolver) createSendTransactionInputType() *graphql.InputObject {
	return graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "SendTransactionInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"to": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"amount": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"fee": &graphql.InputObjectFieldConfig{Type: graphql.String},
			"data": &graphql.InputObjectFieldConfig{Type: graphql.String},
		},
	})
}

func (r *GraphQLResolver) createDeployContractInputType() *graphql.InputObject {
	return graphql.NewInputObject(graphql.InputObjectConfig{
		Name: "DeployContractInput",
		Fields: graphql.InputObjectConfigFieldMap{
			"code": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"owner": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.String)},
			"initialBalance": &graphql.InputObjectFieldConfig{Type: graphql.String},
			"gasLimit": &graphql.InputObjectFieldConfig{Type: graphql.Int},
		},
	})
}

// Result types
func (r *GraphQLResolver) createTransactionResultType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "TransactionResult",
		Fields: graphql.Fields{
			"success": &graphql.Field{Type: graphql.Boolean},
			"transaction": &graphql.Field{Type: r.createTransactionType()},
			"error": &graphql.Field{Type: graphql.String},
		},
	})
}

func (r *GraphQLResolver) createContractResultType() *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: "ContractResult",
		Fields: graphql.Fields{
			"success": &graphql.Field{Type: graphql.Boolean},
			"contract": &graphql.Field{Type: r.createContractType()},
			"gasUsed": &graphql.Field{Type: graphql.Int},
			"error": &graphql.Field{Type: graphql.String},
		},
	})
}

// Resolver methods
func (r *GraphQLResolver) getBlockchain(ctx context.Context) (map[string]interface{}, error) {
	// Используем type assertion для доступа к методам
	var height int64
	var stats map[string]interface{}
	
	if bc, ok := r.blockchain.(interface{ GetHeight() int64 }); ok {
		height = bc.GetHeight()
	} else {
		height = 0
	}
	
	if statsBC, ok := r.blockchain.(interface{ GetStats() map[string]interface{} }); ok {
		stats = statsBC.GetStats()
	} else {
		stats = map[string]interface{}{"cache_hits": 0.0, "cache_misses": 0.0}
	}
	
	return map[string]interface{}{
		"height": height,
		"difficulty": 4, // TODO: Get from blockchain
		"totalSupply": "1000000", // TODO: Calculate from UTXO
		"hashRate": 0.0, // TODO: Calculate from mining
		"stats": map[string]interface{}{
			"totalBlocks": height + 1,
			"totalTransactions": 0, // TODO: Get from blockchain
			"totalAddresses": 0, // TODO: Get from UTXO
			"averageBlockTime": 10.0, // TODO: Calculate
			"cacheHitRate": stats["cache_hits"].(float64) / (stats["cache_hits"].(float64) + stats["cache_misses"].(float64)),
		},
	}, nil
}

func (r *GraphQLResolver) getBlock(ctx context.Context, height int) (map[string]interface{}, error) {
	var block interface{}
	
	if bc, ok := r.blockchain.(interface{ GetBlockByHeight(int64) interface{} }); ok {
		block = bc.GetBlockByHeight(int64(height))
	} else {
		return nil, fmt.Errorf("blockchain does not support GetBlockByHeight")
	}
	
	if block == nil {
		return nil, fmt.Errorf("block not found")
	}
	
	// Используем type assertion для доступа к полям блока
	if blockData, ok := block.(interface{ 
		GetHash() string
		GetHeight() int64
		GetTimestamp() int64
		GetPreviousHash() string
		GetMerkleRoot() string
		GetNonce() int64
		GetDifficulty() int
	}); ok {
		return map[string]interface{}{
			"hash": blockData.GetHash(),
			"height": blockData.GetHeight(),
			"timestamp": time.Unix(blockData.GetTimestamp(), 0).Format(time.RFC3339),
			"previousHash": blockData.GetPreviousHash(),
			"merkleRoot": blockData.GetMerkleRoot(),
			"nonce": blockData.GetNonce(),
			"difficulty": blockData.GetDifficulty(),
			"transactions": []map[string]interface{}{}, // TODO: Convert transactions
			"size": len(blockData.GetHash()),
			"gasUsed": 0,
			"gasLimit": 1000000,
		}, nil
	}
	
	// Fallback для простых случаев
	return map[string]interface{}{
		"hash": "unknown",
		"height": height,
		"timestamp": time.Now().Format(time.RFC3339),
		"previousHash": "unknown",
		"merkleRoot": "unknown",
		"nonce": 0,
		"difficulty": 4,
		"transactions": []map[string]interface{}{},
		"size": 0,
		"gasUsed": 0,
		"gasLimit": 1000000,
	}, nil
}

func (r *GraphQLResolver) getTransaction(ctx context.Context, hash string) (map[string]interface{}, error) {
	// TODO: Implement transaction lookup
	return map[string]interface{}{
		"hash": hash,
		"from": "",
		"to": "",
		"amount": "0",
		"fee": "0",
		"timestamp": time.Now().Format(time.RFC3339),
		"blockHeight": 0,
		"status": "PENDING",
		"gasUsed": 0,
		"gasPrice": "0",
		"data": "",
		"signature": "",
	}, nil
}

func (r *GraphQLResolver) getWallet(ctx context.Context, address string) (map[string]interface{}, error) {
	wallet, exists := r.walletManager.GetWallet(address)
	if !exists {
		return nil, fmt.Errorf("wallet not found")
	}
	
	return map[string]interface{}{
		"address": wallet.Address,
		"balance": "0", // TODO: Get from UTXO
		"nonce": 0,
		"transactions": []map[string]interface{}{},
		"createdAt": time.Now().Format(time.RFC3339),
	}, nil
}

func (r *GraphQLResolver) getContract(ctx context.Context, address string) (map[string]interface{}, error) {
	contract, exists := r.vm.GetContract(address)
	if !exists {
		return nil, fmt.Errorf("contract not found")
	}
	
	return map[string]interface{}{
		"address": contract.Address,
		"code": string(contract.Code),
		"owner": contract.Owner,
		"balance": contract.Balance.String(),
		"createdAt": contract.CreatedAt.Format(time.RFC3339),
		"updatedAt": contract.UpdatedAt.Format(time.RFC3339),
		"storage": map[string]interface{}{
			"address": contract.Address,
			"values": []map[string]interface{}{}, // TODO: Get storage values
		},
	}, nil
}

func (r *GraphQLResolver) getToken(ctx context.Context, address string) (map[string]interface{}, error) {
	// TODO: Implement token lookup
	return map[string]interface{}{
		"address": address,
		"name": "Unknown Token",
		"symbol": "UNK",
		"decimals": 18,
		"totalSupply": "0",
		"owner": "",
		"createdAt": time.Now().Format(time.RFC3339),
	}, nil
}

func (r *GraphQLResolver) getPeers(ctx context.Context) ([]map[string]interface{}, error) {
	// TODO: Get peers from network manager
	return []map[string]interface{}{
		{
			"id": "peer1",
			"address": "127.0.0.1",
			"port": 8080,
			"lastSeen": time.Now().Format(time.RFC3339),
			"latency": 10.5,
		},
	}, nil
}

func (r *GraphQLResolver) getNetworkStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"totalPeers": 1,
		"connectedPeers": 1,
		"averageLatency": 10.5,
		"bandwidth": 1000.0,
	}, nil
}

func (r *GraphQLResolver) getConsensusInfo(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"algorithm": "PoW",
		"validators": []string{},
		"currentRound": 1,
		"nextValidator": "",
	}, nil
}

func (r *GraphQLResolver) sendTransaction(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// TODO: Implement transaction sending
	return map[string]interface{}{
		"success": true,
		"transaction": map[string]interface{}{
			"hash": "tx_hash_123",
			"from": "from_address",
			"to": input["to"],
			"amount": input["amount"],
			"fee": input["fee"],
			"timestamp": time.Now().Format(time.RFC3339),
			"blockHeight": 0,
			"status": "PENDING",
			"gasUsed": 0,
			"gasPrice": "0",
			"data": input["data"],
			"signature": "signature_123",
		},
		"error": "",
	}, nil
}

func (r *GraphQLResolver) deployContract(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	code := input["code"].(string)
	owner := input["owner"].(string)
	
	// Compile contract
	compiler := vm.NewCompiler()
	instructions, err := compiler.Compile(code)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"contract": nil,
			"gasUsed": 0,
			"error": err.Error(),
		}, nil
	}
	
	// Deploy contract
	initialBalance := big.NewInt(0)
	if balance, ok := input["initialBalance"].(string); ok && balance != "" {
		initialBalance, _ = new(big.Int).SetString(balance, 10)
	}
	
	contract, err := r.vm.DeployContract(instructions, owner, initialBalance)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"contract": nil,
			"gasUsed": 0,
			"error": err.Error(),
		}, nil
	}
	
	return map[string]interface{}{
		"success": true,
		"contract": map[string]interface{}{
			"address": contract.Address,
			"code": string(contract.Code),
			"owner": contract.Owner,
			"balance": contract.Balance.String(),
			"createdAt": contract.CreatedAt.Format(time.RFC3339),
			"updatedAt": contract.UpdatedAt.Format(time.RFC3339),
			"storage": map[string]interface{}{
				"address": contract.Address,
				"values": []map[string]interface{}{},
			},
		},
		"gasUsed": r.vm.GetGasUsed(),
		"error": "",
	}, nil
}
