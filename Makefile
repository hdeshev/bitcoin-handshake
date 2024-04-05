# Makefile - use `make` or `make help` to get a list of commands.
#
# Note - Comments inside this makefile should be made using a single
# hashtag `#`, lines with double hash-tags will be the messages that
# printed in the help command

# Name of the current directory
PROJECTNAME="bitcoin-handshake"

# List of all Go-files to be processed
GOFILES=$(wildcard *.go)

# Docker image variables
IMAGE := $(PROJECTNAME)
VERSION := latest

# Ensures firing a blank `make` command default to help
.DEFAULT_GOAL := help

# Make is verbose in Linux. Make it silent
MAKEFLAGS += --silent


.PHONY: help
## `help`: Generates this help dialog for the Makefile
help: Makefile
	echo
	echo " Commands available in \`"$(PROJECTNAME)"\`:"
	echo
	sed -n 's/^[ \t]*##//p' $< | column -t -s ':' |  sed -e 's/^//'
	echo


.PHONY: local-setup
## `local-setup`: Setup development environment locally
local-setup:
	echo "  >  Ensuring directory is a git repository"
	git init &> /dev/null
	echo "  >  Installing pre-commit hooks"
	pre-commit install


# Will install missing dependencies
.PHONY: install
## `install`: Fetch dependencies needed to run `bitcoin-handshake`
install:
	echo "  >  Getting dependencies..."
	go get -v $(get)
	go mod tidy


.PHONY: codestyle
## :
## `codestyle`: Run code formatter(s)
codestyle:
	golangci-lint run --fix


.PHONY: lint
## `lint`: Run linters and check code-style
lint:
	golangci-lint run


# No `help` message for this command - designed to be consumed internally
.PHONY: test
test:
	go test ./... -race -covermode=atomic -coverprofile=./coverage/coverage.txt
	go tool cover -html=./coverage/coverage.txt -o ./coverage/coverage.html


.PHONY: run
## :
## `run`: Run `bitcoin-handshake`
run:
	go run main.go $(q)
