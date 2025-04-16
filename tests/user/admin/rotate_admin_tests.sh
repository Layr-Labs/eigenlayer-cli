#!/usr/bin/env bats

@test "Rotate admins so first admin address is the only admin" {
  run $CLI_PATH user admin add-pending-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$FIRST_ADMIN_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$ACCOUNT_PRIVATE_KEY" \
    --broadcast
  [ "$status" -eq 0 ]

  run $CLI_PATH user admin accept-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$FIRST_ADMIN_PRIVATE_KEY" \
    --broadcast
  [ "$status" -eq 0 ]

  run $CLI_PATH user admin list-admins \
    --account-address "$ACCOUNT_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"
  [ "$status" -eq 0 ]

  [[ "$output" == *"$FIRST_ADMIN_ADDRESS"* ]]
}

@test "Rotate admins again so root account is the only admin" {
  run $CLI_PATH user admin add-pending-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$ACCOUNT_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$FIRST_ADMIN_PRIVATE_KEY" \
    --broadcast
  [ "$status" -eq 0 ]

  run $CLI_PATH user admin accept-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$ACCOUNT_PRIVATE_KEY" \
    --broadcast
  [ "$status" -eq 0 ]

  run $CLI_PATH user admin remove-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$FIRST_ADMIN_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$ACCOUNT_PRIVATE_KEY" \
    --broadcast
  [ "$status" -eq 0 ]

  run $CLI_PATH user admin list-admins \
    --account-address "$ACCOUNT_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"
  [ "$status" -eq 0 ]

  [[ "$output" == *"$ACCOUNT_ADDRESS"* ]]
  [[ "$output" != *"$FIRST_ADMIN_ADDRESS"* ]]
}

