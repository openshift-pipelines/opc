Proof of concept of a single binary including pac and tkn functionalities

Usage: `go build -o tkno` 

```
$ ./tkno
CLI to manage Openshift Pipelines resources

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
  pac                   Manage Pipelines as Code resources
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

```
$ ./tkno pac --help
Manage your Pipelines as Code installation and resources
See https://pipelinesascode.com for more details

Usage:
tkn pac [command]

Available Commands:
  bootstrap   Bootstrap Pipelines as Code.
  completion  Prints shell completion scripts
  create      Create Pipelines as Code resources
  delete      Delete Pipelines as Code resources
  describe    Describe a repository
  generate    Generate PipelineRun
  list        List Pipelines as Code Repository
  logs        Display the PipelineRun logs from a Repository
  resolve     Embed PipelineRun references as a single resource.
  setup       Setup provider app or webhook
  version     Print tkn pac version

Available Plugins:
  watch

Flags:
  -h, --help                help for pac
  -k, --kubeconfig string   Path to the kubeconfig file to use for CLI requests (default: /Users/chmouel/.kube/config.kind) (default "/Users/chmouel/.kube/config.kind")
  -n, --namespace string    If present, the namespace scope for this CLI request

Use "tkn pac [command] --help" for more information about a command.
```

### NOTES

Only add 18mb : 

```
% du $GOPATH/src/github.com/tektoncd/cli/bin/tkn
120M	tkn
% du tkno
138M	tkno
```

### TODO
* avoid tkn-pac plugin showing up in the plugin section
