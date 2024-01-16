/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"github.com/spf13/cobra"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

func createConfigAdd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <key> <path>",
		Short: "add config for given path under the supplied key",
		Long:  `add config for given path under the supplied key. You should provide the full path to the config file`,
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
