package cli

import (
	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
	"github.com/spf13/cobra"
)

func createConfigActivate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "activate config",
		Long:  `activate config`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			configKey := args[0]
			configs, err := cli.LoadSeldonCLIConfigs()
			if err != nil {
				return err
			}
			return configs.Activate(configKey)
		},
	}

	return cmd
}
