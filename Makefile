.PHONY: build

all: build

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
	rm dist/*
