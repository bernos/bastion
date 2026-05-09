package main

import (
	"fmt"

	"github.com/bernos/bastion/internal/config"
	"github.com/spf13/cobra"
)

func NewDownCommand(cfg *config.Config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Down command short",
		Long:  "Down command long",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("This is the down command")
			return nil
		},
	}

	return cmd, nil
}
