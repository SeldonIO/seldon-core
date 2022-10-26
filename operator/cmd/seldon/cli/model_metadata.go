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
			inferenceHost, err := cmd.Flags().GetString(flagInferenceHost)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(flagAuthority)
			if err != nil {
				return err
			}
			modelName := args[0]

			inferenceClient, err := cli.NewInferenceClient(inferenceHost)
			if err != nil {
				return err
			}

			err = inferenceClient.ModelMetadata(modelName, authority)
			return err
		},
	}

	cmdModelMeta.Flags().String(flagInferenceHost, env.GetString(envInfer, defaultInferHost), helpInferenceHost)
	cmdModelMeta.Flags().String(flagAuthority, "", helpAuthority)

	return cmdModelMeta
}
