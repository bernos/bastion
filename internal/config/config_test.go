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

// reset clears Viper's global state and redirects the user config dir to an
// empty temp dir before and after each test. The redirect prevents Viper's
// automatic config-file search from picking up a real config file from the
// developer's machine. XDG_CONFIG_HOME covers Linux; HOME covers macOS.
func reset(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("HOME", tmp)
}

func TestInitialize_FromConfigFile(t *testing.T) {
	reset(t)
	cfgFile := writeTempConfig(t, `
name: file-name
subnet-id: subnet-from-file
vpc-id: vpc-from-file
`)
	cfg := &config.Config{}
	if err := config.Initialize(cfg, newTestCmd(), cfgFile); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "file-name" {
		t.Errorf("Name: want %q, got %q", "file-name", cfg.Name)
	}
	if cfg.SubnetID != "subnet-from-file" {
		t.Errorf("SubnetID: want %q, got %q", "subnet-from-file", cfg.SubnetID)
	}
	if cfg.VPCID != "vpc-from-file" {
		t.Errorf("VPCID: want %q, got %q", "vpc-from-file", cfg.VPCID)
	}
}

func TestInitialize_FromEnvVars(t *testing.T) {
	reset(t)
	t.Setenv("BASTION_NAME", "env-name")
	t.Setenv("BASTION_SUBNET_ID", "subnet-from-env")
	t.Setenv("BASTION_VPC_ID", "vpc-from-env")

	cfg := &config.Config{}
	if err := config.Initialize(cfg, newTestCmd(), ""); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "env-name" {
		t.Errorf("Name: want %q, got %q", "env-name", cfg.Name)
	}
	if cfg.SubnetID != "subnet-from-env" {
		t.Errorf("SubnetID: want %q, got %q", "subnet-from-env", cfg.SubnetID)
	}
	if cfg.VPCID != "vpc-from-env" {
		t.Errorf("VPCID: want %q, got %q", "vpc-from-env", cfg.VPCID)
	}
}

func TestInitialize_FromFlags(t *testing.T) {
	reset(t)
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
	if err := config.Initialize(cfg, cmd, ""); err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "flag-name" {
		t.Errorf("Name: want %q, got %q", "flag-name", cfg.Name)
	}
	if cfg.SubnetID != "subnet-from-flag" {
		t.Errorf("SubnetID: want %q, got %q", "subnet-from-flag", cfg.SubnetID)
	}
	if cfg.VPCID != "vpc-from-flag" {
		t.Errorf("VPCID: want %q, got %q", "vpc-from-flag", cfg.VPCID)
	}
}

func TestInitialize_Precedence_FlagsOverEnv(t *testing.T) {
	reset(t)
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
	if err := config.Initialize(cfg, cmd, ""); err != nil {
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
	reset(t)
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

	cfg := &config.Config{}
	if err := config.Initialize(cfg, cmd, cfgFile); err != nil {
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
	reset(t)
	cfgFile := writeTempConfig(t, `
name: file-name
subnet-id: subnet-from-file
vpc-id: vpc-from-file
`)
	t.Setenv("BASTION_NAME", "env-name")
	t.Setenv("BASTION_SUBNET_ID", "subnet-from-env")
	t.Setenv("BASTION_VPC_ID", "vpc-from-env")

	cfg := &config.Config{}
	if err := config.Initialize(cfg, newTestCmd(), cfgFile); err != nil {
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

// TestInitialize_NoConfigFile_NoError verifies that when no config file path is
// given and none is found during automatic search, Initialize succeeds.
func TestInitialize_NoConfigFile_NoError(t *testing.T) {
	reset(t)
	cfg := &config.Config{}
	if err := config.Initialize(cfg, newTestCmd(), ""); err != nil {
		t.Errorf("expected no error when no config file exists, got: %v", err)
	}
}

// TestInitialize_ExplicitMissingConfigFile_Error verifies that pointing Initialize
// at a specific path that does not exist returns an error.
func TestInitialize_ExplicitMissingConfigFile_Error(t *testing.T) {
	reset(t)
	cfg := &config.Config{}
	if err := config.Initialize(cfg, newTestCmd(), filepath.Join(t.TempDir(), "nonexistent.yaml")); err == nil {
		t.Error("expected error for explicit non-existent config file path, got nil")
	}
}

func TestInitialize_InvalidConfigFile_Error(t *testing.T) {
	reset(t)
	cfgFile := writeTempConfig(t, `{this is not: valid yaml:`)
	cfg := &config.Config{}
	if err := config.Initialize(cfg, newTestCmd(), cfgFile); err == nil {
		t.Error("expected error for invalid config file, got nil")
	}
}
