package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createModelMetadata() *cobra.Command {
	cmdModelMeta := &cobra.Command{
		Use:   "metadata <modelName>",
		Short: "get model metadata",
		Long:  `get model metadata`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inferenceHost, err := cmd.Flags().GetString(inferenceHostFlag)
			if err != nil {
				return err
			}
			modelName := args[0]
			inferenceClient := cli.NewInferenceClient(inferenceHost)

			err = inferenceClient.ModelMetadata(modelName)
			return err
		},
	}
	cmdModelMeta.Flags().String(inferenceHostFlag, env.GetString(EnvInfer, DefaultInferHost), "seldon inference host")
	return cmdModelMeta
}
