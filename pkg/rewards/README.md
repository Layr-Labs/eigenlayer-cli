## Rewards

### Command
```bash
eigenlayer rewards claim --help
NAME:
   eigenlayer rewards claim - Claim rewards for the operator

USAGE:
   eigenlayer rewards claim [command options]

OPTIONS:
   --verbose, -v                                    Enable verbose logging (default: false) [$VERBOSE]
   --network value, -n value                        Network to use. Currently supports 'preprod', 'holesky' and 'mainnet' (default: "holesky") [$NETWORK]
   --eth-rpc-url value, -r value                    URL of the Ethereum RPC [$ETH_RPC_URL]
   --earner-address value, --ea value               Address of the earner (this is your staker/operator address) [$EARNER_ADDRESS]
   --output-file value, -o value                    Output file to write the data [$OUTPUT_FILE]
   --path-to-key-store value, -k value              Path to the key store [$PATH_TO_KEY_STORE]
   --ecdsa-private-key value, -e value              ECDSA private key hex to send transaction [$ECDSA_PRIVATE_KEY]
   --broadcast, -b                                  Use this flag to broadcast the transaction (default: false) [$BROADCAST]
   --recipient-address value, --ra value            Specify the address of the recipient. If this is not provided, the earner address will be used [$RECIPIENT_ADDRESS]
   --token-addresses value, -t value                Specify the addresses of the tokens to claim. Comma separated list of addresses [$TOKEN_ADDRESSES]
   --rewards-coordinator-address value, --rc value  Specify the address of the rewards coordinator. If not provided, the address will be used based on provided network [$REWARDS_COORDINATOR_ADDRESS]
   --claim-timestamp value, -c value                Specify the timestamp. Only 'latest' is supported (default: "latest") [$CLAIM_TIMESTAMP]
   --proof-store-base-url value, --psbu value       Specify the base URL of the proof store. If not provided, the value based on network will be used [$PROOF_STORE_BASE_URL]
   --help, -h                                       show help
```

### Example
### Preprod
```bash
eigenlayer rewards claim \
  --network holesky \
  --env preprod \
  --eth-rpc-url https://rpc.ankr.com/eth_holesky/<> \
  --earner-address 0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f \
  --path-to-key-store /path/to/key \
  --token-addresses 0x554c393923c753d146aa34608523ad7946b61662 \
  --rewards-coordinator-address 0xb22Ef643e1E067c994019A4C19e403253C05c2B0 \
  --proof-store-base-url https://eigenlabs-rewards-preprod-holesky.s3.amazonaws.com
  --broadcast
```

### Testnet
```bash
eigenlayer rewards claim \
  --network holesky \
  --eth-rpc-url https://rpc.ankr.com/eth_holesky/<> \
  --earner-address 0x111116fe4f8c2f83e3eb2318f090557b7cd0bf76 \
  --recipient-address 0x2222AAC0C980Cc029624b7ff55B88Bc6F63C538f \
  --path-to-key-store /path/to/key/store \
  --token-addresses 0xdeeeeE2b48C121e6728ed95c860e296177849932 --broadcast
```