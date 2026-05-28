## MODIFIED Requirements

### Requirement: Stack resources are tagged with Name and Owner plus any user-supplied tags
The EC2 instance and security group SHALL be tagged with `Name` (set to `BastionName`) and `Owner` (set to the `Owner` parameter) via resource-level tags in the CloudFormation template. In addition, any user-supplied tags SHALL be applied as CloudFormation stack-level tags on the `CreateChangeSet` call; CloudFormation SHALL propagate these stack-level tags to all supported resources in the stack.

#### Scenario: Built-in tags applied on stack creation
- **WHEN** the CloudFormation stack is deployed with a BastionName and Owner
- **THEN** the EC2 instance and security group have matching `Name` and `Owner` tags

#### Scenario: User-supplied tags applied as stack-level tags
- **WHEN** `DeployBastionInput.Tags` contains `{"env": "prod", "team": "platform"}`
- **THEN** the CloudFormation `CreateChangeSet` call includes stack-level tags with those key-value pairs and the EC2 instance receives both the template tags and the stack-level tags

#### Scenario: No user tags — stack-level tags list is empty
- **WHEN** `DeployBastionInput.Tags` is nil or empty
- **THEN** the CloudFormation `CreateChangeSet` call receives an empty stack-level `Tags` slice and only the template-defined `Name` and `Owner` tags are present on resources
