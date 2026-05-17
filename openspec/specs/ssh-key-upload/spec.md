### Requirement: `up` command accepts an optional public key path
The `up` command SHALL accept a `--public-key-path` flag, a `BASTION_PUBLIC_KEY_PATH` environment variable, and a `public-key-path` key in the config YAML. If none of these are provided the command SHALL deploy the bastion without uploading any SSH key.

#### Scenario: public-key-path resolved from CLI flag
- **WHEN** `bastion up` is invoked with `--public-key-path /path/to/key.pub` and all other required flags
- **THEN** the value `/path/to/key.pub` is used as the public key path

#### Scenario: public-key-path resolved from environment variable
- **WHEN** `BASTION_PUBLIC_KEY_PATH=/path/to/key.pub` is set in the environment and `bastion up` is invoked without `--public-key-path`
- **THEN** the value `/path/to/key.pub` is used as the public key path

#### Scenario: public-key-path omitted — deploy proceeds normally
- **WHEN** `bastion up` is invoked without `--public-key-path` and without `BASTION_PUBLIC_KEY_PATH`
- **THEN** the bastion is deployed and the command exits successfully without calling SendSSHPublicKey

### Requirement: Public key file is validated before deploy
If `public-key-path` is set, the `up` command SHALL read and validate the file contents before initiating the CloudFormation deploy. Validation SHALL: (1) return an error if the file does not exist or cannot be read; (2) return an error if the content contains the substring `PRIVATE KEY`, indicating a PEM private key was supplied; (3) return an error if `golang.org/x/crypto/ssh.ParseAuthorizedKey` fails to parse the content as a valid SSH public key. All errors SHALL be returned before any CloudFormation operation begins.

#### Scenario: File not found before deploy
- **WHEN** `--public-key-path /nonexistent/key.pub` is supplied
- **THEN** the command returns a non-zero exit code with an error message before any CloudFormation operation begins

#### Scenario: Private key file rejected with clear error
- **WHEN** `--public-key-path` points to a file whose content contains `PRIVATE KEY`
- **THEN** the command returns a non-zero exit code with an error message indicating a private key was detected, before any CloudFormation operation begins

#### Scenario: Malformed public key file rejected
- **WHEN** `--public-key-path` points to a readable file that does not contain a valid SSH public key in authorized-keys format
- **THEN** the command returns a non-zero exit code with a parse error before any CloudFormation operation begins

#### Scenario: Valid public key file accepted
- **WHEN** `--public-key-path` points to a readable file containing a valid SSH public key in authorized-keys format
- **THEN** the file is read without error and the deploy proceeds

### Requirement: SSM agent readiness is confirmed before key upload
If `PublicKeyContent` is non-empty, `BastionService.DeployBastion` SHALL poll `ssm:DescribeInstanceInformation` (filtered by instance ID) every 10 seconds until `PingStatus` is `Online`, before calling `SendSSHPublicKey`. The poll SHALL time out after 2 minutes and return an error if the agent does not come online in that window.

#### Scenario: SSM agent comes online before timeout
- **WHEN** `DeployBastionInput.PublicKeyContent` is non-empty and the SSM agent reports `PingStatus == Online` within 2 minutes
- **THEN** `DeployBastion` proceeds to call `SendSSHPublicKey`

#### Scenario: SSM agent does not come online within timeout
- **WHEN** `DeployBastionInput.PublicKeyContent` is non-empty and the SSM agent does not report `PingStatus == Online` within 2 minutes
- **THEN** `DeployBastion` returns an error without calling `SendSSHPublicKey`

#### Scenario: SSM check skipped when no key is being uploaded
- **WHEN** `DeployBastionInput.PublicKeyContent` is empty
- **THEN** `DescribeInstanceInformation` is NOT called

### Requirement: SSH public key is uploaded after SSM agent is confirmed online
If `public-key-path` is set, `BastionService.DeployBastion` SHALL call `ec2instanceconnect:SendSSHPublicKey` with the deployed instance ID, the instance availability zone, OS user `ec2-user`, and the public key content, after the SSM agent readiness check passes.

#### Scenario: Key uploaded after deploy completes
- **WHEN** `DeployBastionInput.PublicKeyContent` is non-empty and the CloudFormation deploy succeeds
- **THEN** `SendSSHPublicKey` is called once with `InstanceId` equal to the deployed instance ID, `AvailabilityZone` equal to the deployed AZ, `InstanceOSUser` equal to `ec2-user`, and `SSHPublicKey` equal to the supplied key content

#### Scenario: Key upload failure returns an error
- **WHEN** `SendSSHPublicKey` returns an error
- **THEN** `BastionService.DeployBastion` returns that error and the `up` command exits with a non-zero status

#### Scenario: Key not uploaded when PublicKeyContent is empty
- **WHEN** `DeployBastionInput.PublicKeyContent` is empty
- **THEN** `SendSSHPublicKey` is NOT called
