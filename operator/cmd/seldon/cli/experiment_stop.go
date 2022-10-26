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
			schedulerHost, err := cmd.Flags().GetString(flagSchedulerHost)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(flagAuthority)
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

			experimentName := args[0]
			err = schedulerClient.StopExperiment(experimentName, showRequest, showResponse)
			return err
		},
	}

	cmdStopExperiment.Flags().String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	cmdStopExperiment.Flags().String(flagAuthority, "", helpAuthority)

	return cmdStopExperiment
}
