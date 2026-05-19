package bastion

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	ec2ic "github.com/aws/aws-sdk-go-v2/service/ec2instanceconnect"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/bernos/bastion/pkg/aws/cloudformationservice"
)

// --- mocks ---

type mockCloudFormationService struct {
	DeployFn      func(context.Context, *cloudformationservice.DeployInput, ...func(*cloudformationservice.DeployOptions)) (*cloudformationservice.DeployOutput, error)
	DeleteStackFn func(context.Context, string) error
	StackExistsFn func(context.Context, string) (bool, error)
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

type mockEC2InstanceConnectClient struct {
	SendSSHPublicKeyFn func(context.Context, *ec2ic.SendSSHPublicKeyInput, ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error)
}

func (m *mockEC2InstanceConnectClient) SendSSHPublicKey(ctx context.Context, params *ec2ic.SendSSHPublicKeyInput, optFns ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
	return m.SendSSHPublicKeyFn(ctx, params, optFns...)
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

func successfulCFNMock(instanceID, az string) *mockCloudFormationService {
	return &mockCloudFormationService{
		DeployFn: func(_ context.Context, input *cloudformationservice.DeployInput, _ ...func(*cloudformationservice.DeployOptions)) (*cloudformationservice.DeployOutput, error) {
			return &cloudformationservice.DeployOutput{
				Outputs: []types.Output{
					{OutputKey: aws.String("InstanceId"), OutputValue: aws.String(instanceID)},
					{OutputKey: aws.String("AvailabilityZone"), OutputValue: aws.String(az)},
				},
			}, nil
		},
	}
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

	if got := paramValue(captured.Parameters, "InstanceType"); got != "t3.micro" {
		t.Errorf("default InstanceType: want %q, got %q", "t3.micro", got)
	}

	wantAMI := "/aws/service/ami-amazon-linux-latest/amzn2-ami-hvm-x86_64-gp2"
	if got := paramValue(captured.Parameters, "AmiId"); got != wantAMI {
		t.Errorf("default AmiId: want %q, got %q", wantAMI, got)
	}
}

// --- new tests for SSH key upload ---

func Test_BastionService_DeployBastion_NoKeyContent_SkipsSSMAndEC2IC(t *testing.T) {
	ssmCalled := false
	ec2icCalled := false

	svc := NewBastionService(
		successfulCFNMock("i-001", "ap-southeast-2a"),
		&mockEC2InstanceConnectClient{
			SendSSHPublicKeyFn: func(_ context.Context, _ *ec2ic.SendSSHPublicKeyInput, _ ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
				ec2icCalled = true
				return &ec2ic.SendSSHPublicKeyOutput{}, nil
			},
		},
		&mockSSMClient{
			DescribeInstanceInformationFn: func(_ context.Context, _ *ssm.DescribeInstanceInformationInput, _ ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error) {
				ssmCalled = true
				return &ssm.DescribeInstanceInformationOutput{}, nil
			},
		},
	)

	if _, err := svc.DeployBastion(context.Background(), &DeployBastionInput{
		BastionName: "bastion",
		SubnetID:    "subnet-000",
		VPCID:       "vpc-000",
	}); err != nil {
		t.Fatal(err)
	}

	if ssmCalled {
		t.Error("DescribeInstanceInformation should not be called when PublicKeyContent is empty")
	}
	if ec2icCalled {
		t.Error("SendSSHPublicKey should not be called when PublicKeyContent is empty")
	}
}

func Test_BastionService_DeployBastion_SSMOnline_UploadsKey(t *testing.T) {
	var capturedKey *ec2ic.SendSSHPublicKeyInput

	svc := NewBastionService(
		successfulCFNMock("i-abc001", "ap-southeast-2a"),
		&mockEC2InstanceConnectClient{
			SendSSHPublicKeyFn: func(_ context.Context, params *ec2ic.SendSSHPublicKeyInput, _ ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
				capturedKey = params
				return &ec2ic.SendSSHPublicKeyOutput{}, nil
			},
		},
		onlineSSMMock(),
	)

	if _, err := svc.DeployBastion(context.Background(), &DeployBastionInput{
		BastionName:      "bastion",
		SubnetID:         "subnet-000",
		VPCID:            "vpc-000",
		PublicKeyContent: "ssh-ed25519 AAAA test",
	}); err != nil {
		t.Fatal(err)
	}

	if capturedKey == nil {
		t.Fatal("SendSSHPublicKey was not called")
	}
	if aws.ToString(capturedKey.InstanceId) != "i-abc001" {
		t.Errorf("InstanceId: want %q, got %q", "i-abc001", aws.ToString(capturedKey.InstanceId))
	}
	if aws.ToString(capturedKey.AvailabilityZone) != "ap-southeast-2a" {
		t.Errorf("AvailabilityZone: want %q, got %q", "ap-southeast-2a", aws.ToString(capturedKey.AvailabilityZone))
	}
	if aws.ToString(capturedKey.InstanceOSUser) != "ec2-user" {
		t.Errorf("InstanceOSUser: want %q, got %q", "ec2-user", aws.ToString(capturedKey.InstanceOSUser))
	}
	if aws.ToString(capturedKey.SSHPublicKey) != "ssh-ed25519 AAAA test" {
		t.Errorf("SSHPublicKey: want %q, got %q", "ssh-ed25519 AAAA test", aws.ToString(capturedKey.SSHPublicKey))
	}
}

func Test_BastionService_DeployBastion_SSMTimeout_ReturnsError(t *testing.T) {
	ec2icCalled := false

	svc := NewBastionService(
		successfulCFNMock("i-001", "ap-southeast-2a"),
		&mockEC2InstanceConnectClient{
			SendSSHPublicKeyFn: func(_ context.Context, _ *ec2ic.SendSSHPublicKeyInput, _ ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
				ec2icCalled = true
				return &ec2ic.SendSSHPublicKeyOutput{}, nil
			},
		},
		neverOnlineSSMMock(),
	)

	// Short-lived context to trigger the timeout quickly.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := svc.DeployBastion(ctx, &DeployBastionInput{
		BastionName:      "bastion",
		SubnetID:         "subnet-000",
		VPCID:            "vpc-000",
		PublicKeyContent: "ssh-ed25519 AAAA test",
	})

	if err == nil {
		t.Fatal("expected error on SSM timeout, got nil")
	}
	if ec2icCalled {
		t.Error("SendSSHPublicKey should not be called when SSM times out")
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

func Test_BastionService_DeployBastion_EC2ICError_Propagated(t *testing.T) {
	uploadErr := errors.New("ec2ic failure")

	svc := NewBastionService(
		successfulCFNMock("i-001", "ap-southeast-2a"),
		&mockEC2InstanceConnectClient{
			SendSSHPublicKeyFn: func(_ context.Context, _ *ec2ic.SendSSHPublicKeyInput, _ ...func(*ec2ic.Options)) (*ec2ic.SendSSHPublicKeyOutput, error) {
				return nil, uploadErr
			},
		},
		onlineSSMMock(),
	)

	_, err := svc.DeployBastion(context.Background(), &DeployBastionInput{
		BastionName:      "bastion",
		SubnetID:         "subnet-000",
		VPCID:            "vpc-000",
		PublicKeyContent: "ssh-ed25519 AAAA test",
	})

	if !errors.Is(err, uploadErr) {
		t.Errorf("expected uploadErr to be wrapped in returned error, got: %v", err)
	}
}
