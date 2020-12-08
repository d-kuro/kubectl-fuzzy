module github.com/d-kuro/kubectl-fuzzy

go 1.15

require (
	github.com/ktr0731/go-fuzzyfinder v0.2.1
	github.com/moby/term v0.0.0-20201110203204-bea5bbe245bf
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/tidwall/sjson v1.1.2
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	k8s.io/api v0.20.0
	k8s.io/apimachinery v0.20.0
	k8s.io/cli-runtime v0.19.4
	k8s.io/client-go v0.19.4
	k8s.io/klog/v2 v2.4.0
	k8s.io/kubectl v0.19.4
	sigs.k8s.io/yaml v1.2.0
)
