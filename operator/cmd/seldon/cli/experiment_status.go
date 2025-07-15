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

const (
	flagExperimentWait = "wait"
)

func createExperimentStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <experimentName>",
		Short: "get status for experiment",
		Long:  `get status for experiment`,
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
			wait, err := flags.GetBool(flagExperimentWait)
			if err != nil {
				return err
			}
			timeout, err := flags.GetInt64(flagTimeout)
			if err != nil {
				return err
			}
			experimentName := args[0]

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, schedulerHostIsSet, authority, verbose)
			if err != nil {
				return err
			}

			res, err := schedulerClient.ExperimentStatus(experimentName, wait, time.Duration(timeout*int64(time.Second)))
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
	flags.BoolP(flagExperimentWait, "w", false, "wait for experiment to be active")

	return cmd
}
