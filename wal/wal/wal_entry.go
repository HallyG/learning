package wal

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"math"
)

const (
	Version1   = 1
	MaxSize    = math.MaxUint32
	headerSize = 1 + 8 + 4
)

var (
	ErrChecksumMismatch = errors.New("checksum mismatch")
	ErrCorruptRecord    = errors.New("corrupt wal record")
)

type Entry struct {
	Version        byte
	SequenceNumber uint64
	Data           []byte
	Checksum       uint32
}

func (e *Entry) Encode(w io.Writer) error {
	if len(e.Data) > MaxSize {
		return fmt.Errorf("invalid length %d", len(e.Data))
	}

	if e.Version == 0 {
		e.Version = Version1
	}

	if err := binary.Write(w, binary.BigEndian, e.Version); err != nil {
		return fmt.Errorf("writing version: %w", err)
	}

	if err := binary.Write(w, binary.BigEndian, e.SequenceNumber); err != nil {
		return fmt.Errorf("writing sequence number: %w", err)
	}

	if err := binary.Write(w, binary.BigEndian, uint32(len(e.Data))); err != nil { //nolint:gosec
		return fmt.Errorf("writing data length: %w", err)
	}

	if _, err := w.Write(e.Data); err != nil {
		return fmt.Errorf("writing data: %w", err)
	}

	checksum := calculateChecksum(e.Version, e.SequenceNumber, uint32(len(e.Data)), e.Data)
	if err := binary.Write(w, binary.BigEndian, checksum); err != nil {
		return fmt.Errorf("writing checksum: %w", err)
	}

	e.Checksum = checksum

	return nil
}

func (e *Entry) Decode(r io.Reader) error {
	var version byte
	if err := binary.Read(r, binary.BigEndian, &version); err != nil {
		return fmt.Errorf("reading version: %w", err)
	}
	if version != 1 {
		return fmt.Errorf("%w: unknown version %d", ErrCorruptRecord, version)
	}

	var seqNum uint64
	if err := binary.Read(r, binary.BigEndian, &seqNum); err != nil {
		return fmt.Errorf("reading sequence number: %w", err)
	}

	var dataLen uint32
	if err := binary.Read(r, binary.BigEndian, &dataLen); err != nil {
		return fmt.Errorf("reading data length: %w", err)
	}

	if dataLen > MaxSize {
		return fmt.Errorf("%w: invalid length %d", ErrCorruptRecord, len(e.Data))
	}

	data := make([]byte, dataLen)
	if _, err := io.ReadFull(r, data); err != nil {
		return fmt.Errorf("reading data: %w", err)
	}

	var checksum uint32
	if err := binary.Read(r, binary.BigEndian, &checksum); err != nil {
		return fmt.Errorf("reading checksum: %w", err)
	}

	expectedChecksum := calculateChecksum(version, seqNum, dataLen, data)
	if expectedChecksum != checksum {
		return fmt.Errorf("%w at sequence %d", ErrChecksumMismatch, seqNum)
	}

	e.Version = version
	e.SequenceNumber = seqNum
	e.Data = data
	e.Checksum = checksum

	return nil
}

func calculateChecksum(version byte, seq uint64, dataLen uint32, data []byte) uint32 {
	buf := make([]byte, headerSize+len(data))

	buf[0] = version
	binary.BigEndian.PutUint64(buf[1:], seq)
	binary.BigEndian.PutUint32(buf[9:], uint32(len(data)))
	copy(buf[13:], data)

	return crc32.ChecksumIEEE(buf)
}
