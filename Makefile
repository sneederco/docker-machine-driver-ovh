# Makefile for docker-machine-driver-ovh

.PHONY: all build test lint clean install deps

# Binary name
BINARY := docker-machine-driver-ovh

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build flags
LDFLAGS := -w -s

all: build

build:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	$(GOTEST) -v ./...

lint:
	golangci-lint run

clean:
	$(GOCLEAN)
	rm -f $(BINARY)

install:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(GOPATH)/bin/$(BINARY) .

deps:
	$(GOMOD) download
	$(GOMOD) tidy

.DEFAULT_GOAL := build
