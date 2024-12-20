#!/usr/bin/env bats

load './admin_utils.sh'

setup() {
  echo "Setting up test environment..."
  rm -f "$OUTPUT_FILE_FOLDER/output_*.txt"

  echo "Listing admins for account $ACCOUNT_ADDRESS..."
  output_list_admins=$($CLI_PATH user admin list-admins \
    --account-address "$ACCOUNT_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK")
  echo "$output_list_admins"

  echo "Running conditional_add_admin..."
  output_conditional_add=$(conditional_add_admin \
    "$ACCOUNT_ADDRESS" \
    "$ACCOUNT_ADDRESS" \
    "$ACCOUNT_PRIVATE_KEY" \
    "$PERMISSION_CONTROLLER_ADDRESS" \
    "$RPC_URL" \
    "$NETWORK" \
    "$OUTPUT_ADD_FILE" \
    "$OUTPUT_ACCEPT_FILE")
  echo "$output_conditional_add"
}

teardown() {
  echo "Cleaning up test environment..."
  rm -f "$OUTPUT_FILE_FOLDER/output_*.txt"
}

@test "Add $FIRST_ADMIN_ADDRESS as admin (calldata)" {
  run $CLI_PATH user admin add-pending-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$FIRST_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$ACCOUNT_PRIVATE_KEY" \
    --output-type "calldata" \
    --output-file "$OUTPUT_FILE_FOLDER/output_add_first_admin.txt"

  [ "$status" -eq 0 ]
  [[ -s "$OUTPUT_FILE_FOLDER/output_add_first_admin.txt" ]]
}

@test "Add $FIRST_ADMIN_ADDRESS as admin (broadcast)" {
  run $CLI_PATH user admin add-pending-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$FIRST_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$ACCOUNT_PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Verify $FIRST_ADMIN_ADDRESS is a pending admin" {
  run $CLI_PATH user admin is-pending-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --pending-admin-address "$FIRST_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  echo "$output"

  [ "$status" -eq 0 ]
  [[ "$output" == *"Address provided is a pending admin"* ]]
}

@test "Remove $FIRST_ADMIN_ADDRESS as pending admin (calldata)" {
  run $CLI_PATH user admin remove-pending-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$FIRST_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$ACCOUNT_PRIVATE_KEY" \
    --output-type "calldata" \
    --output-file "$OUTPUT_FILE_FOLDER/output_remove_pending_first_admin.txt"

  echo "$output"

  [ "$status" -eq 0 ]
  [[ -s "$OUTPUT_FILE_FOLDER/output_remove_pending_first_admin.txt" ]]
}

@test "Remove $FIRST_ADMIN_ADDRESS as pending admin (broadcast)" {
  run $CLI_PATH user admin remove-pending-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$FIRST_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$ACCOUNT_PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Add $FIRST_ADMIN_ADDRESS as admin after removal (broadcast)" {
  run $CLI_PATH user admin add-pending-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$FIRST_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$ACCOUNT_PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Accept $FIRST_ADMIN_ADDRESS as admin (calldata)" {
  run $CLI_PATH user admin accept-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --caller-address "$FIRST_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$FIRST_ADMIN_PRIVATE_KEY" \
    --output-type "calldata" \
    --output-file "$OUTPUT_FILE_FOLDER/output_accept_first_admin.txt"

  [ "$status" -eq 0 ]
  [[ -s "$OUTPUT_FILE_FOLDER/output_accept_first_admin.txt" ]]
}

@test "Accept $FIRST_ADMIN_ADDRESS as admin (broadcast)" {
  run $CLI_PATH user admin accept-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$FIRST_ADMIN_PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Add $SECOND_ADMIN_ADDRESS as admin (calldata)" {
  run $CLI_PATH user admin add-pending-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$SECOND_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$FIRST_ADMIN_PRIVATE_KEY" \
    --output-type "calldata" \
    --output-file "$OUTPUT_FILE_FOLDER/output_add_second_admin.txt"

  [ "$status" -eq 0 ]
  [[ -s "$OUTPUT_FILE_FOLDER/output_add_second_admin.txt" ]]
}

@test "Add $SECOND_ADMIN_ADDRESS as admin (broadcast)" {
  run $CLI_PATH user admin add-pending-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$SECOND_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$FIRST_ADMIN_PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Verify $SECOND_ADMIN_ADDRESS is listed as a pending admin" {
  run $CLI_PATH user admin list-pending-admins \
    --account-address "$ACCOUNT_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \

  [ "$status" -eq 0 ]
  [[ "$output" != *"$ACCOUNT_ADDRESS"* ]]
  [[ "$output" != *"$FIRST_ADMIN_ADDRESS"* ]]
  [[ "$output" == *"$SECOND_ADMIN_ADDRESS"* ]]
}

@test "Verify $SECOND_ADMIN_ADDRESS is a pending admin" {
  run $CLI_PATH user admin is-pending-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --pending-admin-address "$SECOND_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  echo "$output"

  [ "$status" -eq 0 ]
  [[ "$output" == *"Address provided is a pending admin"* ]]
}

@test "Accept $SECOND_ADMIN_ADDRESS as admin (calldata)" {
  run $CLI_PATH user admin accept-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --caller-address "$SECOND_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$SECOND_ADMIN_PRIVATE_KEY" \
    --output-type "calldata" \
    --output-file "$OUTPUT_FILE_FOLDER/output_accept_second_admin.txt"

  [ "$status" -eq 0 ]
  [[ -s "$OUTPUT_FILE_FOLDER/output_accept_second_admin.txt" ]]
}

@test "Accept $SECOND_ADMIN_ADDRESS as admin (broadcast)" {
  run $CLI_PATH user admin accept-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$SECOND_ADMIN_PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Verify all three admins are listed after acceptance" {
  run $CLI_PATH user admin list-admins \
    --account-address "$ACCOUNT_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  echo "$output"

  [ "$status" -eq 0 ]
  [[ "$output" == *"$ACCOUNT_ADDRESS"* ]]
  [[ "$output" == *"$FIRST_ADMIN_ADDRESS"* ]]
  [[ "$output" == *"$SECOND_ADMIN_ADDRESS"* ]]
}

@test "Remove $SECOND_ADMIN_ADDRESS as admin (calldata)" {
  run $CLI_PATH user admin remove-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$SECOND_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$FIRST_ADMIN_PRIVATE_KEY" \
    --output-type "calldata" \
    --output-file "$OUTPUT_FILE_FOLDER/output_remove_second_admin.txt"

  [ "$status" -eq 0 ]
  [[ -s "$OUTPUT_FILE_FOLDER/output_remove_second_admin.txt" ]]
}

@test "Remove $SECOND_ADMIN_ADDRESS as admin (broadcast)" {
  run $CLI_PATH user admin remove-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$SECOND_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$FIRST_ADMIN_PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Remove $FIRST_ADMIN_ADDRESS as admin (calldata)" {
  run $CLI_PATH user admin remove-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$FIRST_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$ACCOUNT_PRIVATE_KEY" \
    --output-type "calldata" \
    --output-file "$OUTPUT_FILE_FOLDER/output_remove_first_admin.txt"

  [ "$status" -eq 0 ]
  [[ -s "$OUTPUT_FILE_FOLDER/output_remove_first_admin.txt" ]]
}

@test "Remove $FIRST_ADMIN_ADDRESS as admin (broadcast)" {
  run $CLI_PATH user admin remove-admin \
    --account-address "$ACCOUNT_ADDRESS" \
    --admin-address "$FIRST_ADMIN_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK" \
    --ecdsa-private-key "$ACCOUNT_PRIVATE_KEY" \
    --broadcast

  [ "$status" -eq 0 ]
}

@test "Verify only root admins remains" {
  run $CLI_PATH user admin list-admins \
    --account-address "$ACCOUNT_ADDRESS" \
    --permission-controller-address "$PERMISSION_CONTROLLER_ADDRESS" \
    --eth-rpc-url "$RPC_URL" \
    --network "$NETWORK"

  [ "$status" -eq 0 ]
  [[ "$output" == *"$ACCOUNT_ADDRESS"* ]]
  [[ "$output" != *"$FIRST_ADMIN_ADDRESS"* ]]
  [[ "$output" != *"$SECOND_ADMIN_ADDRESS"* ]]
}
