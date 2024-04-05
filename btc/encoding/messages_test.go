package encoding

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Header_Encode(t *testing.T) {
	versionCommand := [12]byte{}
	commandStr := "version"
	copy(versionCommand[:], commandStr)

	payload := []byte("test data ")

	tests := []struct {
		name   string
		header *Header
		want   string
	}{
		{
			name: "mainnet",
			header: noErr(t, func() (*Header, error) {
				return NewHeader(NetworkMainnet, VersionCommand, payload)
			}),
			want: strip(`
			F9 BE B4 D9 76 65 72 73 69 6F 6E 00 00 00 00 00
			0A 00 00 00 6E D5 BA D9`,
			),
		},
		{
			name: "testnet",
			header: noErr(t, func() (*Header, error) {
				return NewHeader(NetworkTestnet3, VersionCommand, payload)
			}),
			want: strip(`
			0B 11 09 07 76 65 72 73 69 6F 6E 00 00 00 00 00
			0A 00 00 00 6E D5 BA D9`,
			),
		},
		{
			name: "regtest",
			header: noErr(t, func() (*Header, error) {
				return NewHeader(NetworkRegtest, VersionCommand, payload)
			}),
			want: strip(`
			FA BF B5 DA 76 65 72 73 69 6F 6E 00 00 00 00 00
			0A 00 00 00 6E D5 BA D9`,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := tt.header.Encode(buf)
			assert.NoError(t, err)

			got := formatBinary(buf.Bytes())
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Header_Decode(t *testing.T) {
	versionCommand := [12]byte{}
	commandStr := "version"
	copy(versionCommand[:], commandStr)

	payload := []byte("test data ")

	tests := []struct {
		name  string
		input []byte
		want  *Header
	}{
		{
			name: "from binary",
			input: unformatBinary(`
			F9 BE B4 D9 76 65 72 73 69 6F 6E 00 00 00 00 00
			0A 00 00 00 6E D5 BA D9`,
			),
			want: noErr(t, func() (*Header, error) {
				return NewHeader(NetworkMainnet, VersionCommand, payload)
			}),
		},
		{
			name: "roundtrip",
			input: func() []byte {
				h, err := NewHeader(NetworkMainnet, VersionCommand, payload)
				assert.NoError(t, err)
				b := bytes.NewBuffer(nil)
				err = h.Encode(b)
				assert.NoError(t, err)
				return b.Bytes()
			}(),
			want: noErr(t, func() (*Header, error) {
				return NewHeader(NetworkMainnet, VersionCommand, payload)
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tt.input)

			got := &Header{}
			err := got.Decode(buf)

			assert.NoError(t, err)
			assert.Equal(t, tt.want.Magic, got.Magic)
			assert.Equal(t, tt.want.Command, got.Command)
			assert.Equal(t, tt.want.PayloadSize, got.PayloadSize)
			assert.Equal(t, tt.want.Checksum, got.Checksum)
		})
	}
}

func noErr[T any](t *testing.T, f func() (T, error)) T {
	t.Helper()

	r, err := f()
	assert.NoError(t, err)
	return r
}

func formatBinary(buf []byte) string {
	var out []string
	var line []string
	for i, b := range buf {
		if i > 0 && i%16 == 0 {
			out = append(out, strings.Join(line, " "))
			line = []string{}
		}

		byteStr := fmt.Sprintf("%02X", b)
		line = append(line, byteStr)
	}
	out = append(out, strings.Join(line, " "))
	return strings.Join(out, "\n")
}

func unformatBinary(s string) []byte {
	s = strings.ReplaceAll(strip(s), "\n", " ")
	parts := strings.Split(s, " ")
	buf := make([]byte, len(parts))
	for i, p := range parts {
		b, err := hex.DecodeString(p)
		if err != nil {
			panic("decode error: " + err.Error())
		}
		if len(b) != 1 {
			panic("invalid hex string: " + p)
		}
		buf[i] = b[0]
	}
	return buf
}

func strip(s string) string {
	return regexp.MustCompile(`(?m)^\s+`).ReplaceAllString(s, "")
}
