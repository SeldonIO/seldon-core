package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createServerStatus() *cobra.Command {
	cmdServerStatus := &cobra.Command{
		Use:   "status",
		Short: "get status for server",
		Long:  `get the status for a server`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(schedulerHostFlag)
			if err != nil {
				return err
			}
			showRequest, err := cmd.Flags().GetBool(showRequestFlag)
			if err != nil {
				return err
			}
			showResponse, err := cmd.Flags().GetBool(showResponseFlag)
			if err != nil {
				return err
			}
			serverName := args[0]

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost)
			if err != nil {
				return err
			}
			err = schedulerClient.ServerStatus(serverName, showRequest, showResponse)
			return err
		},
	}

	cmdServerStatus.Flags().String(schedulerHostFlag, env.GetString(EnvScheduler, DefaultScheduleHost), "seldon scheduler host")

	return cmdServerStatus
}
