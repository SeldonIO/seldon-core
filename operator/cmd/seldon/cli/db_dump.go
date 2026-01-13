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
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/utils/env"

	"github.com/seldonio/seldon-core/operator/v2/pkg/cli"
)

func dbDump() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dbdump",
		Short: "download a database dump",
		Long:  `Download a database dump to a file in protobuf or json format. File will be created automatically if it does not exist`,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()

			dbHostIsSet := flags.Changed(flagDBHost)
			dbHost, err := flags.GetString(flagDBHost)
			if err != nil {
				return err
			}
			authority, err := flags.GetString(flagAuthority)
			if err != nil {
				return err
			}
			verbose, err := flags.GetBool(flagVerbose)
			if err != nil {
				return err
			}

			outputFile, err := flags.GetString(flagDBOutputFile)
			if err != nil {
				return err
			}

			outputFormat, err := flags.GetString(flagDBOutputFormat)
			if err != nil {
				return err
			}

			recordTypeStr, err := flags.GetString(flagDBRecordType)
			if err != nil {
				return err
			}

			var recordType cli.DBRecordType
			switch strings.ToLower(recordTypeStr) {
			case "all":
				recordType = cli.DBRecordTypeAll
			case "models":
				recordType = cli.DBRecordTypeModels
			case "servers":
				recordType = cli.DBRecordTypeServers
			default:
				return fmt.Errorf("invalid record type %q, must be one of: all, models, servers", recordTypeStr)
			}

			var format cli.DBDumpOutputFormat
			switch strings.ToLower(outputFormat) {
			case "json":
				format = cli.DBDumpOutputFormatJson
			case "proto":
				format = cli.DBDumpOutputFormatProto
			default:
				format = cli.DBDumpOutputFormatProto
			}

			dbClient, err := cli.NewDBClient(dbHost, dbHostIsSet, authority, verbose)
			if err != nil {
				return err
			}

			return dbClient.DumpDatabase(cli.DumpDBRequest{
				Action:       recordType,
				OutputFile:   outputFile,
				OutputFormat: format,
			})
		},
	}

	flags := cmd.Flags()
	flags.BoolP(flagVerbose, "v", false, "verbose output")
	flags.String(flagDBHost, env.GetString(envDBHost, defaultDBHost), helpDBHost)
	flags.String(flagAuthority, "", helpAuthority)
	flags.StringP(flagDBOutputFile, "o", "", helpDBOutputFile)
	flags.StringP(flagDBOutputFormat, "f", string(cli.DBDumpOutputFormatProto), helpDBOutputFormat)
	flags.StringP(flagDBRecordType, "r", string(cli.DBRecordTypeAll), helpDBRecordType)

	_ = cmd.MarkFlagRequired(flagDBOutputFile)

	return cmd
}
