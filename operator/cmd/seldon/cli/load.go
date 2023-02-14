package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
	"github.com/spf13/cobra"
)

func createLoad() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "load resources",
		Long:  `load resources`,
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
			showRequest, err := flags.GetBool(flagShowRequest)
			if err != nil {
				return err
			}
			showResponse, err := flags.GetBool(flagShowResponse)
			if err != nil {
				return err
			}

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, schedulerHostIsSet, authority)
			if err != nil {
				return err
			}

			var dataFile []byte
			if filename != "" {
				dataFile = loadFile(filename)
			}
			err = schedulerClient.Load(dataFile, showRequest, showResponse)
			return err
		},
	}

	flags := cmd.Flags()
	flags.BoolP(flagShowRequest, "r", false, "show request")
	flags.BoolP(flagShowResponse, "o", false, "show response")
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagAuthority, "", helpAuthority)
	flags.StringP(flagFile, "f", "", "model manifest file (YAML)")

	return cmd
}
