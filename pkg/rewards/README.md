## Rewards

### Claim Command
```bash
eigenlayer rewards claim --help
NAME:
   eigenlayer rewards claim - Claim rewards for any earner

USAGE:
   eigenlayer rewards claim [command options]

OPTIONS:
   --broadcast, -b                                      Use this flag to broadcast the transaction (default: false) [$BROADCAST]
   --claim-timestamp value, -c value                    Specify the timestamp. Only 'latest' and 'latest_active' are supported. 'latest' can be an inactive root which you can't claim yet. (default: "latest_active") [$CLAIM_TIMESTAMP]
   --claimer-address value, -a value                    Address of the claimer [$REWARDS_CLAIMER_ADDRESS]
   --earner-address value, --ea value                   Address of the earner [$REWARDS_EARNER_ADDRESS]
   --ecdsa-private-key value, -e value                  ECDSA private key hex to send transaction [$ECDSA_PRIVATE_KEY]
   --environment value, --env value                     Environment to use. Currently supports 'preprod' ,'testnet' and 'prod'. If not provided, it will be inferred based on network [$ENVIRONMENT]
   --eth-rpc-url value, -r value                        URL of the Ethereum RPC [$ETH_RPC_URL]
   --fireblocks-api-key value, --ff value               Fireblocks API key [$FIREBLOCKS_API_KEY]
   --fireblocks-aws-region value, --fa value            AWS region if secret is stored in AWS KMS (default: "us-east-1") [$FIREBLOCKS_AWS_REGION]
   --fireblocks-base-url value, --fb value              Fireblocks base URL [$FIREBLOCKS_BASE_URL]
   --fireblocks-secret-key value, --fs value            Fireblocks secret key. If you are using AWS Secret Manager, this should be the secret name. [$FIREBLOCKS_SECRET_KEY]
   --fireblocks-secret-storage-type value, --fst value  Fireblocks secret storage type. Supported values are 'plaintext' and 'aws_secret_manager' [$FIREBLOCKS_SECRET_STORAGE_TYPE]
   --fireblocks-timeout value, --ft value               Fireblocks timeout (default: 30) [$FIREBLOCKS_TIMEOUT]
   --fireblocks-vault-account-name value, --fv value    Fireblocks vault account name [$FIREBLOCKS_VAULT_ACCOUNT_NAME]
   --network value, -n value                            Network to use. Currently supports 'holesky', 'hoodi', 'sepolia' and 'mainnet' (default: "holesky") [$NETWORK]
   --output-file value, -o value                        Output file to write the data [$OUTPUT_FILE]
   --output-type value, --ot value                      Output format of the command. One of 'pretty', 'json' or 'calldata' (default: "pretty") [$OUTPUT_TYPE]
   --path-to-key-store value, -k value                  Path to the key store used to send transactions [$PATH_TO_KEY_STORE]
   --proof-store-base-url value, --psbu value           Specify the base URL of the proof store. If not provided, the value based on network will be used [$PROOF_STORE_BASE_URL]
   --recipient-address value, --ra value                Specify the address of the recipient. If this is not provided, the earner address will be used [$RECIPIENT_ADDRESS]
   --rewards-coordinator-address value, --rc value      Specify the address of the rewards coordinator. If not provided, the address will be used based on provided network [$REWARDS_COORDINATOR_ADDRESS]
   --silent, -s                                         Suppress unnecessary output (default: false) [$SILENT]
   --token-addresses value, -t value                    Specify the addresses of the tokens to claim. Comma separated list of addresses. Omit to claim all rewards. [$TOKEN_ADDRESSES]
   --verbose, -v                                        Enable verbose logging (default: false) [$VERBOSE]
   --web3signer-url value, -w value                     URL of the Web3Signer [$WEB3SIGNER_URL]
   --help, -h                                           show help
```

#### Example
##### Mainnet
```bash
eigenlayer rewards claim \
  --network mainnet \
  --eth-rpc-url https://rpc.ankr.com/eth/<> \
  --earner-address 0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f \
  --path-to-key-store /path/to/key \
  --broadcast
```
##### Testnet
```bash
eigenlayer rewards claim \
  --network holesky \
  --eth-rpc-url https://rpc.ankr.com/eth_holesky/<> \
  --earner-address 0x111116fe4f8c2f83e3eb2318f090557b7cd0bf76 \
  --recipient-address 0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f \
  --path-to-key-store /path/to/key/store \
```

### Set Claimer Command
```bash
eigenlayer rewards set-claimer --help
NAME:
   eigenlayer rewards set-claimer - Set the claimer address for the earner

USAGE:
   set-claimer

DESCRIPTION:

   Set the rewards claimer address for the earner.


OPTIONS:
   --broadcast, -b                                      Use this flag to broadcast the transaction (default: false) [$BROADCAST]
   --claimer-address value, -a value                    Address of the claimer [$REWARDS_CLAIMER_ADDRESS]
   --earner-address value, --ea value                   Address of the earner [$REWARDS_EARNER_ADDRESS]
   --ecdsa-private-key value, -e value                  ECDSA private key hex to send transaction [$ECDSA_PRIVATE_KEY]
   --eth-rpc-url value, -r value                        URL of the Ethereum RPC [$ETH_RPC_URL]
   --fireblocks-api-key value, --ff value               Fireblocks API key [$FIREBLOCKS_API_KEY]
   --fireblocks-aws-region value, --fa value            AWS region if secret is stored in AWS KMS (default: "us-east-1") [$FIREBLOCKS_AWS_REGION]
   --fireblocks-base-url value, --fb value              Fireblocks base URL [$FIREBLOCKS_BASE_URL]
   --fireblocks-secret-key value, --fs value            Fireblocks secret key. If you are using AWS Secret Manager, this should be the secret name. [$FIREBLOCKS_SECRET_KEY]
   --fireblocks-secret-storage-type value, --fst value  Fireblocks secret storage type. Supported values are 'plaintext' and 'aws_secret_manager' [$FIREBLOCKS_SECRET_STORAGE_TYPE]
   --fireblocks-timeout value, --ft value               Fireblocks timeout (default: 30) [$FIREBLOCKS_TIMEOUT]
   --fireblocks-vault-account-name value, --fv value    Fireblocks vault account name [$FIREBLOCKS_VAULT_ACCOUNT_NAME]
   --network value, -n value                            Network to use. Currently supports 'holesky', 'hoodi', 'sepolia' and 'mainnet' (default: "holesky") [$NETWORK]
   --output-file value, -o value                        Output file to write the data [$OUTPUT_FILE]
   --output-type value, --ot value                      Output format of the command. One of 'pretty', 'json' or 'calldata' (default: "pretty") [$OUTPUT_TYPE]
   --path-to-key-store value, -k value                  Path to the key store used to send transactions [$PATH_TO_KEY_STORE]
   --rewards-coordinator-address value, --rc value      Specify the address of the rewards coordinator. If not provided, the address will be used based on provided network [$REWARDS_COORDINATOR_ADDRESS]
   --verbose, -v                                        Enable verbose logging (default: false) [$VERBOSE]
   --web3signer-url value, -w value                     URL of the Web3Signer [$WEB3SIGNER_URL]
   --help, -h                                           show help
```

#### Example
##### Mainnet
```bash
eigenlayer rewards set-claimer \
  --network mainnet \
  --eth-rpc-url https://rpc.ankr.com/eth/<> \
  --earner-address 0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f \
  --claimer-address 0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f \
  --path-to-key-store /path/to/key/store \
  --broadcast
```
##### Testnet
```bash
eigenlayer rewards set-claimer \
  --network holesky \
  --eth-rpc-url https://rpc.ankr.com/eth_holesky/<> \
  --earner-address 0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f \
  --claimer-address 0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f \
  --path-to-key-store /path/to/key/store \
  --broadcast
```

### Show Rewards
```bash
eigenlayer rewards show --help
NAME:
   eigenlayer rewards show - Show rewards for an address against the `DistributionRoot` posted on-chain by the rewards updater

USAGE:
   show

DESCRIPTION:

   Command to show rewards for earners

   Helpful flags
   - claim-type: Type of rewards to show. Can be 'all', 'claimed' or 'unclaimed'
   - claim-timestamp: Timestamp of the claim distribution root to use. Can be 'latest' or 'latest_active'.
     - 'latest' will show rewards for the latest root (can contain non-claimable rewards)
     - 'latest_active' will show rewards for the latest active root (only claimable rewards)


OPTIONS:
   --claim-timestamp value, -c value           Specify the timestamp. Only 'latest' and 'latest_active' are supported. 'latest' can be a from an inactive root which you can't claim yet. (default: "latest_active") [$CLAIM_TIMESTAMP]
   --claim-type value, --ct value              Type of claim you want to see. Can be 'all', 'unclaimed', or 'claimed' (default: "all") [$REWARDS_CLAIM_TYPE]
   --earner-address value, --ea value          Address of the earner [$REWARDS_EARNER_ADDRESS]
   --environment value, --env value            Environment to use. Currently supports 'preprod' ,'testnet' and 'prod'. If not provided, it will be inferred based on network [$ENVIRONMENT]
   --eth-rpc-url value, -r value               URL of the Ethereum RPC [$ETH_RPC_URL]
   --network value, -n value                   Network to use. Currently supports 'holesky', 'hoodi', 'sepolia' and 'mainnet' (default: "holesky") [$NETWORK]
   --output-file value, -o value               Output file to write the data [$OUTPUT_FILE]
   --output-type value, --ot value             Output format of the command. One of 'pretty', 'json' or 'calldata' (default: "pretty") [$OUTPUT_TYPE]
   --proof-store-base-url value, --psbu value  Specify the base URL of the proof store. If not provided, the value based on network will be used [$PROOF_STORE_BASE_URL]
   --verbose, -v                               Enable verbose logging (default: false) [$VERBOSE]
   --help, -h                                  show help
```

#### Example
Show all Rewards
```bash
./bin/eigenlayer rewards show \
  --network mainnet \
  --earner-address 0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f \
  --eth-rpc-url https://rpc.ankr.com/eth/<> \
  --claim-type all --verbose
```

Show claimed Rewards
```bash
./bin/eigenlayer rewards show \
  --network mainnet \
  --earner-address 0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f \
  --eth-rpc-url https://rpc.ankr.com/eth/<> \
  --claim-type claomed --verbose
```

Show unclaimed Rewards
```bash
./bin/eigenlayer rewards show \
  --network mainnet \
  --earner-address 0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f \
  --eth-rpc-url https://rpc.ankr.com/eth/<> \
  --claim-type unclaimed --verbose
```
