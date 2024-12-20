#!/usr/bin/env bats

verify_listed_permissions() {
  local selector_1=$1
  local selector_2=$2
  local appointee=$3
  local account=$4
  local target=$5
  local output=$6

  local selector_1_no_prefix=${selector_1#0x}
  local selector_2_no_prefix=${selector_2#0x}

  [ "$status" -eq 0 ]
  [[ "$output" == *"Appointee address: $appointee"* ]]
  [[ "$output" == *"Appointed by: $account"* ]]
  [[ "$output" == *"Target: $target, Selector: $selector_1_no_prefix"* ]]
  [[ "$output" == *"Target: $target, Selector: $selector_2_no_prefix"* ]]
}

@test "Verify initial empty list of appointees" {
  run $CLI_PATH user appointee list \
    --account-address "$ACCOUNT_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" != *"$ACCOUNT_ADDRESS"* ]]
  [[ "$output" != *"$APPOINTEE_1"* ]]
}

@test "Add first appointee for selector 1" {
  run $CLI_PATH user appointee set \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_1" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Add second appointee for selector 1" {
  run $CLI_PATH user appointee set \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_2" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Add first appointee for selector 2" {
  run $CLI_PATH user appointee set \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_1" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_2" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Add second appointee for selector 2" {
  run $CLI_PATH user appointee set \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_2" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_2" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Verify multiple appointees for selector 1" {
  run $CLI_PATH user appointee list \
    --account-address "$ACCOUNT_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" == *"$APPOINTEE_1"* ]]
  [[ "$output" == *"$APPOINTEE_2"* ]]
}

@test "Verify multiple appointees for selector 2" {
  run $CLI_PATH user appointee list \
    --account-address "$ACCOUNT_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_2" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" == *"$APPOINTEE_1"* ]]
  [[ "$output" == *"$APPOINTEE_2"* ]]
}

@test "Verify appointee1's listed permissions" {
  run $CLI_PATH user appointee list-permissions \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  SELECTOR_1_NO_PREFIX=${SELECTOR_1#0x}
  SELECTOR_2_NO_PREFIX=${SELECTOR_2#0x}
  verify_listed_permissions \
      "$SELECTOR_1_NO_PREFIX" \
      "$SELECTOR_2_NO_PREFIX" \
      "$APPOINTEE_1" \
      "$ACCOUNT_ADDRESS" \
      "$TARGET_ADDRESS" \
      "$output"
}

@test "Verify appointee2's listed permissions" {
  run $CLI_PATH user appointee list-permissions \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_2" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  SELECTOR_1_NO_PREFIX=${SELECTOR_1#0x}
  SELECTOR_2_NO_PREFIX=${SELECTOR_2#0x}
  verify_listed_permissions \
      "$SELECTOR_1_NO_PREFIX" \
      "$SELECTOR_2_NO_PREFIX" \
      "$APPOINTEE_2" \
      "$ACCOUNT_ADDRESS" \
      "$TARGET_ADDRESS" \
      "$output"
}

@test "Test canCall permissions for selectors is true" {
  run $CLI_PATH user appointee can-call \
    --appointee-address "$APPOINTEE_1" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" == *"CanCall Result: true"* ]]

    run $CLI_PATH user appointee can-call \
      --appointee-address "$APPOINTEE_1" \
      --target-address "$TARGET_ADDRESS" \
      --selector "$SELECTOR_1" \
      --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
      --eth-rpc-url "$RPC_URL" \
      --network "$NETWORK"

    [ "$status" -eq 0 ]
    [[ "$output" == *"CanCall Result: true"* ]]

  run $CLI_PATH user appointee can-call \
    --appointee-address "$APPOINTEE_2" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" == *"CanCall Result: true"* ]]

    run $CLI_PATH user appointee can-call \
      --appointee-address "$APPOINTEE_2" \
      --target-address "$TARGET_ADDRESS" \
      --selector "$SELECTOR_2" \
      --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
      --eth-rpc-url "$RPC_URL" \
      --network "$NETWORK"

    [ "$status" -eq 0 ]
    [[ "$output" == *"CanCall Result: true"* ]]
}

@test "Remove first appointee for selector 1" {
  run $CLI_PATH user appointee remove \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_1" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Verify only second appointee remains for selector 1" {
  run $CLI_PATH user appointee list \
    --account-address "$ACCOUNT_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" != *"$APPOINTEE_1"* ]]
  [[ "$output" == *"$APPOINTEE_2"* ]]
}

@test "Remove second appointee for selector 1" {
  run $CLI_PATH user appointee remove \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_2" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Verify list is empty after selector 1 removals" {
  run $CLI_PATH user appointee list \
    --account-address "$ACCOUNT_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" != *"$APPOINTEE_1"* ]]
  [[ "$output" != *"$APPOINTEE_2"* ]]
}

@test "Remove first appointee for selector 2" {
  run $CLI_PATH user appointee remove \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_1" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_2" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Verify only second appointee remains for selector 2" {
  run $CLI_PATH user appointee list \
    --account-address "$ACCOUNT_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_2" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" != *"$APPOINTEE_1"* ]]
  [[ "$output" == *"$APPOINTEE_2"* ]]
}

@test "Remove second appointee for selector 2" {
  run $CLI_PATH user appointee remove \
    --account-address "$ACCOUNT_ADDRESS" \
    --appointee-address "$APPOINTEE_2" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_2" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Verify list is empty after selector 2 removals" {
  run $CLI_PATH user appointee list \
    --account-address "$ACCOUNT_ADDRESS" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_2" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" != *"$APPOINTEE_1"* ]]
  [[ "$output" != *"$APPOINTEE_2"* ]]
}

@test "Test canCall permissions for selectors" {
  run $CLI_PATH user appointee can-call \
    --appointee-address "$APPOINTEE_1" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_1" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" == *"CanCall Result: false"* ]]

  run $CLI_PATH user appointee can-call \
    --appointee-address "$APPOINTEE_2" \
    --target-address "$TARGET_ADDRESS" \
    --selector "$SELECTOR_2" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" == *"CanCall Result: false"* ]]
}

