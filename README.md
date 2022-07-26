Proof of concept of a single binary including pac and tkn functionalities

Usage: `go build -o tkno` 

```
$ ./tkno
CLI for tekton pipelines

Usage:
tkn [flags]
tkn [command]

Available Commands:
  bundle                Manage Tekton Bundles
  chain                 Manage Chains
  clustertask           Manage ClusterTasks
  clustertriggerbinding Manage ClusterTriggerBindings
  eventlistener         Manage EventListeners
  hub                   Interact with tekton hub
  pac                   Pipelines as Code CLI
  pipeline              Manage pipelines
  pipelinerun           Manage PipelineRuns
  resource              Manage pipeline resources
  task                  Manage Tasks
  taskrun               Manage TaskRuns
  triggerbinding        Manage TriggerBindings
  triggertemplate       Manage TriggerTemplates

Other Commands:
  completion            Prints shell completion scripts
  version               Prints version information

Available Plugins:
  watch

Flags:
  -h, --help   help for tkn

Use "tkn [command] --help" for more information about a command.
```

TODO: avoid tkn-pac plugin showing up in the plugin section
