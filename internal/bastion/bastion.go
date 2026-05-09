package bastion

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
)

const (
	CFNValidationError = "ValidationError"
)

//go:embed stack.yaml
var stackTemplate string

type CloudFormationClient interface {
	CreateChangeSet(context.Context, *cloudformation.CreateChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error)
	CreateStack(context.Context, *cloudformation.CreateStackInput, ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	DescribeChangeSet(context.Context, *cloudformation.DescribeChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error)
	DescribeStacks(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	ExecuteChangeSet(context.Context, *cloudformation.ExecuteChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.ExecuteChangeSetOutput, error)
}

type BastionService interface {
	DeployBastion(context.Context, *DeployBastionInput) (*DeployBastionOutput, error)
}

type bastionService struct {
	// TODO: this should use the CloudformationClient interface defined locally in this pkg
	cloudFormationClient CloudFormationClient
}

func NewBastionService(cloudFormationClient *cloudformation.Client) BastionService {
	return &bastionService{
		cloudFormationClient: cloudFormationClient,
	}
}

type DeployBastionInput struct {
	BastionName string
	Owner       string
}

type DeployBastionOutput struct {
	StackName string
}

func (svc *bastionService) DeployBastion(ctx context.Context, input *DeployBastionInput) (*DeployBastionOutput, error) {
	stackName := fmt.Sprintf("%s-stack", input.BastionName)

	stackExists, err := svc.stackExists(ctx, stackName)
	if err != nil {
		return nil, fmt.Errorf("failed to determine whether bastion cloudformation stack %s exists: %w", stackName, err)
	}

	if !stackExists {

		// if it doesnt exist call create stack
		input := &cloudformation.CreateStackInput{
			StackName:    aws.String(stackName),
			Parameters:   []types.Parameter{},
			TemplateBody: aws.String(stackTemplate),
		}

		_, err := svc.cloudFormationClient.CreateStack(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to create cloudformation stack: %w", err)
		}

		waiter := cloudformation.NewStackCreateCompleteWaiter(svc.cloudFormationClient)

		if err := waiter.Wait(ctx, &cloudformation.DescribeStacksInput{StackName: aws.String(stackName)}, time.Minute*15); err != nil {
			return nil, fmt.Errorf("failed to wait for cloudformation stack creation to complete: %w", err)
		}

	} else {

		changeSetName := "foo"

		// if it does exist create changeset
		input := &cloudformation.CreateChangeSetInput{
			StackName:     aws.String(stackName),
			ChangeSetType: "UPDATE",
			ChangeSetName: aws.String(changeSetName),
			Parameters:    []types.Parameter{},
			TemplateBody:  aws.String(stackTemplate),
		}

		_, err := svc.cloudFormationClient.CreateChangeSet(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("failed to create cloudformation change set: %w", err)
		}

		waiter := cloudformation.NewChangeSetCreateCompleteWaiter(svc.cloudFormationClient)

		output, err := waiter.WaitForOutput(ctx, &cloudformation.DescribeChangeSetInput{
			ChangeSetName: aws.String(changeSetName),
			StackName:     aws.String(stackName),
		}, time.Minute*15)

		if err != nil {
			return nil, fmt.Errorf("failed to wait for change set to create: %w", err)
		}

		if output.Status == types.ChangeSetStatusFailed {
			if output.StatusReason != nil && !strings.Contains(*output.StatusReason, "No changes to be made") {
				return nil, fmt.Errorf("cloudformation change set failed: %s", *output.StatusReason)
			}
			// No changes detected, skip the changeset
		}

		// execute the changeset
		_, err = svc.cloudFormationClient.ExecuteChangeSet(ctx, &cloudformation.ExecuteChangeSetInput{
			ChangeSetName: aws.String(changeSetName),
			StackName:     aws.String(stackName),
		})

		updateWaiter := cloudformation.NewStackUpdateCompleteWaiter(svc.cloudFormationClient)
		if err := updateWaiter.Wait(ctx, &cloudformation.DescribeStacksInput{StackName: aws.String(stackName)}, time.Minute*15); err != nil {
			return nil, fmt.Errorf("failed to update cloudformation stack: %w", err)
		}

	}

	return &DeployBastionOutput{}, nil
}

func (svc *bastionService) stackExists(ctx context.Context, stackName string) (bool, error) {

	// paginator := cloudformation.NewDescribeStacksPaginator(svc.cloudFormationClient, &cloudformation.DescribeStacksInput{
	// 	StackName: aws.String(stackName),
	// })

	// for paginator.HasMorePages() {
	// 	page, err := paginator.NextPage(ctx)
	// }

	_, err := svc.cloudFormationClient.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})

	if err != nil {
		if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
			if apiErr.ErrorCode() == CFNValidationError &&
				strings.Contains(apiErr.ErrorMessage(), "does not exist") {
				return false, nil
			}
		}

		return false, err // just return the error
	}

	return true, nil

}
