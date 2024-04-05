# Architecture

The project is built as a single executable that connects to a running Bitcoin node.

The code has two major components:
- The network client that deals with the networking and message passing part.
- The encoding subpackage that encodes and decodes messages to and from binary according to the [Bitcoin protocol](https://en.bitcoin.it/wiki/Protocol_documentation).

### Network Client

The client (`btc/client/client.go`) sets up the TCP connection, and kicks off the handshake process by sending the first version message. It has a background receive goroutine that parses messages and performs different actions according to the current state. Once we complete the handshake, we switch to "application" mode and start forwarding any other packets to an outgoing channel.

Losing connectivity, canceling the client parent context, or other network or parse errors will close the outgoing channel and the connection.

### Encoding and Decoding

I tried to isolate the different types of objects that can be encoded or decoded from the network. Those can be of different types:

- primitives: numbers and fixed-size strings.
- common objects: network addresses, varint, varstr, etc.
- messages: version, verack

I have added a "raw" message type that only reads the full message from the network and passes it to the handler without parsing the body. This is useful for testing and debugging.

## Deployment

The client can be deployed to any container runtime. We have a working Docker image builder that can be extended with a Helm chart.

Configuration is done via environment variables, [12-factor style](https://12factor.net/config). See `config/config.go` for the full list. Those have been kept to the bare minimum like the Bitcoin node endpoint.

## Extensibility

Adding extended support for other messages should be a matter of adding a new message struct and implementing `Encode` and `Decode`.

The client network resiliency should be tweaked with some real-life tests. I would like to set up the right network timeouts and make sure we don't hang unnecessarily long in cases of slow networks or nonconformant peers.

The client handshake state management is pretty ad hoc and uses two boolean flags. It should be a proper state machine that would let us test state transitions in isolation and allow us to manage state from previous messages (e.g. being able to construct a "pong" message with the nonce we got in the previous "ping").

The message parser doesn't verify message checksums in the headers either. That should be added as well and we should be returning "reject" messages in that case.

## Security

- The service is meant to be deployed in a private network alongside the Bitcoin node it requires.

## Testing

- I have added unit tests for the individual components. Coverage is not at 100%, but we could easily get there. Looking at the coverage report we are missing mostly error handling and logging branches that should be easy to test.
- We are close to doing a proper integration test. We have the building blocks to build a TCP server that can speak the handshake on the remote part. The `Test_Client_Connect_and_Cleanup` in `client_test.go` can be used as a starting point for that.
- E2E tests should be possible. We can have a test suite running in Docker Compose that spins up a real Bitcoin node and runs tests connecting to it.

## Monitoring and Logging

- Logging is implemented using the relatively new `log/slog` Go stdlib package. The root logger is created in `main.go` and propagated to downstream components, so we can easily change log configuration and say easily switch to JSON-based log lines.
- If this gets to be a proper product, I'd add custom metrics for sent and received messages, peer connections, etc. I would track connection and handshake latencies.
