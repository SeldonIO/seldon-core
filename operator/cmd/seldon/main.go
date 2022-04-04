package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const (
	fileFlag          = "file-path"
	schedulerHostFlag = "scheduler-host"
	schedulerPortFlag = "scheduler-port"
	modelNameFlag     = "model-name"
	experimentFlag    = "experiment-name"
	showResponseFlag  = "show-response"
	showRequestFlag   = "show-request"
	waitConditionFlag = "wait"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
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

	var rootCmd = &cobra.Command{Use: "app"}

	rootCmd.PersistentFlags().BoolP(showRequestFlag, "r", false, "show request")
	rootCmd.PersistentFlags().BoolP(showResponseFlag, "s", true, "show response")

	rootCmd.AddCommand(cmdModel, cmdServer, cmdExperiment, cmdPipeline)
	cmdModel.AddCommand(cmdModelLoad, cmdModelUnload, cmdModelStatus, cmdModelInfer)
	cmdServer.AddCommand(cmdServerStatus)
	cmdExperiment.AddCommand(cmdExperimentStart, cmdExperimentStop, cmdExperimentStatus)
	cmdPipeline.AddCommand(cmdPipelineLoad, cmdPipelineUnload, cmdPipelineStatus, cmdPipelineInfer)
	return rootCmd.Execute()
}

func loadFile(filename string) []byte {
	dat, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	return dat
}

func main() {
	if err := Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}
