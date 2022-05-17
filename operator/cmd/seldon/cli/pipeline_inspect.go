package cli

import (
	"fmt"

	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

const (
	OffsetFlag = "offset"
)

func createPipelineInspect() *cobra.Command {
	cmdPipelineInspect := &cobra.Command{
		Use:   "inspect",
		Short: "inspect data in a pipeline",
		Long:  `inspect data in a pipeline. Specify as <pipelineName> or <pipelinename>.(inputs|outputs) or  <pipeineName>.<stepName> or <pipelineName>.<stepName>.(inputs|outputs) or <pipelineName>.<stepName>.(inputs|outputs).tensorName`,
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(schedulerHostFlag)
			if err != nil {
				return err
			}
			kafkaBroker, err := cmd.Flags().GetString(kafkaBrokerFlag)
			if err != nil {
				return err
			}
			offset, err := cmd.Flags().GetInt64(OffsetFlag)
			if err != nil {
				return err
			}
			var data []byte
			if len(args) > 0 {
				data = []byte(args[0])
			} else {
				return fmt.Errorf("pipeline step name required")
			}

			kc, err := cli.NewKafkaClient(kafkaBroker, schedulerHost)
			if err != nil {
				return err
			}
			err = kc.InspectStep(string(data), offset)
			return err
		},
	}
	cmdPipelineInspect.Flags().String(kafkaBrokerFlag, env.GetString(EnvKafka, DefaultKafkaHost), "kafka broker")
	cmdPipelineInspect.Flags().Int64(OffsetFlag, 1, "Message offset to start reading from, i.e. default 1 is the last message only")
	cmdPipelineInspect.Flags().String(schedulerHostFlag, env.GetString(EnvScheduler, DefaultScheduleHost), "seldon scheduler host")
	return cmdPipelineInspect
}
