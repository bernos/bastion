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

func Initialize(cfg *Config, cmd *cobra.Command, cfgFile string) error {
	viper.SetEnvPrefix(EnvVarPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserConfigDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(".")
		viper.AddConfigPath(home + "/bastion")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &notFoundErr) {
			return err
		}
	}

	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		return err
	}

	return viper.Unmarshal(cfg)
}
