package main

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
			experimentName, err := cmd.Flags().GetString(experimentFlagName)
			if err != nil {
				return err
			}
			verbose, err := cmd.Flags().GetBool(verboseFlag)
			if err != nil {
				return err
			}
			schedulerClient := cli.NewSchedulerClient(schedulerHost, schedulerPort)
			err = schedulerClient.StopExperiment(experimentName, verbose)
			return err
		},
	}
	cmdStopExperiment.Flags().StringP(experimentFlagName, "e", "", "experiment to stop")
	if err := cmdStopExperiment.MarkFlagRequired(experimentFlagName); err != nil {
		os.Exit(-1)
	}
	cmdStopExperiment.Flags().String(schedulerHostFlag, "0.0.0.0", "seldon scheduler host")
	cmdStopExperiment.Flags().Int(schedulerPortFlag, 9004, "seldon scheduler port")
	return cmdStopExperiment
}
