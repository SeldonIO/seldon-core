package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createModelMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metadata <modelName>",
		Short: "get model metadata",
		Long:  `get model metadata`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			inferenceHostIsSet := flags.Changed(flagInferenceHost)
			inferenceHost, err := flags.GetString(flagInferenceHost)
			if err != nil {
				return err
			}
			authority, err := flags.GetString(flagAuthority)
			if err != nil {
				return err
			}
			modelName := args[0]

			inferenceClient, err := cli.NewInferenceClient(inferenceHost, inferenceHostIsSet)
			if err != nil {
				return err
			}

			err = inferenceClient.ModelMetadata(modelName, authority)
			return err
		},
	}

	flags := cmd.Flags()
	flags.String(flagInferenceHost, env.GetString(envInfer, defaultInferHost), helpInferenceHost)
	flags.String(flagAuthority, "", helpAuthority)

	return cmd
}
