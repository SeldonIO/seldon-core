package cli

import (
	"os"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createExperimentStop() *cobra.Command {
	cmdStopExperiment := &cobra.Command{
		Use:   "stop",
		Short: "stop an experiment",
		Long:  `stop an experiment`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(schedulerHostFlag)
			if err != nil {
				return err
			}
			schedulerPort, err := cmd.Flags().GetInt(schedulerPortFlag)
			if err != nil {
				return err
			}
			experimentName, err := cmd.Flags().GetString(experimentFlag)
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
			schedulerClient := cli.NewSchedulerClient(schedulerHost, schedulerPort)
			err = schedulerClient.StopExperiment(experimentName, showRequest, showResponse)
			return err
		},
	}
	cmdStopExperiment.Flags().StringP(experimentFlag, "e", "", "experiment to stop")
	if err := cmdStopExperiment.MarkFlagRequired(experimentFlag); err != nil {
		os.Exit(-1)
	}
	cmdStopExperiment.Flags().String(schedulerHostFlag, "0.0.0.0", "seldon scheduler host")
	cmdStopExperiment.Flags().Int(schedulerPortFlag, 9004, "seldon scheduler port")
	return cmdStopExperiment
}
