//go:build integration

package bastion_test

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/bernos/bastion/internal/bastion"
	cfnsvc "github.com/bernos/bastion/pkg/aws/cloudformationservice"
	"github.com/bernos/bastion/testhelpers"
	"github.com/testcontainers/testcontainers-go"
)

var (
	cfg aws.Config
)

func TestMain(m *testing.M) {
	flag.Parse()

	ctx := context.Background()

	c, localstackContainer, err := testhelpers.StartLocalStack(ctx)
	if err != nil {
		log.Fatalf("failed to start LocalStack: %s", err)
	}

	cfg = c

	code := m.Run()

	if err := testcontainers.TerminateContainer(localstackContainer); err != nil {
		log.Printf("failed to terminate LocalStack container: %s", err)
	}

	os.Exit(code)
}

func TestBastionService_DeployBastion(t *testing.T) {
	ctx := t.Context()

	ec2Client := ec2.NewFromConfig(cfg)
	ssmClient := ssm.NewFromConfig(cfg)

	vpcOut, err := ec2Client.CreateVpc(ctx, &ec2.CreateVpcInput{
		CidrBlock: aws.String("10.0.0.0/16"),
	})
	if err != nil {
		t.Fatalf("create vpc: %v", err)
	}
	vpcID := aws.ToString(vpcOut.Vpc.VpcId)

	subnetOut, err := ec2Client.CreateSubnet(ctx, &ec2.CreateSubnetInput{
		VpcId:     aws.String(vpcID),
		CidrBlock: aws.String("10.0.1.0/24"),
	})
	if err != nil {
		t.Fatalf("create subnet: %v", err)
	}
	subnetID := aws.ToString(subnetOut.Subnet.SubnetId)

	amiParamName := "/bastion/test/ami-id"
	_, err = ssmClient.PutParameter(ctx, &ssm.PutParameterInput{
		Name:  aws.String(amiParamName),
		Value: aws.String("ami-00000000"),
		Type:  "String",
	})
	if err != nil {
		t.Fatalf("put ssm parameter: %v", err)
	}

	cfnClient := cloudformation.NewFromConfig(cfg)
	svc := bastion.NewBastionService(cfnsvc.New(cfnClient))

	name := "bastion-" + testhelpers.UID(t)

	out, err := svc.DeployBastion(ctx, &bastion.DeployBastionInput{
		BastionName:      name,
		Owner:            "test",
		SubnetID:         subnetID,
		VPCID:            vpcID,
		InstanceType:     "t3.micro",
		AMIParameterName: amiParamName,
	})
	if err != nil {
		t.Fatalf("DeployBastion: %v", err)
	}

	if out.StackName != name+"-stack" {
		t.Errorf("want stack name %q, got %q", name+"-stack", out.StackName)
	}

	stacks, err := cfnClient.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: aws.String(out.StackName),
	})
	if err != nil {
		t.Fatalf("describe stacks: %v", err)
	}

	if len(stacks.Stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks.Stacks))
	}

	t.Logf("stack status: %s", stacks.Stacks[0].StackStatus)
}
