package cli

import (
	"os"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createModelMetadata() *cobra.Command {
	cmdModelMeta := &cobra.Command{
		Use:   "metadata",
		Short: "get model metadata",
		Long:  `get model metadata`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			inferenceHost, err := cmd.Flags().GetString(inferenceHostFlag)
			if err != nil {
				return err
			}
			inferencePort, err := cmd.Flags().GetInt(inferencePortFlag)
			if err != nil {
				return err
			}
			modelName, err := cmd.Flags().GetString(modelNameFlag)
			if err != nil {
				return err
			}
			inferenceClient := cli.NewInferenceClient(inferenceHost, inferencePort)

			err = inferenceClient.ModelMetadata(modelName)
			return err
		},
	}
	cmdModelMeta.Flags().StringP(modelNameFlag, "m", "", "model name for inference")
	if err := cmdModelMeta.MarkFlagRequired(modelNameFlag); err != nil {
		os.Exit(-1)
	}
	cmdModelMeta.Flags().String(inferenceHostFlag, "0.0.0.0", "seldon inference host")
	cmdModelMeta.Flags().Int(inferencePortFlag, 9000, "seldon scheduler port")
	return cmdModelMeta
}
