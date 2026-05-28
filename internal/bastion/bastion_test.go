package bastion

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/bernos/bastion/pkg/aws/cloudformationservice"
)

// --- mocks ---

type mockCloudFormationService struct {
	DeployFn          func(context.Context, *cloudformationservice.DeployInput, ...func(*cloudformationservice.DeployOptions)) (*cloudformationservice.DeployOutput, error)
	DeleteStackFn     func(context.Context, string) error
	StackExistsFn     func(context.Context, string) (bool, error)
	GetStackOutputsFn func(context.Context, string) (map[string]string, error)
}

func (m *mockCloudFormationService) Deploy(ctx context.Context, input *cloudformationservice.DeployInput, opts ...func(*cloudformationservice.DeployOptions)) (*cloudformationservice.DeployOutput, error) {
	return m.DeployFn(ctx, input, opts...)
}

func (m *mockCloudFormationService) DeleteStack(ctx context.Context, stackName string) error {
	if m.DeleteStackFn != nil {
		return m.DeleteStackFn(ctx, stackName)
	}
	return nil
}

func (m *mockCloudFormationService) StackExists(ctx context.Context, stackName string) (bool, error) {
	if m.StackExistsFn != nil {
		return m.StackExistsFn(ctx, stackName)
	}
	return false, nil
}

func (m *mockCloudFormationService) GetStackOutputs(ctx context.Context, stackName string) (map[string]string, error) {
	if m.GetStackOutputsFn != nil {
		return m.GetStackOutputsFn(ctx, stackName)
	}
	return map[string]string{}, nil
}

type mockSSMClient struct {
	DescribeInstanceInformationFn func(context.Context, *ssm.DescribeInstanceInformationInput, ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error)
}

func (m *mockSSMClient) DescribeInstanceInformation(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
	return m.DescribeInstanceInformationFn(ctx, params, optFns...)
}

// --- helpers ---

func paramValue(params []types.Parameter, key string) string {
	for _, p := range params {
		if aws.ToString(p.ParameterKey) == key {
			return aws.ToString(p.ParameterValue)
		}
	}
	return ""
}

func onlineSSMMock() *mockSSMClient {
	return &mockSSMClient{
		DescribeInstanceInformationFn: func(_ context.Context, _ *ssm.DescribeInstanceInformationInput, _ ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
			return &ssm.DescribeInstanceInformationOutput{
				InstanceInformationList: []ssmtypes.InstanceInformation{
					{PingStatus: ssmtypes.PingStatusOnline},
				},
			}, nil
		},
	}
}

func neverOnlineSSMMock() *mockSSMClient {
	return &mockSSMClient{
		DescribeInstanceInformationFn: func(_ context.Context, _ *ssm.DescribeInstanceInformationInput, _ ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
			return &ssm.DescribeInstanceInformationOutput{}, nil
		},
	}
}

// --- existing tests (updated to pass nil for new clients) ---

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

	svc := NewBastionService(mock, nil, nil)

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

	svc := NewBastionService(mock, nil, nil)

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

	if got := paramValue(captured.Parameters, "InstanceType"); got != "t4g.micro" {
		t.Errorf("default InstanceType: want %q, got %q", "t4g.micro", got)
	}

	wantAMI := "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64"
	if got := paramValue(captured.Parameters, "AmiId"); got != wantAMI {
		t.Errorf("default AmiId: want %q, got %q", wantAMI, got)
	}
}

func Test_BastionService_DeployBastion_UserTagsAppliedAsStackTags(t *testing.T) {
	var captured *cloudformationservice.DeployInput

	mock := &mockCloudFormationService{
		DeployFn: func(_ context.Context, input *cloudformationservice.DeployInput, _ ...func(*cloudformationservice.DeployOptions)) (*cloudformationservice.DeployOutput, error) {
			captured = input
			return &cloudformationservice.DeployOutput{}, nil
		},
	}

	svc := NewBastionService(mock, nil, nil)

	_, err := svc.DeployBastion(context.Background(), &DeployBastionInput{
		BastionName: "bastion",
		SubnetID:    "subnet-000",
		VPCID:       "vpc-000",
		Tags:        map[string]string{"env": "prod", "team": "platform"},
	})
	if err != nil {
		t.Fatal(err)
	}

	tagMap := make(map[string]string, len(captured.Tags))
	for _, tag := range captured.Tags {
		tagMap[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	if tagMap["env"] != "prod" {
		t.Errorf("stack tag env: want %q, got %q", "prod", tagMap["env"])
	}
	if tagMap["team"] != "platform" {
		t.Errorf("stack tag team: want %q, got %q", "platform", tagMap["team"])
	}
}

func Test_BastionService_DeployBastion_NilTagsProducesEmptyStackTags(t *testing.T) {
	var captured *cloudformationservice.DeployInput

	mock := &mockCloudFormationService{
		DeployFn: func(_ context.Context, input *cloudformationservice.DeployInput, _ ...func(*cloudformationservice.DeployOptions)) (*cloudformationservice.DeployOutput, error) {
			captured = input
			return &cloudformationservice.DeployOutput{}, nil
		},
	}

	svc := NewBastionService(mock, nil, nil)

	_, err := svc.DeployBastion(context.Background(), &DeployBastionInput{
		BastionName: "bastion",
		SubnetID:    "subnet-000",
		VPCID:       "vpc-000",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(captured.Tags) != 0 {
		t.Errorf("stack tags: want empty, got %v", captured.Tags)
	}
}

func Test_BastionService_DeleteBastion_Success(t *testing.T) {
	var deletedStack string

	cfn := &mockCloudFormationService{
		StackExistsFn: func(_ context.Context, stackName string) (bool, error) {
			return true, nil
		},
		DeleteStackFn: func(_ context.Context, stackName string) error {
			deletedStack = stackName
			return nil
		},
	}

	svc := NewBastionService(cfn, nil, nil)

	if err := svc.DeleteBastion(context.Background(), &DeleteBastionInput{BastionName: "my-bastion"}); err != nil {
		t.Fatal(err)
	}

	if deletedStack != "my-bastion-stack" {
		t.Errorf("want deleted stack %q, got %q", "my-bastion-stack", deletedStack)
	}
}

func Test_BastionService_DeleteBastion_StackNotFound_ReturnsError(t *testing.T) {
	cfn := &mockCloudFormationService{
		StackExistsFn: func(_ context.Context, stackName string) (bool, error) {
			return false, nil
		},
		DeleteStackFn: func(_ context.Context, stackName string) error {
			t.Error("DeleteStack should not be called when stack does not exist")
			return nil
		},
	}

	svc := NewBastionService(cfn, nil, nil)

	err := svc.DeleteBastion(context.Background(), &DeleteBastionInput{BastionName: "missing"})
	if err == nil {
		t.Fatal("expected error for missing stack, got nil")
	}
}

func Test_BastionService_DescribeBastion_ReturnsInstanceDetails(t *testing.T) {
	cfn := &mockCloudFormationService{
		GetStackOutputsFn: func(_ context.Context, stackName string) (map[string]string, error) {
			if stackName != "my-bastion-stack" {
				return nil, fmt.Errorf("unexpected stack name: %s", stackName)
			}
			return map[string]string{
				"InstanceId":       "i-abc001",
				"AvailabilityZone": "ap-southeast-2a",
			}, nil
		},
	}

	svc := NewBastionService(cfn, nil, nil)

	out, err := svc.DescribeBastion(context.Background(), &DescribeBastionInput{
		BastionName: "my-bastion",
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
}

func Test_BastionService_DescribeBastion_StackNotFound_ReturnsError(t *testing.T) {
	cfn := &mockCloudFormationService{
		GetStackOutputsFn: func(_ context.Context, _ string) (map[string]string, error) {
			return nil, fmt.Errorf("stack does not exist")
		},
	}

	svc := NewBastionService(cfn, nil, nil)

	_, err := svc.DescribeBastion(context.Background(), &DescribeBastionInput{
		BastionName: "missing",
	})
	if err == nil {
		t.Fatal("expected error for missing stack, got nil")
	}
}

func Test_BastionService_WaitForSSMReady_OnlineImmediately_ReturnsNil(t *testing.T) {
	svc := NewBastionService(nil, nil, onlineSSMMock())

	err := svc.WaitForSSMReady(context.Background(), &WaitForSSMReadyInput{
		InstanceID: "i-abc001",
	})
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
}

func Test_BastionService_WaitForSSMReady_Timeout_ReturnsError(t *testing.T) {
	svc := NewBastionService(nil, nil, neverOnlineSSMMock())

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := svc.WaitForSSMReady(ctx, &WaitForSSMReadyInput{
		InstanceID: "i-abc001",
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func Test_BastionService_WaitForSSMReady_SSMError_Propagated(t *testing.T) {
	ssmErr := errors.New("ssm unavailable")
	svc := NewBastionService(nil, nil, &mockSSMClient{
		DescribeInstanceInformationFn: func(_ context.Context, _ *ssm.DescribeInstanceInformationInput, _ ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
			return nil, ssmErr
		},
	})

	err := svc.WaitForSSMReady(context.Background(), &WaitForSSMReadyInput{
		InstanceID: "i-abc001",
	})
	if !errors.Is(err, ssmErr) {
		t.Errorf("expected ssmErr to be wrapped, got: %v", err)
	}
}
