# Self Updater

![self modifying code](doc/self-modifying-code.jpg)

This repository contains an example agent and server that demonstrate self-updating capabilities

## Requirements

The following tools were used to develop this:

* golang 1.24
* make
* docker

Additionally, this code was built using an Intel-based Mac, but should work on other OSes

## Setup

The agent is built locally:

```
make
make install
```

We embed the version file in the binary, so different agents can be identified, for example:

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
docker compose up --build
```

It currently supports these methods:

```
GET /ping               -> {"response":"pong"}
GET /versions           -> {"versions":["v0.9.0","v1.0.0","v1.0.1"]}
GET /latest             -> {"latest":"v1.0.1"}
GET /download/:filename -> Downloads the specified version
```

The `dist/` directory is mounted to the running container, so new versions can be added live to the server

## Usage

Once the server is started, the agent can be run from the `remote` directory. We'll simulate the binary being
installed on a client's remote server and assume that we have no additional access or tooling other than
the directory it's installed in:

```
./remote/agent
{"level":"info","ts":1744672808.46314,"caller":"agent/supervisor.go:24","msg":"Starting supervisor","process":"supervisor","version":"v0.9.0","pid":33357}
{"level":"info","ts":1744672808.4633322,"caller":"agent/supervisor.go:44","msg":"Starting agent","process":"supervisor"}
{"level":"info","ts":1744672808.472203,"caller":"agent/agent.go:14","msg":"Starting agent","process":"agent","version":"v0.9.0","pid":33358}
{"level":"info","ts":1744672808.472416,"caller":"agent/agent.go:39","msg":"Working","process":"agent"}
{"level":"info","ts":1744672809.473462,"caller":"agent/agent.go:39","msg":"Working","process":"agent"}
{"level":"info","ts":1744672810.4743838,"caller":"agent/agent.go:39","msg":"Working","process":"agent"}
{"level":"info","ts":1744672832.464385,"caller":"agent/supervisor.go:106","msg":"Checking for update","process":"supervisor"}
{"level":"info","ts":1744672832.467602,"caller":"agent/supervisor.go:154","msg":"Latest version available: v0.9.0","process":"supervisor"}
```

The program starts a supervisor process first, which then starts an agent process, using the same binary with an env variable to indicate the behavior.
The agent process runs in a loop and "works" each second. The supervisor checks the server for updates every three seconds, comparing the version it receives back from `/latest` and the current version embedded in the binary. If the latest version is newer, it is downloaded into the `remote/` directory, the running binary is renamed to `agent.old`, the new binary is renamed to `agent` and the supervisor sends an OS syscall to restart the agent. If successful, the new version is updated in memory for the supervisor, which will continue to check for updates.


```
# In a separate window
make build-v1.0.0
```

```
{"level":"info","ts":1744681579.589848,"caller":"agent/supervisor.go:154","msg":"Latest version available: v1.0.0","process":"supervisor"}
{"level":"info","ts":1744681579.589886,"caller":"agent/supervisor.go:108","msg":"Newer version available","process":"supervisor"}
{"level":"info","ts":1744681579.638844,"caller":"agent/agent.go:39","msg":"Working","process":"agent"}
{"level":"info","ts":1744681579.64734,"caller":"agent/supervisor.go:199","msg":"Downloaded file","process":"supervisor"}
{"level":"info","ts":1744681579.647647,"caller":"agent/supervisor.go:215","msg":"Updated running version to v1.0.0","process":"supervisor"}
{"level":"info","ts":1744681579.647857,"caller":"agent/supervisor.go:114","msg":"Reloading agent","process":"supervisor"}
{"level":"info","ts":1744681579.6478798,"caller":"agent/supervisor.go:83","msg":"Update completed, restarting agent","process":"supervisor"}
{"level":"info","ts":1744681580.6397262,"caller":"agent/agent.go:28","msg":"Received signal from supervisor, shutting down","process":"agent"}
{"level":"info","ts":1744681580.641085,"caller":"agent/supervisor.go:44","msg":"Starting agent","process":"supervisor"}
{"level":"info","ts":1744681581.0677052,"caller":"agent/agent.go:14","msg":"Starting agent","process":"agent","version":"v1.0.0","pid":34644}
{"level":"info","ts":1744681581.0679371,"caller":"agent/agent.go:39","msg":"Working","process":"agent"}
{"level":"info","ts":1744681582.06807,"caller":"agent/agent.go:39","msg":"Working","process":"agent"}
{"level":"info","ts":1744681582.5858371,"caller":"agent/supervisor.go:106","msg":"Checking for update","process":"supervisor"}
{"level":"info","ts":1744681582.5889149,"caller":"agent/supervisor.go:154","msg":"Latest version available: v1.0.0","process":"supervisor"}
```

## Design

In the above example, the supervisor process will continue to update the agent subprocess as long as there are updates available. If an update fails during any step other than the final restart of the agent, the supervisor will continue and try and perform the update again on the next loop. If the update succeeds, but for some reason the new binary crashes, the previous version is left on disk in the remote directory and can be re-run as `agent.old`.

I intentionally restricted myself to a solution that required only a single binary as both an attempt to rely on nothing external from the client's side, and also as a way to increase the portability of the demo. In a production setting, I think the best solution would be to leverage either a package manager or an install script that would be able to install additional tools and/or dependencies, and would leverage SystemD and cron (or similar) to handle the operational load of updating the binary and switching between versions. This would come with the tradeoff in managing OS and tooling compatibility to enable more robust control over updating.

Additional considerations that were not included in the demo include PGP signing, which is an important security step when downloading and executing code without human interaction. The demo also does not mark versions as bad if/when they fail; in a production setting these should not be retried and beyond that, some form of telemetry should probably be sent out indicating a version was excluded.