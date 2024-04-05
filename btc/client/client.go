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
	messageC chan encoding.Message

	ctx         context.Context
	log         *slog.Logger
	nodeAddress string
	reader      io.Reader
	writer      io.Writer

	handShakeVersion bool
	handShakeVerack  bool
}

const messageBufferSize = 10

func New(ctx context.Context, log *slog.Logger, cfg *config.Config) *BTCClient {
	return &BTCClient{
		nodeAddress: cfg.BTCNodeAddress,
		ctx:         ctx,
		log:         log,
		messageC:    make(chan encoding.Message, messageBufferSize),
	}
}

func (c *BTCClient) Connect() (<-chan encoding.Message, error) {
	c.log.Info("connecting to bitcoin node", "address", c.nodeAddress)

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(c.ctx, "tcp", c.nodeAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to bitcoin node: %w", err)
	}

	c.reader = conn
	c.writer = conn
	go c.cleanup(conn)
	go c.receiveMessages()

	err = c.startHandshake()
	if err != nil {
		return nil, fmt.Errorf("failed to start handshake: %w", err)
	}

	return c.messageC, nil
}

func (c *BTCClient) startHandshake() error {
	version, err := c.createConnectMessage()
	if err != nil {
		return errors.Wrap(err, "failed to create version message")
	}

	c.log.Info("sending handshake version message")
	err = encoding.SendMessage(encoding.NetworkRegtest, version, c.writer)
	if err != nil {
		return errors.Wrap(err, "failed sending version")
	}

	return nil
}

func (c *BTCClient) receiveMessages() {
	for {
		_, msg, err := encoding.ReceiveMessage(c.reader)
		if err != nil {
			c.log.Error("failed receiving message", "error", err)
			close(c.messageC)
			return
		}
		err = c.processMessage(msg)
		if err != nil {
			c.log.Error("failed processing message", "error", err)
			close(c.messageC)
			return
		}
	}
}

func (c *BTCClient) processMessage(msg encoding.Message) error {
	switch msg.GetCommand() {
	case encoding.VersionCommand:
		if c.handShakeVersion {
			return errors.New("received duplicate version message")
		}
		c.handShakeVersion = true
		c.log.Info("received handshake version message")

		verack, err := encoding.NewVerackMsg()
		if err != nil {
			return errors.Wrap(err, "failed to create verack message")
		}
		c.log.Info("sending handshake verack message")
		err = encoding.SendMessage(encoding.NetworkRegtest, verack, c.writer)
		if err != nil {
			return errors.Wrap(err, "failed sending verack message")
		}
	case encoding.VerackCommand:
		if c.handShakeVerack {
			return errors.New("received duplicate verack message")
		}
		c.handShakeVerack = true
		c.log.Info("received handshake verack message")
	default:
		handshakeDone := c.handShakeVersion && c.handShakeVerack
		if !handshakeDone {
			return fmt.Errorf("received unexpected message before completing handshake: %s", msg.GetCommand())
		}
		c.log.Debug("received message", "command", string(msg.GetCommand()), "handshake_done", handshakeDone)
		c.messageC <- msg
	}
	return nil
}

func (c *BTCClient) createConnectMessage() (encoding.Message, error) {
	addrFrom, err := encoding.NewIP4Address(0, "0.0.0.0:0")
	if err != nil {
		return nil, errors.Wrap(err, "failed to create from address")
	}
	addrRecv, err := encoding.NewIP4Address(0, c.nodeAddress)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create recv address")
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
		return nil, errors.Wrap(err, "failed to create version message")
	}
	return version, nil
}

func (c *BTCClient) cleanup(conn net.Conn) {
	<-c.ctx.Done()
	c.log.Info("terminating client")
	conn.Close()
}
