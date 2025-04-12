# Self Updater

![self modifying code](doc/self-modifying-code.jpg)

This repository contains an example agent and server that demonstrate self-updating capabilities

## Requirements

The following tools were used to develop this:

* golang 1.24
* make
* docker

Additionally, this code was built using an Intel-based Mac, but should work on other OSes

## Usage

The agent is built locally:

```
make build
```

The server is built and run using docker compose:

```
docker compose up
```
