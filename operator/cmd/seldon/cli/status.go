package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
	"github.com/spf13/cobra"
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
			showRequest, err := flags.GetBool(flagShowRequest)
			if err != nil {
				return err
			}
			showResponse, err := flags.GetBool(flagShowResponse)
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
			var dataFile []byte
			if filename != "" {
				dataFile = loadFile(filename)
			}
			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, schedulerHostIsSet, authority)
			if err != nil {
				return err
			}

			err = schedulerClient.Status(dataFile, showRequest, showResponse, waitCondition)
			return err
		},
	}

	flags := cmd.Flags()
	flags.BoolP(flagShowRequest, "r", false, "show request")
	flags.BoolP(flagShowResponse, "o", false, "show response")
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagAuthority, "", helpAuthority)
	flags.BoolP(flagWaitCondition, "w", false, "wait for resources to be ready")
	flags.StringP(flagFile, "f", "", "model manifest file (YAML)")

	return cmd
}
