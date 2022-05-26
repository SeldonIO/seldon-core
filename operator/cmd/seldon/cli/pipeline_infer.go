package cli

import (
	"fmt"

	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createPipelineInfer() *cobra.Command {
	cmdModelInfer := &cobra.Command{
		Use:   "infer <pipelineName> (data)",
		Short: "run inference on a pipeline",
		Long:  `call a pipeline with a given input and get a prediction`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inferenceHost, err := cmd.Flags().GetString(inferenceHostFlag)
			if err != nil {
				return err
			}
			filename, err := cmd.Flags().GetString(fileFlag)
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
			inferMode, err := cmd.Flags().GetString(inferenceModeFlag)
			if err != nil {
				return err
			}
			inferenceClient := cli.NewInferenceClient(inferenceHost)
			iterations, err := cmd.Flags().GetInt(inferenceIterationsFlag)
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

			err = inferenceClient.Infer(pipelineName, inferMode, data, showRequest, showResponse, iterations, cli.InferPipeline)
			return err
		},
	}
	cmdModelInfer.Flags().StringP(fileFlag, "f", "", "inference payload file")
	cmdModelInfer.Flags().String(inferenceHostFlag, env.GetString(EnvInfer, DefaultInferHost), "seldon inference host")
	cmdModelInfer.Flags().String(inferenceModeFlag, "rest", "inference mode rest or grpc")
	cmdModelInfer.Flags().IntP(inferenceIterationsFlag, "i", 1, "inference iterations")
	return cmdModelInfer
}
