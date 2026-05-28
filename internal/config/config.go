package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	EnvVarPrefix                      = "BASTION"
	DefaultAMIParameterStoreParamName = "/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64"
)

type Config struct {
	Name                       string            `mapstructure:"name"`
	Owner                      string            `mapstructure:"owner"`
	SubnetID                   string            `mapstructure:"subnet-id"`
	VPCID                      string            `mapstructure:"vpc-id"`
	Region                     string            `mapstructure:"region"`
	AMIParameterStoreParamName string            `mapstructure:"ami-parameter-store-param-name"`
	Tags                       map[string]string `mapstructure:"tags"`
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

	if err := v.Unmarshal(cfg, func(dc *mapstructure.DecoderConfig) {
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			stringToTagMapHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	}); err != nil {
		return err
	}

	if cfg.AMIParameterStoreParamName == "" {
		cfg.AMIParameterStoreParamName = DefaultAMIParameterStoreParamName
	}

	return nil
}

// stringToTagMapHookFunc returns a mapstructure decode hook that converts a
// "key=value,key2=value2" string into map[string]string. This handles env var
// and CLI flag inputs where Viper stores the value as a string.
func stringToTagMapHookFunc() mapstructure.DecodeHookFuncType {
	tagMapType := reflect.TypeOf(map[string]string{})
	return func(from, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() != reflect.String || to != tagMapType {
			return data, nil
		}
		s := data.(string)
		if s == "" {
			return map[string]string{}, nil
		}
		result := make(map[string]string)
		for _, pair := range strings.Split(s, ",") {
			parts := strings.SplitN(pair, "=", 2)
			if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
				return nil, fmt.Errorf("invalid tag format: %q (must be key=value)", pair)
			}
			result[parts[0]] = parts[1]
		}
		return result, nil
	}
}
