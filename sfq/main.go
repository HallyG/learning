package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/HallyG/learning/sfq/sfq"
)

type Options struct {
}

func (opts *Options) RegisterFlags(flags *flag.FlagSet) {

}

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, output io.Writer) error {
	_, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	flags := flag.NewFlagSet("wal", flag.ContinueOnError)
	flags.SetOutput(output)

	var opts Options
	opts.RegisterFlags(flags)
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}

		return fmt.Errorf("parsing flags: %w", err)
	}

	queue := sfq.NewSliceQueue[*sfq.Request](5)
	queue2 := sfq.NewSliceQueue[*sfq.Request](5)

	scheduler := sfq.NewSFQScheduler(queue, queue2)
	scheduler.Perturb()

	requests := 10
	for i := range requests {
		req := &sfq.Request{
			ID: fmt.Sprintf("%04d", i),
		}
		scheduler.Enqueue(req)
	}

	for range requests {
		fmt.Println(scheduler.Dequeue())
	}

	return nil
}
