package dependencies

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/bernos/bastion/internal/bastion"
	"github.com/bernos/bastion/internal/config"
	"github.com/bernos/bastion/pkg/aws/cloudformationservice"
)

type Dependencies struct {
	awsConfig            *aws.Config
	cloudFormationClient *cloudformation.Client
	stsClient            *sts.Client
	bastionService       bastion.BastionService
}

func New(cfg *config.Config) *Dependencies {
	return &Dependencies{}
}

func (d *Dependencies) AwsConfig(ctx context.Context) (aws.Config, error) {
	if d.awsConfig != nil {
		return *d.awsConfig, nil
	}

	c, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, err
	}

	d.awsConfig = &c

	return c, nil
}

func (d *Dependencies) CloudFormationClient(ctx context.Context) (*cloudformation.Client, error) {
	if d.cloudFormationClient != nil {
		return d.cloudFormationClient, nil
	}

	c, err := d.AwsConfig(ctx)
	if err != nil {
		return nil, err
	}

	d.cloudFormationClient = cloudformation.NewFromConfig(c)

	return d.cloudFormationClient, nil
}

func (d *Dependencies) BastionService(ctx context.Context) (bastion.BastionService, error) {
	if d.bastionService != nil {
		return d.bastionService, nil
	}
	cfnClient, err := d.CloudFormationClient(ctx)
	if err != nil {
		return nil, err
	}
	d.bastionService = bastion.NewBastionService(cloudformationservice.New(cfnClient))
	return d.bastionService, nil
}

func (d *Dependencies) STSClient(ctx context.Context) (*sts.Client, error) {
	if d.stsClient != nil {
		return d.stsClient, nil
	}

	c, err := d.AwsConfig(ctx)
	if err != nil {
		return nil, err
	}

	d.stsClient = sts.NewFromConfig(c)

	return d.stsClient, nil
}
