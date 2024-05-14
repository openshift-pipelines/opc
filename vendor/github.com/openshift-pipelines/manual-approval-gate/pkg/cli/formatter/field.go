package formatter

import (
	"github.com/fatih/color"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/apis/approvaltask/v1alpha1"
)

var ConditionColor = map[string]color.Attribute{
	"Rejected": color.FgHiRed,
	"Approved": color.FgHiGreen,
	"Pending":  color.FgHiYellow,
}

func ColorStatus(status string) string {
	return color.New(ConditionColor[status]).Sprint(status)
}

func State(at *v1alpha1.ApprovalTask) string {
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
