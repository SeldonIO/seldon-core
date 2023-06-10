/*
Copyright 2023 Seldon Technologies Ltd.

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
	"time"

	"github.com/spf13/cobra"
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

func createStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <pipelineName>",
		Short: "status of a pipeline",
		Long:  `status of a pipeline`,
		Args:  cobra.ExactArgs(0),
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
			waitCondition, err := flags.GetBool(flagWaitCondition)
			if err != nil {
				return err
			}
			filename, err := flags.GetString(flagFile)
			if err != nil {
				return err
			}
			timeout, err := flags.GetInt64(flagTimeout)
			if err != nil {
				return err
			}
			var dataFile []byte
			if filename != "" {
				dataFile = loadFile(filename)
			}
			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, schedulerHostIsSet, authority, verbose)
			if err != nil {
				return err
			}

			responses, err := schedulerClient.Status(dataFile, waitCondition, time.Duration(timeout*int64(time.Second)))
			if err == nil {
				for _, res := range responses {
					cli.PrintProto(res)
				}
			}
			return err
		},
	}

	flags := cmd.Flags()
	flags.Int64P(flagTimeout, "t", flagTimeoutDefault, "timeout seconds")
	flags.BoolP(flagVerbose, "v", false, "verbose output")
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagAuthority, "", helpAuthority)
	flags.BoolP(flagWaitCondition, "w", false, "wait for resources to be ready")
	flags.StringP(flagFile, "f", "", "model manifest file (YAML)")

	return cmd
}
