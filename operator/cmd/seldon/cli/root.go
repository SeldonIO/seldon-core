package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func GetCmd() *cobra.Command {
	var cmdModel = &cobra.Command{
		Use:   "model <subcomand>",
		Short: "manage models",
		Long:  `load and unload and get status for models`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("model subcommand required")
		},
	}

	var cmdServer = &cobra.Command{
		Use:   "server <subcomand>",
		Short: "manage servers",
		Long:  `get status for servers`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("server subcommand required")
		},
	}

	var cmdExperiment = &cobra.Command{
		Use:   "experiment <subcomand>",
		Short: "manage experiments",
		Long:  `experiments allow you to test models by splitting traffic between candidates.`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("experiment subcommand required")
		},
	}

	var cmdPipeline = &cobra.Command{
		Use:   "pipeline <subcomand>",
		Short: "manage pipelines",
		Long:  `pipelines allow you to join modles together into inference graphs.`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("pipeline subcommand required")
		},
	}

	// Model commands
	cmdModelLoad := createModelLoad()
	cmdModelUnload := createModelUnload()
	cmdModelInfer := createModelInfer()
	cmdModelStatus := createModelStatus()
	cmdModelMeta := createModelMetadata()

	// Server commands
	cmdServerStatus := createServerStatus()

	// experiment commands
	cmdExperimentStart := createExperimentStart()
	cmdExperimentStop := createExperimentStop()
	cmdExperimentStatus := createExperimentStatus()

	// pipeline commands
	cmdPipelineLoad := createPipelineLoad()
	cmdPipelineUnload := createPipelineUnload()
	cmdPipelineStatus := createPipelineStatus()
	cmdPipelineInfer := createPipelineInfer()

	var rootCmd = &cobra.Command{Use: "seldon"}

	rootCmd.PersistentFlags().BoolP(showRequestFlag, "r", false, "show request")
	rootCmd.PersistentFlags().BoolP(showResponseFlag, "o", true, "show response")

	rootCmd.AddCommand(cmdModel, cmdServer, cmdExperiment, cmdPipeline)
	cmdModel.AddCommand(cmdModelLoad, cmdModelUnload, cmdModelStatus, cmdModelInfer, cmdModelMeta)
	cmdServer.AddCommand(cmdServerStatus)
	cmdExperiment.AddCommand(cmdExperimentStart, cmdExperimentStop, cmdExperimentStatus)
	cmdPipeline.AddCommand(cmdPipelineLoad, cmdPipelineUnload, cmdPipelineStatus, cmdPipelineInfer)

	return rootCmd
}
