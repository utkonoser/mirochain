# Security Guidelines

## Wallet Files Security

⚠️ **IMPORTANT**: Wallet files contain private keys and are extremely sensitive!

### What's Protected
The following files and directories are automatically ignored by Git to prevent accidental exposure of private keys:

- `**/wallets/` - Any wallets directory
- `**/wallet_*.json` - Individual wallet files
- `**/wallets.json` - Wallet registry files
- `wallet_data/` - Wallet data directory

### Security Best Practices

1. **Never commit wallet files** - They contain private keys that could compromise your funds
2. **Use environment variables** for sensitive configuration
3. **Keep test wallets separate** from production wallets
4. **Regularly rotate keys** in development environments
5. **Use strong passwords** for wallet encryption

### Development Setup

When setting up the development environment:

1. Create test wallets using the wallet CLI:
   ```bash
   ./build/mirochain-wallet create --name test-wallet
   ```

2. Test wallets will be created in `wallet_data/` directory (ignored by Git)

3. For testing, use the provided test utilities that generate temporary wallets

### File Structure

```
project/
├── wallet_data/          # Production wallet data (ignored)
├── tests/test_data/      # Test data directory
│   └── wallets/          # Test wallets (ignored)
│       ├── wallet_*.json # Individual test wallets
│       └── wallets.json  # Test wallet registry
└── .gitignore           # Contains wallet ignore patterns
```

### Emergency Response

If you accidentally committed wallet files:

1. **Immediately rotate all affected keys**
2. **Remove files from Git history**:
   ```bash
   git filter-branch --force --index-filter \
   'git rm --cached --ignore-unmatch tests/test_data/wallets/*.json' \
   --prune-empty --tag-name-filter cat -- --all
   ```
3. **Force push to remote** (⚠️ This rewrites history)
4. **Notify team members** to update their local repositories

### Contact

For security concerns, please contact the development team immediately.
