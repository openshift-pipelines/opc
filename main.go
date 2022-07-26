package main

import (
	"os"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/cmd/tknpac"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params"
	"github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/cmd"
)

const (
	pacLongDesc = `Manage your Pipelines as Code installation and resources
See https://pipelinesascode.com for more details`
	pacShortdesc = "Manage Pipelines as Code resources"
	tknShortDesc = `CLI to manage Openshift Pipelines resources`
)

func main() {
	tp := &cli.TektonParams{}
	tkn := cmd.Root(tp)
	tkn.Short = tknShortDesc
	clients := params.New()
	pac := tknpac.Root(clients)
	pac.Use = "pac"
	pac.Short = pacShortdesc
	pac.Long = pacLongDesc
	tkn.AddCommand(pac)

	if err := tkn.Execute(); err != nil {
		os.Exit(1)
	}
}
