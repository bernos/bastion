package config

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	EnvVarPrefix = "BASTION"
)

type Config struct {
	Name     string `mapstructure:"name"`
	SubnetID string `mapstructure:"subnet-id"`
	VPCID    string `mapstructure:"vpc-id"`
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

	return v.Unmarshal(cfg)
}
