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
	"time"
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

	limiter := NewRateLimiter(1, 2)

	fmt.Println(limiter.Allow("1")) // true
	fmt.Println(limiter.Allow("1")) // true
	fmt.Println(limiter.Allow("1")) // false

	fmt.Println("sleep")
	time.Sleep(time.Second)

	fmt.Println(limiter.Allow("1")) // true
	fmt.Println(limiter.Allow("1")) // false

	fmt.Println("sleep")
	time.Sleep(2 * time.Second)

	fmt.Println(limiter.Allow("1")) // true
	fmt.Println(limiter.Allow("1")) // true
	fmt.Println(limiter.Allow("1")) // false

	return nil
}
