package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createExperimentStop() *cobra.Command {
	cmdStopExperiment := &cobra.Command{
		Use:   "stop <experimentName>",
		Short: "stop an experiment",
		Long:  `stop an experiment`,
		Args:  cobra.ExactArgs(1),
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
			schedulerClient := cli.NewSchedulerClient(schedulerHost)
			experimentName := args[0]
			err = schedulerClient.StopExperiment(experimentName, showRequest, showResponse)
			return err
		},
	}
	cmdStopExperiment.Flags().String(schedulerHostFlag, env.GetString(EnvScheduler, DefaultScheduleHost), "seldon scheduler host")
	return cmdStopExperiment
}
