package cli

import (
	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
	"k8s.io/utils/env"
)

func createExperimentList() *cobra.Command {
	cmdExperimentList := &cobra.Command{
		Use:   "list",
		Short: "get list of experiments",
		Long:  `get list of experiments and whether they are active`,
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
			err = schedulerClient.ListExperiments()
			return err
		},
	}
	cmdExperimentList.Flags().String(schedulerHostFlag, env.GetString(EnvScheduler, DefaultScheduleHost), "seldon scheduler host")
	return cmdExperimentList
}
