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
	"fmt"

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
	flagTimeoutDefault = int64(60)
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
			data := []byte(args[0])
			kc, err := cli.NewKafkaClient(kafkaBroker, kafkaBrokerIsSet, schedulerHost, schedulerHostIsSet)
			if err != nil {
				return err
			}
			err = kc.InspectStep(string(data), offset, requestId, format, verbose, truncateData, namespace)
			return err
		},
	}

	flags := cmd.Flags()
	flags.String(flagKafkaBroker, env.GetString(envKafka, defaultKafkaHost), "kafka broker")
	flags.Int64(flagOffset, 1, "message offset to start reading from, i.e. default 1 is the last message only")
	flags.String(flagRequestId, "", "request id to show, if not specified will be all messages in offset range")
	flags.String(flagSchedulerHost, env.GetString(envScheduler, defaultSchedulerHost), helpSchedulerHost)
	flags.String(flagOutputFormat, cli.InspectFormatRaw, fmt.Sprintf("inspect output format: raw or json. Default %s", cli.InspectFormatRaw))
	flags.String(flagNamespace, "", fmt.Sprintf("Kubernetes namespace. Default %s", cli.DefaultNamespace))
	flags.BoolP(flagVerbose, "v", false, "display more details, such as headers")
	flags.BoolP(flagTruncate, "t", false, "truncate data")
	return cmd
}
