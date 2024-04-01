{
  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    systems.url = "github:nix-systems/default";
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, flake-parts, systems, nixpkgs, utils }:
    utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        go = pkgs.go_1_22;
        bitcoin-bin = pkgs.bitcoin;
        shell = pkgs.mkShell {
          buildInputs = [
            go

            bitcoind
            bitcoin-cli
            bitcoin-bin
          ] ++ (with pkgs; [
            gnumake
            pre-commit
            golangci-lint
            jq
            curl
          ]);
        };

        bitcoind = pkgs.writeShellScriptBin "bitcoind" ''
        PROJECT_ROOT=$(git rev-parse --show-toplevel)
        exec ${bitcoin-bin}/bin/bitcoind -regtest -datadir="$PROJECT_ROOT/.bitcoin" -conf="$PROJECT_ROOT/.bitcoin/bitcoin.conf" "$@"
        '';

        bitcoin-cli = pkgs.writeShellScriptBin "bitcoin-cli" ''
        PROJECT_ROOT=$(git rev-parse --show-toplevel)
        exec ${bitcoin-bin}/bin/bitcoin-cli -regtest -conf="$PROJECT_ROOT/.bitcoin/bitcoin.conf" $@
        '';

        shellHook = ''
        echo "self: ${self}"
        '';
      in
      rec {
        name = "flake 1";
        description = "eth-watcher 1";
        devShells.default = shell;
      }
    );
}
