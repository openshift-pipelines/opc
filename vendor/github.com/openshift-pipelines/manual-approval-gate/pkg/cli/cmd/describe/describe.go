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
   * {{ .Name }}{{if eq .Type "Group"}} (Group){{end}}
{{- end }}


{{- if gt (len .ApprovalTask.Status.ApproversResponse) 0 }}

👨‍💻 ApproverResponse

Name	ApproverResponse	Message
{{- $userGroups := userGroups .ApprovalTask.Status.ApproversResponse}}
{{- range $user, $groups := $userGroups}}
{{$user}}{{if gt (len $groups.Groups) 0}}({{$groups.GroupsStr}}){{end}}	{{response $groups.Response}}	{{message $groups.Message}}
{{- end}}
{{- range .ApprovalTask.Status.ApproversResponse}}
{{- if eq .Type "User"}}
{{.Name}}	{{response .Response}}	{{message .Message}}
{{- end}}
{{- end}}
{{- end}}

🌡️  Status

NumberOfApprovalsRequired	PendingApprovals	STATUS
{{.ApprovalTask.Spec.NumberOfApprovalsRequired}}	{{pendingApprovals .ApprovalTask}}	{{state .ApprovalTask}}
`

var (
	taskGroupResource = schema.GroupVersionResource{Group: "openshift-pipelines.org", Resource: "approvaltasks"}
)

func pendingApprovals(at *v1alpha1.ApprovalTask) int {
	// Count unique users who have responded (approved or rejected)
	respondedUsers := make(map[string]bool)

	for _, approver := range at.Status.ApproversResponse {
		if v1alpha1.DefaultedApproverType(approver.Type) == "User" {
			respondedUsers[approver.Name] = true
		} else if v1alpha1.DefaultedApproverType(approver.Type) == "Group" {
			// Count individual group members who have responded
			for _, member := range approver.GroupMembers {
				if member.Response == "approved" || member.Response == "rejected" {
					respondedUsers[member.Name] = true
				}
			}
		}
	}

	return at.Spec.NumberOfApprovalsRequired - len(respondedUsers)
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

// UserGroupInfo holds information about a user's group memberships and responses
type UserGroupInfo struct {
	Groups    []string
	GroupsStr string
	Response  string
	Message   string
}

// userGroups processes ApproversResponse to group users by name across multiple groups
func userGroups(approversResponse []v1alpha1.ApproverState) map[string]UserGroupInfo {
	userMap := make(map[string]UserGroupInfo)
	
	// Process group members
	for _, approver := range approversResponse {
		if approver.Type == "Group" {
			for _, member := range approver.GroupMembers {
				if existing, exists := userMap[member.Name]; exists {
					// User already exists, add this group to their list
					existing.Groups = append(existing.Groups, approver.Name)
					userMap[member.Name] = existing
				} else {
					// New user, create entry
					userMap[member.Name] = UserGroupInfo{
						Groups:   []string{approver.Name},
						Response: member.Response,
						Message:  member.Message,
					}
				}
			}
		}
	}
	
	// Create comma-separated group strings
	for userName, info := range userMap {
		groupsStr := ""
		for i, group := range info.Groups {
			if i > 0 {
				groupsStr += ", "
			}
			groupsStr += group
		}
		info.GroupsStr = groupsStr
		userMap[userName] = info
	}
	
	return userMap
}

func Command(p cli.Params) *cobra.Command {
	opts := &cli.Options{}

	funcMap := template.FuncMap{
		"pipelineRunRef":   pipelineRunRef,
		"pendingApprovals": pendingApprovals,
		"message":          message,
		"response":         response,
		"state":            formatter.State,
		"userGroups":       userGroups,
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

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 8, 5, ' ', tabwriter.TabIndent)
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
