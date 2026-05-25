## Why

The test VPC stack (`cloudformation/test-vpc.yaml`) exists to give users a realistic private network for exercising the bastion CLI, but it currently contains no workload to reach. Adding a private web server gives a concrete target that demonstrates the value of bastion access patterns (SSM port forwarding, SSH tunnelling) against a real HTTP endpoint reachable only from inside the VPC.

## What Changes

- Add a `t4g.micro` EC2 instance running a static HTTP server (Python built-in, port 80) to `cloudformation/test-vpc.yaml`
- Add a Route 53 private hosted zone (`bastion-test.internal`) with an A record resolving `webserver.bastion-test.internal` to the instance's private IP
- Add a security group allowing HTTP (port 80) ingress from within the VPC CIDR
- Add new stack outputs: `WebServerUrl` and `WebServerPrivateIp`
- No changes to the bastion CLI itself; this is purely test infrastructure

## Capabilities

### New Capabilities

- `test-private-web-server`: A private EC2 web server in the test VPC — a static HTTP target reachable only from within the VPC, used to demonstrate bastion access patterns

### Modified Capabilities

_(none — no existing spec requirements are changing)_

## Impact

- `cloudformation/test-vpc.yaml`: new resources (EC2 instance, security group, Route 53 hosted zone + record set, AMI parameter)
- `cloudformation/deploy-test-vpc.sh`: no changes needed; new outputs will appear automatically
- No Go code changes
- Existing stack deployments will require a CloudFormation update; no destructive changes to existing resources
