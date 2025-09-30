// Copyright 2025 The Tekton Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"fmt"
	"os"

	"github.com/openshift-pipelines/tekton-assist/pkg/cli/cmd/pipelinerun"
	"github.com/openshift-pipelines/tekton-assist/pkg/cli/cmd/taskrun"
	"github.com/openshift-pipelines/tekton-assist/pkg/cli/common"
	"github.com/spf13/cobra"
)

const (
	// Version information - will be set during build
	version = "dev"
)

// RootCommand creates the root command for tkn-assist
func RootCommand() *cobra.Command {
	// Create common params
	params := common.NewParams()

	rootCmd := &cobra.Command{
		Use:   "tkn-assist",
		Short: "Tekton Assistant CLI - AI-powered diagnosis for Tekton resources",
		Long: `tkn-assist is a CLI tool that helps diagnose and troubleshoot Tekton TaskRuns and PipelineRuns.
It uses AI to analyze failures and provide actionable remediation suggestions.

This tool can be used as a tkn plugin by naming the binary 'tkn-assist'.`,
		Example: `  # Diagnose a failed TaskRun
  tkn-assist taskrun diagnose my-failed-taskrun

  # Diagnose a failed PipelineRun
  tkn-assist pipelinerun diagnose my-failed-pipelinerun

  # Diagnose a TaskRun in a specific namespace
  tkn-assist taskrun diagnose my-taskrun -n my-namespace`,
		Annotations: map[string]string{
			"commandType": "main",
		},
	}

	// Add global flags using common params
	params.AddFlags(rootCmd)

	// Add subcommands
	rootCmd.AddCommand(taskrun.TaskRunCommand(params))
	rootCmd.AddCommand(pipelinerun.PipelineRunCommand(params))
	rootCmd.AddCommand(versionCommand())

	return rootCmd
}

// versionCommand creates the version command
func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("tkn-assist version %s\n", version)
		},
	}
}

// Execute runs the root command
func Execute() {
	if err := RootCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
