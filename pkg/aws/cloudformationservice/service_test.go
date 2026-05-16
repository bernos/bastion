package cloudformationservice

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/smithy-go"
)

func outputByKey(outputs []types.Output) map[string]string {
	m := make(map[string]string, len(outputs))
	for _, o := range outputs {
		m[aws.ToString(o.OutputKey)] = aws.ToString(o.OutputValue)
	}
	return m
}

func Test_CloudFormationService_Deploy(t *testing.T) {
	cases := []struct {
		name        string
		setupMock   func(*mockCloudFormationClient)
		wantOutputs map[string]string
	}{
		{
			name: "create happy path",
			wantOutputs: map[string]string{
				"InstanceId":       "i-create001",
				"AvailabilityZone": "us-east-1a",
			},
			setupMock: func(m *mockCloudFormationClient) {
				executeChangeSetCalled := false

				m.CreateChangeSetFn = func(ctx context.Context, input *cloudformation.CreateChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error) {
					if input.ChangeSetType == types.ChangeSetTypeCreate {
						return &cloudformation.CreateChangeSetOutput{}, nil
					}
					return nil, fmt.Errorf("unexpected changeset type %s", input.ChangeSetType)
				}

				m.DescribeChangeSetFn = func(ctx context.Context, input *cloudformation.DescribeChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error) {
					return &cloudformation.DescribeChangeSetOutput{
						Status: types.ChangeSetStatusCreateComplete,
					}, nil
				}

				m.DescribeStacksFn = func(ctx context.Context, input *cloudformation.DescribeStacksInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
					if executeChangeSetCalled {
						return &cloudformation.DescribeStacksOutput{
							Stacks: []types.Stack{
								{
									StackStatus: types.StackStatusCreateComplete,
									Outputs: []types.Output{
										{OutputKey: aws.String("InstanceId"), OutputValue: aws.String("i-create001")},
										{OutputKey: aws.String("AvailabilityZone"), OutputValue: aws.String("us-east-1a")},
									},
								},
							},
						}, nil
					}

					return nil, &smithy.GenericAPIError{
						Code:    "ValidationError",
						Message: "does not exist",
					}
				}

				m.ExecuteChangeSetFn = func(ctx context.Context, input *cloudformation.ExecuteChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.ExecuteChangeSetOutput, error) {
					executeChangeSetCalled = true
					return &cloudformation.ExecuteChangeSetOutput{}, nil
				}
			},
		},

		{
			name: "update happy path",
			wantOutputs: map[string]string{
				"InstanceId":       "i-update001",
				"AvailabilityZone": "us-east-1b",
			},
			setupMock: func(m *mockCloudFormationClient) {
				m.CreateChangeSetFn = func(ctx context.Context, input *cloudformation.CreateChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error) {
					if input.ChangeSetType == types.ChangeSetTypeUpdate {
						return &cloudformation.CreateChangeSetOutput{}, nil
					}
					return nil, fmt.Errorf("unexpected changeset type %s", input.ChangeSetType)
				}

				m.DescribeChangeSetFn = func(ctx context.Context, input *cloudformation.DescribeChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error) {
					return &cloudformation.DescribeChangeSetOutput{
						Status: types.ChangeSetStatusCreateComplete,
					}, nil
				}

				m.DescribeStacksFn = func(ctx context.Context, input *cloudformation.DescribeStacksInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
					return &cloudformation.DescribeStacksOutput{
						Stacks: []types.Stack{
							{
								StackStatus: types.StackStatusUpdateComplete,
								Outputs: []types.Output{
									{OutputKey: aws.String("InstanceId"), OutputValue: aws.String("i-update001")},
									{OutputKey: aws.String("AvailabilityZone"), OutputValue: aws.String("us-east-1b")},
								},
							},
						},
					}, nil
				}

				m.ExecuteChangeSetFn = func(ctx context.Context, input *cloudformation.ExecuteChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.ExecuteChangeSetOutput, error) {
					return &cloudformation.ExecuteChangeSetOutput{}, nil
				}
			},
		},

		{
			name: "empty changeset returns current stack outputs",
			wantOutputs: map[string]string{
				"InstanceId":       "i-existing001",
				"AvailabilityZone": "us-east-1c",
			},
			setupMock: func(m *mockCloudFormationClient) {
				stackOutputs := []types.Output{
					{OutputKey: aws.String("InstanceId"), OutputValue: aws.String("i-existing001")},
					{OutputKey: aws.String("AvailabilityZone"), OutputValue: aws.String("us-east-1c")},
				}

				// Stack exists → UPDATE changeset type, and provides outputs on the final DescribeStacks call.
				m.DescribeStacksFn = func(ctx context.Context, input *cloudformation.DescribeStacksInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
					return &cloudformation.DescribeStacksOutput{
						Stacks: []types.Stack{{StackStatus: types.StackStatusUpdateComplete, Outputs: stackOutputs}},
					}, nil
				}

				m.CreateChangeSetFn = func(ctx context.Context, input *cloudformation.CreateChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error) {
					return &cloudformation.CreateChangeSetOutput{}, nil
				}

				// Waiter polls DescribeChangeSet; returning FAILED with the no-changes reason triggers the empty-changeset path.
				m.DescribeChangeSetFn = func(ctx context.Context, input *cloudformation.DescribeChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error) {
					return &cloudformation.DescribeChangeSetOutput{
						Status:       types.ChangeSetStatusFailed,
						StatusReason: aws.String("The submitted information didn't contain changes. Submit different information to create a change set."),
					}, nil
				}

				m.DeleteChangeSetFn = func(ctx context.Context, input *cloudformation.DeleteChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.DeleteChangeSetOutput, error) {
					return &cloudformation.DeleteChangeSetOutput{}, nil
				}
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockCloudFormationClient{}
			tt.setupMock(mock)
			svc := &cloudFormationService{mock}

			out, err := svc.Deploy(t.Context(), &DeployInput{
				StackName:      aws.String("test-stack"),
				DeploymentName: aws.String("my-deployment"),
			})

			if err != nil {
				t.Fatal(err)
			}

			got := outputByKey(out.Outputs)
			for key, want := range tt.wantOutputs {
				if got[key] != want {
					t.Errorf("output %s: want %q, got %q", key, want, got[key])
				}
			}
		})
	}

	// t.Run("create happy path", func(t *testing.T) {
	// 	executeChangeSetCalled := false

	// 	mock := &mockCloudFormationClient{
	// 		DescribeStacksFn: func(ctx context.Context, input *cloudformation.DescribeStacksInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeStacksOutput, error) {
	// 			if executeChangeSetCalled {
	// 				return &cloudformation.DescribeStacksOutput{
	// 					Stacks: []types.Stack{
	// 						{
	// 							StackStatus: types.StackStatusUpdateComplete,
	// 						},
	// 					},
	// 				}, nil
	// 			}

	// 			return nil, &smithy.GenericAPIError{
	// 				Code:    "ValidationError",
	// 				Message: "does not exist",
	// 			}
	// 		},
	// 		CreateChangeSetFn: func(ctx context.Context, input *cloudformation.CreateChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.CreateChangeSetOutput, error) {
	// 			if input.ChangeSetType == types.ChangeSetTypeCreate {
	// 				return &cloudformation.CreateChangeSetOutput{}, nil
	// 			}
	// 			return nil, fmt.Errorf("unexpected changeset type %s", input.ChangeSetType)
	// 		},
	// 		DescribeChangeSetFn: func(ctx context.Context, input *cloudformation.DescribeChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.DescribeChangeSetOutput, error) {
	// 			return &cloudformation.DescribeChangeSetOutput{
	// 				Status: types.ChangeSetStatusCreateComplete,
	// 			}, nil
	// 		},
	// 		ExecuteChangeSetFn: func(ctx context.Context, input *cloudformation.ExecuteChangeSetInput, o ...func(*cloudformation.Options)) (*cloudformation.ExecuteChangeSetOutput, error) {
	// 			executeChangeSetCalled = true
	// 			return &cloudformation.ExecuteChangeSetOutput{}, nil
	// 		},
	// 	}

	// 	svc := &cloudFormationService{mock}

	// 	_, err := svc.Deploy(t.Context(), &DeployInput{
	// 		StackName:      aws.String("test-stack"),
	// 		DeploymentName: aws.String("my-deployment"),
	// 	}, time.Second*10)

	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// })
}

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
