// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cli

import (
	"fmt"

	"github.com/clivern/finn/core"

	"github.com/spf13/cobra"
)

var runnerCmd = &cobra.Command{
	Use:   "runner",
	Short: "Start the finn runner server",
	Run: func(_ *cobra.Command, _ []string) {
		if err := core.Load(config); err != nil {
			panic(err.Error())
		}

		if err := core.SetupLogging(); err != nil {
			panic(err.Error())
		}

		db, err := core.InitRunner()
		if err != nil {
			panic(fmt.Sprintf("Runner init: %s", err.Error()))
		}

		r := core.SetupRunner(db)
		if err := core.RunRunner(r, db); err != nil {
			panic(fmt.Sprintf("Runner error: %s", err.Error()))
		}
	},
}

func init() {
	runnerCmd.Flags().StringVarP(
		&config,
		"config",
		"c",
		"runner_config.dist.yml",
		"Absolute path to config file (required)",
	)
	runnerCmd.MarkFlagRequired("config")
	rootCmd.AddCommand(runnerCmd)
}
