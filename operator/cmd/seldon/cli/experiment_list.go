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
			schedulerHost, err := cmd.Flags().GetString(flagSchedulerHost)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(flagAuthority)
			if err != nil {
				return err
			}

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, authority)
			if err != nil {
				return err
			}

			err = schedulerClient.ListExperiments()
			return err
		},
	}

	cmdExperimentList.Flags().String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	cmdExperimentList.Flags().String(flagAuthority, "", helpAuthority)

	return cmdExperimentList
}
