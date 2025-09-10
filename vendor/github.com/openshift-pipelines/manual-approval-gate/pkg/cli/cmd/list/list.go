package list

import (
	"fmt"
	"log"
	"text/tabwriter"
	"text/template"

	"github.com/fatih/color"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/actions"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/apis/approvaltask/v1alpha1"
	cli "github.com/openshift-pipelines/manual-approval-gate/pkg/cli"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/cli/flags"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ListOptions struct {
	AllNamespaces bool
}

var (
	taskGroupResource = schema.GroupVersionResource{Group: "openshift-pipelines.org", Resource: "approvaltasks"}
)

var ConditionColor = map[string]color.Attribute{
	"Rejected": color.FgHiRed,
	"Approved": color.FgHiGreen,
	"Pending":  color.FgHiYellow,
}

const listTemplate = `{{- $at := len .ApprovalTasks.Items }}{{ if eq $at 0 -}}
No ApprovalTasks found
{{else -}}
NAME	NumberOfApprovalsRequired	PendingApprovals	Rejected	STATUS
{{range .ApprovalTasks.Items -}}
{{.Name}}	{{.Spec.NumberOfApprovalsRequired}}	{{pendingApprovals .}}	{{rejected .}}	{{state .}}
{{end}}
{{- end -}}
`

func pendingApprovals(at *v1alpha1.ApprovalTask) int {
	return at.Spec.NumberOfApprovalsRequired - len(at.Status.ApproversResponse)
}

func rejected(at *v1alpha1.ApprovalTask) int {
	count := 0
	for _, approver := range at.Status.ApproversResponse {
		if approver.Response == "rejected" {
			count = count + 1
		}
	}
	return count
}

func ColorStatus(status string) string {
	return color.New(ConditionColor[status]).Sprint(status)
}

func state(at *v1alpha1.ApprovalTask) string {
	var state string

	switch at.Status.State {
	case "approved":
		state = "Approved"
	case "rejected":
		state = "Rejected"
	case "pending":
		state = "Pending"
	}
	return ColorStatus(state)
}

func Command(p cli.Params) *cobra.Command {
	opts := &ListOptions{}
	funcMap := template.FuncMap{
		"pendingApprovals": pendingApprovals,
		"state":            state,
		"rejected":         rejected,
	}

	c := &cobra.Command{
		Use:   "list",
		Short: "List all approval tasks",
		Long:  `This command lists all the approval tasks.`,
		Annotations: map[string]string{
			"commandType": "main",
		},
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

			var at *v1alpha1.ApprovalTaskList
			if err := actions.List(taskGroupResource, cs, metav1.ListOptions{}, ns, &at); err != nil {
				return fmt.Errorf("failed to list Tasks from namespace %s: %v", ns, err)
			}

			var data = struct {
				ApprovalTasks *v1alpha1.ApprovalTaskList
			}{
				ApprovalTasks: at,
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 5, 3, ' ', tabwriter.TabIndent)
			t := template.Must(template.New("List ApprovalTasks").Funcs(funcMap).Parse(listTemplate))

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

	c.Flags().BoolVarP(&opts.AllNamespaces, "all-namespaces", "A", opts.AllNamespaces, "list Tasks from all namespaces")

	return c
}
