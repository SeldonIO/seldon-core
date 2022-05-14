package cli

import (
	"os"

	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createModelStatus() *cobra.Command {
	cmdModelStatus := &cobra.Command{
		Use:   "status",
		Short: "get status for model",
		Long:  `get the status for a model`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(schedulerHostFlag)
			if err != nil {
				return err
			}
			modelName, err := cmd.Flags().GetString(modelNameFlag)
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
			modelWaitCondition, err := cmd.Flags().GetString(waitConditionFlag)
			if err != nil {
				return err
			}
			schedulerClient := cli.NewSchedulerClient(schedulerHost)
			err = schedulerClient.ModelStatus(modelName, showRequest, showResponse, modelWaitCondition)
			return err
		},
	}
	cmdModelStatus.Flags().StringP(modelNameFlag, "m", "", "model name to unload")
	if err := cmdModelStatus.MarkFlagRequired(modelNameFlag); err != nil {
		os.Exit(-1)
	}
	cmdModelStatus.Flags().String(schedulerHostFlag, env.GetString(EnvScheduler, DefaultScheduleHost), "seldon scheduler host")
	cmdModelStatus.Flags().StringP(waitConditionFlag, "w", "", "model wait condition")
	return cmdModelStatus
}
