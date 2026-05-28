## ADDED Requirements

### Requirement: Web server is deployed in a private subnet with no public access
The test VPC stack SHALL include an EC2 instance deployed into a private subnet with no public IP address, no internet gateway route, and no inbound SSH rule. The instance SHALL use instance type `t4g.micro` and the latest Amazon Linux 2023 arm64 AMI resolved via SSM parameter at deploy time.

#### Scenario: Instance has no public IP
- **WHEN** the CloudFormation stack is deployed
- **THEN** the web server EC2 instance has `MapPublicIpOnLaunch: false` and no Elastic IP association

#### Scenario: Instance is reachable only from within the VPC
- **WHEN** an HTTP request is made to the instance's private IP on port 80
- **THEN** the request succeeds from within the VPC and is unreachable from outside the VPC

### Requirement: Web server serves a static HTML page on port 80
The web server SHALL serve a static HTML page on port 80 using Python's built-in HTTP server (`python3 -m http.server 80`) managed by a systemd unit. The page SHALL remain available after an instance reboot without manual intervention.

#### Scenario: Static page is served on port 80
- **WHEN** an HTTP GET request is made to the instance's private IP on port 80
- **THEN** a 200 response is returned with an HTML body

#### Scenario: Server survives reboot
- **WHEN** the EC2 instance is rebooted
- **THEN** the HTTP server starts automatically via systemd and serves the page without manual intervention

### Requirement: Web server security group allows HTTP from VPC CIDR only
A dedicated security group SHALL be attached to the web server instance allowing TCP ingress on port 80 from the VPC CIDR block only. No other ingress rules SHALL be present.

#### Scenario: HTTP ingress from VPC CIDR is allowed
- **WHEN** an HTTP request on port 80 originates from an address within the VPC CIDR
- **THEN** the security group permits the connection

#### Scenario: No SSH ingress rule
- **WHEN** the security group rules are inspected
- **THEN** there is no inbound rule permitting TCP port 22

### Requirement: Web server is reachable via a stable private DNS name
The stack SHALL create a Route 53 private hosted zone (`bastion-test.internal`) associated with the VPC and an A record resolving `webserver.bastion-test.internal` to the instance's private IP address.

#### Scenario: Private DNS name resolves within the VPC
- **WHEN** a DNS lookup for `webserver.bastion-test.internal` is performed from within the VPC
- **THEN** it resolves to the web server's private IP address

#### Scenario: DNS name is not resolvable outside the VPC
- **WHEN** a DNS lookup for `webserver.bastion-test.internal` is performed from outside the VPC
- **THEN** the name does not resolve (private hosted zone is VPC-scoped)

### Requirement: Stack exports web server connection details as outputs
The CloudFormation stack SHALL export two new outputs: `WebServerUrl` containing the full HTTP URL (`http://webserver.bastion-test.internal`) and `WebServerPrivateIp` containing the instance's private IP address.

#### Scenario: Outputs are present after stack deployment
- **WHEN** the stack is successfully deployed or updated
- **THEN** `WebServerUrl` and `WebServerPrivateIp` are present in the stack outputs
