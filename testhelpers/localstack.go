package testhelpers

import (
	"context"
	"fmt"
	"net"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

func StartLocalStack(ctx context.Context) (aws.Config, *localstack.LocalStackContainer, error) {
	container, err := localstack.Run(ctx, "localstack/localstack:3.8")
	if err != nil {
		return aws.Config{}, nil, fmt.Errorf("failed to start localstack container: %w", err)
	}

	mappedPort, err := container.MappedPort(ctx, "4566/tcp")
	if err != nil {
		return aws.Config{}, nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return aws.Config{}, nil, fmt.Errorf("failed to get host: %w", err)
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("abc", "123", "456")),
		config.WithBaseEndpoint(fmt.Sprintf("http://%s", net.JoinHostPort(host, mappedPort.Port()))))

	return cfg, container, err
}
