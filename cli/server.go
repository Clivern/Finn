// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"

	"github.com/clivern/finn/core"

	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the finn management server",
	Run: func(_ *cobra.Command, _ []string) {
		if err := core.Load(config); err != nil {
			panic(err.Error())
		}

		if err := core.SetupLogging(); err != nil {
			panic(err.Error())
		}

		r := core.SetupServer(Static)

		if err := core.RunServer(r); err != nil {
			panic(fmt.Sprintf("Server error: %s", err.Error()))
		}
	},
}

// init registers the server subcommand and flags.
func init() {
	serverCmd.Flags().StringVarP(
		&config,
		"config",
		"c",
		"mgmt_config.prod.yml",
		"Absolute path to config file (required)",
	)
	serverCmd.MarkFlagRequired("config")
	rootCmd.AddCommand(serverCmd)
}
