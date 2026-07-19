package wal_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/HallyG/learning/wal/wal"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeEntry(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input                *wal.Entry
		expected             *wal.Entry
		expectedError        error
		modifyEncodedPayload func(buf *bytes.Buffer)
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
					fmt.Println("RUN")
					data = data[:len(data)-5]
					buf.Reset()
					buf.Write(data)
				}
			},
			expectedError: io.ErrUnexpectedEOF,
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
			} else {
				require.NoError(t, err)

				require.Equal(t, test.expected.Version, got.Version)
				require.Equal(t, test.expected.SequenceNumber, got.SequenceNumber)
				require.Equal(t, test.expected.Data, got.Data)
			}
		})
	}
}
