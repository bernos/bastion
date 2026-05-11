package cloudformationservice

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
)

var errMockNotImplemented = fmt.Errorf("not implemented")

type mockCloudFormationClient struct {
	CreateChangeSetFn   func(context.Context, *cloudformation.CreateChangeSetInput, ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error)
	CreateStackFn       func(context.Context, *cloudformation.CreateStackInput, ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error)
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

func Test_CloudFormationService_StackExists(t *testing.T) {

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

	t.Run("should ignore empty changeset error", func(t *testing.T) {

		mock := &mockCloudFormationClient{
			CreateChangeSetFn: func(ctx context.Context, input *cloudformation.CreateChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error) {
				return nil, nil
			},
			DescribeChangeSetFn: func(ctx context.Context, input *cloudformation.DescribeChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error) {

				return &cloudformation.DescribeChangeSetOutput{
					Status:       types.ChangeSetStatusFailed,
					StatusReason: aws.String("No changes to be made"),
				}, nil
			},
		}

		svc := &cloudFormationService{mock}
		input := &cloudformation.CreateChangeSetInput{
			StackName:     aws.String("test"),
			ChangeSetName: aws.String("changeset"),
		}

		_, err := svc.createChangeSetAndWait(t.Context(), input, time.Minute, true)

		if err != nil {
			t.Fatalf("got unexpected error: %s", err)
		}

	})

}
