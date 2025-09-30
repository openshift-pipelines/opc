package opc

import (
	"encoding/json"
	"fmt"
	"html/template"

	_ "embed"

	paccli "github.com/openshift-pipelines/pipelines-as-code/pkg/cli"
	tkncli "github.com/tektoncd/cli/pkg/cli"
	tknversion "github.com/tektoncd/cli/pkg/version"

	"github.com/spf13/cobra"
)

var serverFlag = "server"

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
	Assist             string `json:"assist"`
}

func getLiveInformations(iostreams *paccli.IOStreams) error {
	tp := &tkncli.TektonParams{}
	cs, err := tp.Clients()
	if err != nil {
		return err
	}
	namespace := "openshift-pipelines"
	operatorNamespace := "openshift-operators"

	chainsVersion, _ := tknversion.GetChainsVersion(cs, namespace)
	if chainsVersion != "" {
		fmt.Fprintf(iostreams.Out, "Chains version: %s\n", chainsVersion)
	}
	pipelineVersion, _ := tknversion.GetPipelineVersion(cs, namespace)
	if pipelineVersion == "" {
		pipelineVersion = "unknown, " +
			"pipeline controller may be installed in another namespace."
	}
	fmt.Fprintf(iostreams.Out, "Pipeline version: %s\n", pipelineVersion)
	triggersVersion, _ := tknversion.GetTriggerVersion(cs, namespace)
	if triggersVersion != "" {
		fmt.Fprintf(iostreams.Out, "Triggers version: %s\n", triggersVersion)
	}
	operatorVersion, _ := tknversion.GetOperatorVersion(cs, operatorNamespace)
	if operatorVersion != "" {
		fmt.Fprintf(iostreams.Out, "Operator version: %s\n", operatorVersion)
	}
	hubVersion, _ := tknversion.GetHubVersion(cs, namespace)
	if hubVersion != "" {
		fmt.Fprintf(iostreams.Out, "Hub version: %s\n", hubVersion)
	}
	return nil
}

func VersionCommand(ioStreams *paccli.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print opc version",
		Long:  "Print OpenShift Pipeline Client version",
		RunE: func(cmd *cobra.Command, args []string) error {
			var v versions
			server, err := cmd.Flags().GetBool(serverFlag)
			if err != nil {
				return err
			}
			if server {
				// TODO(chmouel): pac version when it's  refactored in pac code
				return getLiveInformations(ioStreams)
			}
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
				case "assist":
					fmt.Fprintln(ioStreams.Out, v.Assist)
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

	cmd.Flags().BoolP(serverFlag, "s", false, "Get the services version information from cluster instead of the client version.")
	return cmd
}
