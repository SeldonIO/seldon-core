package cli

import (
	"os"

	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createPipelineStatus() *cobra.Command {
	cmdPipelineStatus := &cobra.Command{
		Use:   "status",
		Short: "status of a pipeline",
		Long:  `status of a pipeline`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(schedulerHostFlag)
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
			waitCondition, err := cmd.Flags().GetString(waitConditionFlag)
			if err != nil {
				return err
			}
			schedulerClient := cli.NewSchedulerClient(schedulerHost)
			err = schedulerClient.PipelineStatus(pipelineName, showRequest, showResponse, waitCondition)
			return err
		},
	}
	cmdPipelineStatus.Flags().StringP(pipelineNameFlag, "p", "", "pipeline name for status")
	if err := cmdPipelineStatus.MarkFlagRequired(pipelineNameFlag); err != nil {
		os.Exit(-1)
	}
	cmdPipelineStatus.Flags().String(schedulerHostFlag, env.GetString(EnvScheduler, DefaultScheduleHost), "seldon scheduler host")
	cmdPipelineStatus.Flags().StringP(waitConditionFlag, "w", "", "pipeline wait condition")
	return cmdPipelineStatus
}
