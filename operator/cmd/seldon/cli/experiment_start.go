package cli

import (
	"os"

	"k8s.io/utils/env"

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
			filename, err := cmd.Flags().GetString(fileFlag)
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
			schedulerClient, err := cli.NewSchedulerClient(schedulerHost)
			if err != nil {
				return err
			}
			err = schedulerClient.StartExperiment(loadFile(filename), showRequest, showResponse)
			return err
		},
	}
	cmdStartExperiment.Flags().StringP(fileFlag, "f", "", "model file to load")
	if err := cmdStartExperiment.MarkFlagRequired(fileFlag); err != nil {
		os.Exit(-1)
	}
	cmdStartExperiment.Flags().String(schedulerHostFlag, env.GetString(EnvScheduler, DefaultScheduleHost), "seldon scheduler host")
	return cmdStartExperiment
}
