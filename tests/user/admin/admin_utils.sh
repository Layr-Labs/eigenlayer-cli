#!/bin/bash

broadcast_add_pending_admin() {
  local account="$1"
  local admin="$2"
  local private_key="$3"
  local permission_controller_address="$4"
  local rpc_url="$5"
  local network="$6"

  echo "Broadcasting add-pending-admin for $admin to $account..."
  $CLI_PATH user admin add-pending-admin \
    --account-address "$account" \
    --admin-address "$admin" \
    --permission-controller-address "$permission_controller_address" \
    --eth-rpc-url "$rpc_url" \
    --network "$network" \
    --ecdsa-private-key "$private_key" \
    --broadcast
}

broadcast_accept_admin() {
  local account="$1"
  local private_key="$2"
  local permission_controller_address="$3"
  local rpc_url="$4"
  local network="$5"

  echo "Broadcasting accept-admin for $account..."
  $CLI_PATH user admin accept-admin \
    --account-address "$account" \
    --permission-controller-address "$permission_controller_address" \
    --eth-rpc-url "$rpc_url" \
    --network "$network" \
    --ecdsa-private-key "$private_key" \
    --broadcast
}

conditional_add_admin() {
  local account="$1"
  local admin="$2"
  local private_key="$3"
  local permission_controller_address="$4"
  local rpc_url="$5"
  local network="$6"
  local output_add_file="$7"
  local output_accept_file="$8"
  local is_admin_string="is an admin"

  echo "Checking if $admin is an admin for $account..."
  local is_admin_output=$($CLI_PATH user admin is-admin \
    --account-address "$account" \
    --admin-address "$admin" \
    --permission-controller-address "$permission_controller_address" \
    --eth-rpc-url "$rpc_url" \
    --network "$network" 2>&1)

  if echo "$is_admin_output" | grep -q "$is_admin_string"; then
    echo "$admin is already an admin for $account."
  else
    echo "$admin is not an admin for $account. Adding as an admin..."
    broadcast_add_pending_admin "$account" "$admin" "$private_key" "$permission_controller_address" "$rpc_url" "$network" "$output_add_file"
    broadcast_accept_admin "$account" "$admin" "$private_key" "$permission_controller_address" "$rpc_url" "$network" "$output_accept_file"
  fi
}

