package cli

import (
	"os"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createModelUnload() *cobra.Command {
	cmdModelUnload := &cobra.Command{
		Use:   "unload",
		Short: "unload a model",
		Long:  `unload a model`,
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
			schedulerClient := cli.NewSchedulerClient(schedulerHost, schedulerPort)
			err = schedulerClient.UnloadModel(modelName, showRequest, showResponse)
			return err
		},
	}
	cmdModelUnload.Flags().StringP(modelNameFlag, "m", "", "model name to unload")
	if err := cmdModelUnload.MarkFlagRequired(modelNameFlag); err != nil {
		os.Exit(-1)
	}
	cmdModelUnload.Flags().String(schedulerHostFlag, "0.0.0.0", "seldon scheduler host")
	cmdModelUnload.Flags().Int(schedulerPortFlag, 9004, "seldon scheduler port")
	return cmdModelUnload
}
