package cmd

import (
	cli "github.com/openshift-pipelines/manual-approval-gate/pkg/cli"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/cli/cmd/approve"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/cli/cmd/describe"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/cli/cmd/list"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/cli/cmd/reject"
	"github.com/openshift-pipelines/manual-approval-gate/pkg/cli/flags"
	"github.com/spf13/cobra"
)

func Root(p cli.Params) *cobra.Command {
	c := &cobra.Command{
		Use:   "tkn-approvaltask",
		Short: "Approval Task CLI",
		Long:  `tkn plugin to use approval task as CLI`,
		Annotations: map[string]string{
			"commandType": "main",
		},
		PersistentPreRunE: flags.PersistentPreRunE(p),
	}

	c.AddCommand(list.Command(p))
	c.AddCommand(approve.Command(p))
	c.AddCommand(describe.Command(p))
	c.AddCommand(reject.Command(p))

	return c
}
