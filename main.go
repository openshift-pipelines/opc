package main

import (
	"os"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/cmd/tknpac"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/params"
	"github.com/tektoncd/cli/pkg/cli"
	"github.com/tektoncd/cli/pkg/cmd"
)

func main() {
	tp := &cli.TektonParams{}
	tkn := cmd.Root(tp)
	clients := params.New()
	pac := tknpac.Root(clients)
	pac.Use = "pac"
	tkn.AddCommand(pac)

	if err := tkn.Execute(); err != nil {
		os.Exit(1)
	}
}
