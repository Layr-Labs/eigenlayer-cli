#!/usr/bin/env bats

setup() {
  echo "Setting up test environment..."
  rm -f "$OUTPUT_FILE_FOLDER/output_*.txt"
}

teardown() {
  echo "Cleaning up test environment..."
  rm -f "$OUTPUT_FILE_FOLDER/output_*.txt"
}

@test "Verify canCall permissions for account" {
  run $CLI_PATH user appointee can-call \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$ACCOUNT_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" == *"CanCall Result: true"* ]]
}

@test "Set appointee and verify calldata output (calldata)" {
  run $CLI_PATH user appointee set \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --output-type "calldata" \
    --output-file "$OUTPUT_FILE_FOLDER/output_set.txt"

  [ "$status" -eq 0 ]
  [[ -s "$OUTPUT_FILE_FOLDER/output_set.txt" ]]
}

@test "Broadcast set appointee command (broadcast)" {
  run $CLI_PATH user appointee set \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Remove appointee and verify calldata output (calldata)" {
  run $CLI_PATH user appointee remove \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --output-type "calldata" \
    --output-file "$OUTPUT_FILE_FOLDER/output_remove.txt"

  [ "$status" -eq 0 ]
  [[ -s "$OUTPUT_FILE_FOLDER/output_remove.txt" ]]
}

@test "Broadcast remove appointee command (broadcast)" {
  run $CLI_PATH user appointee remove \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Verify appointee removal" {
  run $CLI_PATH user appointee list \
    --account-address "$ACCOUNT_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" != *"$APPOINTEE_ADDRESS"* ]]
}

@test "Verify canCall permissions for removed appointee" {
  run $CLI_PATH user appointee can-call \
    --appointee-address "$APPOINTEE_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" == *"CanCall Result: false"* ]]
}
