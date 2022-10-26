package cli

import (
	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
	"k8s.io/utils/env"
)

func createServerList() *cobra.Command {
	cmdServerList := &cobra.Command{
		Use:   "list",
		Short: "get list of servers",
		Long:  `get the available servers, their replicas and loaded models`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(flagSchedulerHost)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(flagAuthority)
			if err != nil {
				return err
			}

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, authority)
			if err != nil {
				return err
			}

			err = schedulerClient.ListServers()
			return err
		},
	}

	cmdServerList.Flags().String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	cmdServerList.Flags().String(flagAuthority, "", helpAuthority)

	return cmdServerList
}
