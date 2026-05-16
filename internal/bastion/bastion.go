package bastion

import (
	"context"
	"fmt"

	_ "embed"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/bernos/bastion/pkg/aws/cloudformationservice"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	CFNValidationError = "ValidationError"
)

var (
	//go:embed stack.yaml
	stackTemplate string
)

type BastionService interface {
	DeployBastion(context.Context, *DeployBastionInput) (*DeployBastionOutput, error)
}

type bastionService struct {
	cloudFormationService cloudformationservice.CloudFormationService
}

func NewBastionService(cloudFormationService cloudformationservice.CloudFormationService) BastionService {
	return &bastionService{
		cloudFormationService: cloudFormationService,
	}
}

type DeployBastionInput struct {
	BastionName      string
	Owner            string
	SubnetID         string
	AMIParameterName string
	InstanceType     string
	VPCID            string
}

type DeployBastionOutput struct {
	StackName string
}

func (svc *bastionService) DeployBastion(ctx context.Context, input *DeployBastionInput) (*DeployBastionOutput, error) {
	id, err := gonanoid.New(8)
	if err != nil {
		return nil, err
	}

	stackName := fmt.Sprintf("%s-stack", input.BastionName)
	deploymentName := fmt.Sprintf("%s-%s", input.BastionName, id)

	_, err = svc.cloudFormationService.Deploy(ctx, &cloudformationservice.DeployInput{
		StackName:      aws.String(stackName),
		DeploymentName: aws.String(deploymentName),
		TemplateBody:   aws.String(stackTemplate),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to deploy bastion: %w", err)
	}

	return &DeployBastionOutput{
		StackName: stackName,
	}, nil
}
