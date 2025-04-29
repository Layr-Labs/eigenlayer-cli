#!/bin/bash

echo "Starting User command integration tests."

export $(cat .env | xargs)

any_test_failures=0

run_bats_test() {
  local test_file=$1
  bats "$test_file"
  if [ $? -ne 0 ]; then
    any_test_failures=1
  fi
}

run_bats_test "admin/rotate_admin_tests.sh"
run_bats_test "admin/admin_lifecycle_tests.sh"
run_bats_test "appointee/appointee_lifecycle_tests.sh"
run_bats_test "appointee/appointee_output_tests.sh"

rm -rf output/

if [ $any_test_failures -eq 0 ]; then
  echo "All tests passed."
else
  echo "Some tests failed."
fi

exit $any_test_failures