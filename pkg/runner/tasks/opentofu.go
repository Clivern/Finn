// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"
	"strings"

	"github.com/clivern/finn/pkg/runner"
)

var (
	opentofuImages = map[string]string{
		"":     "ghcr.io/opentofu/opentofu:1.9",
		"1.8":  "ghcr.io/opentofu/opentofu:1.8",
		"1.9":  "ghcr.io/opentofu/opentofu:1.9",
		"1.10": "ghcr.io/opentofu/opentofu:1.10",
	}
)

func RunOpenTofu(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	execCmd, err := BuildIaCExec(dir, task.Payload)
	if err != nil {
		return err
	}

	if IsDockerAvailable() {
		return runIaCDocker(ctx, deps, task, execCmd, "tofu", opentofuDockerImage)
	}

	return runIaCLocal(ctx, deps, task, execCmd, "tofu")
}

func opentofuDockerImage(version string) (string, error) {
	version = strings.TrimSpace(version)
	image, ok := opentofuImages[version]
	if !ok {
		return "", fmt.Errorf("iac: unsupported opentofu version %q", version)
	}
	return image, nil
}
