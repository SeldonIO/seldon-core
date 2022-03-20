package main

import (
	"os"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createExperimentStart() *cobra.Command {
	cmdStartExperiment := &cobra.Command{
		Use:   "start",
		Short: "start an experiment",
		Long:  `start an experiment`,
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
			filename, err := cmd.Flags().GetString(fileFlag)
			if err != nil {
				return err
			}
			verbose, err := cmd.Flags().GetBool(verboseFlag)
			if err != nil {
				return err
			}
			schedulerClient := cli.NewSchedulerClient(schedulerHost, schedulerPort)
			err = schedulerClient.StartExperiment(loadFile(filename), verbose)
			return err
		},
	}
	cmdStartExperiment.Flags().StringP(fileFlag, "f", "", "model file to load")
	if err := cmdStartExperiment.MarkFlagRequired(fileFlag); err != nil {
		os.Exit(-1)
	}
	cmdStartExperiment.Flags().String(schedulerHostFlag, "0.0.0.0", "seldon scheduler host")
	cmdStartExperiment.Flags().Int(schedulerPortFlag, 9004, "seldon scheduler port")
	return cmdStartExperiment
}
