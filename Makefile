# Makefile for docker-machine-driver-ovh

.PHONY: all build test test-coverage test-integration lint clean install deps

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

test-coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration:
	@echo "Running integration tests (requires OVH credentials)..."
	$(GOTEST) -v -tags=integration ./...

lint:
	golangci-lint run

clean:
	$(GOCLEAN)
	rm -f $(BINARY)
	rm -f coverage.out coverage.html

install:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(GOPATH)/bin/$(BINARY) .

deps:
	$(GOMOD) download
	$(GOMOD) tidy

.DEFAULT_GOAL := build
