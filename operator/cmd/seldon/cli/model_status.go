package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

func createModelStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <modelName>",
		Short: "get status for model",
		Long:  `get the status for a model`,
		Args:  cobra.ExactArgs(1),
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
			modelWaitCondition, err := flags.GetString(flagWaitCondition)
			if err != nil {
				return err
			}
			modelName := args[0]

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, schedulerHostIsSet, authority)
			if err != nil {
				return err
			}

			err = schedulerClient.ModelStatus(modelName, showRequest, showResponse, modelWaitCondition)
			return err
		},
	}

	flags := cmd.Flags()
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagAuthority, "", helpAuthority)
	flags.StringP(flagWaitCondition, "w", "", "model wait condition")

	return cmd
}
