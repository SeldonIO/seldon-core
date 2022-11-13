package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createExperimentStop() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop <experimentName>",
		Short: "stop an experiment",
		Long:  `stop an experiment`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			schedulerHostIsSet := flags.Changed(flagSchedulerHost)
			schedulerHost, err := flags.GetString(flagSchedulerHost)
			if err != nil {
				return err
			}
			authority, err := flags.GetString(flagAuthority)
			if err != nil {
				return err
			}
			showRequest, err := flags.GetBool(flagShowRequest)
			if err != nil {
				return err
			}
			showResponse, err := flags.GetBool(flagShowResponse)
			if err != nil {
				return err
			}

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, schedulerHostIsSet, authority)
			if err != nil {
				return err
			}

			experimentName := args[0]
			err = schedulerClient.StopExperiment(experimentName, showRequest, showResponse)
			return err
		},
	}

	flags := cmd.Flags()
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagAuthority, "", helpAuthority)

	return cmd
}
