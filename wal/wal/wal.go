package wal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var _ WAL = &simpleWAL{}

type (
	WAL interface {
		Log(data []byte) error
		Replay(fn func(*Entry) error) error
		Close() error
	}

	simpleWAL struct {
		file   *os.File
		writer *bufio.Writer

		nextSeqNum uint64

		mu sync.Mutex
	}
)

func NewFromFile(path string) (*simpleWAL, error) {
	return nil, nil
}

func Open(path string) (*simpleWAL, error) {
	f, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_RDWR, 0600) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("opening wal: %w", err)
	}

	w := &simpleWAL{
		file:   f,
		writer: bufio.NewWriter(f),
	}

	if err := w.Replay(func(e *Entry) error {
		w.nextSeqNum = e.SequenceNumber + 1
		return nil
	}); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("replaying wal: %w", err)
	}

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("seeking to file end: %w", err)
	}

	return w, nil
}

func (w *simpleWAL) Log(data []byte) error {
	if data == nil {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	entry := &Entry{
		Version:        1,
		SequenceNumber: w.nextSeqNum,
		Data:           data,
	}

	if err := entry.Encode(w.writer); err != nil {
		return fmt.Errorf("encoding wal entry: %w", err)
	}

	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("flushing wal: %w", err)
	}

	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("syncing wal: %w", err)
	}

	w.nextSeqNum++

	return nil
}

func (w *simpleWAL) Replay(fn func(*Entry) error) error {
	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seeking wal: %w", err)
	}

	return ReplayReader(w.file, fn)
}

func (w *simpleWAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("flushing writer: %w", err)
	}

	if err := w.file.Close(); err != nil {
		return fmt.Errorf("closing wal: %w", err)
	}

	return nil
}

func ReplayReader(r io.Reader, fn func(*Entry) error) error {
	br := bufio.NewReader(r)

	for {
		var ent Entry
		if err := ent.Decode(br); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("decoding entry: %w", err)
		}

		if err := fn(&ent); err != nil {
			return fmt.Errorf("processing entry: %w", err)
		}
	}
}
