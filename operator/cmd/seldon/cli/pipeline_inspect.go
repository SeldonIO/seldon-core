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
	"time"

	"github.com/spf13/cobra"
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

const (
	flagOffset         = "offset"
	flagRequestId      = "request-id"
	flagOutputFormat   = "format"
	flagTruncate       = "truncate"
	flagNamespace      = "namespace"
	flagTimeoutDefault = int64(5)
)

func createPipelineInspect() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect <expression>",
		Short: "inspect data in a pipeline",
		Long:  `inspect data in a pipeline. Specify as pipelineName or pipelineName.(inputs|outputs) or pipeineName.stepName or pipelineName.stepName.(inputs|outputs) or pipelineName.stepName.(inputs|outputs).tensorName`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			schedulerHostIsSet := flags.Changed(flagSchedulerHost)
			schedulerHost, err := flags.GetString(flagSchedulerHost)
			if err != nil {
				return err
			}
			kafkaBrokerIsSet := flags.Changed(flagKafkaBroker)
			kafkaBroker, err := flags.GetString(flagKafkaBroker)
			if err != nil {
				return err
			}
			offset, err := flags.GetInt64(flagOffset)
			if err != nil {
				return err
			}
			requestId, err := flags.GetString(flagRequestId)
			if err != nil {
				return err
			}
			format, err := flags.GetString(flagOutputFormat)
			if err != nil {
				return err
			}
			verbose, err := flags.GetBool(flagVerbose)
			if err != nil {
				return err
			}
			truncateData, err := flags.GetBool(flagTruncate)
			if err != nil {
				return err
			}
			namespace, err := flags.GetString(flagNamespace)
			if err != nil {
				return err
			}
			kafkaConfigPath, err := flags.GetString(flagKafkaConfigPath)
			if err != nil {
				return err
			}
			timeoutSecs, err := flags.GetInt64(flagTimeout)
			if err != nil {
				return err
			}
			kc, err := cli.NewKafkaClient(kafkaBroker, kafkaBrokerIsSet, schedulerHost, schedulerHostIsSet, kafkaConfigPath)
			if err != nil {
				return err
			}
			data := []byte(args[0])
			err = kc.InspectStep(string(data), offset, requestId, format, verbose, truncateData, namespace, time.Duration(timeoutSecs)*time.Second)
			return err
		},
	}

	flags := cmd.Flags()
	flags.String(flagKafkaBroker, env.GetString(envKafka, defaultKafkaHost), "kafka broker")
	flags.Int64(flagOffset, 1, "message offset to start reading from, i.e. default 1 is the last message only")
	flags.String(flagRequestId, "", "request id to show, if not specified will be all messages in offset range")
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagOutputFormat, cli.InspectFormatRaw, fmt.Sprintf("inspect output format: raw or json. Default %s", cli.InspectFormatRaw))
	flags.String(flagNamespace, env.GetString(envNamespace, cli.DefaultNamespace), fmt.Sprintf("Kubernetes namespace. Default %s", cli.DefaultNamespace))
	flags.BoolP(flagVerbose, "v", false, "display more details, such as headers")
	flags.BoolP(flagTruncate, "t", false, "truncate data")
	flags.String(flagKafkaConfigPath, env.GetString(envKafkaConfigPath, ""), "path to kafka config file")
	flags.Int64P(flagTimeout, "d", flagTimeoutDefault, "timeout seconds for kafka operations")
	return cmd
}
