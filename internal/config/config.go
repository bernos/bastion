package config

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	EnvVarPrefix                      = "BASTION"
	DefaultAMIParameterStoreParamName = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64"
)

type Config struct {
	Name                       string `mapstructure:"name"`
	Owner                      string `mapstructure:"owner"`
	SubnetID                   string `mapstructure:"subnet-id"`
	VPCID                      string `mapstructure:"vpc-id"`
	Region                     string `mapstructure:"region"`
	AMIParameterStoreParamName string `mapstructure:"ami-parameter-store-param-name"`
}

func Initialize(cfg *Config, v *viper.Viper, cmd *cobra.Command) error {
	v.SetEnvPrefix(EnvVarPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return err
		}
	}

	if err := v.BindPFlags(cmd.Flags()); err != nil {
		return err
	}

	if err := v.Unmarshal(cfg); err != nil {
		return err
	}

	if cfg.AMIParameterStoreParamName == "" {
		cfg.AMIParameterStoreParamName = DefaultAMIParameterStoreParamName
	}

	return nil
}
