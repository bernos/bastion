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
}

func TestInitialize_FromEnvVars(t *testing.T) {
	t.Setenv("BASTION_NAME", "env-name")
	t.Setenv("BASTION_OWNER", "owner-from-env")
	t.Setenv("BASTION_SUBNET_ID", "subnet-from-env")
	t.Setenv("BASTION_VPC_ID", "vpc-from-env")

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

func TestInitialize_InvalidConfigFile_Error(t *testing.T) {
	cfgFile := writeTempConfig(t, `{this is not: valid yaml:`)
	v := viper.New()
	v.SetConfigFile(cfgFile)

	cfg := &config.Config{}
	if err := config.Initialize(cfg, v, newTestCmd()); err == nil {
		t.Error("expected error for invalid config file, got nil")
	}
}
