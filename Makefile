.PHONY: build

all: build

build:
	@current=$(shell cat agent/VERSION); \
	go build -o dist/agent-$$current ./agent
