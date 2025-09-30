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

package common

import "github.com/spf13/cobra"

// TektonParams holds common CLI parameters
type TektonParams struct {
	kubeconfig string
	context    string
	namespace  string
	verbose    bool
}

// NewParams creates a new TektonParams instance
func NewParams() *TektonParams {
	return &TektonParams{}
}

// AddFlags adds common flags to the command
func (p *TektonParams) AddFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&p.kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	cmd.PersistentFlags().StringVar(&p.context, "context", "", "Kubernetes context to use")
	cmd.PersistentFlags().StringVarP(&p.namespace, "namespace", "n", "", "Kubernetes namespace")
	cmd.PersistentFlags().BoolVarP(&p.verbose, "verbose", "v", false, "Enable verbose output")
}

// Kubeconfig returns the path to the kubeconfig file
func (p *TektonParams) Kubeconfig() string {
	return p.kubeconfig
}

// Context returns the kubernetes context to use
func (p *TektonParams) Context() string {
	return p.context
}

// Namespace returns the kubernetes namespace
func (p *TektonParams) Namespace() string {
	return p.namespace
}

// Verbose returns whether verbose output is enabled
func (p *TektonParams) Verbose() bool {
	return p.verbose
}
