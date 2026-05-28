package bastion

import (
	"context"
	"fmt"
	"time"

	_ "embed"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	ec2ic "github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
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

type ec2InstanceConnectClient interface {
	SendSSHPublicKey(ctx context.Context, params *ec2ic.SendSSHPublicKeyInput, optFns ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error)
}

type ssmClient interface {
	DescribeInstanceInformation(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error)
}

type BastionService interface {
	DeployBastion(context.Context, *DeployBastionInput) (*DeployBastionOutput, error)
	DeleteBastion(context.Context, *DeleteBastionInput) error
	DescribeBastion(context.Context, *DescribeBastionInput) (*DescribeBastionOutput, error)
	WaitForSSMReady(context.Context, *WaitForSSMReadyInput) error
}

type bastionService struct {
	cloudFormationService cloudformationservice.CloudFormationService
	ec2ic                 ec2InstanceConnectClient
	ssm                   ssmClient
}

func NewBastionService(
	cloudFormationService cloudformationservice.CloudFormationService,
	ec2ic ec2InstanceConnectClient,
	ssm ssmClient,
) BastionService {
	return &bastionService{
		cloudFormationService: cloudFormationService,
		ec2ic:                 ec2ic,
		ssm:                   ssm,
	}
}

type DeployBastionInput struct {
	BastionName      string
	Owner            string
	SubnetID         string
	AMIParameterName string
	InstanceType     string
	VPCID            string
	Tags             map[string]string
}

type DeployBastionOutput struct {
	StackName        string
	InstanceID       string
	AvailabilityZone string
}

type DeleteBastionInput struct {
	BastionName string
}

type DescribeBastionInput struct {
	BastionName string
}

type DescribeBastionOutput struct {
	InstanceID       string
	AvailabilityZone string
}

type WaitForSSMReadyInput struct {
	InstanceID string
}

func (svc *bastionService) DeployBastion(ctx context.Context, input *DeployBastionInput) (*DeployBastionOutput, error) {
	id, err := gonanoid.Generate("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", 8)
	if err != nil {
		return nil, err
	}

	stackName := fmt.Sprintf("%s-stack", input.BastionName)
	deploymentName := fmt.Sprintf("%s-%s", input.BastionName, id)

	amiParam := input.AMIParameterName
	if amiParam == "" {
		amiParam = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64"
	}

	instanceType := input.InstanceType
	if instanceType == "" {
		instanceType = "t4g.micro"
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
		Tags: stackTagsFromMap(input.Tags),
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

func (svc *bastionService) DeleteBastion(ctx context.Context, input *DeleteBastionInput) error {
	stackName := fmt.Sprintf("%s-stack", input.BastionName)

	exists, err := svc.cloudFormationService.StackExists(ctx, stackName)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("stack %q does not exist", stackName)
	}

	return svc.cloudFormationService.DeleteStack(ctx, stackName)
}

func (svc *bastionService) DescribeBastion(ctx context.Context, input *DescribeBastionInput) (*DescribeBastionOutput, error) {
	stackName := fmt.Sprintf("%s-stack", input.BastionName)

	outputs, err := svc.cloudFormationService.GetStackOutputs(ctx, stackName)
	if err != nil {
		return nil, fmt.Errorf("describing bastion %q: %w", input.BastionName, err)
	}

	return &DescribeBastionOutput{
		InstanceID:       outputs["InstanceId"],
		AvailabilityZone: outputs["AvailabilityZone"],
	}, nil
}

func (svc *bastionService) WaitForSSMReady(ctx context.Context, input *WaitForSSMReadyInput) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	check := func() (bool, error) {
		out, err := svc.ssm.DescribeInstanceInformation(ctx, &ssm.DescribeInstanceInformationInput{
			InstanceInformationFilterList: []ssmtypes.InstanceInformationFilter{
				{Key: ssmtypes.InstanceInformationFilterKeyInstanceIds, ValueSet: []string{input.InstanceID}},
			},
		})
		if err != nil {
			return false, fmt.Errorf("describing SSM instance information: %w", err)
		}
		for _, info := range out.InstanceInformationList {
			if info.PingStatus == ssmtypes.PingStatusOnline {
				return true, nil
			}
		}
		return false, nil
	}

	if ready, err := check(); err != nil || ready {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for SSM agent on instance %s to come online: check SSM Fleet Manager for details", input.InstanceID)
		case <-ticker.C:
			if ready, err := check(); err != nil || ready {
				return err
			}
		}
	}
}

func stackTagsFromMap(m map[string]string) []types.Tag {
	tags := make([]types.Tag, 0, len(m))
	for k, v := range m {
		tags = append(tags, types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	return tags
}
