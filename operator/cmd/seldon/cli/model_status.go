/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"time"

	"github.com/spf13/cobra"
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

func createModelStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <modelName>",
		Short: "get status for model",
		Long:  `get the status for a model`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			schedulerHostIsSet := flags.Changed(flagSchedulerHost)
			schedulerHost, err := flags.GetString(flagSchedulerHost)
			if err != nil {
				return err
			}
			authority, err := flags.GetString(flagAuthority)
			if err != nil {
				return err
			}
			verbose, err := flags.GetBool(flagVerbose)
			if err != nil {
				return err
			}
			modelWaitCondition, err := flags.GetString(flagWaitCondition)
			if err != nil {
				return err
			}
			timeout, err := flags.GetInt64(flagTimeout)
			if err != nil {
				return err
			}
			modelName := args[0]

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, schedulerHostIsSet, authority, verbose)
			if err != nil {
				return err
			}

			res, err := schedulerClient.ModelStatus(modelName, modelWaitCondition, time.Duration(timeout*int64(time.Second)))
			if err == nil {
				cli.PrintProto(res)
			}
			return err
		},
	}

	flags := cmd.Flags()
	flags.Int64P(flagTimeout, "t", flagTimeoutDefault, "timeout seconds")
	flags.BoolP(flagVerbose, "v", false, "verbose output")
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagAuthority, "", helpAuthority)
	flags.StringP(flagWaitCondition, "w", "", "model wait condition")

	return cmd
}
