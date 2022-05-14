package cli

import (
	"os"

	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

const serverNameFlag = "server-name"

func createServerStatus() *cobra.Command {
	cmdServerStatus := &cobra.Command{
		Use:   "status",
		Short: "get status for server",
		Long:  `get the status for a server`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(schedulerHostFlag)
			if err != nil {
				return err
			}
			serverName, err := cmd.Flags().GetString(serverNameFlag)
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
			schedulerClient := cli.NewSchedulerClient(schedulerHost)
			err = schedulerClient.ServerStatus(serverName, showRequest, showResponse)
			return err
		},
	}
	cmdServerStatus.Flags().StringP(serverNameFlag, "s", "", "server name")
	if err := cmdServerStatus.MarkFlagRequired(serverNameFlag); err != nil {
		os.Exit(-1)
	}
	cmdServerStatus.Flags().String(schedulerHostFlag, env.GetString(EnvScheduler, DefaultScheduleHost), "seldon scheduler host")
	return cmdServerStatus
}
