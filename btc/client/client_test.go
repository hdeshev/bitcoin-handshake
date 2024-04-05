package client

import (
	"bytes"
	"context"
	"log/slog"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"deshev.com/bitcoin-handshake/btc/encoding"
	"deshev.com/bitcoin-handshake/config"
)

func Test_Client_Connect_and_Cleanup(t *testing.T) {
	listenerStarted := make(chan struct{})
	var listener net.Listener
	go func() {
		var err error
		listener, err = net.Listen("tcp", "127.0.0.1:58333")
		assert.NoError(t, err)
		listenerStarted <- struct{}{}

		conn, err := listener.Accept()
		assert.NoError(t, err)
		assert.NotNil(t, conn)
	}()

	<-listenerStarted
	defer listener.Close()

	cfg := &config.Config{
		BTCNodeAddress: "127.0.0.1:58333",
	}
	log := slog.Default()
	ctx, cancel := context.WithCancel(context.Background())

	c := New(ctx, log, cfg)
	messageC, err := c.Connect()
	assert.NoError(t, err)
	assert.NotNil(t, messageC)

	cancel()
	_, stillOpen := <-messageC
	assert.False(t, stillOpen)
}

func Test_Client_HandshakeStates(t *testing.T) {
	cfg := config.New()
	log := slog.Default()
	ctx := context.Background()
	pingCommand := [12]byte{}
	copy(pingCommand[:], "ping")

	tests := []struct {
		name                 string
		messages             []encoding.Message
		wantHandShakeVersion bool
		wantHandShakeVerack  bool
		wantErr              string
		wantAppMessages      []string
	}{
		{
			name:                 "initial state",
			messages:             []encoding.Message{},
			wantHandShakeVersion: false,
			wantHandShakeVerack:  false,
		},
		{
			name: "version message",
			messages: []encoding.Message{
				&encoding.MsgVersion{},
			},
			wantHandShakeVersion: true,
			wantHandShakeVerack:  false,
		},
		{
			name: "version and verack",
			messages: []encoding.Message{
				&encoding.MsgVersion{},
				&encoding.MsgVerack{},
			},
			wantHandShakeVersion: true,
			wantHandShakeVerack:  true,
		},
		{
			name: "duplicate version",
			messages: []encoding.Message{
				&encoding.MsgVersion{},
				&encoding.MsgVersion{},
			},
			wantHandShakeVersion: true,
			wantHandShakeVerack:  false,
			wantErr:              "received duplicate version message",
		},
		{
			name: "duplicate verack",
			messages: []encoding.Message{
				&encoding.MsgVersion{},
				&encoding.MsgVerack{},
				&encoding.MsgVerack{},
			},
			wantHandShakeVersion: true,
			wantHandShakeVerack:  true,
			wantErr:              "received duplicate verack message",
		},
		{
			name: "one app message after handshake",
			messages: []encoding.Message{
				&encoding.MsgVersion{},
				&encoding.MsgVerack{},
				&encoding.MsgRaw{Header: &encoding.Header{Command: pingCommand}},
			},
			wantHandShakeVersion: true,
			wantHandShakeVerack:  true,
			wantAppMessages:      []string{"ping"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(ctx, log, cfg)
			c.writer = bytes.NewBuffer(nil)
			c.reader = bytes.NewBuffer(nil)
			c.messageC = make(chan encoding.Message, 5)

			assert.False(t, c.handShakeVersion)
			assert.False(t, c.handShakeVerack)

			var err error
			for _, message := range tt.messages {
				err = c.processMessage(message)
				if err != nil {
					break
				}
			}

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantHandShakeVersion, c.handShakeVersion)
			assert.Equal(t, tt.wantHandShakeVerack, c.handShakeVerack)

			var appMessages []string
			close(c.messageC)
			for message := range c.messageC {
				appMessages = append(appMessages, string(message.GetCommand()))
			}
			assert.Equal(t, tt.wantAppMessages, appMessages)
		})
	}
}
