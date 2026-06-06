package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"
)

var _ WAL = &simpleWAL{}

type (
	WAL interface {
		Log(data []byte) error
		Replay(fn func(*entry) error) error
		Close() error
	}

	simpleWAL struct {
		file   *os.File
		writer *bufio.Writer

		nextSeqNum uint64

		mu sync.Mutex
	}
)

func Open(path string) (*simpleWAL, error) {
	f, err := os.OpenFile(filepath.Clean(path), os.O_CREATE|os.O_RDWR, 0600) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("opening wal: %w", err)
	}

	w := &simpleWAL{
		file:   f,
		writer: bufio.NewWriter(f),
	}

	if err := w.Replay(func(e *entry) error {
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

type entry struct {
	Version        byte
	SequenceNumber uint64
	Data           []byte
	Checksum       uint32
}

func (e *entry) encode(w io.Writer) error {
	buf := make([]byte, 9+len(e.Data)) // 1 version + 8 seqnum + data
	buf[0] = e.Version
	binary.BigEndian.PutUint64(buf[1:], e.SequenceNumber)
	copy(buf[9:], e.Data)

	e.Checksum = crc32.ChecksumIEEE(buf)

	if err := binary.Write(w, binary.BigEndian, e.Version); err != nil {
		return fmt.Errorf("version: %w", err)
	}

	if err := binary.Write(w, binary.BigEndian, e.SequenceNumber); err != nil {
		return fmt.Errorf("sequence number: %w", err)
	}

	if len(e.Data) > math.MaxUint32 {
		return fmt.Errorf("data too large: %d bytes", len(e.Data))
	}
	if err := binary.Write(w, binary.BigEndian, uint32(len(e.Data))); err != nil { //nolint:gosec
		return fmt.Errorf("data length: %w", err)
	}

	if _, err := w.Write(e.Data); err != nil {
		return fmt.Errorf("data: %w", err)
	}

	if err := binary.Write(w, binary.BigEndian, e.Checksum); err != nil {
		return fmt.Errorf("checksum: %w", err)
	}

	return nil
}

func (e *entry) decode(r io.Reader) error {
	var version byte
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return fmt.Errorf("version: %w", err)
	}
	if version != 1 {
		return fmt.Errorf("unknown version %d", version)
	}

	var seqNum uint64
	if err := binary.Read(r, binary.BigEndian, &seqNum); err != nil {
		return fmt.Errorf("sequence number: %w", err)
	}

	var dataLen uint32
	if err := binary.Read(r, binary.BigEndian, &dataLen); err != nil {
		return fmt.Errorf("data length: %w", err)
	}

	data := make([]byte, dataLen)
	if _, err := io.ReadFull(r, data); err != nil {
		return fmt.Errorf("data: %w", err)
	}

	var checksum uint32
	if err := binary.Read(r, binary.BigEndian, &checksum); err != nil {
		return fmt.Errorf("checksum: %w", err)
	}

	buf := make([]byte, 9+dataLen) // 1 version + 8 seqnum + data
	buf[0] = version
	binary.BigEndian.PutUint64(buf[1:], seqNum)
	copy(buf[9:], data)
	expectedChecksum := crc32.ChecksumIEEE(buf)
	if expectedChecksum != checksum {
		return fmt.Errorf("checksum mismatch at seq %d", seqNum)
	}

	e.Version = version
	e.SequenceNumber = seqNum
	e.Data = data
	e.Checksum = checksum

	return nil
}

func (w *simpleWAL) Log(data []byte) error {
	if data == nil {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	entry := &entry{
		Version:        1,
		SequenceNumber: w.nextSeqNum,
		Data:           data,
	}

	if err := entry.encode(w.writer); err != nil {
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

func (w *simpleWAL) Replay(fn func(*entry) error) error {
	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seeking to file start: %w", err)
	}

	r := bufio.NewReader(w.file)
	for {
		var ent entry
		if err := ent.decode(r); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("reading wal entry: %w", err)
		}

		if err := fn(&ent); err != nil {
			return fmt.Errorf("processing latest wal entry: %w", err)
		}
	}
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
