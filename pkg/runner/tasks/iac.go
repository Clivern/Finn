// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import (
	"context"
	"fmt"

	"github.com/clivern/finn/pkg/runner"
)

// RunIaC runs an IaC template task.
func RunIaC(ctx context.Context, deps *runner.Deps, task *runner.Task, dir string) error {
	engine := StringFromPayload(task.Payload, "iac.engine")
	switch engine {
	case "terraform", "":
		return RunTerraform(ctx, deps, task, dir)
	case "opentofu":
		return RunOpenTofu(ctx, deps, task, dir)
	default:
		return fmt.Errorf("iac: unsupported engine %q", engine)
	}
}
