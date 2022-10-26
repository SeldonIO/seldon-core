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
			schedulerHost, err := cmd.Flags().GetString(flagSchedulerHost)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(flagAuthority)
			if err != nil {
				return err
			}
			filename, err := cmd.Flags().GetString(flagFile)
			if err != nil {
				return err
			}
			showRequest, err := cmd.Flags().GetBool(flagShowRequest)
			if err != nil {
				return err
			}
			showResponse, err := cmd.Flags().GetBool(flagShowResponse)
			if err != nil {
				return err
			}

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, authority)
			if err != nil {
				return err
			}

			err = schedulerClient.StartExperiment(loadFile(filename), showRequest, showResponse)
			return err
		},
	}

	cmdStartExperiment.Flags().String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	cmdStartExperiment.Flags().String(flagAuthority, "", helpAuthority)
	cmdStartExperiment.Flags().StringP(flagFile, "f", "", "experiment manifest file (YAML)")
	if err := cmdStartExperiment.MarkFlagRequired(flagFile); err != nil {
		os.Exit(-1)
	}

	return cmdStartExperiment
}
