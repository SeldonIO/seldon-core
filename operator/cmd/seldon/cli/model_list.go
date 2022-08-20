package cli

import (
	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
	"k8s.io/utils/env"
)

func createModelList() *cobra.Command {
	cmdModelList := &cobra.Command{
		Use:   "list",
		Short: "get list of models",
		Long:  `get the list of all models with their status`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(schedulerHostFlag)
			if err != nil {
				return err
			}
			schedulerClient, err := cli.NewSchedulerClient(schedulerHost)
			if err != nil {
				return err
			}
			err = schedulerClient.ListModels()
			return err
		},
	}
	cmdModelList.Flags().String(schedulerHostFlag, env.GetString(EnvScheduler, DefaultScheduleHost), "seldon scheduler host")
	return cmdModelList
}
