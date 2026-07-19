package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"math"
)

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
