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
	prcmd "github.com/openshift-pipelines/tekton-assist/pkg/cli/pipelinerun"
	trcmd "github.com/openshift-pipelines/tekton-assist/pkg/cli/taskrun"
	"github.com/spf13/cobra"
)

// RootCommand returns the root command for the assist CLI. Consumers (like OPC)
// can import this package and mount the returned command under their own root.
func RootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "tkn-assist",
		Short: "AI-assisted diagnosis for Tekton",
		Long:  `tkn plugin to use AI-assisted diagnosis for Tekton`,
		Annotations: map[string]string{
			"commandType": "main",
		},
	}

	// Add top-level groups
	root.AddCommand(trcmd.TaskRunCommand())
	root.AddCommand(prcmd.PipelineRunCommand())

	return root
}
