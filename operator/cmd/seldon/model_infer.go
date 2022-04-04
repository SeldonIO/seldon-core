package main

import (
	"fmt"
	"os"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

const (
	inferenceHostFlag       = "inference-host"
	inferencePortFlag       = "inference-port"
	inferenceModeFlag       = "inference-mode"
	inferenceIterationsFlag = "iterations"
)

func createModelInfer() *cobra.Command {
	cmdModelInfer := &cobra.Command{
		Use:   "infer",
		Short: "run inference on a model",
		Long:  `call a model with a given input and get a prediction`,
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

			err = inferenceClient.Infer(modelName, inferMode, data, showRequest, showResponse, iterations, cli.InferModel)
			return err
		},
	}
	cmdModelInfer.Flags().StringP(fileFlag, "f", "", "inference payload file")
	cmdModelInfer.Flags().StringP(modelNameFlag, "m", "", "model name for inference")
	if err := cmdModelInfer.MarkFlagRequired(modelNameFlag); err != nil {
		os.Exit(-1)
	}
	cmdModelInfer.Flags().String(inferenceHostFlag, "0.0.0.0", "seldon inference host")
	cmdModelInfer.Flags().Int(inferencePortFlag, 9000, "seldon scheduler port")
	cmdModelInfer.Flags().String(inferenceModeFlag, "rest", "inference mode rest or grpc")
	cmdModelInfer.Flags().IntP(inferenceIterationsFlag, "i", 1, "inference iterations")
	return cmdModelInfer
}
