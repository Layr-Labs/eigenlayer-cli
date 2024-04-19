#!/bin/bash

# Run with ./scripts/build.sh <optional_version>
if ! [[ "$0" =~ scripts/build.sh ]]; then
  echo "must be run from repository root"
  exit 1
fi

go build -ldflags "-X main.Version=$1" -v -o bin/eigenlayer cmd/eigenlayer/main.go