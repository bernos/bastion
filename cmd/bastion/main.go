package main

import (
	"fmt"
	"os"

	"github.com/bernos/bastion/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewRootCommand(cfg *config.Config) (*cobra.Command, error) {
	var cfgFile string

	v := viper.New()
	cmd := &cobra.Command{
		Use: "bastion",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cfgFile != "" {
				v.SetConfigFile(cfgFile)
			} else {
				home, err := os.UserConfigDir()
				if err != nil {
					return err
				}
				v.AddConfigPath(".")
				v.AddConfigPath(home + "/bastion")
				v.SetConfigName("config")
				v.SetConfigType("yaml")
			}
			return config.Initialize(cfg, v, cmd)
		},
	}

	up, err := NewUpCommand(cfg)
	if err != nil {
		return nil, err
	}

	down, err := NewDownCommand(cfg)
	if err != nil {
		return nil, err
	}

	sshCmd, err := NewSSHCommand(cfg)
	if err != nil {
		return nil, err
	}

	proxyCmd, err := NewProxyCommand(cfg)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(up, down, sshCmd, proxyCmd)

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")

	return cmd, nil
}

func main() {
	cfg := &config.Config{}

	cmd, err := NewRootCommand(cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
