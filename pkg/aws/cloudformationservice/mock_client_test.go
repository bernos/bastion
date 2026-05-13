package cloudformationservice

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
)

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
