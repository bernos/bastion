package bastion

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/smithy-go"
)

type mockCloudFormationClient struct {
	DescribeStacksFn func(context.Context, *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error)
	CreateStackFn    func(context.Context, *cloudformation.CreateStackInput) (*cloudformation.CreateStackOutput, error)
}

func (m *mockCloudFormationClient) DescribeStacks(ctx context.Context, input *cloudformation.DescribeStacksInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	if m.DescribeStacksFn != nil {
		return m.DescribeStacksFn(ctx, input)
	}

	return nil, fmt.Errorf("not implemented")
}

func (m *mockCloudFormationClient) CreateStack(ctx context.Context, input *cloudformation.CreateStackInput, o ...func(*cloudformation.Options)) (*cloudformation.CreateStackOutput, error) {
	if m.CreateStackFn != nil {
		return m.CreateStackFn(ctx, input)
	}

	return nil, fmt.Errorf("not implemented")
}

var _ CloudFormationClient = (*mockCloudFormationClient)(nil)

func Test_stackExists(t *testing.T) {

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
				Code:    CFNValidationError,
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
				DescribeStacksFn: func(ctx context.Context, input *cloudformation.DescribeStacksInput) (*cloudformation.DescribeStacksOutput, error) {
					return nil, tt.mockError
				},
			}

			svc := &bastionService{mock}

			got, err := svc.stackExists(ctx, tt.name)

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
