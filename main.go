package main

import (
	"fmt"
	"os"
	"syscall"

	paccli "github.com/openshift-pipelines/pipelines-as-code/pkg/cli"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/cmd/tknpac"
	pacversion "github.com/openshift-pipelines/pipelines-as-code/pkg/cmd/tknpac/version"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/params"
	"github.com/spf13/cobra"
	"github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/cmd"

	"github.com/tektoncd/cli/pkg/plugins"
)

const (
	pacLongDesc = `Manage your Pipelines as Code installation and resources
See https://pipelinesascode.com for more details`
	pacShortdesc = "Manage Pipelines as Code resources"
	tknShortDesc = `CLI to manage Openshift Pipelines resources`
	binaryName   = `opc`
)

func main() {
	tp := &cli.TektonParams{}
	tkn := cmd.Root(tp)
	tkn.Use = binaryName
	tkn.Short = tknShortDesc
	clients := params.New()
	pac := tknpac.Root(clients)
	pac.Use = "pac"
	pac.Short = pacShortdesc
	pac.Long = pacLongDesc
	tkn.AddCommand(pac)
	pluginList := plugins.GetAllTknPluginFromPaths()
	newPluginList := []string{}
	// remove pac from the plugin list
	for _, value := range pluginList {
		if value != "pac" {
			newPluginList = append(newPluginList, value)
		}
	}
	cobra.AddTemplateFunc("pluginList", func() []string { return newPluginList })
	paciostreams := paccli.NewIOStreams()
	tkn.RemoveCommand(pacversion.Command(paciostreams))

	args := os.Args[1:]
	cmd, _, _ := tkn.Find(args)

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
	if err := tkn.Execute(); err != nil {
		os.Exit(1)
	}
}
