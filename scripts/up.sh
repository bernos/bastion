#!/usr/bin/env bash
set -euo pipefail

VPC_STACK="${VPC_STACK:-bastion-test-vpc}"
BASTION_NAME="${BASTION_NAME:-bastion-test}"
REGION="${REGION:-ap-southeast-2}"
SSH_USER="${SSH_USER:-ec2-user}"
OWNER="${OWNER:-bernos}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BASTION="${SCRIPT_DIR}/../bastion"

# Resolve VPC and subnet from the test VPC stack
echo "Fetching VPC stack outputs from ${VPC_STACK}..." >&2

stack_outputs=$(aws cloudformation describe-stacks \
  --stack-name "$VPC_STACK" \
  --query 'Stacks[0].Outputs' \
  --output json | jq 'map({(.OutputKey): .OutputValue}) | add')

vpc_id=$(echo "$stack_outputs" | jq -r '.VpcId')
subnet_id=$(echo "$stack_outputs" | jq -r '.PrivateSubnet1Id')

echo "VPC: ${vpc_id}  Subnet: ${subnet_id}" >&2

# Deploy (or update) the bastion and upload the SSH public key.
# The uploaded key expires after 60s — SSH must follow immediately.
echo "Running bastion up..." >&2

bastion_json=$("$BASTION" up \
  --name "$BASTION_NAME" \
  --vpc-id "$vpc_id" \
  --subnet-id "$subnet_id" \
  --owner "$OWNER" \
  --region "$REGION")

instance_id=$(echo "$bastion_json" | jq -r '.instanceId')

echo "Instance: ${instance_id}" >&2
# echo "Connecting..." >&2

# ssh \
#   -o StrictHostKeyChecking=no \
#   -o ProxyCommand="aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters portNumber=%p" \
#   "${SSH_USER}@${instance_id}"
