package cloudformationservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
)

type CloudFormationService interface {
	Deploy(ctx context.Context, input *DeployInput, opts ...func(*DeployOptions)) (*DeployOutput, error)
	StackExists(ctx context.Context, stackName string) (bool, error)
}

type cloudFormationService struct {
	client CloudFormationClient
}

type CloudFormationClient interface {
	CreateChangeSet(context.Context, *cloudformation.CreateChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error)
	CreateStack(context.Context, *cloudformation.CreateStackInput, ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	DeleteChangeSet(context.Context, *cloudformation.DeleteChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.DeleteChangeSetOutput, error)
	DescribeChangeSet(context.Context, *cloudformation.DescribeChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error)
	DescribeStacks(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	ExecuteChangeSet(context.Context, *cloudformation.ExecuteChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.ExecuteChangeSetOutput, error)
}

var _ CloudFormationService = (*cloudFormationService)(nil)

func New(client *cloudformation.Client) CloudFormationService {
	return &cloudFormationService{
		client: client,
	}
}

type DeployInput struct {
	// The name of the change set. The name must be unique among all change sets that
	// are associated with the specified stack.
	//
	// A change set name can contain only alphanumeric, case sensitive characters, and
	// hyphens. It must start with an alphabetical character and can't exceed 128
	// characters.
	//
	// This member is required.
	DeploymentName *string

	// The name or the unique ID of the stack for which you are creating a change set.
	// CloudFormation generates the change set by comparing this stack's information
	// with the information that you submit, such as a modified template or different
	// parameter input values.
	//
	// This member is required.
	StackName *string

	// In some cases, you must explicitly acknowledge that your stack template
	// contains certain capabilities in order for CloudFormation to create the stack.
	//
	//   - CAPABILITY_IAM and CAPABILITY_NAMED_IAM
	//
	// Some stack templates might include resources that can affect permissions in
	//   your Amazon Web Services account, for example, by creating new IAM users. For
	//   those stacks, you must explicitly acknowledge this by specifying one of these
	//   capabilities.
	//
	// The following IAM resources require you to specify either the CAPABILITY_IAM or
	//   CAPABILITY_NAMED_IAM capability.
	//
	//   - If you have IAM resources, you can specify either capability.
	//
	//   - If you have IAM resources with custom names, you must specify
	//   CAPABILITY_NAMED_IAM .
	//
	//   - If you don't specify either of these capabilities, CloudFormation returns
	//   an InsufficientCapabilities error.
	//
	// If your stack template contains these resources, we suggest that you review all
	//   permissions associated with them and edit their permissions if necessary.
	//
	// [AWS::IAM::AccessKey]
	//
	// [AWS::IAM::Group]
	//
	// [AWS::IAM::InstanceProfile]
	//
	// [AWS::IAM::ManagedPolicy]
	//
	// [AWS::IAM::Policy]
	//
	// [AWS::IAM::Role]
	//
	// [AWS::IAM::User]
	//
	// [AWS::IAM::UserToGroupAddition]
	//
	// For more information, see [Acknowledging IAM resources in CloudFormation templates].
	//
	//   - CAPABILITY_AUTO_EXPAND
	//
	// Some template contain macros. Macros perform custom processing on templates;
	//   this can include simple actions like find-and-replace operations, all the way to
	//   extensive transformations of entire templates. Because of this, users typically
	//   create a change set from the processed template, so that they can review the
	//   changes resulting from the macros before actually creating the stack. If your
	//   stack template contains one or more macros, and you choose to create a stack
	//   directly from the processed template, without first reviewing the resulting
	//   changes in a change set, you must acknowledge this capability. This includes the
	//   [AWS::Include]and [AWS::Serverless]transforms, which are macros hosted by CloudFormation.
	//
	// This capacity doesn't apply to creating change sets, and specifying it when
	//   creating change sets has no effect.
	//
	// If you want to create a stack from a stack template that contains macros and
	//   nested stacks, you must create or update the stack directly from the template
	//   using the CreateStackor UpdateStackaction, and specifying this capability.
	//
	// For more information about macros, see [Perform custom processing on CloudFormation templates with template macros].
	//
	// Only one of the Capabilities and ResourceType parameters can be specified.
	//
	// [AWS::IAM::ManagedPolicy]: https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-iam-managedpolicy.html
	// [AWS::IAM::AccessKey]: https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-iam-accesskey.html
	// [AWS::Include]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/transform-aws-include.html
	// [AWS::IAM::User]: https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-iam-user.html
	// [AWS::IAM::InstanceProfile]: https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-iam-instanceprofile.html
	// [Acknowledging IAM resources in CloudFormation templates]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/control-access-with-iam.html#using-iam-capabilities
	// [Perform custom processing on CloudFormation templates with template macros]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/template-macros.html
	// [AWS::IAM::Policy]: https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-iam-policy.html
	// [AWS::IAM::Group]: https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-iam-group.html
	// [AWS::IAM::UserToGroupAddition]: https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-iam-usertogroupaddition.html
	// [AWS::IAM::Role]: https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-resource-iam-role.html
	// [AWS::Serverless]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/transform-aws-serverless.html
	Capabilities []types.Capability

	// A unique identifier for this CreateChangeSet request. Specify this token if you
	// plan to retry requests so that CloudFormation knows that you're not attempting
	// to create another change set with the same name. You might retry CreateChangeSet
	// requests to ensure that CloudFormation successfully received them.
	ClientToken *string

	// Determines how CloudFormation handles configuration drift during deployment.
	//
	//   - REVERT_DRIFT – Creates a drift-aware change set that brings actual resource
	//   states in line with template definitions. Provides a three-way comparison
	//   between actual state, previous deployment state, and desired state.
	//
	// For more information, see [Using drift-aware change sets] in the CloudFormation User Guide.
	//
	// [Using drift-aware change sets]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/drift-aware-change-sets.html
	DeploymentMode types.DeploymentMode

	// A description to help you identify this change set.
	Description *string

	// Preserves the state of previously provisioned resources when an operation
	// fails. This parameter can't be specified when the OnStackFailure parameter to
	// the [CreateChangeSet]API operation was specified.
	//
	//   - True - if the stack creation fails, do nothing. This is equivalent to
	//   specifying DO_NOTHING for the OnStackFailure parameter to the [CreateChangeSet]API operation.
	//
	//   - False - if the stack creation fails, roll back the stack. This is equivalent
	//   to specifying ROLLBACK for the OnStackFailure parameter to the [CreateChangeSet]API operation.
	//
	// Default: True
	//
	// [CreateChangeSet]: https://docs.aws.amazon.com/AWSCloudFormation/latest/APIReference/API_CreateChangeSet.html
	DisableRollback *bool

	// Indicates if the change set auto-imports resources that already exist. For more
	// information, see [Import Amazon Web Services resources into a CloudFormation stack automatically]in the CloudFormation User Guide.
	//
	// This parameter can only import resources that have custom names in templates.
	// For more information, see [name type]in the CloudFormation User Guide. To import resources
	// that do not accept custom names, such as EC2 instances, use the
	// ResourcesToImport parameter instead.
	//
	// [Import Amazon Web Services resources into a CloudFormation stack automatically]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/import-resources-automatically.html
	// [name type]: https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/aws-properties-name.html
	ImportExistingResources *bool

	// Creates a change set for the all nested stacks specified in the template. The
	// default behavior of this action is set to False . To include nested sets in a
	// change set, specify True .
	IncludeNestedStacks *bool

	// The Amazon Resource Names (ARNs) of Amazon SNS topics that CloudFormation
	// associates with the stack. To remove all associated notification topics, specify
	// an empty list.
	NotificationARNs []string

	// Determines what action will be taken if stack creation fails. If this parameter
	// is specified, the DisableRollback parameter to the [ExecuteChangeSet] API operation must not be
	// specified. This must be one of these values:
	//
	//   - DELETE - Deletes the change set if the stack creation fails. This is only
	//   valid when the ChangeSetType parameter is set to CREATE . If the deletion of
	//   the stack fails, the status of the stack is DELETE_FAILED .
	//
	//   - DO_NOTHING - if the stack creation fails, do nothing. This is equivalent to
	//   specifying true for the DisableRollback parameter to the [ExecuteChangeSet]API operation.
	//
	//   - ROLLBACK - if the stack creation fails, roll back the stack. This is
	//   equivalent to specifying false for the DisableRollback parameter to the [ExecuteChangeSet]API
	//   operation.
	//
	// For nested stacks, when the OnStackFailure parameter is set to DELETE for the
	// change set for the parent stack, any failure in a child stack will cause the
	// parent stack creation to fail and all stacks to be deleted.
	//
	// [ExecuteChangeSet]: https://docs.aws.amazon.com/AWSCloudFormation/latest/APIReference/API_ExecuteChangeSet.html
	OnStackFailure types.OnStackFailure

	// A list of Parameter structures that specify input parameters for the change
	// set. For more information, see the Parameterdata type.
	Parameters []types.Parameter

	// Specifies which resource types you can work with, such as AWS::EC2::Instance or
	// Custom::MyCustomInstance .
	//
	// If the list of resource types doesn't include a resource type that you're
	// updating, the stack update fails. By default, CloudFormation grants permissions
	// to all resource types. IAM uses this parameter for condition keys in IAM
	// policies for CloudFormation. For more information, see [Control CloudFormation access with Identity and Access Management]in the CloudFormation
	// User Guide.
	//
	// Only one of the Capabilities and ResourceType parameters can be specified.
	//
	// [Control CloudFormation access with Identity and Access Management]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/control-access-with-iam.html
	ResourceTypes []string

	// The resources to import into your stack.
	ResourcesToImport []types.ResourceToImport

	// When set to true , newly created resources are deleted when the operation rolls
	// back. This includes newly created resources marked with a deletion policy of
	// Retain .
	//
	// Default: false
	RetainExceptOnCreate *bool

	// The Amazon Resource Name (ARN) of an IAM role that CloudFormation assumes when
	// executing the change set. CloudFormation uses the role's credentials to make
	// calls on your behalf. CloudFormation uses this role for all future operations on
	// the stack. Provided that users have permission to operate on the stack,
	// CloudFormation uses this role even if the users don't have permission to pass
	// it. Ensure that the role grants least permission.
	//
	// If you don't specify a value, CloudFormation uses the role that was previously
	// associated with the stack. If no role is available, CloudFormation uses a
	// temporary session that is generated from your user credentials.
	RoleARN *string

	// The rollback triggers for CloudFormation to monitor during stack creation and
	// updating operations, and for the specified monitoring period afterwards.
	RollbackConfiguration *types.RollbackConfiguration

	// Key-value pairs to associate with this stack. CloudFormation also propagates
	// these tags to resources in the stack. You can specify a maximum of 50 tags.
	Tags []types.Tag

	// A structure that contains the body of the revised template, with a minimum
	// length of 1 byte and a maximum length of 51,200 bytes. CloudFormation generates
	// the change set by comparing this template with the template of the stack that
	// you specified.
	//
	// Conditional: You must specify only one of the following parameters: TemplateBody
	// , TemplateURL , or set the UsePreviousTemplate to true .
	TemplateBody *string

	// The URL of the file that contains the revised template. The URL must point to a
	// template (max size: 1 MB) that's located in an Amazon S3 bucket or a Systems
	// Manager document. CloudFormation generates the change set by comparing this
	// template with the stack that you specified. The location for an Amazon S3 bucket
	// must start with https:// . URLs from S3 static websites are not supported.
	//
	// Conditional: You must specify only one of the following parameters: TemplateBody
	// , TemplateURL , or set the UsePreviousTemplate to true .
	TemplateURL *string

	// Whether to reuse the template that's associated with the stack to create the
	// change set.
	//
	// When using templates with the AWS::LanguageExtensions transform, provide the
	// template instead of using UsePreviousTemplate to ensure new parameter values
	// and Systems Manager parameter updates are applied correctly. For more
	// information, see [AWS::LanguageExtensions transform].
	//
	// Conditional: You must specify only one of the following parameters: TemplateBody
	// , TemplateURL , or set the UsePreviousTemplate to true .
	//
	// [AWS::LanguageExtensions transform]: https://docs.aws.amazon.com/AWSCloudFormation/latest/TemplateReference/transform-aws-languageextensions.html
	UsePreviousTemplate *bool
}

func (d *DeployInput) AsCreateCreateChangeSetInput(changeSetType types.ChangeSetType) *cloudformation.CreateChangeSetInput {
	return &cloudformation.CreateChangeSetInput{
		ChangeSetName:           d.DeploymentName,
		StackName:               d.StackName,
		Capabilities:            d.Capabilities,
		ChangeSetType:           changeSetType,
		ClientToken:             d.ClientToken,
		DeploymentMode:          d.DeploymentMode,
		Description:             d.Description,
		ImportExistingResources: d.ImportExistingResources,
		IncludeNestedStacks:     d.IncludeNestedStacks,
		NotificationARNs:        d.NotificationARNs,
		OnStackFailure:          d.OnStackFailure,
		Parameters:              d.Parameters,
		ResourceTypes:           d.ResourceTypes,
		ResourcesToImport:       d.ResourcesToImport,
		RoleARN:                 d.RoleARN,
		RollbackConfiguration:   d.RollbackConfiguration,
		Tags:                    d.Tags,
		TemplateBody:            d.TemplateBody,
		TemplateURL:             d.TemplateURL,
		UsePreviousTemplate:     d.UsePreviousTemplate,
	}
}

func (d *DeployInput) AsExecuteChangeSetInput() *cloudformation.ExecuteChangeSetInput {
	return &cloudformation.ExecuteChangeSetInput{
		ChangeSetName:        d.DeploymentName,
		ClientRequestToken:   d.ClientToken,
		DisableRollback:      d.DisableRollback,
		RetainExceptOnCreate: d.RetainExceptOnCreate,
		StackName:            d.StackName,
	}
}

func (d *DeployInput) AsDescribeStacksInput() *cloudformation.DescribeStacksInput {
	return &cloudformation.DescribeStacksInput{
		StackName: d.StackName,
	}
}

func (d *DeployInput) Validate() error {
	if d.StackName == nil {
		return fmt.Errorf("DeployInput.StackName is required")
	}

	if d.DeploymentName == nil {
		return fmt.Errorf("DeployInput.DeploymentName is required")
	}

	return nil
}

type DeployOutput struct {
	CreateChangeSetInput *cloudformation.CreateChangeSetInput
	Changes              []types.Change
	Outputs              []types.Output
}

type DeployOptions struct {
	Timeout time.Duration
}

func (s *cloudFormationService) Deploy(ctx context.Context, input *DeployInput, opts ...func(*DeployOptions)) (*DeployOutput, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	o := &DeployOptions{
		Timeout: time.Minute * 15,
	}

	for _, opt := range opts {
		opt(o)
	}

	changeSetType, err := s.calculateChangeSetType(ctx, input)
	if err != nil {
		return nil, err
	}

	createChangeSetInput := input.AsCreateCreateChangeSetInput(changeSetType)

	describeChangeSetOutput, err := s.createChangeSetAndWait(ctx, createChangeSetInput, o.Timeout)
	if err != nil {
		if strings.Contains(err.Error(), "waiter state transitioned to Failure") {
			isEmpty, err := s.isEmptyChangeSet(ctx, &cloudformation.DescribeChangeSetInput{
				StackName:     input.StackName,
				ChangeSetName: input.DeploymentName,
			})

			if err != nil {
				return nil, err
			}

			// If the changeset was empty, clean up after ourselves and then just
			// return successfully
			if isEmpty {
				_, err := s.client.DeleteChangeSet(ctx, &cloudformation.DeleteChangeSetInput{
					StackName:     input.StackName,
					ChangeSetName: input.DeploymentName,
				})

				if err != nil {
					return nil, err
				}

				describeOut, err := s.client.DescribeStacks(ctx, input.AsDescribeStacksInput())
				if err != nil {
					return nil, err
				}

				var outputs []types.Output
				if len(describeOut.Stacks) > 0 {
					outputs = describeOut.Stacks[0].Outputs
				}

				return &DeployOutput{
					CreateChangeSetInput: createChangeSetInput,
					Changes:              []types.Change{},
					Outputs:              outputs,
				}, nil
			}
		} else {
			return nil, fmt.Errorf("failed to create changeset %s: %w", *input.DeploymentName, err)
		}
	}

	describeStacksOut, err := s.executeChangeSetAndWait(ctx, input.AsExecuteChangeSetInput(), changeSetType, o.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to execute changeset %s: %w", *input.DeploymentName, err)
	}

	var outputs []types.Output
	if describeStacksOut != nil && len(describeStacksOut.Stacks) > 0 {
		outputs = describeStacksOut.Stacks[0].Outputs
	}

	return &DeployOutput{
		CreateChangeSetInput: createChangeSetInput,
		Changes:              describeChangeSetOutput.Changes,
		Outputs:              outputs,
	}, nil
}

func (s *cloudFormationService) StackExists(ctx context.Context, stackName string) (bool, error) {
	_, err := s.client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
			if apiErr.ErrorCode() == "ValidationError" &&
				strings.Contains(apiErr.ErrorMessage(), "does not exist") {
				return false, nil
			}
		}

		return false, err
	}

	return true, nil
}

func (s *cloudFormationService) calculateChangeSetType(ctx context.Context, input *DeployInput) (types.ChangeSetType, error) {
	stackExists, err := s.StackExists(ctx, *input.StackName)
	if err != nil {
		return types.ChangeSetTypeCreate, err
	}

	if stackExists {
		return types.ChangeSetTypeUpdate, nil
	}

	return types.ChangeSetTypeCreate, nil
}

func (s *cloudFormationService) createChangeSetAndWait(ctx context.Context, input *cloudformation.CreateChangeSetInput, timeout time.Duration) (*cloudformation.DescribeChangeSetOutput, error) {
	_, err := s.client.CreateChangeSet(ctx, input)
	if err != nil {
		return nil, err
	}

	waiter := cloudformation.NewChangeSetCreateCompleteWaiter(s.client)
	params := &cloudformation.DescribeChangeSetInput{
		ChangeSetName: input.ChangeSetName,
		StackName:     input.StackName,
	}

	return waiter.WaitForOutput(ctx, params, timeout)
}

func (s *cloudFormationService) executeChangeSetAndWait(ctx context.Context, input *cloudformation.ExecuteChangeSetInput, changeSetType types.ChangeSetType, timeout time.Duration) (*cloudformation.DescribeStacksOutput, error) {

	_, err := s.client.ExecuteChangeSet(ctx, input)
	if err != nil {
		return nil, err
	}

	describeStackInput := &cloudformation.DescribeStacksInput{
		StackName: input.StackName,
	}

	switch changeSetType {
	case types.ChangeSetTypeUpdate:
		w := cloudformation.NewStackUpdateCompleteWaiter(s.client)
		return w.WaitForOutput(ctx, describeStackInput, timeout)
	case types.ChangeSetTypeCreate:
		w := cloudformation.NewStackCreateCompleteWaiter(s.client)
		return w.WaitForOutput(ctx, describeStackInput, timeout)
	}

	return nil, fmt.Errorf("unsupported change set type: %s", changeSetType)
}

func (s *cloudFormationService) isEmptyChangeSet(ctx context.Context, input *cloudformation.DescribeChangeSetInput) (bool, error) {
	output, err := s.client.DescribeChangeSet(ctx, input)
	if err != nil {
		return false, err
	}

	if output.Status == types.ChangeSetStatusFailed &&
		output.StatusReason != nil &&
		strings.Contains(*output.StatusReason, "The submitted information didn't contain changes") {
		return true, nil
	}

	return false, nil
}
