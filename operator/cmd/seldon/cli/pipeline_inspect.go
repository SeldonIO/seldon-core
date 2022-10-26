package cli

import (
	"fmt"

	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operatorv2/pkg/cli"
	"github.com/spf13/cobra"
)

const (
	flagOffset       = "offset"
	flagRequestId    = "request-id"
	flagOutputFormat = "format"
	flagVerbose      = "verbose"
	flagNamespace    = "namespace"
	defaultNamespace = "default"
)

func createPipelineInspect() *cobra.Command {
	cmdPipelineInspect := &cobra.Command{
		Use:   "inspect <expression>",
		Short: "inspect data in a pipeline",
		Long:  `inspect data in a pipeline. Specify as pipelineName or pipelineName.(inputs|outputs) or pipeineName.stepName or pipelineName.stepName.(inputs|outputs) or pipelineName.stepName.(inputs|outputs).tensorName`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			schedulerHost, err := cmd.Flags().GetString(flagSchedulerHost)
			if err != nil {
				return err
			}
			kafkaBroker, err := cmd.Flags().GetString(flagKafkaBroker)
			if err != nil {
				return err
			}
			offset, err := cmd.Flags().GetInt64(flagOffset)
			if err != nil {
				return err
			}
			requestId, err := cmd.Flags().GetString(flagRequestId)
			if err != nil {
				return err
			}
			format, err := cmd.Flags().GetString(flagOutputFormat)
			if err != nil {
				return err
			}
			verbose, err := cmd.Flags().GetBool(flagVerbose)
			if err != nil {
				return err
			}
			namespace, err := cmd.Flags().GetString(flagNamespace)
			if err != nil {
				return err
			}
			data := []byte(args[0])
			kc, err := cli.NewKafkaClient(kafkaBroker, schedulerHost)
			if err != nil {
				return err
			}
			err = kc.InspectStep(string(data), offset, requestId, format, verbose, namespace)
			return err
		},
	}
	cmdPipelineInspect.Flags().String(flagKafkaBroker, env.GetString(envKafka, defaultKafkaHost), "kafka broker")
	cmdPipelineInspect.Flags().Int64(flagOffset, 1, "message offset to start reading from, i.e. default 1 is the last message only")
	cmdPipelineInspect.Flags().String(flagRequestId, "", "request id to show, if not specified will be all messages in offset range")
	cmdPipelineInspect.Flags().String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	cmdPipelineInspect.Flags().String(flagOutputFormat, cli.InspectFormatRaw, fmt.Sprintf("inspect output format: raw or json. Default %s", cli.InspectFormatRaw))
	cmdPipelineInspect.Flags().String(flagNamespace, defaultNamespace, fmt.Sprintf("namespace. Default %s", defaultNamespace))
	cmdPipelineInspect.Flags().Bool(flagVerbose, false, "display more details, such as headers")
	return cmdPipelineInspect
}
