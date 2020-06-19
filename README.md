# kubectl-fzf

Fuzzy Finder kubectl!

[![asciicast](https://asciinema.org/a/kMNLBIDAGLaNl6JcgJnUACCUr.svg)](https://asciinema.org/a/kMNLBIDAGLaNl6JcgJnUACCUr)

## Summary

`kubectl-fzf` is a kubectl plugin providing a fuzzy finder selector.
Uses [go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder), so there is no dependency on fzf binaries or anything else.

> 📝 Notes
>
> kubectl >= v1.12.0 is required for plugins to work.
>
> For more information on kuberctl plugins see [documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)

## Install

```shell
git clone https://github.com/d-kuro/kubectl-fzf.git
cd kubectl-fzf
make install
```

or

Please download the binaries from the [release page](https://github.com/d-kuro/kubectl-fzf/releases).

## Usage

```console
$ kubectl fzf -h
Fuzzy Finder kubectl

Usage:
  kubectl-fzf [flags]
  kubectl-fzf [command]

Available Commands:
  exec        Selecting a Pod with the fuzzy finder and execute a command in a container
  help        Help about any command
  logs        Selecting a Pod with the fuzzy finder and view the log
  version     Show version

Flags:
  -h, --help   help for kubectl-fzf

Use "kubectl-fzf [command] --help" for more information about a command.
```

## Support

* [x] `kubectl logs`
* [x] `kubectl exec`
* anything else...
