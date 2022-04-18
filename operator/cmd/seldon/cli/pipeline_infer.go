package cli

import (
	"fmt"
	"os"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createPipelineInfer() *cobra.Command {
	cmdModelInfer := &cobra.Command{
		Use:   "infer",
		Short: "run inference on a pipeline",
		Long:  `call a pipeline with a given input and get a prediction`,
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
			filename, err := cmd.Flags().GetString(fileFlag)
			if err != nil {
				return err
			}
			pipelineName, err := cmd.Flags().GetString(pipelineNameFlag)
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
			inferenceClient := cli.NewInferenceClient(inferenceHost, inferencePort)
			iterations, err := cmd.Flags().GetInt(inferenceIterationsFlag)
			if err != nil {
				return err
			}
			// Get inference data
			var data []byte
			if len(args) > 0 {
				data = []byte(args[0])
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
	cmdModelInfer.Flags().StringP(pipelineNameFlag, "p", "", "pipeline name for inference")
	if err := cmdModelInfer.MarkFlagRequired(pipelineNameFlag); err != nil {
		os.Exit(-1)
	}
	cmdModelInfer.Flags().String(inferenceHostFlag, "0.0.0.0", "seldon inference host")
	cmdModelInfer.Flags().Int(inferencePortFlag, 9000, "seldon scheduler port")
	cmdModelInfer.Flags().String(inferenceModeFlag, "rest", "inference mode rest or grpc")
	cmdModelInfer.Flags().IntP(inferenceIterationsFlag, "i", 1, "inference iterations")
	return cmdModelInfer
}
