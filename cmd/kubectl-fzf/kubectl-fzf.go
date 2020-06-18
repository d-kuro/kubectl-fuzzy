package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/d-kuro/kubectl-fzf/pkg/cmd"
	"github.com/d-kuro/kubectl-fzf/pkg/signal"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-fzf", pflag.ExitOnError)
	pflag.CommandLine = flags

	ctx := context.Background()
	root := cmd.NewCmdRoot(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})

	done := make(chan struct{})
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer close(done)
		return root.ExecuteContext(ctx)
	})
	eg.Go(func() error {
		return signal.Handler(ctx, done)
	})

	if err := eg.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
