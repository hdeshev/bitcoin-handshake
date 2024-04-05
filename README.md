# bitcoin-handshake

Build a rudimentary client for the Bitcoin P2P binary protocol so we can complete the peer handshake and possibly use it as a starting point for other things.

Architecture description in [ARCHITECTURE](ARCHITECTURE.md)

🚀 Looking for the quickstart? Scroll to the Docker Compose section below.

## Prerequisites

1. Install `golangci-lint`
2. Install `pre-commit`

[Nix](https://nixos.org/) users can use the provided `flake.nix` and enter a dev shell with all dependencies in place by running:

```sh
nix develop
```

Editors with [Direnv](https://direnv.net/) support/plugins (VS Code, Emacs, NeoVim) can pick up the dev shell environment and use the tools in a reproducible manner.

Tested on Linux, but the Nix setup should work on macOS too.

### Installing dependencies

To install Go dependencies associated with `eth-address-watch`, run the
command

```sh
make install
```

### Setting up `pre-commit`

Install the git hooks to run checks before committing via `pre-commit`.

```sh
make local-setup
```

### Using Code Formatters

Format code with

```sh
make codestyle
```

### Using Code Linters

Run code linters with

```sh
make lint
```

### Running Tests

Run tests with

```sh
make test
```

### Running the client

The client entrypoint is `main.go`, which you can build and run directly, but we have several helpers.

#### Running the client (manual)

Local developer run, assuming you have all the dependencies, the Go toolchain and a bitcoin node running on your machine

```sh
make run
```

#### Running the client (Docker Compose)

Docker-based run -- the only dependency is a working Docker installation and Docker Compose

```sh
docker compose run client
```

The `docker-compose.yml` file defines two services:
- `node` running a bitcoin node in regtest mode.
- `client` running the client app.

The command above will start the node, wait for it to boot, and then run the client. You should see it connecting

```
❯ docker compose run client
[+] Building 0.0s (0/0)
[+] Creating 1/0
 ✔ Container bitcoin-handshake-node-1  Running
[+] Building 0.0s (0/0)
2024/04/05 14:45:55 INFO starting bitcoin-handshake
2024/04/05 14:45:55 INFO connecting to bitcoin node address=node:18444
2024/04/05 14:45:55 INFO sending handshake version message
2024/04/05 14:45:55 INFO received handshake version message
2024/04/05 14:45:55 INFO sending handshake verack message
2024/04/05 14:45:55 INFO received handshake verack message
2024/04/05 14:45:55 INFO app received message command=sendcmpct
2024/04/05 14:45:55 INFO app received message command=ping
2024/04/05 14:45:55 INFO app received message command=feefilter
```

The Bitcoin node is running with `-debug=net`, so you should be able to see the connection in its logs

```
❯ docker compose logs node

...

node-1  | 2024-04-05T14:45:55Z New inbound v1 peer connected: version: 70015, blocks=1, peer=1
node-1  | 2024-04-05T14:45:55Z [net] sending sendcmpct (9 bytes) peer=1
node-1  | 2024-04-05T14:45:55Z [net] sending ping (8 bytes) peer=1
node-1  | 2024-04-05T14:45:55Z [net] sending feefilter (8 bytes) peer=1
node-1  | 2024-04-05T14:45:56Z [net] socket closed for peer=1
node-1  | 2024-04-05T14:45:56Z [net] disconnecting peer=1
node-1  | 2024-04-05T14:45:56Z [net] Cleared nodestate for peer=1
```

You can flip to a separate console and ask the Bitcoin node for its peer state to confirm our client has connected using the `getpeerinfo` command

```
❯ bitcoin-cli -regtest -conf="$(pwd)/.bitcoin/bitcoin.conf" getpeerinfo
[
  {
    "id": 2,
    "addr": "192.168.16.3:35492",
    "addrbind": "192.168.16.2:18444",
    "addrlocal": "192.168.16.2:18444",
    "network": "not_publicly_routable",
    "services": "0000000000000000",
    "servicesnames": [
    ],
    "relaytxes": false,
    "lastsend": 1712330273,
    "lastrecv": 1712330273,
    "last_transaction": 0,
    "last_block": 0,
    "bytessent": 247,
    "bytesrecv": 152,
    "conntime": 1712330273,
    "timeoffset": 0,
    "pingwait": 6.55446,
    "version": 70015,
    "subver": "/MemeClient:0.0.1/",
    "inbound": true,
    "bip152_hb_to": false,
    "bip152_hb_from": false,
    "startingheight": 1,
    "presynced_headers": -1,
    "synced_headers": -1,
    "synced_blocks": -1,
    "inflight": [
    ],
    "addr_relay_enabled": false,
    "addr_processed": 0,
    "addr_rate_limited": 0,
    "permissions": [
    ],
    "minfeefilter": 0.00000000,
    "bytessent_per_msg": {
      "feefilter": 32,
      "ping": 32,
      "sendcmpct": 33,
      "verack": 24,
      "version": 126
    },
    "bytesrecv_per_msg": {
      "verack": 24,
      "version": 128
    },
    "connection_type": "inbound",
    "transport_protocol_type": "v1",
    "session_id": ""
  }
]
```
