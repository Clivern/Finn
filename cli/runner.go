// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/clivern/finn/core"
	"github.com/clivern/finn/service/runner"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

var runnerIdentityOutput string

var runnerIdentityCmd = &cobra.Command{
	Use:   "identity",
	Short: "Output runner identity token and public key",
	Run: func(_ *cobra.Command, _ []string) {
		if err := core.Load(config); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}

		path := viper.GetString("app.database.path")
		if path == "" {
			fmt.Fprintln(os.Stderr, "runner identity: app.database.path is required")
			os.Exit(1)
		}

		db := &runner.Bolt{Path: path}
		if err := db.Open(); err != nil {
			fmt.Fprintf(os.Stderr, "runner identity: %s\n", err.Error())
			os.Exit(1)
		}
		defer db.Close()

		mod := &runner.Module{Bolt: db, Key: &runner.Key{}}
		identity, err := mod.GetIdentity()
		if err != nil || identity == nil {
			identity, err = mod.Init()
			if err != nil {
				fmt.Fprintf(os.Stderr, "runner identity: %s\n", err.Error())
				os.Exit(1)
			}
		}

		switch runnerIdentityOutput {
		case "json":
			out := map[string]string{
				"token":      identity.ID,
				"public_key": string(identity.PublicKey),
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetEscapeHTML(false)
			if err := enc.Encode(out); err != nil {
				fmt.Fprintf(os.Stderr, "runner identity: %s\n", err.Error())
				os.Exit(1)
			}
		default:
			fmt.Printf("Token: %s\n\n", identity.ID)
			fmt.Println("Public Key:")
			fmt.Print(string(identity.PublicKey))
			if len(identity.PublicKey) > 0 && identity.PublicKey[len(identity.PublicKey)-1] != '\n' {
				fmt.Println()
			}
		}
	},
}

func init() {
	runnerCmd.PersistentFlags().StringVarP(
		&config,
		"config",
		"c",
		"runner_config.dist.yml",
		"Absolute path to config file (required)",
	)
	runnerCmd.MarkPersistentFlagRequired("config")
	runnerIdentityCmd.Flags().StringVarP(&runnerIdentityOutput, "output", "o", "text", "Output format: text (default) or json")
	runnerCmd.AddCommand(runnerIdentityCmd)
	rootCmd.AddCommand(runnerCmd)
}
