package cli

import (
	"fmt"

	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createModelInfer() *cobra.Command {
	cmdModelInfer := &cobra.Command{
		Use:   "infer <modelName> (data)",
		Short: "run inference on a model",
		Long:  `call a model with a given input and get a prediction`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inferenceHost, err := cmd.Flags().GetString(flagInferenceHost)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(flagAuthority)
			if err != nil {
				return err
			}
			filename, err := cmd.Flags().GetString(flagFile)
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
			stickySession, err := cmd.Flags().GetBool(flagStickySession)
			if err != nil {
				return err
			}
			showHeaders, err := cmd.Flags().GetBool(flagShowHeaders)
			if err != nil {
				return err
			}
			inferMode, err := cmd.Flags().GetString(flagInferenceMode)
			if err != nil {
				return err
			}
			iterations, err := cmd.Flags().GetInt(flagInferenceIterations)
			if err != nil {
				return err
			}
			headers, err := cmd.Flags().GetStringArray(flagAddHeader)
			if err != nil {
				return err
			}
			modelName := args[0]

			// Get inference data
			var data []byte
			if len(args) > 1 {
				data = []byte(args[1])
			} else if filename != "" {
				data = loadFile(filename)
			} else {
				return fmt.Errorf("required inline data or from file with -f <file-path>")
			}

			inferenceClient, err := cli.NewInferenceClient(inferenceHost)
			if err != nil {
				return err
			}

			callOpts := &cli.CallOptions{
				InferProtocol: inferMode,
				InferType:     cli.InferModel,
				StickySession: stickySession,
				Iterations:    iterations,
			}
			logOpts := &cli.LogOptions{
				ShowHeaders:  showHeaders,
				ShowRequest:  showRequest,
				ShowResponse: showResponse,
			}

			err = inferenceClient.Infer(
				modelName,
				data,
				headers,
				authority,
				callOpts,
				logOpts,
			)
			return err
		},
	}

	cmdModelInfer.Flags().StringP(flagFile, "f", "", helpFileInference)
	cmdModelInfer.Flags().BoolP(flagStickySession, "s", false, helpStickySession)
	cmdModelInfer.Flags().String(flagInferenceHost, env.GetString(envInfer, defaultInferHost), helpInferenceHost)
	cmdModelInfer.Flags().String(flagInferenceMode, "rest", helpInferenceMode)
	cmdModelInfer.Flags().IntP(flagInferenceIterations, "i", 1, helpInferenceIterations)
	cmdModelInfer.Flags().Bool(flagShowHeaders, false, helpShowHeaders)
	cmdModelInfer.Flags().StringArray(flagAddHeader, []string{}, helpAddHeader)
	cmdModelInfer.Flags().String(flagAuthority, "", helpAuthority)

	return cmdModelInfer
}
