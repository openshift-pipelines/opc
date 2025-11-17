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
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// DiagnoseOptions holds options specific to the diagnose command
type DiagnoseOptions struct {
	TaskRunName   string
	Output        string
	Namespace     string
	Verbose       bool
	Kubeconfig    string
	KubeContext   string
	LightspeedURL string
	BearerToken   string
	TokenFile     string
	InsecureTLS   bool
	Timeout       time.Duration
}

// DiagnoseCommand creates the diagnose command for TaskRuns
func DiagnoseCommand() *cobra.Command {
	opts := &DiagnoseOptions{
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
		Annotations: map[string]string{"commandType": "main"},
		Args:        cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.TaskRunName = args[0]
			return runDiagnose(cmd.Context(), opts)
		},
	}

	// Command-specific flags
	diagnoseCmd.Flags().StringVarP(&opts.Output, "output", "o", "text", "Output format (text, json, yaml)")
	diagnoseCmd.Flags().StringVarP(&opts.Namespace, "namespace", "n", "", "Kubernetes namespace")
	diagnoseCmd.Flags().StringVar(&opts.Kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	diagnoseCmd.Flags().StringVar(&opts.KubeContext, "context", "", "Kubernetes context to use")
	diagnoseCmd.Flags().StringVar(&opts.LightspeedURL, "lightspeed-url", "", "Lightspeed service base URL (default: https://localhost:8443)")
	diagnoseCmd.Flags().StringVar(&opts.BearerToken, "token", "", "Bearer token for Lightspeed service (or set LIGHTSPEED_TOKEN)")
	diagnoseCmd.Flags().StringVar(&opts.TokenFile, "token-file", "", "Path to a file containing the bearer token")
	diagnoseCmd.Flags().BoolVarP(&opts.InsecureTLS, "insecure-skip-tls-verify", "k", false, "Skip TLS certificate verification (insecure)")
	diagnoseCmd.Flags().DurationVar(&opts.Timeout, "timeout", 30*time.Second, "Timeout for API requests")

	return diagnoseCmd
}

// runDiagnose executes the diagnosis workflow
func runDiagnose(ctx context.Context, opts *DiagnoseOptions) error {
	if opts.Verbose {
		fmt.Printf("Diagnosing TaskRun: %s\n", opts.TaskRunName)
		if opts.Namespace != "" {
			fmt.Printf("Namespace: %s\n", opts.Namespace)
		}
		fmt.Printf("Output format: %s\n", opts.Output)
		if opts.LightspeedURL != "" {
			fmt.Printf("Lightspeed URL: %s\n", opts.LightspeedURL)
		}
	}

	// Determine the Lightspeed base URL
	baseURL := opts.LightspeedURL
	if baseURL == "" {
		baseURL = "https://localhost:8443"
	}

	if opts.Verbose {
		fmt.Printf("Connecting to Lightspeed at: %s\n", baseURL)
	}

	// Resolve namespace
	namespace := opts.Namespace
	if namespace == "" {
		namespace = "default"
		if opts.Verbose {
			fmt.Printf("Using default namespace: %s\n", namespace)
		}
	}

	// Build query payload (chat-style phrasing + ask for solutions + JSON shape)
	query := fmt.Sprintf(
		"Why is my Tekton TaskRun '%s' failing in namespace '%s'? "+
			"Provide a brief summary, a clear root-cause analysis, and 3-5 actionable solutions. "+
			"If possible, respond as a JSON object with fields: response (string), analysis (string), solutions (array of strings).",
		opts.TaskRunName, namespace,
	)
	if opts.Verbose {
		fmt.Printf("Query: %s\n", query)
	}

	payload := map[string]interface{}{
		"query": query,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Prepare HTTP client
	httpClient := &http.Client{Timeout: opts.Timeout}
	if opts.InsecureTLS {
		transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		httpClient.Transport = transport
	}

	// Prepare request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, joinURL(baseURL, "/v1/query"), bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// Resolve token
	token := resolveToken(opts.BearerToken, opts.TokenFile)
	if token == "" {
		token = resolveTokenFromKubeconfig(opts.Kubeconfig, opts.KubeContext)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Execute request
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request to Lightspeed failed: %w", err)
	}
	defer safeClose(resp.Body)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("lightspeed returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Format and display the response based on output format
	return formatOutput(string(respBody), opts.Output)
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

	printed := false

	// Prefer top-level LLM response if present. Handle embedded fenced JSON blocks.
	if resp, ok := data["response"].(string); ok && resp != "" {
		if openIdx, contentStart, closeStart, okFence := findFence(resp); okFence {
			preface := strings.TrimSpace(resp[:openIdx])
			if preface != "" {
				fmt.Printf("Summary:\n%s\n\n", preface)
				printed = true
			}
			inner := strings.TrimSpace(resp[contentStart:closeStart])
			inner = stripFenceLanguage(inner)
			inner = strings.TrimSpace(inner)
			var embedded interface{}
			if len(inner) > 0 && (inner[0] == '{' || inner[0] == '[') && json.Unmarshal([]byte(inner), &embedded) == nil {
				if obj, ok := embedded.(map[string]interface{}); ok {
					if s, ok := obj["response"].(string); ok && s != "" && preface == "" {
						fmt.Printf("Summary:\n%s\n\n", s)
						printed = true
					}
					if a, ok := obj["analysis"].(string); ok && a != "" {
						fmt.Printf("Analysis & Suggested Remediation:\n%s\n\n", a)
						printed = true
					}
					if sols, ok := obj["solutions"].([]interface{}); ok && len(sols) > 0 {
						fmt.Println("Solutions:")
						for i, s := range sols {
							if str, ok := s.(string); ok && str != "" {
								fmt.Printf("  %d. %s\n", i+1, str)
							}
						}
						fmt.Println()
						printed = true
					}
				}
			}
			// Do not print the fenced block itself
		} else {
			clean := stripCodeFence(resp)
			clean = truncateAtFence(clean)
			if clean != "" {
				fmt.Printf("Summary:\n%s\n\n", clean)
				printed = true
			}
		}
	}

	// Print references if available
	if refs, ok := data["referenced_documents"].([]interface{}); ok && len(refs) > 0 {
		fmt.Println("References:")
		count := 0
		for _, r := range refs {
			if rm, ok := r.(map[string]interface{}); ok {
				title, _ := rm["doc_title"].(string)
				url, _ := rm["doc_url"].(string)
				if title != "" || url != "" {
					fmt.Printf("  - %s%s\n", title, func() string {
						if url != "" {
							return " (" + url + ")"
						}
						return ""
					}())
					count++
				}
			}
			if count >= 5 { // avoid overly long lists
				break
			}
		}
		fmt.Println()
	}

	// Token usage (optional diagnostics)
	if inTok, ok := data["input_tokens"].(float64); ok {
		if outTok, ok := data["output_tokens"].(float64); ok {
			fmt.Printf("Token usage: input %.0f, output %.0f\n\n", inTok, outTok)
		}
	}

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
		printed = true
	}

	// Display analysis if present
	if analysis, ok := data["analysis"].(string); ok && analysis != "" {
		fmt.Printf("\nAnalysis & Suggested Remediation:\n%s\n", analysis)
		printed = true
	}

	// Display solutions if present
	if sols, ok := data["solutions"].([]interface{}); ok {
		if len(sols) > 0 {
			fmt.Println("\nSolutions:")
			for i, s := range sols {
				if str, ok := s.(string); ok && str != "" {
					fmt.Printf("  %d. %s\n", i+1, str)
				}
			}
			printed = true
		}
	}

	// Try generic response keys from Lightspeed or LLM-like responses
	if !printed {
		for _, key := range []string{"answer", "response", "result", "message", "content", "text", "output"} {
			if v, ok := data[key].(string); ok && v != "" {
				fmt.Printf("\nResponse:\n%s\n", v)
				printed = true
				break
			}
		}
	}

	// Handle OpenAI-like choices
	if !printed {
		if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
			var combined string
			for _, ch := range choices {
				if m, ok := ch.(map[string]interface{}); ok {
					// message.content
					if msg, ok := m["message"].(map[string]interface{}); ok {
						if c, ok := msg["content"].(string); ok && c != "" {
							combined += c + "\n"
						}
					}
					// or text
					if t, ok := m["text"].(string); ok && t != "" {
						combined += t + "\n"
					}
				}
			}
			if combined != "" {
				fmt.Printf("\nResponse:\n%s", combined)
				printed = true
			}
		}
	}

	// Fallback: pretty-print JSON if nothing recognized
	if !printed {
		b, err := json.MarshalIndent(data, "", "  ")
		if err == nil {
			fmt.Println("API Response:")
			fmt.Println("=============")
			fmt.Println(string(b))
		}
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

// --- helpers ---

func joinURL(base, path string) string {
	if base == "" {
		return path
	}
	if len(base) > 0 && base[len(base)-1] == '/' {
		base = base[:len(base)-1]
	}
	if len(path) > 0 && path[0] == '/' {
		return base + path
	}
	return base + "/" + path
}

func resolveToken(tokenFlag, tokenFile string) string {
	if tokenFlag != "" {
		return tokenFlag
	}
	if tokenFile != "" {
		if b, err := os.ReadFile(tokenFile); err == nil {
			return string(bytes.TrimSpace(b))
		}
	}
	if env := os.Getenv("LIGHTSPEED_TOKEN"); env != "" {
		return env
	}
	return ""
}

func safeClose(c io.Closer) {
	_ = c.Close()
}

// stripCodeFence removes leading/trailing markdown code fences if present
func stripCodeFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		if nl := strings.Index(s, "\n"); nl != -1 {
			s2 := s[nl+1:]
			if end := strings.LastIndex(s2, "```"); end != -1 {
				s = s2[:end]
			} else {
				s = s2
			}
		}
		s = strings.TrimSpace(s)
	}
	return s
}

// truncateAtFence removes any trailing markdown fence and following content
func truncateAtFence(s string) string {
	if idx := strings.Index(s, "```"); idx != -1 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}

// findFence locates the first ``` fenced code block and returns indexes to its contents
func findFence(s string) (openIdx, contentStart, closeStart int, ok bool) {
	openIdx = strings.Index(s, "```")
	if openIdx == -1 {
		return 0, 0, 0, false
	}
	// find end of opening fence line (may include a language token)
	nl := strings.Index(s[openIdx+3:], "\n")
	if nl == -1 {
		return 0, 0, 0, false
	}
	contentStart = openIdx + 3 + nl + 1
	j := strings.Index(s[contentStart:], "```")
	if j == -1 {
		return 0, 0, 0, false
	}
	closeStart = contentStart + j
	return openIdx, contentStart, closeStart, true
}

// stripFenceLanguage removes a leading language id from a fenced block (e.g., json)
func stripFenceLanguage(s string) string {
	// If the first line looks like a language tag, drop it
	if ln := strings.Index(s, "\n"); ln != -1 {
		first := strings.TrimSpace(s[:ln])
		if first == "json" || first == "yaml" || first == "yml" || first == "bash" || first == "txt" {
			return s[ln+1:]
		}
	}
	return s
}

// resolveTokenFromKubeconfig tries to extract a bearer token from kubeconfig via YAML parsing
func resolveTokenFromKubeconfig(kubeconfigPath, contextName string) string {
	// Determine kubeconfig path
	if kubeconfigPath == "" {
		if env := os.Getenv("KUBECONFIG"); env != "" {
			// If multiple paths, take the first
			parts := strings.Split(env, string(os.PathListSeparator))
			if len(parts) > 0 {
				kubeconfigPath = parts[0]
			}
		} else {
			if home, err := os.UserHomeDir(); err == nil {
				kubeconfigPath = filepath.Join(home, ".kube", "config")
			}
		}
	}
	if kubeconfigPath == "" {
		return ""
	}

	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return ""
	}

	// Minimal kubeconfig model
	type kcUser struct {
		Token     string `yaml:"token"`
		TokenFile string `yaml:"token-file"`
	}
	type kcUserEntry struct {
		Name string `yaml:"name"`
		User kcUser `yaml:"user"`
	}
	type kcContext struct {
		User string `yaml:"user"`
	}
	type kcContextEntry struct {
		Name    string    `yaml:"name"`
		Context kcContext `yaml:"context"`
	}
	type kubeconfig struct {
		CurrentContext string           `yaml:"current-context"`
		Contexts       []kcContextEntry `yaml:"contexts"`
		Users          []kcUserEntry    `yaml:"users"`
	}

	var cfg kubeconfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return ""
	}

	current := contextName
	if current == "" {
		current = cfg.CurrentContext
	}
	if current == "" {
		return ""
	}

	var userName string
	for _, c := range cfg.Contexts {
		if c.Name == current {
			userName = c.Context.User
			break
		}
	}
	if userName == "" {
		return ""
	}

	for _, u := range cfg.Users {
		if u.Name == userName {
			if u.User.Token != "" {
				return u.User.Token
			}
			if u.User.TokenFile != "" {
				if b, err := os.ReadFile(u.User.TokenFile); err == nil {
					return string(bytes.TrimSpace(b))
				}
			}
		}
	}
	return ""
}
