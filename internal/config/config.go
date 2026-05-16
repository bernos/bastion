package config

import (
	"errors"
	"os"
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

func Initialize(cfg *Config, v *viper.Viper, cmd *cobra.Command, cfgFile string) error {
	v.SetEnvPrefix(EnvVarPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserConfigDir()
		cobra.CheckErr(err)

		v.AddConfigPath(".")
		v.AddConfigPath(home + "/bastion")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

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
