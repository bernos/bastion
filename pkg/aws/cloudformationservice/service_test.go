package cloudformationservice

import (
	"context"
	_ "embed"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	// "github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
)

//go:embed testdata/stack.yaml
var stackTemplate string

var errMockNotImplemented = fmt.Errorf("not implemented")

type mockCloudFormationClient struct {
	CreateChangeSetFn   func(context.Context, *cloudformation.CreateChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error)
	CreateStackFn       func(context.Context, *cloudformation.CreateStackInput, ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
	DeleteChangeSetFn   func(context.Context, *cloudformation.DeleteChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.DeleteChangeSetOutput, error)
	DescribeChangeSetFn func(context.Context, *cloudformation.DescribeChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error)
	DescribeStacksFn    func(context.Context, *cloudformation.DescribeStacksInput, ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error)
	ExecuteChangeSetFn  func(context.Context, *cloudformation.ExecuteChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.ExecuteChangeSetOutput, error)
}

func (m *mockCloudFormationClient) CreateChangeSet(ctx context.Context, input *cloudformation.CreateChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error) {
	if m.CreateChangeSetFn != nil {
		return m.CreateChangeSetFn(ctx, input, o...)
	}
	return nil, errMockNotImplemented
}

func (m *mockCloudFormationClient) CreateStack(ctx context.Context, input *cloudformation.CreateStackInput, o ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
	if m.CreateStackFn != nil {
		return m.CreateStackFn(ctx, input, o...)
	}

	return nil, errMockNotImplemented
}

func (m *mockCloudFormationClient) DeleteChangeSet(ctx context.Context, input *cloudformation.DeleteChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.DeleteChangeSetOutput, error) {
	if m.DeleteChangeSetFn != nil {
		return m.DeleteChangeSetFn(ctx, input, o...)
	}
	return nil, errMockNotImplemented
}

func (m *mockCloudFormationClient) DescribeChangeSet(ctx context.Context, input *cloudformation.DescribeChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error) {
	if m.DescribeChangeSetFn != nil {
		return m.DescribeChangeSetFn(ctx, input, o...)
	}
	return nil, errMockNotImplemented
}

func (m *mockCloudFormationClient) DescribeStacks(ctx context.Context, input *cloudformation.DescribeStacksInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	if m.DescribeStacksFn != nil {
		return m.DescribeStacksFn(ctx, input, o...)
	}

	return nil, errMockNotImplemented
}

func (m *mockCloudFormationClient) ExecuteChangeSet(ctx context.Context, input *cloudformation.ExecuteChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.ExecuteChangeSetOutput, error) {
	if m.ExecuteChangeSetFn != nil {
		return m.ExecuteChangeSetFn(ctx, input, o...)
	}
	return nil, errMockNotImplemented
}

var _ CloudFormationClient = (*mockCloudFormationClient)(nil)

func Test_CloudFormationService_Deploy(t *testing.T) {
	t.Run("create happy path", func(t *testing.T) {
		executeChangeSetCalled := false

		mock := &mockCloudFormationClient{
			DescribeStacksFn: func(ctx context.Context, input *cloudformation.DescribeStacksInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
				t.Log(">>> DescribeStacks")
				if executeChangeSetCalled {
					return &cloudformation.DescribeStacksOutput{
						Stacks: []types.Stack{
							{
								StackStatus: types.StackStatusUpdateComplete,
							},
						},
					}, nil
				}

				return nil, &smithy.GenericAPIError{
					Code:    "ValidationError",
					Message: "does not exist",
				}
			},
			CreateChangeSetFn: func(ctx context.Context, input *cloudformation.CreateChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error) {
				t.Log(">>> CreateChangeSet")
				if input.ChangeSetType == types.ChangeSetTypeCreate {
					return &cloudformation.CreateChangeSetOutput{}, nil
				}
				return nil, fmt.Errorf("unexpected changeset type %s", input.ChangeSetType)
			},
			DescribeChangeSetFn: func(ctx context.Context, input *cloudformation.DescribeChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error) {
				t.Log(">>> DescribeChangeSets")

				return &cloudformation.DescribeChangeSetOutput{
					Status: types.ChangeSetStatusCreateComplete,
				}, nil
			},
			ExecuteChangeSetFn: func(ctx context.Context, input *cloudformation.ExecuteChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.ExecuteChangeSetOutput, error) {
				executeChangeSetCalled = true
				t.Log(">>> ExecuteChangeSet")

				return &cloudformation.ExecuteChangeSetOutput{}, nil
			},
		}

		svc := &cloudFormationService{mock}

		_, err := svc.Deploy(t.Context(), &DeployInput{
			StackName:      aws.String("test-stack"),
			DeploymentName: aws.String("my-deployment"),
		}, time.Second*10)

		if err != nil {
			t.Fatal(err)
		}
	})
}

func Test_CloudFormationService_StackExists(t *testing.T) {
	t.Skip()
	cases := []struct {
		name       string
		mockError  error
		wantResult bool
		wantError  bool
	}{
		{
			name:       "i exist",
			mockError:  nil,
			wantResult: true,
			wantError:  false,
		},
		{
			name: "i dont exist",
			mockError: &smithy.GenericAPIError{
				Code:    "ValidationError",
				Message: "does not exist",
			},
			wantResult: false,
			wantError:  false,
		},
		{
			name:       "i should error",
			mockError:  fmt.Errorf("Im just a regular error"),
			wantResult: false,
			wantError:  true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {

			ctx := t.Context()

			mock := &mockCloudFormationClient{
				DescribeStacksFn: func(ctx context.Context, input *cloudformation.DescribeStacksInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
					return nil, tt.mockError
				},
			}

			svc := &cloudFormationService{mock}

			got, err := svc.StackExists(ctx, tt.name)

			if tt.wantError && err == nil {
				t.Fatal("expected error but got none")
			}

			if !tt.wantError && err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if got != tt.wantResult {
				t.Fatalf("want: %t, got %t", tt.wantResult, got)
			}
		})
	}
}

func Test_CloudFormationService_createChangeSetAndWait(t *testing.T) {
	t.Skip()
	t.Run("should ignore empty changeset error", func(t *testing.T) {

		mock := &mockCloudFormationClient{
			CreateChangeSetFn: func(ctx context.Context, input *cloudformation.CreateChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error) {
				return nil, nil
			},
			DescribeChangeSetFn: func(ctx context.Context, input *cloudformation.DescribeChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error) {

				return &cloudformation.DescribeChangeSetOutput{
					Status:       types.ChangeSetStatusFailed,
					StatusReason: aws.String("The submitted information didn't contain changes"),
				}, nil
			},
		}

		svc := &cloudFormationService{mock}
		input := &cloudformation.CreateChangeSetInput{
			StackName:     aws.String("test"),
			ChangeSetName: aws.String("changeset"),
		}

		_, err := svc.createChangeSetAndWait(t.Context(), input, time.Minute)

		if err != nil {
			t.Fatalf("got unexpected error: %s", err)
		}

	})

}
