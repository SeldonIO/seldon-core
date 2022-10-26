package cli

import (
	"fmt"

	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createPipelineInfer() *cobra.Command {
	cmdPipelineInfer := &cobra.Command{
		Use:   "infer <pipelineName> (data)",
		Short: "run inference on a pipeline",
		Long:  `call a pipeline with a given input and get a prediction`,
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
			stickySession, err := cmd.Flags().GetBool(flagStickySession)
			if err != nil {
				return err
			}
			showResponse, err := cmd.Flags().GetBool(flagShowResponse)
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
			headers, err := cmd.Flags().GetStringArray(flagAddHeader)
			if err != nil {
				return err
			}
			iterations, err := cmd.Flags().GetInt(flagInferenceIterations)
			if err != nil {
				return err
			}
			pipelineName := args[0]

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
				InferType:     cli.InferPipeline,
				StickySession: stickySession,
				Iterations:    iterations,
			}
			logOpts := &cli.LogOptions{
				ShowHeaders:  showHeaders,
				ShowRequest:  showRequest,
				ShowResponse: showResponse,
			}

			err = inferenceClient.Infer(pipelineName, data, headers, authority, callOpts, logOpts)
			return err
		},
	}

	cmdPipelineInfer.Flags().StringP(flagFile, "f", "", helpFileInference)
	cmdPipelineInfer.Flags().BoolP(flagStickySession, "s", false, helpStickySession)
	cmdPipelineInfer.Flags().String(flagInferenceHost, env.GetString(envInfer, defaultInferHost), helpInferenceHost)
	cmdPipelineInfer.Flags().String(flagInferenceMode, "rest", helpInferenceMode)
	cmdPipelineInfer.Flags().IntP(flagInferenceIterations, "i", 1, helpInferenceIterations)
	cmdPipelineInfer.Flags().Bool(flagShowHeaders, false, helpShowHeaders)
	cmdPipelineInfer.Flags().StringArray(flagAddHeader, []string{}, helpAddHeader)
	cmdPipelineInfer.Flags().String(flagAuthority, "", helpAuthority)

	return cmdPipelineInfer
}
