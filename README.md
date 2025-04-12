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

We embed the version file in the binary, so different agents can be identified:

```
make build-v1.0.0
make build-v1.0.1
```

All versions can be removed as well:

```
make clean
```

The server is built and run using docker compose:

```
docker compose up
```

It currently supports these methods:

```
GET /ping               -> {"response":"pong"}
GET /versions           -> {"versions":["v0.9.0","v1.0.0","v1.0.1"]}
GET /latest             -> {"latest":"v1.0.1"}
GET /download/:filename -> Downloads the specified version
```

The `dist/` directory is mounted to the running container, so new versions can be added live to the server

The agent currently loops, printing its current version, sleeping, then checking for an update:

```
./dist/v0.9.0
```
