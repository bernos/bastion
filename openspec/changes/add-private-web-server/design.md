## Context

The test VPC (`cloudformation/test-vpc.yaml`) is a fully private network: no internet gateway, no NAT gateway, no public subnets. It exists so users can run `bastion up` against a realistic isolated environment. Currently the VPC has no workload inside it, so there is nothing to demonstrate reaching via the bastion.

The VPC already has SSM interface endpoints (SSM, SSMMessages, EC2Messages) so the bastion host itself can reach the AWS control plane. Any new EC2 instance in the VPC inherits this — it can receive SSM sessions but cannot reach the internet.

## Goals / Non-Goals

**Goals:**
- Add a private HTTP server that is reachable only from within the VPC
- Give it a human-readable private DNS name (`webserver.bastion-test.internal`) suitable for use in documentation and examples
- Keep the server alive across reboots without manual intervention
- Minimise cost (smallest viable instance, no additional paid endpoints)

**Non-Goals:**
- HTTPS/TLS termination — a plain HTTP server is sufficient for the demo
- Any IAM role or SSM access on the web server — it is a dumb target, not a managed instance
- Changes to the bastion CLI or Go codebase

## Decisions

### HTTP server: Python built-in over installing httpd

**Decision**: Use `python3 -m http.server 80` managed by a systemd unit written in UserData.

**Rationale**: The VPC has no internet route and no S3 Gateway endpoint, so `dnf install httpd` would fail silently or hang. Python 3 ships pre-installed on Amazon Linux 2023 — no package manager needed. For a static demo page, Python's built-in server is fully sufficient.

**Alternative considered**: Add an S3 Gateway endpoint to the VPC (free) and then `dnf install httpd`. Rejected because it adds resources and complexity beyond what this change needs. If the stack later grows to include instances that need S3 access, the gateway endpoint can be added then.

### Systemd unit over nohup/background process

**Decision**: Write a `/etc/systemd/system/static-web.service` unit in UserData and `systemctl enable + start` it.

**Rationale**: A backgrounded process dies on reboot. Systemd units survive reboots and restart automatically on failure. The unit file is written inline in UserData — no external dependency.

### Route 53 private hosted zone for DNS

**Decision**: Create a `Route53::HostedZone` (private, associated with the VPC) named `bastion-test.internal`, with an A record for `webserver.bastion-test.internal` pointing to the instance's private IP.

**Rationale**: The auto-assigned private DNS name (`ip-10-0-1-x.region.compute.internal`) is opaque and changes on stop/start. A stable, readable name is far better for documentation and `--parameters` flags in SSM port-forwarding examples.

**Alternative considered**: Elastic IP — not applicable for private instances with no internet route.

**Trade-off**: Route 53 private hosted zone costs $0.50/month. Acceptable for a persistent test stack.

### AMI: SSM parameter for latest AL2023 arm64

**Decision**: Use `AWS::SSM::Parameter::Value<AWS::EC2::Image::Id>` with default path `/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64`.

**Rationale**: Avoids hardcoded AMI IDs that rot over time. CloudFormation resolves the latest valid AMI at deploy time. arm64 is required because the instance type is `t4g.micro` (Graviton).

### Instance type: t4g.micro

**Decision**: `t4g.micro` (Graviton2, ARM64).

**Rationale**: Cheapest general-purpose instance type that can serve HTTP. Graviton instances are ~20% cheaper than x86 equivalents. Serving a static HTML file has negligible CPU and memory requirements.

## Risks / Trade-offs

- **Python http.server is single-threaded** → Fine for a demo with one or two concurrent connections; not suitable for production load. This is an accepted limitation of the demo context.
- **Instance costs money while the stack is up** → t4g.micro is ~$0.0084/hr. Users should tear down the stack when not needed. The deploy script already handles this; no additional guidance needed.
- **Private IP changes on stop/start** → The Route 53 A record is set at instance launch via `!GetAtt WebServerInstance.PrivateIp`. If the instance is stopped and restarted, its IP may change and the DNS record will be stale. Mitigation: the instance should stay running; this is a test stack, not a long-lived environment.
- **UserData runs only on first boot** → If someone stops and starts the instance (not reboots), the systemd unit persists, so the server continues to work. If the instance is replaced (e.g., stack update replacing the resource), UserData runs again and the DNS record is updated.

## Migration Plan

1. Run `cloudformation/deploy-test-vpc.sh` — CloudFormation performs a stack update adding all new resources
2. No existing resources are modified or deleted
3. Rollback: if the update fails, CloudFormation automatically rolls back; no manual cleanup needed
4. New outputs (`WebServerUrl`, `WebServerPrivateIp`) appear after successful update
