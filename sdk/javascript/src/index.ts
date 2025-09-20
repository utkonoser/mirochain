import axios, { AxiosInstance } from 'axios';
import { GraphQLClient } from 'graphql-request';

export interface MiroChainConfig {
  baseURL: string;
  apiKey?: string;
  timeout?: number;
}

export interface BlockchainInfo {
  height: number;
  difficulty: number;
  totalSupply: string;
  hashRate: number;
  stats: {
    totalBlocks: number;
    totalTransactions: number;
    totalAddresses: number;
    averageBlockTime: number;
    cacheHitRate: number;
  };
}

export interface Block {
  hash: string;
  height: number;
  timestamp: string;
  previousHash: string;
  merkleRoot: string;
  nonce: number;
  difficulty: number;
  transactions: Transaction[];
  size: number;
  gasUsed: number;
  gasLimit: number;
}

export interface Transaction {
  hash: string;
  from: string;
  to: string;
  amount: string;
  fee: string;
  timestamp: string;
  blockHeight?: number;
  status: 'PENDING' | 'CONFIRMED' | 'FAILED';
  gasUsed?: number;
  gasPrice?: string;
  data?: string;
  signature: string;
}

export interface Wallet {
  address: string;
  balance: string;
  nonce: number;
  transactions: Transaction[];
  createdAt: string;
}

export interface Contract {
  address: string;
  code: string;
  owner: string;
  balance: string;
  createdAt: string;
  updatedAt: string;
  storage?: {
    address: string;
    values: Array<{
      key: string;
      value: string;
    }>;
  };
}

export interface Token {
  address: string;
  name: string;
  symbol: string;
  decimals: number;
  totalSupply: string;
  owner: string;
  createdAt: string;
}

export interface NFT {
  contractAddress: string;
  tokenId: string;
  owner: string;
  metadata: string;
  createdAt: string;
}

export class MiroChainSDK {
  private httpClient: AxiosInstance;
  private graphqlClient: GraphQLClient;

  constructor(config: MiroChainConfig) {
    this.httpClient = axios.create({
      baseURL: config.baseURL,
      timeout: config.timeout || 10000,
      headers: {
        'Content-Type': 'application/json',
        ...(config.apiKey && { 'Authorization': `Bearer ${config.apiKey}` }),
      },
    });

    this.graphqlClient = new GraphQLClient(`${config.baseURL}/graphql`, {
      headers: {
        ...(config.apiKey && { 'Authorization': `Bearer ${config.apiKey}` }),
      },
    });
  }

  // Blockchain methods
  async getBlockchainInfo(): Promise<BlockchainInfo> {
    const response = await this.httpClient.get('/api/v1/blockchain');
    return response.data;
  }

  async getBlock(height: number): Promise<Block> {
    const response = await this.httpClient.get(`/api/v1/blocks/${height}`);
    return response.data;
  }

  async getTransaction(hash: string): Promise<Transaction> {
    const response = await this.httpClient.get(`/api/v1/transactions/${hash}`);
    return response.data;
  }

  // Wallet methods
  async getWallet(address: string): Promise<Wallet> {
    const response = await this.httpClient.get(`/api/v1/wallets/${address}`);
    return response.data;
  }

  async createWallet(): Promise<Wallet> {
    const response = await this.httpClient.post('/api/v1/wallets');
    return response.data;
  }

  // Contract methods
  async getContract(address: string): Promise<Contract> {
    const response = await this.httpClient.get(`/api/v1/contracts/${address}`);
    return response.data;
  }

  async deployContract(code: string, owner: string, initialBalance?: string): Promise<Contract> {
    const response = await this.httpClient.post('/api/v1/contracts/deploy', {
      code,
      owner,
      initialBalance: initialBalance || '0',
    });
    return response.data;
  }

  async callContract(address: string, function: string, args: string[]): Promise<any> {
    const response = await this.httpClient.post(`/api/v1/contracts/${address}/call`, {
      function,
      args,
    });
    return response.data;
  }

  // Token methods
  async getToken(address: string): Promise<Token> {
    const response = await this.httpClient.get(`/api/v1/tokens/${address}`);
    return response.data;
  }

  async createToken(name: string, symbol: string, decimals: number, totalSupply: string): Promise<Token> {
    const response = await this.httpClient.post('/api/v1/tokens', {
      name,
      symbol,
      decimals,
      totalSupply,
    });
    return response.data;
  }

  async transferToken(tokenAddress: string, to: string, amount: string): Promise<Transaction> {
    const response = await this.httpClient.post(`/api/v1/tokens/${tokenAddress}/transfer`, {
      to,
      amount,
    });
    return response.data;
  }

  // NFT methods
  async getNFT(contractAddress: string, tokenId: string): Promise<NFT> {
    const response = await this.httpClient.get(`/api/v1/nfts/${contractAddress}/${tokenId}`);
    return response.data;
  }

  async createNFTContract(name: string, symbol: string, baseURI?: string): Promise<Contract> {
    const response = await this.httpClient.post('/api/v1/nfts', {
      name,
      symbol,
      baseURI,
    });
    return response.data;
  }

  async mintNFT(contractAddress: string, to: string, metadata: string): Promise<NFT> {
    const response = await this.httpClient.post(`/api/v1/nfts/${contractAddress}/mint`, {
      to,
      metadata,
    });
    return response.data;
  }

  // GraphQL methods
  async graphqlQuery(query: string, variables?: any): Promise<any> {
    return await this.graphqlClient.request(query, variables);
  }

  // Health check
  async healthCheck(): Promise<boolean> {
    try {
      const response = await this.httpClient.get('/health');
      return response.status === 200;
    } catch {
      return false;
    }
  }
}

// GraphQL queries
export const QUERIES = {
  GET_BLOCKCHAIN: `
    query GetBlockchain {
      blockchain {
        height
        difficulty
        totalSupply
        hashRate
        stats {
          totalBlocks
          totalTransactions
          totalAddresses
          averageBlockTime
          cacheHitRate
        }
      }
    }
  `,
  
  GET_BLOCK: `
    query GetBlock($height: Int!) {
      block(height: $height) {
        hash
        height
        timestamp
        previousHash
        merkleRoot
        nonce
        difficulty
        size
        gasUsed
        gasLimit
      }
    }
  `,
  
  GET_WALLET: `
    query GetWallet($address: String!) {
      wallet(address: $address) {
        address
        balance
        nonce
        createdAt
      }
    }
  `,
  
  GET_CONTRACT: `
    query GetContract($address: String!) {
      contract(address: $address) {
        address
        code
        owner
        balance
        createdAt
        updatedAt
        storage {
          address
          values {
            key
            value
          }
        }
      }
    }
  `,
};

// GraphQL mutations
export const MUTATIONS = {
  SEND_TRANSACTION: `
    mutation SendTransaction($input: SendTransactionInput!) {
      sendTransaction(input: $input) {
        success
        transaction {
          hash
          from
          to
          amount
          fee
          timestamp
          status
        }
        error
      }
    }
  `,
  
  DEPLOY_CONTRACT: `
    mutation DeployContract($input: DeployContractInput!) {
      deployContract(input: $input) {
        success
        contract {
          address
          code
          owner
          balance
          createdAt
        }
        gasUsed
        error
      }
    }
  `,
};

export default MiroChainSDK;
