package cli

import (
	"os"

	"k8s.io/utils/env"

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
			schedulerHost, err := cmd.Flags().GetString(flagSchedulerHost)
			if err != nil {
				return err
			}
			authority, err := cmd.Flags().GetString(flagAuthority)
			if err != nil {
				return err
			}
			filename, err := cmd.Flags().GetString(flagFile)
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

			schedulerClient, err := cli.NewSchedulerClient(schedulerHost, authority)
			if err != nil {
				return err
			}

			err = schedulerClient.LoadPipeline(loadFile(filename), showRequest, showResponse)
			return err
		},
	}

	cmdModelLoad.Flags().String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	cmdModelLoad.Flags().String(flagAuthority, "", helpAuthority)
	cmdModelLoad.Flags().StringP(flagFile, "f", "", "pipeline manifest file (YAML)")
	if err := cmdModelLoad.MarkFlagRequired(flagFile); err != nil {
		os.Exit(-1)
	}

	return cmdModelLoad
}
