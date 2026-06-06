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

	wal, err := Open("example.hg.log")
	if err != nil {
		return err
	}
	defer func() {
		_ = wal.Close()
	}()

	fmt.Println("--- Replaying WAL ---")

	if err := wal.Replay(func(e *entry) error {
		fmt.Printf("%d: %s\n", e.SequenceNumber, string(e.Data))
		return nil
	}); err != nil {
		return err
	}

	fmt.Println("--- Writing time to WAL ---")

	if err := wal.Log([]byte(time.Now().Format(time.RFC1123))); err != nil {
		return err
	}

	return nil
}
