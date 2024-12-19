## Allocations Command
### Initialize Delay
```bash
eigenlayer operator allocations initialize-delay --help
NAME:
   eigenlayer operator allocations initialize-delay - Initialize the allocation delay for operator

USAGE:
   initialize-delay [flags] <delay>

DESCRIPTION:
   Initializes the allocation delay for operator. This is a one time command. You can not change the allocation delay once

OPTIONS:
   --broadcast, -b                                         Use this flag to broadcast the transaction (default: false) [$BROADCAST]
   --ecdsa-private-key value, -e value                     ECDSA private key hex to send transaction [$ECDSA_PRIVATE_KEY]
   --environment value, --env value                        environment to use. Currently supports 'preprod' ,'testnet' and 'prod'. If not provided, it will be inferred based on network [$ENVIRONMENT]
   --eth-rpc-url value, -r value                           URL of the Ethereum RPC [$ETH_RPC_URL]
   --fireblocks-api-key value, --ff value                  Fireblocks API key [$FIREBLOCKS_API_KEY]
   --fireblocks-aws-region value, --fa value               AWS region if secret is stored in AWS KMS (default: "us-east-1") [$FIREBLOCKS_AWS_REGION]
   --fireblocks-base-url value, --fb value                 Fireblocks base URL [$FIREBLOCKS_BASE_URL]
   --fireblocks-secret-key value, --fs value               Fireblocks secret key. If you are using AWS Secret Manager, this should be the secret name. [$FIREBLOCKS_SECRET_KEY]
   --fireblocks-secret-storage-type value, --fst value     Fireblocks secret storage type. Supported values are 'plaintext' and 'aws_secret_manager' [$FIREBLOCKS_SECRET_STORAGE_TYPE]
   --fireblocks-timeout value, --ft value                  Fireblocks timeout (default: 30) [$FIREBLOCKS_TIMEOUT]
   --fireblocks-vault-account-name value, --fv value       Fireblocks vault account name [$FIREBLOCKS_VAULT_ACCOUNT_NAME]
   --network value, -n value                               Network to use. Currently supports 'holesky' and 'mainnet' (default: "holesky") [$NETWORK]
   --operator-address value, --oa value, --operator value  Operator address [$OPERATOR_ADDRESS]
   --output-file value, -o value                           Output file to write the data [$OUTPUT_FILE]
   --output-type value, --ot value                         Output format of the command. One of 'pretty', 'json' or 'calldata' (default: "pretty") [$OUTPUT_TYPE]
   --path-to-key-store value, -k value                     Path to the key store used to send transactions [$PATH_TO_KEY_STORE]
   --verbose, -v                                           Enable verbose logging (default: false) [$VERBOSE]
   --web3signer-url value, -w value                        URL of the Web3Signer [$WEB3SIGNER_URL]
   --help, -h                                              show help
```

### Update allocations
```bash
eigenlayer operator allocations update --help
NAME:
   eigenlayer operator allocations update - Update allocations

USAGE:
   update

DESCRIPTION:

       Command to update allocations


OPTIONS:
   --avs-address value, --aa value                                   AVS addresses [$AVS_ADDRESS]
   --bips-to-allocate value, --bta value, --bips value, --bps value  Bips to allocate to the strategy (default: 0) [$BIPS_TO_ALLOCATE]
   --broadcast, -b                                                   Use this flag to broadcast the transaction (default: false) [$BROADCAST]
   --csv-file value, --csv value                                     CSV file to read data from [$CSV_FILE]
   --ecdsa-private-key value, -e value                               ECDSA private key hex to send transaction [$ECDSA_PRIVATE_KEY]
   --environment value, --env value                                  environment to use. Currently supports 'preprod' ,'testnet' and 'prod'. If not provided, it will be inferred based on network [$ENVIRONMENT]
   --eth-rpc-url value, -r value                                     URL of the Ethereum RPC [$ETH_RPC_URL]
   --fireblocks-api-key value, --ff value                            Fireblocks API key [$FIREBLOCKS_API_KEY]
   --fireblocks-aws-region value, --fa value                         AWS region if secret is stored in AWS KMS (default: "us-east-1") [$FIREBLOCKS_AWS_REGION]
   --fireblocks-base-url value, --fb value                           Fireblocks base URL [$FIREBLOCKS_BASE_URL]
   --fireblocks-secret-key value, --fs value                         Fireblocks secret key. If you are using AWS Secret Manager, this should be the secret name. [$FIREBLOCKS_SECRET_KEY]
   --fireblocks-secret-storage-type value, --fst value               Fireblocks secret storage type. Supported values are 'plaintext' and 'aws_secret_manager' [$FIREBLOCKS_SECRET_STORAGE_TYPE]
   --fireblocks-timeout value, --ft value                            Fireblocks timeout (default: 30) [$FIREBLOCKS_TIMEOUT]
   --fireblocks-vault-account-name value, --fv value                 Fireblocks vault account name [$FIREBLOCKS_VAULT_ACCOUNT_NAME]
   --network value, -n value                                         Network to use. Currently supports 'holesky' and 'mainnet' (default: "holesky") [$NETWORK]
   --operator-address value, --oa value, --operator value            Operator address [$OPERATOR_ADDRESS]
   --operator-set-id value, --osid value                             Operator set ID (default: 0) [$OPERATOR_SET_ID]
   --output-file value, -o value                                     Output file to write the data [$OUTPUT_FILE]
   --output-type value, --ot value                                   Output format of the command. One of 'pretty', 'json' or 'calldata' (default: "pretty") [$OUTPUT_TYPE]
   --path-to-key-store value, -k value                               Path to the key store used to send transactions [$PATH_TO_KEY_STORE]
   --strategy-address value, --sa value                              Strategy addresses [$STRATEGY_ADDRESS]
   --verbose, -v                                                     Enable verbose logging (default: false) [$VERBOSE]
   --web3signer-url value, -w value                                  URL of the Web3Signer [$WEB3SIGNER_URL]
   --help, -h                                                        show help
```