.PHONY: build

all: build

format:
	go fmt ./...

build:
	@base=$(shell echo v0.9.0 > agent/VERSION); \
	go build -o dist/v0.9.0 ./agent

build-v1.0.0:
	@update=$(shell echo v1.0.0 > agent/VERSION); \
	go build -o dist/v1.0.0 ./agent

build-v1.0.1:
	@update=$(shell echo v1.0.1 > agent/VERSION); \
	go build -o dist/v1.0.1 ./agent

clean:
	rm -f dist/*
	rm -rf remote/

install: build
	mkdir -p remote
	cp dist/v0.9.0 remote/agent
