package encoding

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_SerializeHeader(t *testing.T) {
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
			name: "version only",
			header: noErr(t, func() (*Header, error) {
				return NewHeader(NetworkMainnet, VersionCommand, payload)
			}),
			want: strip(`
			F9 BE B4 D9 76 65 72 73 69 6F 6E 00 00 00 00 00
			0A 00 00 00 6E D5 BA D9`,
			),
		},
		{
			name: "version only",
			header: noErr(t, func() (*Header, error) {
				return NewHeader(NetworkTestnet3, VersionCommand, payload)
			}),
			want: strip(`
			0B 11 09 07 76 65 72 73 69 6F 6E 00 00 00 00 00
			0A 00 00 00 6E D5 BA D9`,
			),
		},
		{
			name: "version only",
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

func Test_SerializeNetworkAddress(t *testing.T) {
	testTime := time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		addr   *NetworkAddress
		want   string
		err    error
		errStr string
	}{
		{
			name: "IPv4",
			addr: noErr(t, func() (*NetworkAddress, error) {
				return NewIP4Address(1, "10.0.0.1:8333")
			}),
			want: strip(`
			01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
			00 00 FF FF 0A 00 00 01 20 8D`),
		},
		{
			name: "IPv4 with time",
			addr: noErr(t, func() (*NetworkAddress, error) {
				a, err := NewIP4Address(1, "10.0.0.1:8333")
				if err != nil {
					return nil, err
				}
				a.Time = UInt32(testTime.Unix())
				return a, nil
			}),
			want: strip(`
			00 F9 09 66 01 00 00 00 00 00 00 00 00 00 00 00
			00 00 00 00 00 00 FF FF 0A 00 00 01 20 8D`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := tt.addr.Encode(buf)
			assert.NoError(t, err)

			got := formatBinary(buf.Bytes())
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_SerializeVersion(t *testing.T) {
	docsTime := time.Unix(0x50D0B211, 0)
	recvAddr := noErr(t, func() (*NetworkAddress, error) {
		return NewIP4Address(ServicesNodeNetwork, "0.0.0.0:0")
	})
	fromAddr := noErr(t, func() (*NetworkAddress, error) {
		return NewIP4Address(ServicesNone, "0.0.0.0:0")
	})
	tests := []struct {
		name    string
		version *MsgVersion
		want    string
	}{
		{
			name: "documentation example",
			version: noErr(t, func() (*MsgVersion, error) {
				v, err := NewVersionMsg(
					NetworkMainnet,
					docsTime,
					ServicesNodeNetwork,
					recvAddr,
					fromAddr,
					0x6517E68C5DB32E3B,
					212672,
				)
				if err != nil {
					return nil, err
				}
				v.Version = 60002
				v.UserAgent = "/Satoshi:0.7.2/"
				return v, nil
			}),
			want: strip(`
			F9 BE B4 D9 76 65 72 73 69 6F 6E 00 00 00 00 00
			64 00 00 00 35 8D 49 32 62 EA 00 00 01 00 00 00
			00 00 00 00 11 B2 D0 50 00 00 00 00 01 00 00 00
			00 00 00 00 00 00 00 00 00 00 00 00 00 00 FF FF
			00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
			00 00 00 00 00 00 00 00 FF FF 00 00 00 00 00 00
			3B 2E B3 5D 8C E6 17 65 0F 2F 53 61 74 6F 73 68
			69 3A 30 2E 37 2E 32 2F C0 3E 03 00`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := tt.version.Encode(buf)
			assert.NoError(t, err)
			got := formatBinary(buf.Bytes())
			assert.Equal(t, tt.want, got)
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

func strip(s string) string {
	return regexp.MustCompile(`(?m)^\s+`).ReplaceAllString(s, "")
}
