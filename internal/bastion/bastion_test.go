package bastion

import (
	"context"
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/bernos/bastion/pkg/aws/cloudformationservice"
)

type mockCloudFormationService struct {
	DeployFn func(context.Context, *cloudformationservice.DeployInput, ...func(*cloudformationservice.DeployOptions)) (*cloudformationservice.DeployOutput, error)
}

func (m *mockCloudFormationService) Deploy(ctx context.Context, input *cloudformationservice.DeployInput, opts ...func(*cloudformationservice.DeployOptions)) (*cloudformationservice.DeployOutput, error) {
	return m.DeployFn(ctx, input, opts...)
}

func (m *mockCloudFormationService) StackExists(ctx context.Context, stackName string) (bool, error) {
	return false, nil
}

func paramValue(params []types.Parameter, key string) string {
	for _, p := range params {
		if aws.ToString(p.ParameterKey) == key {
			return aws.ToString(p.ParameterValue)
		}
	}
	return ""
}

func Test_BastionService_DeployBastion_ForwardsParameters(t *testing.T) {
	var captured *cloudformationservice.DeployInput

	mock := &mockCloudFormationService{
		DeployFn: func(_ context.Context, input *cloudformationservice.DeployInput, _ ...func(*cloudformationservice.DeployOptions)) (*cloudformationservice.DeployOutput, error) {
			captured = input
			return &cloudformationservice.DeployOutput{
				Outputs: []types.Output{
					{OutputKey: aws.String("InstanceId"), OutputValue: aws.String("i-abc001")},
					{OutputKey: aws.String("AvailabilityZone"), OutputValue: aws.String("ap-southeast-2a")},
				},
			}, nil
		},
	}

	svc := NewBastionService(mock)

	out, err := svc.DeployBastion(context.Background(), &DeployBastionInput{
		BastionName:      "my-bastion",
		Owner:            "alice",
		SubnetID:         "subnet-abc123",
		VPCID:            "vpc-def456",
		InstanceType:     "t3.small",
		AMIParameterName: "/my/custom/ami",
	})
	if err != nil {
		t.Fatal(err)
	}

	if out.InstanceID != "i-abc001" {
		t.Errorf("InstanceID: want %q, got %q", "i-abc001", out.InstanceID)
	}
	if out.AvailabilityZone != "ap-southeast-2a" {
		t.Errorf("AvailabilityZone: want %q, got %q", "ap-southeast-2a", out.AvailabilityZone)
	}

	if captured == nil {
		t.Fatal("Deploy was not called")
	}

	if aws.ToString(captured.StackName) != "my-bastion-stack" {
		t.Errorf("want stack name %q, got %q", "my-bastion-stack", aws.ToString(captured.StackName))
	}

	if !slices.Contains(captured.Capabilities, types.CapabilityCapabilityIam) {
		t.Errorf("expected %s in Capabilities, got %v", types.CapabilityCapabilityIam, captured.Capabilities)
	}

	cases := map[string]string{
		"SubnetId":     "subnet-abc123",
		"VpcId":        "vpc-def456",
		"InstanceType": "t3.small",
		"AmiId":        "/my/custom/ami",
		"BastionName":  "my-bastion",
		"Owner":        "alice",
	}

	for key, want := range cases {
		if got := paramValue(captured.Parameters, key); got != want {
			t.Errorf("parameter %s: want %q, got %q", key, want, got)
		}
	}
}

func Test_BastionService_DeployBastion_DefaultsInstanceTypeAndAMI(t *testing.T) {
	var captured *cloudformationservice.DeployInput

	mock := &mockCloudFormationService{
		DeployFn: func(_ context.Context, input *cloudformationservice.DeployInput, _ ...func(*cloudformationservice.DeployOptions)) (*cloudformationservice.DeployOutput, error) {
			captured = input
			return &cloudformationservice.DeployOutput{}, nil
		},
	}

	svc := NewBastionService(mock)

	out, err := svc.DeployBastion(context.Background(), &DeployBastionInput{
		BastionName: "bastion",
		SubnetID:    "subnet-000",
		VPCID:       "vpc-000",
	})
	if err != nil {
		t.Fatal(err)
	}

	if out.InstanceID != "" {
		t.Errorf("InstanceID: want empty string, got %q", out.InstanceID)
	}
	if out.AvailabilityZone != "" {
		t.Errorf("AvailabilityZone: want empty string, got %q", out.AvailabilityZone)
	}

	if got := paramValue(captured.Parameters, "InstanceType"); got != "t3.micro" {
		t.Errorf("default InstanceType: want %q, got %q", "t3.micro", got)
	}

	wantAMI := "/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"
	if got := paramValue(captured.Parameters, "AmiId"); got != wantAMI {
		t.Errorf("default AmiId: want %q, got %q", wantAMI, got)
	}
}
