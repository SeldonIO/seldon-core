package cli

import (
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

const (
	flagExperimentWait = "wait"
)

func createExperimentStatus() *cobra.Command {
	cmdExperimentStatus := &cobra.Command{
		Use:   "status <experimentName>",
		Short: "get status for experiment",
		Long:  `get status for experiment`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(flagSchedulerHost)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(flagAuthority)
			if err != nil {
				return err
			}
			showRequest, err := cmd.Flags().GetBool(flagShowRequest)
			if err != nil {
				return err
			}
			showResponse, err := cmd.Flags().GetBool(flagShowResponse)
			if err != nil {
				return err
			}
			wait, err := cmd.Flags().GetBool(flagExperimentWait)
			if err != nil {
				return err
			}
			experimentName := args[0]

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, authority)
			if err != nil {
				return err
			}

			err = schedulerClient.ExperimentStatus(experimentName, showRequest, showResponse, wait)
			return err
		},
	}

	cmdExperimentStatus.Flags().String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	cmdExperimentStatus.Flags().String(flagAuthority, "", helpAuthority)
	cmdExperimentStatus.Flags().BoolP(flagExperimentWait, "w", false, "wait for experiment to be active")

	return cmdExperimentStatus
}
