/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
	"github.com/spf13/cobra"
)

func createPipelineUnload() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unload <pipelineName> or -f <filename>",
		Short: "unload a pipeline",
		Long:  `unload a pipeline`,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			schedulerHostIsSet := flags.Changed(flagSchedulerHost)
			schedulerHost, err := flags.GetString(flagSchedulerHost)
			if err != nil {
				return err
			}
			authority, err := flags.GetString(flagAuthority)
			if err != nil {
				return err
			}
			verbose, err := flags.GetBool(flagVerbose)
			if err != nil {
				return err
			}
			fileBytes, pipelineName, err := extractFileOrName(flags, args)
			if err != nil {
				return err
			}

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, schedulerHostIsSet, authority)
			if err != nil {
				return err
			}

			res, err := schedulerClient.UnloadPipeline(pipelineName, fileBytes)
			if err == nil && verbose {
				cli.PrintProto(res)
			}
			return err
		},
	}

	flags := cmd.Flags()
	flags.BoolP(flagVerbose, "v", false, "verbose output")
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagAuthority, "", helpAuthority)
	flags.StringP(flagFile, "f", "", "pipeline manifest file (YAML)")

	return cmd
}
