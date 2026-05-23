// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// ScriptExec is the parsed local script command.
type ScriptExec struct {
	Engine     string
	ScriptPath string
	Env        []string
	Args       []string
}

// BuildScriptExec parses a script payload into engine, env, and args for exec.
func BuildScriptExec(dir string, payload map[string]interface{}) (*ScriptExec, error) {
	engine := StringFromPayload(payload, "script.engine")
	if engine == "" {
		return nil, fmt.Errorf("script: engine is required")
	}

	scriptPath := StringFromPayload(payload, "script.scriptPath")
	if scriptPath == "" {
		return nil, fmt.Errorf("script: script path is required")
	}

	env := append(
		keyValueArgs(MapFromPayload(payload, "groupVars.environmentVariables")),
		keyValueArgs(MapFromPayload(payload, "groupVars.secretEnvironmentVariables"))...,
	)
	args := append(
		keyValueArgs(MapFromPayload(payload, "extraArgs")),
		keyValueArgs(MapFromPayload(payload, "groupVars.extraVariables"))...,
	)
	args = append(args, keyValueArgs(MapFromPayload(payload, "groupVars.secretExtraVariables"))...)

	return &ScriptExec{
		Engine:     engine,
		ScriptPath: filepath.Join(dir, filepath.FromSlash(scriptPath)),
		Env:        env,
		Args:       args,
	}, nil
}

// BuildScriptCommand builds shell-style command tokens from a script payload.
func BuildScriptCommand(dir string, payload map[string]interface{}) ([]string, error) {
	execCmd, err := BuildScriptExec(dir, payload)
	if err != nil {
		return nil, err
	}

	command := make([]string, 0, len(execCmd.Env)+2+len(execCmd.Args))
	command = append(command, execCmd.Env...)
	command = append(command, execCmd.Engine, execCmd.ScriptPath)
	command = append(command, execCmd.Args...)
	return command, nil
}

func keyValueArgs(values map[string]interface{}) []string {
	if len(values) == 0 {
		return nil
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	args := make([]string, 0, len(keys))
	for _, key := range keys {
		if values[key] == nil {
			continue
		}
		args = append(args, fmt.Sprintf("%s=%s", key, strings.TrimSpace(fmt.Sprint(values[key]))))
	}
	return args
}
