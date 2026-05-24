// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package tasks

import "testing"

func TestScriptDockerImage(t *testing.T) {
	tests := []struct {
		name    string
		engine  string
		version string
		want    string
		wantErr bool
	}{
		{
			name:   "default bash image",
			engine: "bash",
			want:   "docker.io/bash:5.2",
		},
		{
			name:    "supported bash version",
			engine:  "bash",
			version: "5.3",
			want:    "docker.io/bash:5.3",
		},
		{
			name:    "unsupported bash version",
			engine:  "bash",
			version: "3.2",
			wantErr: true,
		},
		{
			name:    "supported python3 version",
			engine:  "python3",
			version: "3.13",
			want:    "docker.io/python:3.13-slim",
		},
		{
			name:    "python alias",
			engine:  "python",
			version: "3.12",
			want:    "docker.io/python:3.12-slim",
		},
		{
			name:    "supported node version",
			engine:  "node",
			version: "22",
			want:    "docker.io/node:22-bookworm-slim",
		},
		{
			name:    "unsupported engine",
			engine:  "ruby",
			version: "3.3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := scriptDockerImage(tt.engine, tt.version)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("scriptDockerImage() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("scriptDockerImage() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("scriptDockerImage() = %q, want %q", got, tt.want)
			}
		})
	}
}
