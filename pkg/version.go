package opc

import (
	"encoding/json"
	"fmt"

	_ "embed"

	tkncli "github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/version"

	// paccli "github.com/openshift-pipelines/opc/pkg"
	"github.com/spf13/cobra"
)

//go:embed version.json
var versionFile string

var (
	component = ""
	namespace = ""
	err       error
)

type versions struct {
	Opc string `json:"opc"`
	Tkn string `json:"tkn"`
	Pac string `json:"pac"`
}

func VersionCommand(tp tkncli.Params) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print opc version",
		Long:  "Print OpenShift Pipeline Client version",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			component, err = cmd.Flags().GetString("component")
			if err != nil {
				return err
			}
			namespace, err = cmd.Flags().GetString("namespace")
			if err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var v versions
			if err := json.Unmarshal([]byte(versionFile), &v); err != nil {
				return fmt.Errorf("cannot unmarshall versions: %w", err)
			}
			if args[0] == "pac" {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\n", v.Opc)
				return nil
			}

			cs, err := tp.Clients()
			if err == nil {
				switch component {
				case "":
					fmt.Fprintf(cmd.OutOrStdout(), "OpenShift Pipelines Client Verion: %s\n", v.Opc)
					fmt.Fprintf(cmd.OutOrStdout(), "Tekton Client Verion: %s\n", v.Tkn)
					fmt.Fprintf(cmd.OutOrStdout(), "Pipelines as Code CLI Verion: %s\n", v.Tkn)
					chainsVersion, _ := version.GetChainsVersion(cs, namespace)
					if chainsVersion != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "Chains version: %s\n", chainsVersion)
					}

					pipelineVersion, _ := version.GetPipelineVersion(cs, namespace)
					if pipelineVersion == "" {
						pipelineVersion = "unknown, " +
							"pipeline controller may be installed in another namespace please use tkn version -n {namespace}"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "Pipeline version: %s\n", pipelineVersion)

					triggersVersion, _ := version.GetTriggerVersion(cs, namespace)
					if triggersVersion != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "Triggers version: %s\n", triggersVersion)
					}

					dashboardVersion, _ := version.GetDashboardVersion(cs, namespace)
					if dashboardVersion != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "Dashboard version: %s\n", dashboardVersion)
					}

					operatorVersion, _ := version.GetOperatorVersion(cs, namespace)
					if operatorVersion != "" {
						fmt.Fprintf(cmd.OutOrStdout(), "Operator version: %s\n", operatorVersion)
					}
				case "tkn":
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", v.Tkn)
				case "client":
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", v.Opc)
				case "chains":
					chainsVersion, _ := version.GetChainsVersion(cs, namespace)
					if chainsVersion == "" {
						chainsVersion = "unknown"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", chainsVersion)

				case "pipeline":
					pipelineVersion, _ := version.GetPipelineVersion(cs, namespace)
					if pipelineVersion == "" {
						pipelineVersion = "unknown"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", pipelineVersion)

				case "triggers":
					triggersVersion, _ := version.GetTriggerVersion(cs, namespace)
					if triggersVersion == "" {
						triggersVersion = "unknown"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", triggersVersion)

				case "dashboard":
					dashboardVersion, _ := version.GetDashboardVersion(cs, namespace)
					if dashboardVersion == "" {
						dashboardVersion = "unknown"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", dashboardVersion)

				case "operator":
					operatorVersion, _ := version.GetOperatorVersion(cs, namespace)
					if operatorVersion == "" {
						operatorVersion = "unknown"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\n", operatorVersion)

				default:
					fmt.Fprintf(cmd.OutOrStdout(), "Invalid component value\n")
				}
			}
			return nil
		},
		Annotations: map[string]string{
			"commandType": "main",
		},
	}
	cmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "namespace to check installed controller version")
	cmd.Flags().StringVarP(&component, "component", "c", "", "provide a particular component name for its version (client|tkn|chains|pipeline|triggers|dashboard)")
	return cmd
}
