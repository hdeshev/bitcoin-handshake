package encoding

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Verack_Roundtrip(t *testing.T) {
	verack, err := NewVerackMsg()
	assert.NoError(t, err)

	buf := bytes.NewBuffer(nil)

	err = SendMessage(NetworkMainnet, verack, buf)
	assert.NoError(t, err)
	header, got, err := ReceiveMessage(buf)
	assert.NoError(t, err)

	assert.Equal(t, VerackCommand, header.GetCommand())
	assert.Equal(t, verack, got)
}
