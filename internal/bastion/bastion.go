package bastion

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/smithy-go"
)

const (
	CFNValidationError = "ValidationError"
)

type CloudFormationClient interface {
	CreateStack(context.Context, *cloudformation.CreateStackInput, ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	DescribeStacks(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
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

	if stackExists {

		// if it doesnt exist call create stack
	} else {

		// if it does exist create changeset

		// execute the changeset
	}
	return &DeployBastionOutput{}, nil
}

func (svc *bastionService) stackExists(ctx context.Context, stackName string) (bool, error) {
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
