package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
)

const (
	Version1   byte = 1
	MaxSize         = 100 << 20
	headerSize      = 1 + 8 + 4
)

var (
	ErrChecksumMismatch = errors.New("entry checksum mismatch")
	ErrCorruptEntry     = errors.New("corrupt entry")
)

type Entry struct {
	Version        byte
	SequenceNumber uint64
	Data           []byte
}

func (e *Entry) Encode(w io.Writer) error {
	dataLen := len(e.Data)
	if dataLen > MaxSize {
		return fmt.Errorf("invalid length %d", dataLen)
	}

	buf := make([]byte, headerSize+dataLen+4)
	buf[0] = e.Version
	binary.BigEndian.PutUint64(buf[1:9], e.SequenceNumber)
	binary.BigEndian.PutUint32(buf[9:13], uint32(dataLen)) //nolint:gosec
	copy(buf[13:13+dataLen], e.Data)

	checksum := crc32.ChecksumIEEE(buf[:headerSize+dataLen])
	binary.BigEndian.PutUint32(buf[headerSize+dataLen:], checksum)

	if _, err := w.Write(buf); err != nil {
		return fmt.Errorf("writing entry: %w", err)
	}

	return nil
}

func (e *Entry) Decode(r io.Reader) error {
	header := make([]byte, headerSize)
	if _, err := io.ReadFull(r, header); err != nil {
		if errors.Is(err, io.EOF) {
			return io.EOF
		}

		return fmt.Errorf("%w: reading header: %w", ErrCorruptEntry, err)
	}

	version := header[0]
	if version != Version1 {
		return fmt.Errorf("%w: unknown version %d", ErrCorruptEntry, version)
	}

	seqNum := binary.BigEndian.Uint64(header[1:9])
	dataLen := binary.BigEndian.Uint32(header[9:13])
	if dataLen > MaxSize {
		return fmt.Errorf("%w: invalid length %d", ErrCorruptEntry, dataLen)
	}

	buf := make([]byte, headerSize+int(dataLen)+4)
	copy(buf, header)

	if _, err := io.ReadFull(r, buf[headerSize:headerSize+int(dataLen)]); err != nil {
		return fmt.Errorf("%w: reading data: %w", ErrCorruptEntry, err)
	}

	checksumBytes := buf[headerSize+int(dataLen):]
	if _, err := io.ReadFull(r, checksumBytes); err != nil {
		return fmt.Errorf("%w reading checksum: %w", ErrCorruptEntry, err)
	}

	checksum := binary.BigEndian.Uint32(checksumBytes)
	expected := crc32.ChecksumIEEE(buf[:headerSize+int(dataLen)])
	if expected != checksum {
		return fmt.Errorf("%w at sequence %d", ErrChecksumMismatch, seqNum)
	}

	e.Version = version
	e.SequenceNumber = seqNum
	e.Data = buf[headerSize : headerSize+int(dataLen)]

	return nil
}
