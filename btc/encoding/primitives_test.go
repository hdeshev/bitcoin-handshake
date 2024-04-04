package encoding

import (
	"bytes"
	"strings"
	"testing"

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
