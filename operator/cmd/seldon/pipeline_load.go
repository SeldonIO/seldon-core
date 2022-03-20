package main

import (
	"os"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createPipelineLoad() *cobra.Command {
	cmdModelLoad := &cobra.Command{
		Use:   "load",
		Short: "load a pipeline",
		Long:  `load a pipeline`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(schedulerHostFlag)
			if err != nil {
				return err
			}
			schedulerPort, err := cmd.Flags().GetInt(schedulerPortFlag)
			if err != nil {
				return err
			}
			filename, err := cmd.Flags().GetString(fileFlag)
			if err != nil {
				return err
			}
			verbose, err := cmd.Flags().GetBool(verboseFlag)
			if err != nil {
				return err
			}
			schedulerClient := cli.NewSchedulerClient(schedulerHost, schedulerPort)
			err = schedulerClient.LoadPipeline(loadFile(filename), verbose)
			return err
		},
	}
	cmdModelLoad.Flags().StringP(fileFlag, "f", "", "pipeline file to load")
	if err := cmdModelLoad.MarkFlagRequired(fileFlag); err != nil {
		os.Exit(-1)
	}
	cmdModelLoad.Flags().String(schedulerHostFlag, "0.0.0.0", "seldon scheduler host")
	cmdModelLoad.Flags().Int(schedulerPortFlag, 9004, "seldon scheduler port")
	return cmdModelLoad
}
