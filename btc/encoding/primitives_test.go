package encoding

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_VarInt_Encode(t *testing.T) {
	tests := []struct {
		name string
		num  VarInt
		want string
	}{
		{
			name: "short",
			num:  VarInt(0xA),
			want: strip(`0A`),
		},
		{
			name: "medium",
			num:  VarInt(0xF234),
			want: strip(`FD 34 F2`),
		},
		{
			name: "large",
			num:  VarInt(0xF2341020),
			want: strip(`FE 20 10 34 F2`),
		},
		{
			name: "extra large",
			num:  VarInt(0xF2341020F2341020),
			want: strip(`FF 20 10 34 F2 20 10 34 F2`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := tt.num.Encode(buf)
			assert.NoError(t, err)

			got := formatBinary(buf.Bytes())
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_VarStr_Encode(t *testing.T) {
	tests := []struct {
		name string
		str  VarStr
		want string
	}{
		{
			name: "short",
			str:  VarStr("user-agent-1"),
			want: strip(`0C 75 73 65 72 2D 61 67 65 6E 74 2D 31`),
		},
		{
			name: "medium",
			str:  VarStr(strings.Repeat("user-agent-1", 500)),
			want: strip(`FD 70 17 75 73 65 72 2D 61 67 65 6E 74 2D 31`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := tt.str.Encode(buf)
			assert.NoError(t, err)

			got := formatBinary(buf.Bytes())
			assert.Contains(t, got, tt.want)
		})
	}
}

func Test_NetworkAddress_Encode(t *testing.T) {
	testTime := time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		addr *NetworkAddress
		want string
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

func Test_NetworkAddress_Decode(t *testing.T) {
	tests := []struct {
		name  string
		want  *NetworkAddress
		input []byte
	}{
		{
			name: "IPv4",
			want: noErr(t, func() (*NetworkAddress, error) {
				return NewIP4Address(1, "10.0.0.1:8333")
			}),
			input: unformatBinary(`
			01 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
			00 00 FF FF 0A 00 00 01 20 8D`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tt.input)
			got := &NetworkAddress{}
			err := got.Decode(buf)

			assert.NoError(t, err)
			assert.Equal(t, tt.want.Time, got.Time)
			assert.Equal(t, tt.want.Services, got.Services)
			assert.Equal(t, tt.want.IP, got.IP)
			assert.Equal(t, tt.want.Port, got.Port)
		})
	}
}
