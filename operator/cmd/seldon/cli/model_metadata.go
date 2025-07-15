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
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

func createModelMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metadata <modelName>",
		Short: "get model metadata",
		Long:  `get model metadata`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			inferenceHostIsSet := flags.Changed(flagInferenceHost)
			inferenceHost, err := flags.GetString(flagInferenceHost)
			if err != nil {
				return err
			}
			authority, err := flags.GetString(flagAuthority)
			if err != nil {
				return err
			}
			modelName := args[0]

			inferenceClient, err := cli.NewInferenceClient(inferenceHost, inferenceHostIsSet)
			if err != nil {
				return err
			}

			err = inferenceClient.ModelMetadata(modelName, authority)
			return err
		},
	}

	flags := cmd.Flags()
	flags.String(flagInferenceHost, env.GetString(envInfer, defaultInferHost), helpInferenceHost)
	flags.String(flagAuthority, "", helpAuthority)

	return cmd
}
