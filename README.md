# OPC - A CLI for OpenShift Pipeline

The OPC project merge multiple upstream CLI and specific cli features for
OpenShift Pipelines.

It contains :

- TektonCD CLI (tkn) - <https://github.com/tektoncd/cli>
- Pipelines as Code CLI (tkn-pac) - <https://pipelinesascode.com/docs/guide/cli/>

## Build

Use the default target of the Makefile:

i.e:

`make`

## Usage

Same as tkn with the addition of the pac command which redirect to tkn-pac.

## Features

Support completion :

`opc completions [bash|zsh|...]`

Plugins :

`opc foo will resolve to opc-foo`

### TODO

- Versioning are a bit all over the place

### NOTES

Only add 18mb :

```
% du $GOPATH/src/github.com/tektoncd/cli/bin/tkn
120M tkn
% du opc
138M opc
```
