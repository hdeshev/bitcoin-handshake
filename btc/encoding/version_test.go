package encoding

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const docsUserAgent = "/Satoshi:0.7.2/"

func Test_Version_Send(t *testing.T) {
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
				v.UserAgent = docsUserAgent
				return v, nil
			}),
			want: strip(`
			F9 BE B4 D9 76 65 72 73 69 6F 6E 00 00 00 00 00
			65 00 00 00 8A 80 97 A9 62 EA 00 00 01 00 00 00
			00 00 00 00 11 B2 D0 50 00 00 00 00 01 00 00 00
			00 00 00 00 00 00 00 00 00 00 00 00 00 00 FF FF
			00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
			00 00 00 00 00 00 00 00 FF FF 00 00 00 00 00 00
			3B 2E B3 5D 8C E6 17 65 0F 2F 53 61 74 6F 73 68
			69 3A 30 2E 37 2E 32 2F C0 3E 03 00 00`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := SendMessage(NetworkMainnet, tt.version, buf)
			assert.NoError(t, err)
			got := formatBinary(buf.Bytes())
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_Version_Decode(t *testing.T) {
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
	}{
		{
			name: "roundtrip",
			version: noErr(t, func() (*MsgVersion, error) {
				v, err := NewVersionMsg(
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
				v.UserAgent = docsUserAgent
				return v, nil
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)

			err := tt.version.Encode(buf)
			assert.NoError(t, err)

			loaded := &MsgVersion{}
			err = loaded.Decode(buf)
			assert.NoError(t, err)
			assert.Equal(t, tt.version, loaded)
		})
	}
}

func Test_Version_Receive(t *testing.T) {
	docsTime := time.Unix(0x50D0B211, 0)
	recvAddr := noErr(t, func() (*NetworkAddress, error) {
		return NewIP4Address(ServicesNodeNetwork, "0.0.0.0:0")
	})
	fromAddr := noErr(t, func() (*NetworkAddress, error) {
		return NewIP4Address(ServicesNone, "0.0.0.0:0")
	})
	tests := []struct {
		name  string
		want  *MsgVersion
		input []byte
	}{
		{
			name: "documentation example",
			want: noErr(t, func() (*MsgVersion, error) {
				v, err := NewVersionMsg(
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
				v.UserAgent = docsUserAgent
				return v, nil
			}),
			input: unformatBinary(`
			F9 BE B4 D9 76 65 72 73 69 6F 6E 00 00 00 00 00
			65 00 00 00 8A 80 97 A9 62 EA 00 00 01 00 00 00
			00 00 00 00 11 B2 D0 50 00 00 00 00 01 00 00 00
			00 00 00 00 00 00 00 00 00 00 00 00 00 00 FF FF
			00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
			00 00 00 00 00 00 00 00 FF FF 00 00 00 00 00 00
			3B 2E B3 5D 8C E6 17 65 0F 2F 53 61 74 6F 73 68
			69 3A 30 2E 37 2E 32 2F C0 3E 03 00 00`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tt.input)
			header, got, err := ReceiveMessage(buf)
			assert.NoError(t, err)

			assert.Equal(t, VersionCommand, header.GetCommand())
			assert.Equal(t, tt.want, got)
		})
	}
}
