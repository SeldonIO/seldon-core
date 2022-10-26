package cli

import (
	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
	"k8s.io/utils/env"
)

func createModelList() *cobra.Command {
	cmdModelList := &cobra.Command{
		Use:   "list",
		Short: "get list of models",
		Long:  `get the list of all models with their status`,
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

			err = schedulerClient.ListModels()
			return err
		},
	}

	cmdModelList.Flags().String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	cmdModelList.Flags().String(flagAuthority, "", helpAuthority)

	return cmdModelList
}
