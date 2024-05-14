package opc

import (
	"encoding/json"
	"fmt"
	"html/template"

	_ "embed"

	paccli "github.com/openshift-pipelines/pipelines-as-code/pkg/cli"

	// paccli "github.com/openshift-pipelines/opc/pkg"
	"github.com/spf13/cobra"
)

//go:embed version.json
var versionFile string

//go:embed version.tmpl
var versionTmpl string

type versions struct {
	Opc                string `json:"opc"`
	Tkn                string `json:"tkn"`
	Pac                string `json:"pac"`
	Results            string `json:"results"`
	ManualApprovalGate string `json:"manualapprovalgate"`
}

func VersionCommand(ioStreams *paccli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print opc version",
		Long:  "Print OpenShift Pipeline Client version",
		RunE: func(_ *cobra.Command, args []string) error {
			var v versions
			if err := json.Unmarshal([]byte(versionFile), &v); err != nil {
				return fmt.Errorf("cannot unmarshall versions: %w", err)
			}
			if len(args) > 1 {
				switch args[1] {
				case "pac":
					fmt.Fprintln(ioStreams.Out, v.Pac)
				case "tkn":
					fmt.Fprintln(ioStreams.Out, v.Tkn)
				case "opc":
					fmt.Fprintln(ioStreams.Out, v.Opc)
				case "results":
					fmt.Fprintln(ioStreams.Out, v.Results)
				case "manualapprovalgate":
					fmt.Fprintln(ioStreams.Out, v.ManualApprovalGate)
				default:
					return fmt.Errorf("unknown component: %v", args[1])
				}
				return nil
			}

			t, err := template.New("Describe Repository").Parse(versionTmpl)
			if err != nil {
				return err
			}
			return t.Execute(ioStreams.Out, v)
		},
		Annotations: map[string]string{
			"commandType": "main",
		},
	}
	return cmd
}
