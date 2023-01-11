package cli

import (
	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
	"github.com/spf13/cobra"
)

func createConfigList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list configs",
		Long:  `list configs`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			configKey := ""
			if len(args) == 1 {
				configKey = args[0]
			}

			configs, err := cli.LoadSeldonCLIConfigs()
			if err != nil {
				return err
			}
			return configs.List(configKey)
		},
	}

	return cmd
}
