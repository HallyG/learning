package wal_test

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	"github.com/HallyG/learning/wal/wal"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeEntry(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input                 *wal.Entry
		expected              *wal.Entry
		expectedError         error
		expectedErrorContains string
		modifyEncodedPayload  func(buf *bytes.Buffer)
	}{
		"happy path": {
			input: &wal.Entry{
				Version:        1,
				SequenceNumber: 42,
				Data:           []byte("hello world"),
			},
			expected: &wal.Entry{
				Version:        1,
				SequenceNumber: 42,
				Data:           []byte("hello world"),
			},
		},
		"returns error for corrupted entry": {
			input: &wal.Entry{
				Version:        1,
				SequenceNumber: 42,
				Data:           []byte("hello world"),
			},
			modifyEncodedPayload: func(buf *bytes.Buffer) {
				data := buf.Bytes()
				data[len(data)-2] ^= 0xff
			},
			expectedError: wal.ErrChecksumMismatch,
		},
		"returns error for truncated write": {
			input: &wal.Entry{
				Version:        1,
				SequenceNumber: 42,
				Data:           []byte("hello world"),
			},
			modifyEncodedPayload: func(buf *bytes.Buffer) {
				data := buf.Bytes()

				if len(data) >= 5 {
					data = data[:len(data)-5]
					buf.Reset()
					buf.Write(data)
				}
			},
			expectedError: io.ErrUnexpectedEOF,
		},
		"returns error for invalid version": {
			input: &wal.Entry{
				Version:        2,
				SequenceNumber: 42,
				Data:           []byte("hello world"),
			},
			expectedError:         wal.ErrCorruptRecord,
			expectedErrorContains: "unknown version 2",
		},
		"returns error for invalid data length": {
			input: &wal.Entry{
				Version:        wal.Version1,
				SequenceNumber: 42,
				Data:           []byte("hello world22"),
			},
			modifyEncodedPayload: func(buf *bytes.Buffer) {
				// overwrite data length in header
				binary.BigEndian.PutUint32(buf.Bytes()[9:13], wal.MaxSize+1)
			},
			expectedError:         wal.ErrCorruptRecord,
			expectedErrorContains: "invalid length 104857601",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			err := test.input.Encode(&buf)
			require.NoError(t, err)

			if test.modifyEncodedPayload != nil {
				test.modifyEncodedPayload(&buf)
			}

			var got wal.Entry
			err = got.Decode(&buf)

			if test.expectedError != nil {
				require.ErrorIs(t, err, test.expectedError)

				if test.expectedErrorContains != "" {
					require.ErrorContains(t, err, test.expectedErrorContains)
				}
			} else {
				require.NoError(t, err)

				require.Equal(t, test.expected.Version, got.Version)
				require.Equal(t, test.expected.SequenceNumber, got.SequenceNumber)
				require.Equal(t, test.expected.Data, got.Data)
			}
		})
	}
}
