/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func GetCmd() *cobra.Command {
	cmdModel := &cobra.Command{
		Use:   "model <subcomand>",
		Short: "manage models",
		Long:  `load and unload and get status for models`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("model subcommand required")
		},
	}

	cmdServer := &cobra.Command{
		Use:   "server <subcomand>",
		Short: "manage servers",
		Long:  `get status for servers`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("server subcommand required")
		},
	}

	cmdExperiment := &cobra.Command{
		Use:   "experiment <subcomand>",
		Short: "manage experiments",
		Long:  `experiments allow you to test models by splitting traffic between candidates.`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("experiment subcommand required")
		},
	}

	cmdPipeline := &cobra.Command{
		Use:   "pipeline <subcomand>",
		Short: "manage pipelines",
		Long:  `pipelines allow you to join models together into inference graphs.`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("pipeline subcommand required")
		},
	}

	cmdConfig := &cobra.Command{
		Use:   "config <subcomand>",
		Short: "manage configs",
		Long:  `Manage and activate configuration files for the CLI`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("config subcommand required")
		},
	}

	// Model commands
	cmdModelLoad := createModelLoad()
	cmdModelUnload := createModelUnload()
	cmdModelInfer := createModelInfer()
	cmdModelStatus := createModelStatus()
	cmdModelMeta := createModelMetadata()
	cmdModelList := createModelList()

	// Server commands
	cmdServerStatus := createServerStatus()
	cmdServerList := createServerList()

	// experiment commands
	cmdExperimentStart := createExperimentStart()
	cmdExperimentStop := createExperimentStop()
	cmdExperimentStatus := createExperimentStatus()
	cmdExperimentList := createExperimentList()

	// pipeline commands
	cmdPipelineLoad := createPipelineLoad()
	cmdPipelineUnload := createPipelineUnload()
	cmdPipelineStatus := createPipelineStatus()
	cmdPipelineInfer := createPipelineInfer()
	cmdPipelineList := createPipelineList()
	cmdPipelineInspect := createPipelineInspect()

	// config commands
	cmdConfigActivate := createConfigActivate()
	cmdConfigDeactivate := createConfigDeactivate()
	cmdConfigAdd := createConfigAdd()
	cmdConfigRemove := createConfigRemove()
	cmdConfigList := createConfigList()

	// Generic commands
	cmdLoad := createLoad()
	cmdUnload := createUnload()
	cmdStatus := createStatus()

	var rootCmd = &cobra.Command{Use: "seldon", SilenceErrors: false, SilenceUsage: true}

	rootCmd.DisableAutoGenTag = true

	rootCmd.AddCommand(cmdModel, cmdServer, cmdExperiment, cmdPipeline, cmdConfig, cmdLoad, cmdUnload, cmdStatus)
	cmdModel.AddCommand(cmdModelLoad, cmdModelUnload, cmdModelStatus, cmdModelInfer, cmdModelMeta, cmdModelList)
	cmdServer.AddCommand(cmdServerStatus, cmdServerList)
	cmdExperiment.AddCommand(cmdExperimentStart, cmdExperimentStop, cmdExperimentStatus, cmdExperimentList)
	cmdPipeline.AddCommand(cmdPipelineLoad, cmdPipelineUnload, cmdPipelineStatus, cmdPipelineInfer, cmdPipelineList, cmdPipelineInspect)
	cmdConfig.AddCommand(cmdConfigActivate, cmdConfigAdd, cmdConfigDeactivate, cmdConfigList, cmdConfigRemove)

	return rootCmd
}
