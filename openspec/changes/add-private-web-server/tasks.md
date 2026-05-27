## 1. EC2 Instance Resources

- [x] 1.1 Add `WebServerAmiId` parameter (`AWS::SSM::Parameter::Value<AWS::EC2::Image::Id>`) with default `/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64`
- [x] 1.2 Add `WebServerSecurityGroup` resource — TCP port 80 ingress from `!Ref VpcCidr`, attached to `!Ref VPC`
- [x] 1.3 Add `WebServerInstance` resource — `t4g.micro`, `!Ref PrivateSubnet1`, `!Ref WebServerAmiId`, `!Ref WebServerSecurityGroup`, no public IP

## 2. UserData — Static HTTP Server

- [x] 2.1 Write UserData script that creates `/var/www/index.html` with a self-describing static HTML page (mentions the private DNS name and that it is only reachable via the bastion)
- [x] 2.2 Write UserData script that creates `/etc/systemd/system/static-web.service` — `python3 -m http.server 80` with `WorkingDirectory=/var/www`, `Restart=always`
- [x] 2.3 Write UserData script commands: `systemctl daemon-reload`, `systemctl enable static-web`, `systemctl start static-web`

## 3. Route 53 Private DNS

- [x] 3.1 Add `WebServerHostedZone` resource — `AWS::Route53::HostedZone`, name `bastion-test.internal`, VPC association `!Ref VPC` with `!Ref AWS::Region`
- [x] 3.2 Add `WebServerDnsRecord` resource — `AWS::Route53::RecordSet`, name `webserver.bastion-test.internal`, type A, TTL 300, value `!GetAtt WebServerInstance.PrivateIp`, hosted zone `!Ref WebServerHostedZone`

## 4. Stack Outputs

- [x] 4.1 Add `WebServerUrl` output — value `http://webserver.bastion-test.internal`, description explaining it is only reachable from within the VPC via the bastion
- [x] 4.2 Add `WebServerPrivateIp` output — value `!GetAtt WebServerInstance.PrivateIp`
