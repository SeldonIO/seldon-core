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

	"github.com/spf13/cobra"
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

func createModelInfer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "infer <modelName> (data)",
		Short: "run inference on a model",
		Long:  `call a model with a given input and get a prediction`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			inferenceHostIsSet := flags.Changed(flagInferenceHost)
			inferenceHost, err := flags.GetString(flagInferenceHost)
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
			stickySession, err := flags.GetBool(flagStickySession)
			if err != nil {
				return err
			}
			showHeaders, err := flags.GetBool(flagShowHeaders)
			if err != nil {
				return err
			}
			inferMode, err := flags.GetString(flagInferenceMode)
			if err != nil {
				return err
			}
			iterations, err := flags.GetInt(flagInferenceIterations)
			if err != nil {
				return err
			}
			secs, err := flags.GetInt64(flagInferenceSecs)
			if err != nil {
				return err
			}
			headers, err := flags.GetStringArray(flagAddHeader)
			if err != nil {
				return err
			}
			modelName := args[0]

			// Get inference data
			var data []byte
			if len(args) > 1 {
				data = []byte(args[1])
			} else if filename != "" {
				data = loadFile(filename)
			} else {
				return fmt.Errorf("required inline data or from file with -f <file-path>")
			}

			inferModeParsed, err := getInferProtocol(inferMode)
			if err != nil {
				return err
			}

			inferenceClient, err := cli.NewInferenceClient(inferenceHost, inferenceHostIsSet)
			if err != nil {
				return err
			}

			callOpts := &cli.CallOptions{
				InferProtocol: inferModeParsed,
				InferType:     cli.InferModel,
				StickySession: stickySession,
				Iterations:    iterations,
				Seconds:       secs,
			}
			logOpts := &cli.LogOptions{
				ShowHeaders:  showHeaders,
				ShowRequest:  showRequest,
				ShowResponse: showResponse,
			}

			_, err = inferenceClient.Infer(
				modelName,
				data,
				headers,
				authority,
				callOpts,
				logOpts,
			)
			return err
		},
	}

	flags := cmd.Flags()
	flags.BoolP(flagShowRequest, "r", false, "show request")
	flags.BoolP(flagShowResponse, "o", true, "show response")
	flags.StringP(flagFile, "f", "", helpFileInference)
	flags.BoolP(flagStickySession, "s", false, helpStickySession)
	flags.String(flagInferenceHost, env.GetString(envInfer, defaultInferHost), helpInferenceHost)
	flags.String(flagInferenceMode, "rest", helpInferenceMode)
	flags.IntP(flagInferenceIterations, "i", 1, helpInferenceIterations)
	flags.Int64P(flagInferenceSecs, "t", 0, helpInferenceSecs)
	flags.Bool(flagShowHeaders, false, helpShowHeaders)
	flags.StringArray(flagAddHeader, []string{}, helpAddHeader)
	flags.String(flagAuthority, "", helpAuthority)

	return cmd
}
