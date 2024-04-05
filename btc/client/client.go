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

//nolint:funlen,cyclop // TODO
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

	addrFrom, err := encoding.NewIP4Address(0, "0.0.0.0:0")
	if err != nil {
		return errors.Wrap(err, "failed to create from address")
	}
	addrRecv, err := encoding.NewIP4Address(0, c.nodeAddress)
	if err != nil {
		return errors.Wrap(err, "failed to create recv address")
	}
	version, err := encoding.NewVersionMsg(
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

	err = encoding.SendMessage(encoding.NetworkRegtest, version, writer)
	if err != nil {
		return errors.Wrap(err, "failed sending version")
	}

	_, msg, err := encoding.ReceiveMessage(reader)
	if err != nil {
		return errors.Wrap(err, "failed receiving first message")
	}
	c.log.Info("received handshake message", "command", string(msg.GetCommand()))

	verack, err := encoding.NewVerackMsg()
	if err != nil {
		return errors.Wrap(err, "failed to create verack message")
	}
	err = encoding.SendMessage(encoding.NetworkRegtest, verack, writer)
	if err != nil {
		return errors.Wrap(err, "failed sending version")
	}

	_, msg, err = encoding.ReceiveMessage(reader)
	if err != nil {
		return errors.Wrap(err, "failed receiving second message")
	}
	c.log.Info("received handshake message", "command", string(msg.GetCommand()))

	c.log.Info("connected to bitcoin node", "address", c.nodeAddress)

	for {
		_, msg, err = encoding.ReceiveMessage(reader)
		if err != nil {
			return errors.Wrap(err, "failed receiving message")
		}
		c.log.Info("received message", "command", string(msg.GetCommand()))
	}
}

func (c *BTCClient) watchContext() {
	<-c.ctx.Done()
	c.log.Info("terminating client")
	c.conn.Close()
}
