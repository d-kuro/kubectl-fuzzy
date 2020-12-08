package main

import (
	"context"
	"fmt"
	"os"

	"github.com/d-kuro/kubectl-fuzzy/pkg/cmd"
	"github.com/d-kuro/kubectl-fuzzy/pkg/signal"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	//  import the auth plugin package
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	flags := pflag.NewFlagSet("kubectl-fuzzy", pflag.ExitOnError)
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
