# opc - A CLI for OpenShift Pipeline

`opc` make it easy to work with Tekton resources in OpenShift Pipelines. It is built on top of `tkn` and `tkn-pac` and expands their capablities to the functionality and user-experience that is available on OpenShift. 

## Build

Use the default target of the Makefile:

i.e:

`make`

## Useful commands

The following commands help you understand and effectively use the OpenShift Pipelines CLI:

`opc hub`: search and install from Tekton Hub

`opc pac`: add and manage git repositories (pipelines as code)

`opc pipeline`: manage Pipelines
`opc pipelinerun`: manage PipelineRuns
`opc task`: manage Tasks
`opc clustertask`: manage ClusterTasks
`opc taskrun`: manage TaskRuns

`opc triggerbinding`: manage TriggerBindings
`opc clustertriggerbinding`: manage ClusterTriggerBindings
`opc triggertemplate`: manage TriggerTemplates
`opc eventlistener`: manage EventListeners

## Features

Support completion :

`opc completions [bash|zsh|...]`

Plugins :

opc shows tkn plugins, it doesn't try to show opc plugins. (may change).

### TODO

- Versioning are a bit all over the place

### NOTES

Only add 18mb :

```shell
% du $GOPATH/src/github.com/tektoncd/cli/bin/tkn
120M tkn
% du opc
138M opc
```
