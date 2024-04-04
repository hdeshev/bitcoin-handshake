package client

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"time"

	"github.com/pkg/errors"

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

//nolint:funlen // TODO
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

	var reader io.Reader = conn
	var writer io.Writer = conn

	versionCommand := [12]byte{}
	commandStr := "version"
	copy(versionCommand[:], commandStr)

	addrFrom, err := encoding.NewIP4Address(0, "0.0.0.0:0")
	if err != nil {
		return errors.Wrap(err, "failed to create from address")
	}
	addrRecv, err := encoding.NewIP4Address(0, c.nodeAddress)
	if err != nil {
		return errors.Wrap(err, "failed to create recv address")
	}
	version, err := encoding.NewVersionMsg(
		encoding.NetworkRegtest,
		time.Now(),
		0,
		addrRecv,
		addrFrom,
		uint64(rand.Int63()), //nolint:gosec // not a crypto random
		1,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create version message")
	}

	err = version.Encode(writer)
	if err != nil {
		return errors.Wrap(err, "failed serializing version")
	}

	headerSize := 24
	inHeader := make([]byte, headerSize)
	nn, err := io.ReadFull(reader, inHeader)
	c.log.Info("read data", "nn", nn, "size", len(inHeader), "data", string(inHeader), "error", err)

	return fmt.Errorf("TODO")
}

func (c *BTCClient) watchContext() {
	<-c.ctx.Done()
	c.log.Info("context done")
	c.conn.Close()
}
