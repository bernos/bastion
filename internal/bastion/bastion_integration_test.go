//go:build integration

package bastion_test

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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
