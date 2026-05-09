package main

import (
	"fmt"
	"os"

	"github.com/bernos/bastion/internal/config"
	"github.com/spf13/cobra"
)

func NewRootCommand(cfg *config.Config) (*cobra.Command, error) {
	var cfgFile string

	cmd := &cobra.Command{
		Use: "bastion",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return config.Initialize(cfg, cmd, cfgFile)
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

	cmd.AddCommand(up, down)

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
