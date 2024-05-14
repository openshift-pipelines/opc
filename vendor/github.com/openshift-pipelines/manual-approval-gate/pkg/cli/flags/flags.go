package flags

import (
	cli "github.com/openshift-pipelines/manual-approval-gate/pkg/cli"
	"github.com/spf13/cobra"
)

func AddOptions(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(
		"namespace", "n", "",
		"namespace to use (default: from $KUBECONFIG)")
}

func PersistentPreRunE(p cli.Params) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		return InitParams(p, cmd)
	}
}

func InitParams(p cli.Params, cmd *cobra.Command) error {
	ns, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}
	if ns != "" {
		p.SetNamespace(ns)
	}

	return nil
}
