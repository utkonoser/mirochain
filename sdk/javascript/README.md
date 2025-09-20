# MiroChain JavaScript SDK

JavaScript/TypeScript SDK для взаимодействия с MiroChain блокчейном.

## Установка

```bash
npm install mirochain-sdk
```

## Быстрый старт

```typescript
import MiroChainSDK from 'mirochain-sdk';

// Инициализация SDK
const sdk = new MiroChainSDK({
  baseURL: 'http://localhost:8080',
  timeout: 10000,
});

// Получение информации о блокчейне
const blockchainInfo = await sdk.getBlockchainInfo();
console.log('Blockchain height:', blockchainInfo.height);

// Создание кошелька
const wallet = await sdk.createWallet();
console.log('Wallet address:', wallet.address);

// Создание токена
const token = await sdk.createToken(
  'My Token',
  'MTK',
  18,
  '1000000000000000000000000'
);
console.log('Token created:', token.address);
```

## API Reference

### Инициализация

```typescript
const sdk = new MiroChainSDK({
  baseURL: 'http://localhost:8080',  // URL узла MiroChain
  apiKey: 'your-api-key',            // Опциональный API ключ
  timeout: 10000,                    // Таймаут запросов в мс
});
```

### Блокчейн

```typescript
// Получение информации о блокчейне
const info = await sdk.getBlockchainInfo();

// Получение блока по высоте
const block = await sdk.getBlock(0);

// Получение транзакции по хешу
const tx = await sdk.getTransaction('0x...');
```

### Кошельки

```typescript
// Создание нового кошелька
const wallet = await sdk.createWallet();

// Получение информации о кошельке
const walletInfo = await sdk.getWallet('0x...');
```

### Смарт-контракты

```typescript
// Развертывание контракта
const contract = await sdk.deployContract(
  'PUSH 42\nSTORE counter\nRETURN',
  '0x...',  // owner
  '0'       // initial balance
);

// Вызов функции контракта
const result = await sdk.callContract(
  contract.address,
  'increment',
  []
);

// Получение информации о контракте
const contractInfo = await sdk.getContract('0x...');
```

### Токены (ERC-20)

```typescript
// Создание токена
const token = await sdk.createToken(
  'My Token',
  'MTK',
  18,
  '1000000000000000000000000'
);

// Перевод токенов
const tx = await sdk.transferToken(
  token.address,
  '0x...',  // to
  '1000000000000000000'  // amount
);

// Получение информации о токене
const tokenInfo = await sdk.getToken('0x...');
```

### NFT (ERC-721)

```typescript
// Создание NFT контракта
const nftContract = await sdk.createNFTContract(
  'My NFTs',
  'MNFT',
  'https://api.example.com/metadata/'
);

// Минт NFT
const nft = await sdk.mintNFT(
  nftContract.address,
  '0x...',  // to
  JSON.stringify({
    name: 'My NFT',
    description: 'A unique NFT',
    image: 'https://api.example.com/images/1.png'
  })
);

// Получение NFT
const nftInfo = await sdk.getNFT(contractAddress, tokenId);
```

### GraphQL

```typescript
import { QUERIES, MUTATIONS } from 'mirochain-sdk';

// GraphQL запрос
const result = await sdk.graphqlQuery(QUERIES.GET_BLOCKCHAIN);

// GraphQL мутация
const mutationResult = await sdk.graphqlQuery(MUTATIONS.SEND_TRANSACTION, {
  input: {
    to: '0x...',
    amount: '1000000000000000000',
    fee: '1000000000000000'
  }
});
```

## Примеры

Смотрите папку `examples/` для более подробных примеров использования.

## Лицензия

MIT
