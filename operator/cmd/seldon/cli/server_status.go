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

func createServerStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "get status for server",
		Long:  `get the status for a server`,
		Args:  cobra.MinimumNArgs(1),
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
			showRequest, err := flags.GetBool(flagShowRequest)
			if err != nil {
				return err
			}
			showResponse, err := flags.GetBool(flagShowResponse)
			if err != nil {
				return err
			}
			verbose, err := flags.GetBool(flagVerbose)
			if err != nil {
				return err
			}
			serverName := args[0]

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, schedulerHostIsSet, authority, verbose)
			if err != nil {
				return err
			}

			err = schedulerClient.ServerStatus(serverName, showRequest, showResponse)
			return err
		},
	}

	flags := cmd.Flags()
	flags.BoolP(flagVerbose, "v", false, "verbose output")
	flags.BoolP(flagShowRequest, "r", false, "show request")
	flags.BoolP(flagShowResponse, "o", true, "show response")
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagAuthority, "", helpAuthority)

	return cmd
}
