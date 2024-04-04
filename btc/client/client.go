package client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"deshev.com/bitcoin-handshake/btc/encoding"
	"deshev.com/bitcoin-handshake/config"
)

type BTCClient struct {
	ctx         context.Context
	log         *slog.Logger
	nodeAddress string
	conn        net.Conn
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) *BTCClient {
	return &BTCClient{
		nodeAddress: cfg.BTCNodeAddress,
		ctx:         ctx,
		log:         log,
	}
}

func (c *BTCClient) Connect() error {
	c.log.Info("connecting to bitcoin node", "address", c.nodeAddress)

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(c.ctx, "tcp", c.nodeAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to bitcoin node: %w", err)
	}
	defer conn.Close()

	c.conn = conn
	go c.watchContext()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	versionCommand := [12]byte{}
	commandStr := "version"
	copy(versionCommand[:], commandStr)

	headerSize := 24

	version := &encoding.MsgVersion{
		Version:   encoding.ProtocolVersion,
		Services:  0,
		Timestamp: encoding.UInt64(time.Now().Unix()),

		Nonce: 1, // TODO: Generate random nonce
	}
	versionBuf := bytes.NewBuffer(nil)
	err = version.Encode(versionBuf)
	if err != nil {
		c.log.Error("failed serializing version", "error", err)
	}

	n, err := io.Copy(writer, versionBuf)
	c.log.Info("wrote version msg", "n", n, "error", err)
	writer.Flush()

	inHeader := make([]byte, headerSize)
	nn, err := io.ReadFull(reader, inHeader)
	c.log.Info("read data", "nn", nn, "data", string(inHeader), "error", err)

	return fmt.Errorf("TODO")
}

func (c *BTCClient) watchContext() {
	<-c.ctx.Done()
	c.log.Info("context done")
	c.conn.Close()
}
