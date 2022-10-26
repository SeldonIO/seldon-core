package cli

import (
	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
	"k8s.io/utils/env"
)

func createPipelineList() *cobra.Command {
	cmdPipelineList := &cobra.Command{
		Use:   "list",
		Short: "list pipelines",
		Long:  `list pipelines`,
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

			err = schedulerClient.ListPipelines()
			return err
		},
	}

	cmdPipelineList.Flags().String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	cmdPipelineList.Flags().String(flagAuthority, "", helpAuthority)

	return cmdPipelineList
}
