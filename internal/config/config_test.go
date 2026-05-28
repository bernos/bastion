package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bernos/bastion/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// newTestCmd builds a command with the same flags registered by up.go.
func newTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("name", "", "")
	cmd.Flags().String("owner", "", "")
	cmd.Flags().String("subnet-id", "", "")
	cmd.Flags().String("vpc-id", "", "")
	cmd.Flags().String("region", "", "")
	cmd.Flags().String("ami-parameter-store-param-name", "", "")
	cmd.Flags().String("tags", "", "")
	return cmd
}

// writeTempConfig writes a YAML string to a temp file and returns its path.
func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(f, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return f
}

func TestInitialize_FromConfigFile(t *testing.T) {
	cfgFile := writeTempConfig(t, `
name: file-name
owner: owner-from-file
subnet-id: subnet-from-file
vpc-id: vpc-from-file
ami-parameter-store-param-name: param-name-from-file
`)
	v := viper.New()
	v.SetConfigFile(cfgFile)

	cfg := &config.Config{}
	if err := config.Initialize(cfg, v, newTestCmd()); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "file-name" {
		t.Errorf("Name: want %q, got %q", "file-name", cfg.Name)
	}
	if cfg.Owner != "owner-from-file" {
		t.Errorf("Owner: want %q, got %q", "owner-from-file", cfg.Owner)
	}
	if cfg.SubnetID != "subnet-from-file" {
		t.Errorf("SubnetID: want %q, got %q", "subnet-from-file", cfg.SubnetID)
	}
	if cfg.VPCID != "vpc-from-file" {
		t.Errorf("VPCID: want %q, got %q", "vpc-from-file", cfg.VPCID)
	}
	if cfg.AMIParameterStoreParamName != "param-name-from-file" {
		t.Errorf("ParameterStoreParamName: want %q, got %q", "param-name-from-file", cfg.AMIParameterStoreParamName)
	}
}

func TestInitialize_FromEnvVars(t *testing.T) {
	t.Setenv("BASTION_NAME", "env-name")
	t.Setenv("BASTION_OWNER", "owner-from-env")
	t.Setenv("BASTION_SUBNET_ID", "subnet-from-env")
	t.Setenv("BASTION_VPC_ID", "vpc-from-env")
	t.Setenv("BASTION_AMI_PARAMETER_STORE_PARAM_NAME", "param-name-from-env")

	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), newTestCmd()); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "env-name" {
		t.Errorf("Name: want %q, got %q", "env-name", cfg.Name)
	}
	if cfg.Owner != "owner-from-env" {
		t.Errorf("Owner: want %q, got %q", "owner-from-env", cfg.Owner)
	}
	if cfg.SubnetID != "subnet-from-env" {
		t.Errorf("SubnetID: want %q, got %q", "subnet-from-env", cfg.SubnetID)
	}
	if cfg.VPCID != "vpc-from-env" {
		t.Errorf("VPCID: want %q, got %q", "vpc-from-env", cfg.VPCID)
	}
	if cfg.AMIParameterStoreParamName != "param-name-from-env" {
		t.Errorf("ParameterStoreParamName: want %q, got %q", "param-name-from-env", cfg.AMIParameterStoreParamName)
	}
}

func TestInitialize_FromFlags(t *testing.T) {
	cmd := newTestCmd()
	if err := cmd.Flags().Set("name", "flag-name"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("owner", "owner-from-flag"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("subnet-id", "subnet-from-flag"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("vpc-id", "vpc-from-flag"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("ami-parameter-store-param-name", "param-name-from-flag"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), cmd); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "flag-name" {
		t.Errorf("Name: want %q, got %q", "flag-name", cfg.Name)
	}
	if cfg.Owner != "owner-from-flag" {
		t.Errorf("Owner: want %q, got %q", "owner-from-flag", cfg.Owner)
	}
	if cfg.SubnetID != "subnet-from-flag" {
		t.Errorf("SubnetID: want %q, got %q", "subnet-from-flag", cfg.SubnetID)
	}
	if cfg.VPCID != "vpc-from-flag" {
		t.Errorf("VPCID: want %q, got %q", "vpc-from-flag", cfg.VPCID)
	}
	if cfg.AMIParameterStoreParamName != "param-name-from-flag" {
		t.Errorf("ParameterStoreParamName: want %q, got %q", "param-name-from-flag", cfg.AMIParameterStoreParamName)
	}
}

func TestInitialize_Precedence_FlagsOverEnv(t *testing.T) {
	t.Setenv("BASTION_NAME", "env-name")
	t.Setenv("BASTION_SUBNET_ID", "subnet-from-env")
	t.Setenv("BASTION_VPC_ID", "vpc-from-env")

	cmd := newTestCmd()
	if err := cmd.Flags().Set("name", "flag-name"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("subnet-id", "subnet-from-flag"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("vpc-id", "vpc-from-flag"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), cmd); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "flag-name" {
		t.Errorf("Name: want flag value %q, got %q", "flag-name", cfg.Name)
	}
	if cfg.SubnetID != "subnet-from-flag" {
		t.Errorf("SubnetID: want flag value %q, got %q", "subnet-from-flag", cfg.SubnetID)
	}
	if cfg.VPCID != "vpc-from-flag" {
		t.Errorf("VPCID: want flag value %q, got %q", "vpc-from-flag", cfg.VPCID)
	}
}

func TestInitialize_Precedence_FlagsOverConfigFile(t *testing.T) {
	cfgFile := writeTempConfig(t, `
name: file-name
subnet-id: subnet-from-file
vpc-id: vpc-from-file
`)
	cmd := newTestCmd()
	if err := cmd.Flags().Set("name", "flag-name"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("subnet-id", "subnet-from-flag"); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Flags().Set("vpc-id", "vpc-from-flag"); err != nil {
		t.Fatal(err)
	}

	v := viper.New()
	v.SetConfigFile(cfgFile)

	cfg := &config.Config{}
	if err := config.Initialize(cfg, v, cmd); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "flag-name" {
		t.Errorf("Name: want flag value %q, got %q", "flag-name", cfg.Name)
	}
	if cfg.SubnetID != "subnet-from-flag" {
		t.Errorf("SubnetID: want flag value %q, got %q", "subnet-from-flag", cfg.SubnetID)
	}
	if cfg.VPCID != "vpc-from-flag" {
		t.Errorf("VPCID: want flag value %q, got %q", "vpc-from-flag", cfg.VPCID)
	}
}

func TestInitialize_Precedence_EnvOverConfigFile(t *testing.T) {
	cfgFile := writeTempConfig(t, `
name: file-name
subnet-id: subnet-from-file
vpc-id: vpc-from-file
`)
	t.Setenv("BASTION_NAME", "env-name")
	t.Setenv("BASTION_SUBNET_ID", "subnet-from-env")
	t.Setenv("BASTION_VPC_ID", "vpc-from-env")

	v := viper.New()
	v.SetConfigFile(cfgFile)

	cfg := &config.Config{}
	if err := config.Initialize(cfg, v, newTestCmd()); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "env-name" {
		t.Errorf("Name: want env value %q, got %q", "env-name", cfg.Name)
	}
	if cfg.SubnetID != "subnet-from-env" {
		t.Errorf("SubnetID: want env value %q, got %q", "subnet-from-env", cfg.SubnetID)
	}
	if cfg.VPCID != "vpc-from-env" {
		t.Errorf("VPCID: want env value %q, got %q", "vpc-from-env", cfg.VPCID)
	}
}

// TestInitialize_NoConfigFile_NoError verifies that Initialize succeeds when
// the Viper instance has no config paths configured (ReadInConfig returns
// ConfigFileNotFoundError, which Initialize silently ignores).
func TestInitialize_NoConfigFile_NoError(t *testing.T) {
	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), newTestCmd()); err != nil {
		t.Errorf("expected no error when no config file is configured, got: %v", err)
	}
}

// TestInitialize_ExplicitMissingConfigFile_Error verifies that pointing the
// Viper instance at a specific path that does not exist returns an error.
func TestInitialize_ExplicitMissingConfigFile_Error(t *testing.T) {
	v := viper.New()
	v.SetConfigFile(filepath.Join(t.TempDir(), "nonexistent.yaml"))

	cfg := &config.Config{}
	if err := config.Initialize(cfg, v, newTestCmd()); err == nil {
		t.Error("expected error for explicit non-existent config file path, got nil")
	}
}

func TestInitialize_Region_FromFlag(t *testing.T) {
	cmd := newTestCmd()
	if err := cmd.Flags().Set("region", "ap-southeast-2"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), cmd); err != nil {
		t.Fatal(err)
	}

	if cfg.Region != "ap-southeast-2" {
		t.Errorf("Region: want %q, got %q", "ap-southeast-2", cfg.Region)
	}
}

func TestInitialize_Region_FromEnvVar(t *testing.T) {
	t.Setenv("BASTION_REGION", "us-east-1")

	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), newTestCmd()); err != nil {
		t.Fatal(err)
	}

	if cfg.Region != "us-east-1" {
		t.Errorf("Region: want %q, got %q", "us-east-1", cfg.Region)
	}
}

func TestInitialize_Region_FromConfigFile(t *testing.T) {
	cfgFile := writeTempConfig(t, `region: eu-west-1`)
	v := viper.New()
	v.SetConfigFile(cfgFile)

	cfg := &config.Config{}
	if err := config.Initialize(cfg, v, newTestCmd()); err != nil {
		t.Fatal(err)
	}

	if cfg.Region != "eu-west-1" {
		t.Errorf("Region: want %q, got %q", "eu-west-1", cfg.Region)
	}
}

func TestInitialize_Region_OmittedIsEmpty(t *testing.T) {
	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), newTestCmd()); err != nil {
		t.Fatal(err)
	}

	if cfg.Region != "" {
		t.Errorf("Region: want empty string, got %q", cfg.Region)
	}
}

func TestInitialize_AMIParam_Default(t *testing.T) {
	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), newTestCmd()); err != nil {
		t.Fatal(err)
	}

	if cfg.AMIParameterStoreParamName != config.DefaultAMIParameterStoreParamName {
		t.Errorf("AMIParameterStoreParamName: want %q, got %q", config.DefaultAMIParameterStoreParamName, cfg.AMIParameterStoreParamName)
	}
}

func TestInitialize_Tags_FromConfigFile(t *testing.T) {
	cfgFile := writeTempConfig(t, `
tags:
  env: prod
  team: platform
`)
	v := viper.New()
	v.SetConfigFile(cfgFile)

	cfg := &config.Config{}
	if err := config.Initialize(cfg, v, newTestCmd()); err != nil {
		t.Fatal(err)
	}

	if cfg.Tags["env"] != "prod" {
		t.Errorf("Tags[env]: want %q, got %q", "prod", cfg.Tags["env"])
	}
	if cfg.Tags["team"] != "platform" {
		t.Errorf("Tags[team]: want %q, got %q", "platform", cfg.Tags["team"])
	}
}

func TestInitialize_Tags_FromEnvVar(t *testing.T) {
	t.Setenv("BASTION_TAGS", "env=staging,owner=sre")

	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), newTestCmd()); err != nil {
		t.Fatal(err)
	}

	if cfg.Tags["env"] != "staging" {
		t.Errorf("Tags[env]: want %q, got %q", "staging", cfg.Tags["env"])
	}
	if cfg.Tags["owner"] != "sre" {
		t.Errorf("Tags[owner]: want %q, got %q", "sre", cfg.Tags["owner"])
	}
}

func TestInitialize_Tags_FromFlag(t *testing.T) {
	cmd := newTestCmd()
	if err := cmd.Flags().Set("tags", "env=prod,team=ops"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), cmd); err != nil {
		t.Fatal(err)
	}

	if cfg.Tags["env"] != "prod" {
		t.Errorf("Tags[env]: want %q, got %q", "prod", cfg.Tags["env"])
	}
	if cfg.Tags["team"] != "ops" {
		t.Errorf("Tags[team]: want %q, got %q", "ops", cfg.Tags["team"])
	}
}

func TestInitialize_Tags_FlagOverridesEnvVar(t *testing.T) {
	t.Setenv("BASTION_TAGS", "env=prod")

	cmd := newTestCmd()
	if err := cmd.Flags().Set("tags", "env=staging"); err != nil {
		t.Fatal(err)
	}

	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), cmd); err != nil {
		t.Fatal(err)
	}

	if cfg.Tags["env"] != "staging" {
		t.Errorf("Tags[env]: want flag value %q, got %q", "staging", cfg.Tags["env"])
	}
}

func TestInitialize_Tags_AbsentIsEmpty(t *testing.T) {
	cfg := &config.Config{}
	if err := config.Initialize(cfg, viper.New(), newTestCmd()); err != nil {
		t.Fatal(err)
	}

	if len(cfg.Tags) != 0 {
		t.Errorf("Tags: want empty map, got %v", cfg.Tags)
	}
}

func TestInitialize_InvalidConfigFile_Error(t *testing.T) {
	cfgFile := writeTempConfig(t, `{this is not: valid yaml:`)
	v := viper.New()
	v.SetConfigFile(cfgFile)

	cfg := &config.Config{}
	if err := config.Initialize(cfg, v, newTestCmd()); err == nil {
		t.Error("expected error for invalid config file, got nil")
	}
}
