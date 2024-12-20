echo "Starting User command integration tests."

export $(cat .env | xargs)

bats "admin/admin_lifecycle_tests.sh"
bats "appointee/appointee_lifecycle_tests.sh"
bats "appointee/appointee_output_tests.sh"

rm -rf output/