#!/usr/bin/env bash
set -euo pipefail

STACK_NAME="${STACK_NAME:-bastion-test-vpc}"
TEMPLATE="$(cd "$(dirname "$0")" && pwd)/test-vpc.yaml"

aws cloudformation deploy \
  --stack-name "$STACK_NAME" \
  --template-file "$TEMPLATE" \
  >&2

aws cloudformation describe-stacks \
  --stack-name "$STACK_NAME" \
  --query 'Stacks[0].Outputs' \
  --output json \
| jq 'map({(.OutputKey): .OutputValue}) | add'
