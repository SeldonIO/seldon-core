package cli

import (
	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
	"github.com/spf13/cobra"
)

func createConfigAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add config",
		Long:  `add config`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			configKey := args[0]
			configPath := args[1]
			configs, err := cli.LoadSeldonCLIConfigs()
			if err != nil {
				return err
			}
			return configs.Add(configKey, configPath)
		},
	}

	return cmd
}
