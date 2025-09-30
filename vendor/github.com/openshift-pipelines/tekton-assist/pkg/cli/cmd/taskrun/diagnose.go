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

package taskrun

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openshift-pipelines/tekton-assist/pkg/cli/client"
	"github.com/openshift-pipelines/tekton-assist/pkg/cli/common"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// DiagnoseOptions holds options specific to the diagnose command
type DiagnoseOptions struct {
	common.Params
	TaskRunName string
	Output      string
	BaseURL     string
	Timeout     time.Duration
}

// DiagnoseCommand creates the diagnose command for TaskRuns
func DiagnoseCommand(params common.Params) *cobra.Command {
	opts := &DiagnoseOptions{
		Params:  params,
		Output:  "text",
		Timeout: 30 * time.Second,
	}

	diagnoseCmd := &cobra.Command{
		Use:   "diagnose <taskrun-name>",
		Short: "Diagnose a TaskRun and provide AI-powered analysis",
		Long: `Diagnose analyzes a TaskRun's status, logs, and events to identify issues
and provide AI-powered recommendations for fixing failures.

The command will:
1. Fetch the TaskRun status and events
2. Collect relevant logs from failed steps
3. Send data to the Tekton Assistant API for analysis
4. Display actionable recommendations`,
		Example: `  # Diagnose a TaskRun in the current namespace
  tkn-assist taskrun diagnose my-failed-taskrun

  # Diagnose with JSON output
  tkn-assist taskrun diagnose my-taskrun -o json

  # Diagnose with custom base URL
  tkn-assist taskrun diagnose my-taskrun --base-url http://localhost:8080

  # Diagnose with custom timeout
  tkn-assist taskrun diagnose my-taskrun --timeout 60s`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.TaskRunName = args[0]
			return runDiagnose(cmd.Context(), opts)
		},
	}

	// Command-specific flags
	diagnoseCmd.Flags().StringVarP(&opts.Output, "output", "o", "text", "Output format (text, json, yaml)")
	diagnoseCmd.Flags().StringVar(&opts.BaseURL, "base-url", "", "Tekton Assistant API base URL (default: http://localhost:8080)")
	diagnoseCmd.Flags().DurationVar(&opts.Timeout, "timeout", 30*time.Second, "Timeout for API requests")

	return diagnoseCmd
}

// runDiagnose executes the diagnosis workflow
func runDiagnose(ctx context.Context, opts *DiagnoseOptions) error {
	if opts.Verbose() {
		fmt.Printf("Diagnosing TaskRun: %s\n", opts.TaskRunName)
		if opts.Namespace() != "" {
			fmt.Printf("Namespace: %s\n", opts.Namespace())
		}
		fmt.Printf("Output format: %s\n", opts.Output)
		if opts.BaseURL != "" {
			fmt.Printf("Base URL: %s\n", opts.BaseURL)
		}
	}

	// Determine the API base URL
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	// Create the API client
	clientConfig := &client.Config{
		BaseURL: baseURL,
		Timeout: opts.Timeout,
	}
	apiClient := client.NewClient(clientConfig)

	if opts.Verbose() {
		fmt.Printf("Connecting to API at: %s\n", baseURL)
	}

	// Prepare the request
	req := &client.ExplainFailureRequest{
		Namespace: opts.Namespace(),
		TaskRun:   opts.TaskRunName,
	}

	// Handle default namespace
	if req.Namespace == "" {
		req.Namespace = "default"
		if opts.Verbose() {
			fmt.Printf("Using default namespace: %s\n", req.Namespace)
		}
	}

	// Call the API
	if opts.Verbose() {
		fmt.Printf("Calling API: /taskrun/explainFailure?namespace=%s&taskrun=%s\n", req.Namespace, req.TaskRun)
	}

	response, err := apiClient.ExplainFailure(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get diagnosis from API: %w", err)
	}

	// Format and display the response based on output format
	return formatOutput(response, opts.Output)
}

// formatOutput formats the API response according to the specified output format
func formatOutput(response, format string) error {
	switch format {
	case "json":
		return formatJSON(response)
	case "yaml":
		return formatYAML(response)
	case "text":
		fallthrough
	default:
		return formatText(response)
	}
}

// formatJSON pretty-prints the JSON response
func formatJSON(response string) error {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(response), &jsonData); err != nil {
		// If it's not valid JSON, print as-is
		fmt.Println(response)
		return nil
	}

	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Println(response)
		return nil
	}

	fmt.Println(string(prettyJSON))
	return nil
}

// formatYAML converts JSON response to YAML format
func formatYAML(response string) error {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(response), &jsonData); err != nil {
		// If it's not valid JSON, print as-is
		fmt.Println(response)
		return nil
	}

	yamlData, err := yaml.Marshal(jsonData)
	if err != nil {
		fmt.Println(response)
		return nil
	}

	fmt.Print(string(yamlData))
	return nil
}

// formatText displays the response in a human-readable text format
func formatText(response string) error {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(response), &jsonData); err != nil {
		// If it's not valid JSON, print as-is with header
		fmt.Println("API Response:")
		fmt.Println("=============")
		fmt.Println(response)
		return nil
	}

	// Try to parse as structured data for better text formatting
	if data, ok := jsonData.(map[string]interface{}); ok {
		return displayStructuredText(data)
	}

	// Fallback to pretty JSON if we can't structure it
	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Println(response)
		return nil
	}

	fmt.Println("API Response:")
	fmt.Println("=============")
	fmt.Println(string(prettyJSON))
	return nil
}

// displayStructuredText formats structured JSON data as readable text
func displayStructuredText(data map[string]interface{}) error {
	fmt.Println("TaskRun Diagnosis Report")
	fmt.Println("========================")
	fmt.Println()

	// Handle the actual JSON structure from the server
	if debug, ok := data["debug"].(map[string]interface{}); ok {
		// Display basic info
		if taskrun, ok := debug["taskrun"].(string); ok {
			fmt.Printf("TaskRun: %s\n", taskrun)
		}
		if namespace, ok := debug["namespace"].(string); ok {
			fmt.Printf("Namespace: %s\n", namespace)
		}
		if succeeded, ok := debug["succeeded"].(bool); ok {
			if succeeded {
				fmt.Printf("Succeeded: ✅ Yes\n")
			} else {
				fmt.Printf("Succeeded: ❌ No\n")
			}
		}

		// Display failed step info
		if failedStep, ok := debug["failed_step"].(map[string]interface{}); ok {
			if name, ok := failedStep["name"].(string); ok {
				fmt.Printf("Failed Step: %s\n", name)
			}
			if exitCode, ok := failedStep["exit_code"].(float64); ok {
				fmt.Printf("Exit Code: %.0f\n", exitCode)
			}
		}

		// Display error details
		if errorInfo, ok := debug["error"].(map[string]interface{}); ok {
			fmt.Println("\nError Details:")
			if errorType, ok := errorInfo["type"].(string); ok {
				fmt.Printf("Type: %s\n", errorType)
			}
			if status, ok := errorInfo["status"].(string); ok {
				fmt.Printf("Status: %s\n", status)
			}
			if reason, ok := errorInfo["reason"].(string); ok {
				fmt.Printf("Reason: %s\n", reason)
			}
			if message, ok := errorInfo["message"].(string); ok {
				fmt.Printf("Message: %s\n", message)
			}
			if logSnippet, ok := errorInfo["log_snippet"].(string); ok {
				if logSnippet != "" && logSnippet != errorInfo["message"] {
					fmt.Printf("\nLog Snippet:\n%s\n", logSnippet)
				}
			}
		}
	}

	// Display analysis if present
	if analysis, ok := data["analysis"].(string); ok && analysis != "" {
		fmt.Printf("\nAnalysis & Suggested Remediation:\n%s\n", analysis)
	}

	fmt.Println()
	return nil
}

// DiagnoseResult represents the output of a diagnosis
type DiagnoseResult struct {
	TaskRunName   string                 `json:"taskrunName" yaml:"taskrunName"`
	Namespace     string                 `json:"namespace" yaml:"namespace"`
	Status        string                 `json:"status" yaml:"status"`
	FailedSteps   []string               `json:"failedSteps,omitempty" yaml:"failedSteps,omitempty"`
	Analysis      string                 `json:"analysis" yaml:"analysis"`
	Suggestions   []string               `json:"suggestions" yaml:"suggestions"`
	ErrorMessages []string               `json:"errorMessages,omitempty" yaml:"errorMessages,omitempty"`
	Timestamp     time.Time              `json:"timestamp" yaml:"timestamp"`
	Metadata      map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// Display formats and prints the diagnosis result
func (r *DiagnoseResult) Display(format string) error {
	switch format {
	case "json":
		return r.displayJSON()
	case "yaml":
		return r.displayYAML()
	case "text":
		fallthrough
	default:
		return r.displayText()
	}
}

func (r *DiagnoseResult) displayText() error {
	fmt.Printf("TaskRun Diagnosis Report\n")
	fmt.Printf("========================\n\n")
	fmt.Printf("TaskRun: %s\n", r.TaskRunName)
	fmt.Printf("Namespace: %s\n", r.Namespace)
	fmt.Printf("Status: %s\n", r.Status)
	fmt.Printf("Analyzed at: %s\n\n", r.Timestamp.Format(time.RFC3339))

	if len(r.FailedSteps) > 0 {
		fmt.Printf("Failed Steps:\n")
		for _, step := range r.FailedSteps {
			fmt.Printf("  - %s\n", step)
		}
		fmt.Printf("\n")
	}

	if len(r.ErrorMessages) > 0 {
		fmt.Printf("Error Messages:\n")
		for _, msg := range r.ErrorMessages {
			fmt.Printf("  - %s\n", msg)
		}
		fmt.Printf("\n")
	}

	fmt.Printf("Analysis:\n%s\n\n", r.Analysis)

	if len(r.Suggestions) > 0 {
		fmt.Printf("Recommendations:\n")
		for i, suggestion := range r.Suggestions {
			fmt.Printf("  %d. %s\n", i+1, suggestion)
		}
	}

	return nil
}

func (r *DiagnoseResult) displayJSON() error {
	// TODO: Implement JSON output
	return fmt.Errorf("JSON output not implemented yet")
}

func (r *DiagnoseResult) displayYAML() error {
	// TODO: Implement YAML output
	return fmt.Errorf("YAML output not implemented yet")
}
