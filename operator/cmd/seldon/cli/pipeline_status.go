package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createPipelineStatus() *cobra.Command {
	cmdPipelineStatus := &cobra.Command{
		Use:   "status <pipelineName>",
		Short: "status of a pipeline",
		Long:  `status of a pipeline`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(schedulerHostFlag)
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
			pipelineName := args[0]
			err = schedulerClient.PipelineStatus(pipelineName, showRequest, showResponse, waitCondition)
			return err
		},
	}
	cmdPipelineStatus.Flags().String(schedulerHostFlag, env.GetString(EnvScheduler, DefaultScheduleHost), "seldon scheduler host")
	cmdPipelineStatus.Flags().StringP(waitConditionFlag, "w", "", "pipeline wait condition")
	return cmdPipelineStatus
}
