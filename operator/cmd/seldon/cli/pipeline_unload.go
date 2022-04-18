package cli

import (
	"os"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

const (
	pipelineNameFlag = "pipeline-name"
)

func createPipelineUnload() *cobra.Command {
	cmdPipelineUnload := &cobra.Command{
		Use:   "unload",
		Short: "unload a pipeline",
		Long:  `unload a pipeline`,
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
			pipelineName, err := cmd.Flags().GetString(pipelineNameFlag)
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
			schedulerClient := cli.NewSchedulerClient(schedulerHost, schedulerPort)
			err = schedulerClient.UnloadPipeline(pipelineName, showRequest, showResponse)
			return err
		},
	}
	cmdPipelineUnload.Flags().StringP(pipelineNameFlag, "p", "", "pipeline name to unload")
	if err := cmdPipelineUnload.MarkFlagRequired(pipelineNameFlag); err != nil {
		os.Exit(-1)
	}
	cmdPipelineUnload.Flags().String(schedulerHostFlag, "0.0.0.0", "seldon scheduler host")
	cmdPipelineUnload.Flags().Int(schedulerPortFlag, 9004, "seldon scheduler port")
	return cmdPipelineUnload
}
