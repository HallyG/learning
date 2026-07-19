package wal_test

import (
	"bytes"
	"testing"

	"github.com/HallyG/learning/wal/wal"
	"github.com/stretchr/testify/require"
)

func TestReplay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		buildInput      func(t *testing.T) []byte
		expectedEntries []wal.Entry
		expectedErr     error
	}{
		{
			name:            "empty WAL",
			expectedEntries: nil,
			expectedErr:     nil,
		},
		{
			name: "single entry",
			buildInput: func(t *testing.T) []byte {
				t.Helper()

				return buildWAL(
					t,
					wal.Entry{
						Version:        1,
						SequenceNumber: 1,
						Data:           []byte("hello"),
					},
				)
			},
			expectedEntries: []wal.Entry{
				{
					Version:        1,
					SequenceNumber: 1,
					Data:           []byte("hello"),
				},
			},
		},
		{
			name: "multiple entries",
			buildInput: func(t *testing.T) []byte {
				t.Helper()

				return buildWAL(
					t,
					wal.Entry{
						Version:        1,
						SequenceNumber: 1,
						Data:           []byte("one"),
					},
					wal.Entry{
						Version:        1,
						SequenceNumber: 2,
						Data:           []byte("two"),
					},
					wal.Entry{
						Version:        1,
						SequenceNumber: 3,
						Data:           []byte("three"),
					},
				)
			},
			expectedEntries: []wal.Entry{
				{
					Version:        1,
					SequenceNumber: 1,
					Data:           []byte("one"),
				},
				{
					Version:        1,
					SequenceNumber: 2,
					Data:           []byte("two"),
				},
				{
					Version:        1,
					SequenceNumber: 3,
					Data:           []byte("three"),
				},
			},
		},
		{
			name: "returns error when corrupt checksum",
			buildInput: func(t *testing.T) []byte {
				t.Helper()

				data := buildWAL(
					t,
					wal.Entry{
						Version:        1,
						SequenceNumber: 1,
						Data:           []byte("hello"),
					},
				)

				data[len(data)-5] ^= 0xff
				return data
			},
			expectedErr: wal.ErrChecksumMismatch,
		},
		{
			name: "returns error when truncated final record",
			buildInput: func(t *testing.T) []byte {
				t.Helper()

				data := buildWAL(
					t,
					wal.Entry{
						Version:        1,
						SequenceNumber: 1,
						Data:           []byte("hello"),
					},
				)

				// e.g crash during write
				return data[:len(data)-3]
			},
			expectedErr: wal.ErrCorruptEntry,
		},
		{
			name: "return error when valid entries followed by corrupt entry",
			buildInput: func(t *testing.T) []byte {
				t.Helper()

				data := buildWAL(
					t,
					wal.Entry{
						Version:        1,
						SequenceNumber: 1,
						Data:           []byte("one"),
					},
					wal.Entry{
						Version:        1,
						SequenceNumber: 2,
						Data:           []byte("two"),
					},
					wal.Entry{
						Version:        1,
						SequenceNumber: 3,
						Data:           []byte("three"),
					},
				)

				data[len(data)-6] ^= 0xff
				return data
			},
			expectedEntries: []wal.Entry{
				{
					Version:        1,
					SequenceNumber: 1,
					Data:           []byte("one"),
				},
				{
					Version:        1,
					SequenceNumber: 2,
					Data:           []byte("two"),
				},
			},
			expectedErr: wal.ErrChecksumMismatch,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var (
				output []wal.Entry
			)

			var input []byte = nil
			if test.buildInput != nil {
				input = test.buildInput(t)
			}

			err := wal.ReplayReader(
				bytes.NewReader(input),
				func(e *wal.Entry) error {
					output = append(output, *e)
					return nil
				},
			)

			if test.expectedErr != nil {
				require.ErrorIs(t, err, test.expectedErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, test.expectedEntries, output)
		})
	}
}

func buildWAL(t *testing.T, entries ...wal.Entry) []byte {
	t.Helper()

	var buf bytes.Buffer

	for i := range entries {
		require.NoError(
			t,
			entries[i].Encode(&buf),
		)
	}

	return buf.Bytes()
}
