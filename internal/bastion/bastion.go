package bastion

import (
	"context"
	"fmt"

	_ "embed"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
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
	StackName        string
	InstanceID       string
	AvailabilityZone string
}

func (svc *bastionService) DeployBastion(ctx context.Context, input *DeployBastionInput) (*DeployBastionOutput, error) {
	id, err := gonanoid.New(8)
	if err != nil {
		return nil, err
	}

	stackName := fmt.Sprintf("%s-stack", input.BastionName)
	deploymentName := fmt.Sprintf("%s-%s", input.BastionName, id)

	amiParam := input.AMIParameterName
	if amiParam == "" {
		amiParam = "/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"
	}

	instanceType := input.InstanceType
	if instanceType == "" {
		instanceType = "t3.micro"
	}

	out, err := svc.cloudFormationService.Deploy(ctx, &cloudformationservice.DeployInput{
		StackName:      aws.String(stackName),
		DeploymentName: aws.String(deploymentName),
		TemplateBody:   aws.String(stackTemplate),
		Capabilities:   []types.Capability{types.CapabilityCapabilityIam},
		Parameters: []types.Parameter{
			{ParameterKey: aws.String("SubnetId"), ParameterValue: aws.String(input.SubnetID)},
			{ParameterKey: aws.String("VpcId"), ParameterValue: aws.String(input.VPCID)},
			{ParameterKey: aws.String("InstanceType"), ParameterValue: aws.String(instanceType)},
			{ParameterKey: aws.String("AmiId"), ParameterValue: aws.String(amiParam)},
			{ParameterKey: aws.String("BastionName"), ParameterValue: aws.String(input.BastionName)},
			{ParameterKey: aws.String("Owner"), ParameterValue: aws.String(input.Owner)},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to deploy bastion: %w", err)
	}

	result := &DeployBastionOutput{StackName: stackName}
	for _, o := range out.Outputs {
		switch aws.ToString(o.OutputKey) {
		case "InstanceId":
			result.InstanceID = aws.ToString(o.OutputValue)
		case "AvailabilityZone":
			result.AvailabilityZone = aws.ToString(o.OutputValue)
		}
	}
	return result, nil
}
