# kubectl-fzf

Fuzzy Finder kubectl!

## Summary

`kubectl-fzf` is a kubectl plugin providing a fuzzy finder selector.
Uses [go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder), so there is no dependency on fzf binaries or anything else.

> 📝 Notes
>
> kubectl >= v1.12.0 is required for plugins to work.
>
> For more information on kuberctl plugins see [documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)

## Usage

```shell
kubectl fzf -h
Fuzzy Finder kubectl

Usage:
  kubectl-fzf [flags]
  kubectl-fzf [command]

Available Commands:
  help        Help about any command
  logs        Selecting a Pod with the fuzzy finder and view the log
  version     Show version

Flags:
  -h, --help   help for kubectl-fzf

Use "kubectl-fzf [command] --help" for more information about a command.
```

## Support

* [x] `kubectl logs`
* [ ] `kubectl exec`
* anything else...
