//go:build integration

package cloudformationservice_test

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/bernos/bastion/pkg/aws/cloudformationservice"
	"github.com/bernos/bastion/testhelpers"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
)

var (
	//go:embed testdata/stack.yaml
	stackTemplate string
	cfg           aws.Config
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

func TestLocalStack(t *testing.T) {
	ctx := t.Context()

	client := cloudformation.NewFromConfig(cfg)

	result, err := client.ListStacks(ctx, &cloudformation.ListStacksInput{})
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("result:\n%s", testhelpers.AsJSON(result))

	t.Run("deploys a fresh stack", func(t *testing.T) {
		svc := cloudformationservice.New(client)
		suffix := uuid.NewString()[:8]

		output, err := svc.Deploy(t.Context(), &cloudformationservice.DeployInput{
			StackName:      aws.String(fmt.Sprintf("test-stack-%s", suffix)),
			TemplateBody:   aws.String(stackTemplate),
			DeploymentName: aws.String(fmt.Sprintf("test-deployment-%s", suffix)),
		})

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("output:\n%s", testhelpers.AsJSON(output))

		describeStacksOutput, err := client.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
			StackName: aws.String(fmt.Sprintf("test-stack-%s", suffix)),
		})

		if err != nil {
			t.Fatal(err)
		}

		if len(describeStacksOutput.Stacks) != 1 {
			t.Fatalf("expected 1 stack but found %d", len(describeStacksOutput.Stacks))
		}
	})

	t.Run("doesn't fail when deploying noop update", func(t *testing.T) {
		svc := cloudformationservice.New(client)

		stackName := fmt.Sprintf("test-stack-%s", testhelpers.UID(t))

		t.Logf("deploying %s", stackName)

		_, err := svc.Deploy(t.Context(), &cloudformationservice.DeployInput{
			StackName:      aws.String(stackName),
			TemplateBody:   aws.String(stackTemplate),
			DeploymentName: aws.String(fmt.Sprintf("test-deployment-%s", testhelpers.UID(t))),
		})

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("deploying an empty update to %s", stackName)

		_, err = svc.Deploy(t.Context(), &cloudformationservice.DeployInput{
			StackName:      aws.String(stackName),
			TemplateBody:   aws.String(stackTemplate),
			DeploymentName: aws.String(fmt.Sprintf("test-deployment-%s", testhelpers.UID(t))),
		})

		if err != nil {
			t.Fatal(err)
		}

		changeSets, err := client.ListChangeSets(ctx, &cloudformation.ListChangeSetsInput{
			StackName: aws.String(stackName),
		})

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("Change Sets:\n%s", testhelpers.AsJSON(changeSets.Summaries))

		if len(changeSets.Summaries) != 2 {
			t.Fatalf("expected 2 change sets but found %d", len(changeSets.Summaries))
		}

	})

	t.Run("doesn't fail when deploying a parameter update", func(t *testing.T) {
		svc := cloudformationservice.New(client)

		stackName := fmt.Sprintf("test-stack-%s", testhelpers.UID(t))

		t.Logf("deploying %s", stackName)

		_, err := svc.Deploy(t.Context(), &cloudformationservice.DeployInput{
			StackName:      aws.String(stackName),
			TemplateBody:   aws.String(stackTemplate),
			DeploymentName: aws.String(fmt.Sprintf("test-deployment-%s", testhelpers.UID(t))),
			Parameters: []types.Parameter{
				{
					ParameterKey:   aws.String("MyIP"),
					ParameterValue: aws.String("10.0.0.1"),
				},
			},
		})

		if err != nil {
			t.Fatal(err)
		}

		t.Logf("deploying an update to %s", stackName)

		_, err = svc.Deploy(t.Context(), &cloudformationservice.DeployInput{
			StackName:      aws.String(stackName),
			TemplateBody:   aws.String(stackTemplate),
			DeploymentName: aws.String(fmt.Sprintf("test-deployment-%s", testhelpers.UID(t))),
			Parameters: []types.Parameter{
				{
					ParameterKey:   aws.String("MyIP"),
					ParameterValue: aws.String("10.0.0.2"),
				},
			},
		})

		if err != nil {
			t.Fatal(err)
		}

	})
}
