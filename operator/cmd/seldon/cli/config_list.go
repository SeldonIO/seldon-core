/*
Copyright 2023 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"github.com/spf13/cobra"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
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
