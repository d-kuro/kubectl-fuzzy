# Support Commands

* [kubectl create](#create)
* [kubectl delete](#delete)
* [kubectl describe](#describe)
* [kubectl logs](#logs)
* [kubectl exec](#exec)

## Create

Compatibility commands with `kubectl create`.

Available Commands:

* `kubectl fuzzy create job`

## Delete

Compatibility commands with `kubectl delete`.

Usage:

```console
$ kubectl fuzzy delete TYPE [flags]
```

Helps:

<details>

```console
$ kubectl fuzzy delete -h
Selecting an object with the fuzzy finder and delete

Usage:
  kubectl-fuzzy delete [flags]

Examples:

	# Selecting an object with the fuzzy finder and delete
	kubectl fuzzy delete TYPE [flags]


Flags:
  -A, --all-namespaces                 If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.
      --cascade                        If true, cascade the deletion of the resources managed by this resource (e.g. Pods created by a ReplicationController). Default true. (default true)
      --dry-run string[="unchanged"]   Must be "none", "server", or "client". If client strategy, only print the object that would be sent, without sending it. If server strategy, submit server-side request without persisting the resource. (default "none")
      --field-selector string          Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2).The server only supports a limited number of field queries per type.
      --force                          If true, immediately remove resources from API and bypass graceful deletion. Note that immediate deletion of some resources may result in inconsistency or data loss and requires confirmation.
      --grace-period int               Period of time in seconds given to the resource to terminate gracefully. Ignored if negative. Set to 1 for immediate shutdown. Can only be set to 0 when --force is true (force deletion). (default -1)
  -h, --help                           help for delete
      --now                            If true, resources are signaled for immediate shutdown (same as --grace-period=1).
  -o, --output string                  Output mode. Use "-o name" for shorter output (resource/name).
  -P, --preview                        If true, display the object YAML|JSON by preview window for fuzzy finder selector.
      --preview-format string          Preview window output format. One of json|yaml. (default "yaml")
      --raw-preview                    If true, display the unsimplified object in the preview window. (default is simplified)
  -l, --selector string                Selector (label query) to filter on, not including uninitialized ones.
      --timeout duration               The length of time to wait before giving up on a delete, zero means determine a timeout from the size of the object
      --wait                           If true, wait for resources to be gone before returning. This waits for finalizers. (default true)

Global Flags:
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default cache directory (default "/Users/d-kuro/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
```

</details>

### Job

Compatibility commands with `kubectl create job --from=cronjob`.

> 📝 TODO
>
> Currently, `kubectl fuzzy create job` is `--from=cronjob` option required.
> Job creation without a CronJob is not supported.

Usage:

```console
$ kubectl fuzzy create job [jobName] --from=cronjob [flags]
```

Helps:

<details>

```console
$ kubectl fuzzy create job -h
Selecting a CronJob with the fuzzy finder and create job

Usage:
  kubectl-fuzzy create job [NAME] --from=cronjob [flags]

Examples:

	# Selecting a CronJob with the fuzzy finder and create job
	# Only supported cronjob
	# If a jobName is omitted, generated from cronJob name
	kubectl fuzzy create job [jobName] --from=cronjob [flags]


Flags:
      --from string             The name of the resource to create a Job from (only cronjob is supported).
  -h, --help                    help for job
  -P, --preview                 If true, display the object YAML|JSON by preview window for fuzzy finder selector.
      --preview-format string   Preview window output format. One of json|yaml. (default "yaml")
      --raw-preview             If true, display the unsimplified object in the preview window. (default is simplified)

Global Flags:
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default cache directory (default "/Users/d-kuro/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
```

</details>

## Describe

Compatibility commands with `kubectl describe`.

Usage:

```console
$ kubectl fuzzy describe TYPE [flags]
```

Helps:

<details>

```console
$ kubectl fuzzy describe -h
Selecting an object with the fuzzy finder and show details

Usage:
  kubectl-fuzzy describe [flags]

Examples:

	# Selecting an object with the fuzzy finder and view the log and show details
	kubectl fuzzy describe TYPE [flags]


Flags:
  -A, --all-namespaces          If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.
  -h, --help                    help for describe
  -P, --preview                 If true, display the object YAML|JSON by preview window for fuzzy finder selector.
      --preview-format string   Preview window output format. One of json|yaml. (default "yaml")
      --raw-preview             If true, display the unsimplified object in the preview window. (default is simplified)
  -l, --selector string         Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
      --show-events             If true, display events related to the described object. (default true)

Global Flags:
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default cache directory (default "/Users/d-kuro/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
```

</details>

## Logs

Compatibility commands with `kubectl logs`.

> 📝 TODO
>
> Currently, `kubectl fuzzy logs` only supports Pod selection.
> Does not currently support `--selector` options and resource selection such as Deployment.

Usage:

```console
$ kubectl fuzzy logs [flags]
```

Helps:

<details>

```console
$ kubectl fuzzy logs -h
Selecting a Pod with the fuzzy finder and view the log

Usage:
  kubectl-fuzzy logs [flags]

Examples:

	# Selecting a Pod with the fuzzy finder and view the log
	kubectl fuzzy logs [flags]


Flags:
  -A, --all-namespaces          If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.
  -f, --follow                  Specify if the logs should be streamed.
  -h, --help                    help for logs
      --limit-bytes int         Maximum bytes of logs to return. Defaults to no limit.
  -P, --preview                 If true, display the object YAML|JSON by preview window for fuzzy finder selector.
      --preview-format string   Preview window output format. One of json|yaml. (default "yaml")
  -p, --previous                If true, print the logs for the previous instance of the container in a pod if it exists.
      --raw-preview             If true, display the unsimplified object in the preview window. (default is simplified)
      --since duration          Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.
      --since-time string       Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.
      --tail int                Lines of recent log file to display. Defaults to -1 with no selector, showing all log lines otherwise 10, if a selector is provided. (default -1)
      --timestamps              Include timestamps on each line in the log output.

Global Flags:
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default cache directory (default "/Users/d-kuro/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
```

</details>

## Exec

Compatibility commands with `kubectl exec`.

> 📝 TODO
>
> Currently, `kubectl fuzzy exec` only supports Pod selection.

Usage:

```console
$ kubectl fuzzy exec [flags]
```

Helps:

<details>

```console
$ kubectl fuzzy exec -h
Selecting a Pod with the fuzzy finder and execute a command in a container

Usage:
  kubectl-fuzzy exec [flags]

Examples:

	# Selecting a Pod with the fuzzy finder and execute a command in a container
	kubectl fuzzy exec [flags] -- COMMAND [args...]


Flags:
  -A, --all-namespaces          If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.
  -h, --help                    help for exec
  -P, --preview                 If true, display the object YAML|JSON by preview window for fuzzy finder selector.
      --preview-format string   Preview window output format. One of json|yaml. (default "yaml")
      --raw-preview             If true, display the unsimplified object in the preview window. (default is simplified)
  -l, --selector string         Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
  -i, --stdin                   Pass stdin to the container
  -t, --tty                     Stdin is a TTY

Global Flags:
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default cache directory (default "/Users/d-kuro/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
```

</details>
