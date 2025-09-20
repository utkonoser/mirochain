import MiroChainSDK, { QUERIES, MUTATIONS } from '../src/index';

async function main() {
  // Инициализация SDK
  const sdk = new MiroChainSDK({
    baseURL: 'http://localhost:8080',
    timeout: 10000,
  });

  try {
    // Проверка здоровья
    console.log('Checking health...');
    const isHealthy = await sdk.healthCheck();
    console.log('Health status:', isHealthy);

    // Получение информации о блокчейне
    console.log('\nGetting blockchain info...');
    const blockchainInfo = await sdk.getBlockchainInfo();
    console.log('Blockchain info:', blockchainInfo);

    // Создание кошелька
    console.log('\nCreating wallet...');
    const wallet = await sdk.createWallet();
    console.log('Wallet created:', wallet);

    // Получение блока
    console.log('\nGetting block 0...');
    const block = await sdk.getBlock(0);
    console.log('Block 0:', block);

    // GraphQL запрос
    console.log('\nGraphQL query...');
    const graphqlResult = await sdk.graphqlQuery(QUERIES.GET_BLOCKCHAIN);
    console.log('GraphQL result:', graphqlResult);

    // Создание токена
    console.log('\nCreating token...');
    const token = await sdk.createToken(
      'MiroChain Token',
      'MCT',
      18,
      '1000000000000000000000000' // 1M tokens
    );
    console.log('Token created:', token);

    // Создание NFT контракта
    console.log('\nCreating NFT contract...');
    const nftContract = await sdk.createNFTContract(
      'MiroChain NFTs',
      'MCNFT',
      'https://api.mirochain.com/metadata/'
    );
    console.log('NFT contract created:', nftContract);

    // Минт NFT
    console.log('\nMinting NFT...');
    const nft = await sdk.mintNFT(
      nftContract.address,
      wallet.address,
      JSON.stringify({
        name: 'MiroChain Genesis NFT',
        description: 'The first NFT on MiroChain',
        image: 'https://api.mirochain.com/images/genesis.png',
        attributes: [
          { trait_type: 'Rarity', value: 'Legendary' },
          { trait_type: 'Generation', value: 'Genesis' }
        ]
      })
    );
    console.log('NFT minted:', nft);

    // Развертывание смарт-контракта
    console.log('\nDeploying smart contract...');
    const contractCode = `
      PUSH 42
      STORE counter
      PUSH 1
      SLOAD counter
      ADD
      SSTORE counter
      SLOAD counter
      RETURN
    `;
    const contract = await sdk.deployContract(contractCode, wallet.address, '0');
    console.log('Contract deployed:', contract);

    // Вызов контракта
    console.log('\nCalling contract...');
    const result = await sdk.callContract(contract.address, 'increment', []);
    console.log('Contract call result:', result);

  } catch (error) {
    console.error('Error:', error);
  }
}

// Запуск примера
if (require.main === module) {
  main().catch(console.error);
}

export { main };
