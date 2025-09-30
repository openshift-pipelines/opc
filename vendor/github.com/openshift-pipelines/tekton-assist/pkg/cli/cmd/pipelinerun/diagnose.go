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

package pipelinerun

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
	PipelineRunName string
	Output          string
	BaseURL         string
	Timeout         time.Duration
}

// DiagnoseCommand creates the diagnose command for PipelineRuns
func DiagnoseCommand(params common.Params) *cobra.Command {
	opts := &DiagnoseOptions{
		Params:  params,
		Output:  "text",
		Timeout: 30 * time.Second,
	}

	diagnoseCmd := &cobra.Command{
		Use:   "diagnose <pipelinerun-name>",
		Short: "Diagnose a PipelineRun and provide AI-powered analysis",
		Long: `Diagnose analyzes a PipelineRun's status, associated TaskRuns, and events to identify issues
and provide AI-powered recommendations for fixing failures.

This command will:
1. Check the PipelineRun status and conditions
2. Query associated TaskRuns using the pipelineRun label
3. If failed TaskRuns exist, provide a list and guidance to use '/taskrun/explainFailure'
4. If no TaskRuns exist, analyze the PipelineRun failure directly with LLM

The analysis helps identify root causes and provides actionable remediation steps.`,
		Example: `  # Diagnose a failed PipelineRun using default output
  tkn-assist pipelinerun diagnose my-failed-pipelinerun

  # Diagnose with JSON output
  tkn-assist pipelinerun diagnose my-failed-pipelinerun --output json

  # Diagnose in a specific namespace
  tkn-assist pipelinerun diagnose my-failed-pipelinerun --namespace my-namespace

  # Use a custom API server URL
  tkn-assist pipelinerun diagnose my-failed-pipelinerun --url http://custom-server:8080`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.PipelineRunName = args[0]
			return runDiagnose(cmd.Context(), opts)
		},
	}

	// Add flags
	diagnoseCmd.Flags().StringVarP(&opts.Output, "output", "o", opts.Output,
		"Output format. One of: text|json|yaml")
	diagnoseCmd.Flags().StringVar(&opts.BaseURL, "url", "",
		"Base URL of the tekton-assist API server")
	diagnoseCmd.Flags().DurationVar(&opts.Timeout, "timeout", opts.Timeout,
		"Timeout for API requests")

	return diagnoseCmd
}

// runDiagnose executes the diagnosis workflow
func runDiagnose(ctx context.Context, opts *DiagnoseOptions) error {
	if opts.Verbose() {
		fmt.Printf("Diagnosing PipelineRun: %s\n", opts.PipelineRunName)
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
	req := &client.ExplainPipelineRunFailureRequest{
		Namespace:   opts.Namespace(),
		PipelineRun: opts.PipelineRunName,
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
		fmt.Printf("Calling API: /pipelinerun/explainFailure?namespace=%s&name=%s\n", req.Namespace, req.PipelineRun)
	}

	response, err := apiClient.ExplainPipelineRunFailure(ctx, req)
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
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON: %w", err)
	}

	fmt.Println(string(prettyJSON))
	return nil
}

// formatYAML converts JSON response to YAML
func formatYAML(response string) error {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(response), &jsonData); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	yamlData, err := yaml.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("failed to convert to YAML: %w", err)
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

// displayStructuredText formats structured JSON data as readable text for PipelineRun
func displayStructuredText(data map[string]interface{}) error {
	fmt.Println("PipelineRun Diagnosis Report")
	fmt.Println("============================")
	fmt.Println()

	// Display PipelineRun basic info
	if pipelineRun, ok := data["pipelineRun"].(map[string]interface{}); ok {
		if name, ok := pipelineRun["name"].(string); ok {
			fmt.Printf("PipelineRun: %s\n", name)
		}
		if namespace, ok := pipelineRun["namespace"].(string); ok {
			fmt.Printf("Namespace: %s\n", namespace)
		}
		if uid, ok := pipelineRun["uid"].(string); ok {
			fmt.Printf("UID: %s\n", uid)
		}
	}

	// Display status information
	if status, ok := data["status"].(map[string]interface{}); ok {
		fmt.Println()
		if phase, ok := status["phase"].(string); ok {
			switch phase {
			case "Succeeded":
				fmt.Printf("Status: ✅ %s\n", phase)
			case "Failed":
				fmt.Printf("Status: ❌ %s\n", phase)
			case "Running":
				fmt.Printf("Status: 🏃 %s\n", phase)
			default:
				fmt.Printf("Status: %s\n", phase)
			}
		}

		if startTime, ok := status["startTime"].(string); ok {
			fmt.Printf("Start Time: %s\n", startTime)
		}
		if completionTime, ok := status["completionTime"].(string); ok {
			fmt.Printf("Completion Time: %s\n", completionTime)
		}
		if duration, ok := status["durationSeconds"].(float64); ok {
			fmt.Printf("Duration: %.0f seconds\n", duration)
		}

		// Display conditions
		if conditions, ok := status["conditions"].([]interface{}); ok && len(conditions) > 0 {
			fmt.Println("\nConditions:")
			for _, condInterface := range conditions {
				if cond, ok := condInterface.(map[string]interface{}); ok {
					condType, _ := cond["type"].(string)
					condStatus, _ := cond["status"].(string)
					reason, _ := cond["reason"].(string)
					message, _ := cond["message"].(string)

					var statusIcon string
					switch condStatus {
					case "True":
						statusIcon = "✅"
					case "False":
						statusIcon = "❌"
					default:
						statusIcon = "❓"
					}

					fmt.Printf("  %s %s: %s (%s)\n", statusIcon, condType, condStatus, reason)
					if message != "" {
						fmt.Printf("    Message: %s\n", message)
					}
				}
			}
		}
	}

	// Display failed TaskRuns
	if failedTaskRuns, ok := data["failedTaskRuns"].([]interface{}); ok {
		fmt.Println()
		if len(failedTaskRuns) > 0 {
			fmt.Printf("Failed TaskRuns (%d):\n", len(failedTaskRuns))
			for i, taskRunInterface := range failedTaskRuns {
				if taskRun, ok := taskRunInterface.(map[string]interface{}); ok {
					name, _ := taskRun["name"].(string)
					reason, _ := taskRun["reason"].(string)
					message, _ := taskRun["message"].(string)

					fmt.Printf("  %d. ❌ %s\n", i+1, name)
					fmt.Printf("     Reason: %s\n", reason)
					if message != "" {
						// Truncate long messages for better readability
						if len(message) > 100 {
							message = message[:97] + "..."
						}
						fmt.Printf("     Message: %s\n", message)
					}
					fmt.Println()
				}
			}
		} else {
			fmt.Println("Failed TaskRuns: None")
		}
	}

	// Display analysis
	if analysis, ok := data["analysis"].(string); ok && analysis != "" {
		fmt.Printf("Analysis & Recommendations:\n")
		fmt.Printf("===========================\n")
		fmt.Printf("%s\n", analysis)
	}

	fmt.Println()
	return nil
}
