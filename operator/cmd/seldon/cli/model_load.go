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
	"os"

	"github.com/spf13/cobra"
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

func createModelLoad() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "load a model",
		Long:  `load a model`,
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
			filename, err := flags.GetString(flagFile)
			if err != nil {
				return err
			}
			verbose, err := flags.GetBool(flagVerbose)
			if err != nil {
				return err
			}
			force, err := flags.GetBool(flagForceControlPlane)
			if err != nil {
				return err
			}
			fmt.Println(helpForceControlPlaneWarning)
			if !force {
				return nil
			}

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, schedulerHostIsSet, authority, verbose)
			if err != nil {
				return err
			}

			res, err := schedulerClient.LoadModel(loadFile(filename))
			if err == nil {
				cli.PrintProto(res)
			}
			return err
		},
	}

	flags := cmd.Flags()
	flags.BoolP(flagVerbose, "v", false, "verbose output")
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagAuthority, "", helpAuthority)
	flags.StringP(flagFile, "f", "", "model manifest file (YAML)")
	if err := cmd.MarkFlagRequired(flagFile); err != nil {
		os.Exit(-1)
	}
	forceFlag, err := env.GetBool(envForceControlPlane, defaultForceControlPlane)
	if err != nil {
		os.Exit(-1)
	}
	flags.Bool(flagForceControlPlane, forceFlag, helpForceControlPlane)

	return cmd
}
