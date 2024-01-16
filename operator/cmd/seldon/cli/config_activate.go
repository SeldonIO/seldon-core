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

func createConfigActivate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <key>",
		Short: "activate config for given key",
		Long:  `activate config for given key`,
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
