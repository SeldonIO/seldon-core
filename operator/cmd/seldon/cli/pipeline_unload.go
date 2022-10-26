package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createPipelineUnload() *cobra.Command {
	cmdPipelineUnload := &cobra.Command{
		Use:   "unload <pipelineName>",
		Short: "unload a pipeline",
		Long:  `unload a pipeline`,
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
			pipelineName := args[0]

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, authority)
			if err != nil {
				return err
			}

			err = schedulerClient.UnloadPipeline(pipelineName, showRequest, showResponse)
			return err
		},
	}

	cmdPipelineUnload.Flags().String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	cmdPipelineUnload.Flags().String(flagAuthority, "", helpAuthority)

	return cmdPipelineUnload
}
