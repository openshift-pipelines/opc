package opc

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	_ "embed"

	paccli "github.com/openshift-pipelines/pipelines-as-code/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/cli"
	tkncli "github.com/tektoncd/cli/pkg/cli"
	tknversion "github.com/tektoncd/cli/pkg/version"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var serverFlag = "server"

const operatorInfo string = "tekton-operator-info"

//go:embed version.json
var versionFile string

//go:embed version.tmpl
var versionTmpl string

var defaultNamespaces = []string{"tekton-pipelines", "openshift-pipelines", "tekton-chains", "tekton-operator", "openshift-operators"}

type versions struct {
	Opc                      string `json:"opc"`
	Tkn                      string `json:"tkn"`
	Pac                      string `json:"pac"`
	Results                  string `json:"results"`
	ManualApprovalGate       string `json:"manualapprovalgate"`
	OpenShiftPipelines string `json:"openshiftpipelines"`
}

func getConfigMap(c *cli.Clients, name, ns string) (*corev1.ConfigMap, error) {

	var (
		err       error
		configMap *corev1.ConfigMap
	)

	if ns != "" {
		configMap, err = c.Kube.CoreV1().ConfigMaps(ns).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		return configMap, nil
	}

	for _, n := range defaultNamespaces {
		configMap, err = c.Kube.CoreV1().ConfigMaps(n).Get(context.Background(), name, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				continue
			}
			if strings.Contains(err.Error(), fmt.Sprintf(`cannot get resource "configmaps" in API group "" in the namespace "%s"`, n)) {
				continue
			}
			return nil, err
		}
		if configMap != nil {
			break
		}
	}

	if configMap == nil {
		return nil, fmt.Errorf("ConfigMap with name %s not found in the namespace %s", name, ns)
	}
	return configMap, nil
}

func GetRedHatOpenShiftPipelinesVersion(c *cli.Clients, ns string) (string, error) {
	configMap, err := getConfigMap(c, operatorInfo, ns)
	if err != nil {
		return "", nil // Not found or inaccessible, return no version
	}

	// 1. Check for a dedicated "product" field
	if product, exists := configMap.Data["product"]; exists && product != "" {
		return product, nil
	}

	// 2. Check for embedded version in the "version" field
	if version, exists := configMap.Data["version"]; exists && version != "" {
		if strings.Contains(version, "(") && strings.Contains(version, ")") {
			parts := strings.SplitN(version, "(", 2)
			if len(parts) > 1 {
				productVersion := strings.TrimSuffix(strings.TrimSpace(parts[1]), ")")
				if strings.HasPrefix(productVersion, "Red Hat OpenShift Pipelines") {
					productVersion = strings.TrimSpace(
						strings.TrimPrefix(productVersion, "Red Hat OpenShift Pipelines"),
					)
					return productVersion, nil
				}
				return productVersion, nil
			}
		}
	}

	// 3. Fallback to "rhProduct" field
	if rhProduct, exists := configMap.Data["rhProduct"]; exists && rhProduct != "" {
		return rhProduct, nil
	}

	return "", nil
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
	productVersion, _ := GetRedHatOpenShiftPipelinesVersion(cs, namespace)
	if productVersion != "" {
		fmt.Fprintf(iostreams.Out, "OpenShift Pipelines: %s\n", productVersion)
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
				case "openShiftpipelines":
					fmt.Fprintln(ioStreams.Out, v.OpenShiftPipelines)
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
