package describe

import (
	"fmt"
	"log"
	"text/tabwriter"
	"text/template"

	"github.com/openshift-pipelines/manual-approval-gate/pkg/actions"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/apis/approvaltask/v1alpha1"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/cli"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/cli/flags"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/cli/formatter"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var taskTemplate = `📦 Name:            {{ .ApprovalTask.Name }}
🗂  Namespace:       {{ .ApprovalTask.Namespace }}
{{- $pipelineRunRef := pipelineRunRef .ApprovalTask }}
{{- if ne $pipelineRunRef "" }}
🏷️  PipelineRunRef:  {{ $pipelineRunRef }}
{{- end }}

👥 Approvers
{{- range .ApprovalTask.Spec.Approvers }}
   * {{ .Name }}
{{- end }}


{{- if gt (len .ApprovalTask.Status.ApproversResponse) 0 }}

👨‍💻 ApproverResponse

Name	ApproverResponse	Message
{{- range .ApprovalTask.Status.ApproversResponse }}
{{ .Name }}	{{response .Response }}	{{message .Message }}
{{- end }}
{{- end }}

🌡️  Status

NumberOfApprovalsRequired	PendingApprovals	STATUS
{{.ApprovalTask.Spec.NumberOfApprovalsRequired}}	{{pendingApprovals .ApprovalTask}}	{{state .ApprovalTask}}
`

var (
	taskGroupResource = schema.GroupVersionResource{Group: "openshift-pipelines.org", Resource: "approvaltasks"}
)

func pendingApprovals(at *v1alpha1.ApprovalTask) int {
	return at.Spec.NumberOfApprovalsRequired - len(at.Status.ApproversResponse)
}

func pipelineRunRef(at *v1alpha1.ApprovalTask) string {
	var pipelineRunReference string
	for k, v := range at.Labels {
		if k == "tekton.dev/pipelineRun" {
			pipelineRunReference = v
		}
	}

	return pipelineRunReference
}

func message(msg string) string {
	if msg == "" {
		return "---"
	}
	return msg
}

func response(response string) string {
	if response == "approved" {
		return "✅"
	}
	return "❌"
}

func Command(p cli.Params) *cobra.Command {
	opts := &cli.Options{}

	funcMap := template.FuncMap{
		"pipelineRunRef":   pipelineRunRef,
		"pendingApprovals": pendingApprovals,
		"message":          message,
		"response":         response,
		"state":            formatter.State,
	}

	c := &cobra.Command{
		Use:   "describe",
		Short: "Describe approval task",
		Long:  `This command describe the approval task.`,
		Annotations: map[string]string{
			"commandType": "main",
		},
		Args:              cobra.ExactArgs(1),
		PersistentPreRunE: flags.PersistentPreRunE(p),
		RunE: func(cmd *cobra.Command, args []string) error {
			cs, err := p.Clients()
			if err != nil {
				return err
			}

			ns := p.Namespace()
			if opts.AllNamespaces {
				ns = ""
			}

			opts = &cli.Options{
				Namespace: ns,
				Name:      args[0],
			}

			at, err := actions.Get(taskGroupResource, cs, opts)
			if err != nil {
				return fmt.Errorf("failed to Get ApprovalTasks %s from %s namespace", args[0], ns)
			}

			var data = struct {
				ApprovalTask *v1alpha1.ApprovalTask
			}{
				ApprovalTask: at,
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 5, 3, ' ', tabwriter.TabIndent)
			t := template.Must(template.New("Describe ApprovalTask").Funcs(funcMap).Parse(taskTemplate))

			if err != nil {
				return err
			}

			if err := t.Execute(w, data); err != nil {
				log.Fatal(err)
				return err
			}

			return w.Flush()
		},
	}
	flags.AddOptions(c)

	return c
}
