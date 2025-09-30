package main

import (
	"fmt"
	"os"
	"slices"
	"syscall"

	magcli "github.com/openshift-pipelines/manual-approval-gate/pkg/cli"
	magcmd "github.com/openshift-pipelines/manual-approval-gate/pkg/cli/cmd"
	opccli "github.com/openshift-pipelines/opc/pkg"
	paccli "github.com/openshift-pipelines/pipelines-as-code/pkg/cli"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/cmd/tknpac"
	pacversion "github.com/openshift-pipelines/pipelines-as-code/pkg/cmd/tknpac/version"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params"
	assistcli "github.com/openshift-pipelines/tekton-assist/pkg/cli"
	"github.com/spf13/cobra"
	tkncli "github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/cmd"
	tknversion "github.com/tektoncd/cli/pkg/cmd/version"
	"github.com/tektoncd/cli/pkg/plugins"
	resultscmd "github.com/tektoncd/results/pkg/cli/cmd"
	resultscommon "github.com/tektoncd/results/pkg/cli/common"
)

const (
	pacLongDesc = `Manage your Pipelines as Code installation and resources
See https://pipelinesascode.com for more details`
	pacShortdesc     = "CLI to interact with Pipelines as Code resources"
	tknShortDesc     = "CLI to interact with Openshift Pipelines resources"
	resultsShortDesc = "CLI to interact with Tekton Results API."
	magShortDesc     = "CLI to interact with Manual Approval Gate."
	assistShortDesc  = "CLI to analyze and diagnose Tekton failures with AI assistance."
	binaryName       = `opc`
)

func main() {
	tp := &tkncli.TektonParams{}
	tkn := cmd.Root(tp)
	tkn.Use = binaryName
	tkn.Short = tknShortDesc
	clients := params.New()
	pac := tknpac.Root(clients)
	pac.Use = "pac"
	pac.Short = pacShortdesc
	pac.Long = pacLongDesc
	tkn.AddCommand(pac)

	// adding manual approval gate cli
	p := &magcli.ApprovalTaskParams{}
	mag := magcmd.Root(p)
	mag.Use = "approvaltask"
	mag.Short = magShortDesc
	tkn.AddCommand(mag)

	// adding results
	rp := &resultscommon.ResultsParams{}
	results := resultscmd.Root(rp)
	results.Use = "results"
	results.Short = resultsShortDesc
	tkn.AddCommand(results)

	// adding tekton assist
	assist := assistcli.RootCommand()
	assist.Use = "assist"
	assist.Short = assistShortDesc
	tkn.AddCommand(assist)

	pluginList := plugins.GetAllTknPluginFromPaths()
	newPluginList := []string{}
	// remove integrated commands from the plugin list
	excludedPlugins := []string{"pac", "assist", "results"}
	for _, value := range pluginList {
		if !slices.Contains(excludedPlugins, value) {
			newPluginList = append(newPluginList, value)
		}
	}
	cobra.AddTemplateFunc("pluginList", func() []string { return newPluginList })
	paciostreams := paccli.NewIOStreams()
	tkn.RemoveCommand(pacversion.Command(paciostreams)) // does not work 🤷🏽‍♂️, we do the hard way later on
	tkn.RemoveCommand(tknversion.Command(tp))
	tkn.AddCommand(opccli.VersionCommand(paciostreams))

	args := os.Args[1:]
	var cmd *cobra.Command
	if len(args) > 0 && args[0] == "version" {
		cmd = opccli.VersionCommand(paciostreams)
		goto CoreTkn
	}
	if len(args) > 1 && args[0] == "pac" && args[1] == "version" {
		// Arthur: "I've Got Nothing Left To Lose. Nothing Can Hurt Me Anymore. My Life Is Nothing But A Comedy." 🃏
		os.Args = []string{"version"}
		vcmd := opccli.VersionCommand(paciostreams)
		_ = vcmd.Execute()
		os.Exit(0)
	}
	cmd, _, _ = tkn.Find(args)

	if cmd != nil && cmd == tkn && len(args) > 0 {
		exCmd, err := plugins.FindPlugin(os.Args[1])
		// if we can't find command then execute the normal tkn command.
		if err != nil {
			goto CoreTkn
		}

		// if we have found the plugin then sysexec it by replacing current process.
		if err := syscall.Exec(exCmd, append([]string{exCmd}, os.Args[2:]...), os.Environ()); err != nil {
			fmt.Fprintf(os.Stderr, "Command finished with error: %v", err)
			os.Exit(127)
		}
		return
	}

CoreTkn:
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
