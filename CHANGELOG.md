# Changelog

## v1.9.0

### Enhancement

* Bump Go to v1.16 #162
* Bump Kubernetes library to v0.21.0
* Added darwin/arm64 release build

### Bug Fix

* Fix unsupported printer type error #182

#### Fix unsupported printer type error

A bug existed that caused error messages to appear in the preview window due to a Kubernetes upgrade.

> error: unsupported printer type: *printers.OmitManagedFieldsPrinter

Fixed a bug so that the preview window is displayed properly.

Also, `.metadata.managedFields` field is now omitted even when `--raw-preview` is enabled.

## v1.8.1

### Enhancement

None

#### 64-bit ARM build binary released

Added ARM builds binaries to the release.
Currently, only Linux is supported.

Please wait for the 64-bit ARM MacOS build for Apple Silicon to be supported in Go 1.16.

https://github.com/golang/go/blob/869e2957b9f66021581b839cadce6cb48ad46114/doc/go1.16.html#L32-L50

## v1.8.0

### Enhancement

* Add env var to enable preview across all commands #86

#### Add env var to enable preview across all commands

Add environment variables about enable preview mode.
You can use `KUBE_FUZZY_PREVIEW_ENABLED=true`.

## v1.7.0

### Enhancement

* Support delete command #81

#### Support delete command

Added support for the `kubectl delete` command with fuzzy finder selector.
You can use the `kubectl fuzzy delete` command.

## v1.6.1

### Bug Fix

* Fix help message #75

## v1.6.0

### Enhancement

* Support global flags #68

#### Support global flags

Show the common command flags as global flags.

## v1.5.2

### Fix Bugs

* fix: no auth provider error (#50) #51

#### fix: no auth provider error

Added import of the auth plugin.

```go
//  import the auth plugin package
_ "k8s.io/client-go/plugin/pkg/client/auth"
```

refs: https://github.com/kubernetes/client-go/issues/242

## v1.5.1

### Fix Bugs

* Add no resource found error handling #48

#### Add no resource found error handling

Outputs an error message if the resource does not exist.

```console
$ kubectl fuzzy describe pod -n default
no resources found
exit status 1
```

## v1.5.0

### Enhancement

* Support label selector for exec command #43

#### Support label selector for exec command

```diff
  Flags:
+ -l, --selector string                Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
```

for example:

```console
$ kubectl fuzzy exec -it --preview -l app=nginx -- bash
```

## v1.4.0

### Enhancement

* Add create job command to create job from cronjob #32

#### Create Job command

Added support for the `kubectl create job --from=cronjob` command with fuzzy finder selector.
You can use the `kubectl fuzzy create job --from=cronjob` command.
Currently, `kubectl fuzzy create job` is `--from=cronjob` option required. Job creation without a CronJob is not supported.

> 📝 See the [documentation](https://github.com/d-kuro/kubectl-fuzzy/blob/v1.4.0/docs/commands.md#create) for details.

```console
$ kubectl fuzzy create job --from=cronjob -P
job.batch/test-cronjob-kb87r created
```

## v1.3.0

### Enhancement

* Support multi resource type #33

#### Support multi resource type

`kubectl fuzzy describe` command now allows multiple resource types to be entered.

It is displayed with the following rules:

```go
"%s/%s (%s)", Lower(GroupKind), Name, Namespace
```

for example:

```
kubectl fuzzy describe service,configmap -n argocd
```

## v1.2.0

### Enhancement

* Simplify display the object #30

#### Simplify Display Object

Added ability to simplify Kubernetes objects displayed in the preview window.
Simplified by default. Some metadata and statuses have been removed.
Use the `--raw-preview` option to display the unsimplified object.

## v1.1.0

### Enhancement

* Add describe command #24
* Add preview option for logs and exec command #25

#### Describe Command

Added support for the `kubectl describe` command with fuzzy finder selector.
You can use the `kubectl fuzzy describe` command.

#### Preview Window

You can use the `--preview` or `-P` option to display a YAML of the Kubernetes object in a fuzzy finder selector.
You can switch display to YAML or JSON with the `--preview-format` option.

## v1.0.1

### Bug Fix

* fix project name for command example (#17)

## v1.0.0

### Breaking Change

* rename project to kubectl-fuzzy (#15)

#### Renamed this project from `kubectl-fzf` to `kubectl-fuzzy`.

The name `fzf` was not appropriate.
I'm sorry for the breaking changes.

There are no feature changes in this release.

After this release I'm going add the project to the krew index.
This should make it easier to install and so on.

https://github.com/kubernetes-sigs/krew-index/pull/660

## v0.1.0

### Enhancement

* Support kubectl exec command (#14)

## v0.0.2

### Bug Fix

* Fixed the error message displayed twice(https://github.com/d-kuro/kubectl-fzf/pull/12)

## v0.0.1

### Enhancement

* Support for `kubectl logs` command
