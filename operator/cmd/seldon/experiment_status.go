package main

import (
	"os"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

const (
	experimentWaitFlag = "wait"
)

func createExperimentStatus() *cobra.Command {
	cmdExperimentStatus := &cobra.Command{
		Use:   "status",
		Short: "get status for experiment",
		Long:  `get status for experiment`,
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
			wait, err := cmd.Flags().GetBool(experimentWaitFlag)
			if err != nil {
				return err
			}
			schedulerClient := cli.NewSchedulerClient(schedulerHost, schedulerPort)
			err = schedulerClient.ExperimentStatus(experimentName, verbose, wait)
			return err
		},
	}
	cmdExperimentStatus.Flags().StringP(experimentFlagName, "e", "", "experiment to stop")
	if err := cmdExperimentStatus.MarkFlagRequired(experimentFlagName); err != nil {
		os.Exit(-1)
	}
	cmdExperimentStatus.Flags().String(schedulerHostFlag, "0.0.0.0", "seldon scheduler host")
	cmdExperimentStatus.Flags().Int(schedulerPortFlag, 9004, "seldon scheduler port")
	cmdExperimentStatus.Flags().BoolP(experimentWaitFlag, "w", false, "wait for experiment to be active")
	return cmdExperimentStatus
}
