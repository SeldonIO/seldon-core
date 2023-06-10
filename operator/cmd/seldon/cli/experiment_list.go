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
	"github.com/spf13/cobra"
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

func createExperimentList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "get list of experiments",
		Long:  `get list of experiments and whether they are active`,
		Args:  cobra.MinimumNArgs(0),
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

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, schedulerHostIsSet, authority, verbose)
			if err != nil {
				return err
			}

			err = schedulerClient.ListExperiments()
			return err
		},
	}

	flags := cmd.Flags()
	flags.BoolP(flagVerbose, "v", false, "verbose output")
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagAuthority, "", helpAuthority)

	return cmd
}
