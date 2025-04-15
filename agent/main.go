package main

import (
	_ "embed"
	"os"

	"go.uber.org/zap"
)

//go:embed VERSION
var version string

const agentEnv = "AGENT_PROC"

// Setup logging and start supervisor
// Start agent instead if env var is set
func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	sugar := logger.Sugar()
	agentLogger := sugar.With(zap.String("process", "agent"))
	supervisorLogger := sugar.With(zap.String("process", "supervisor"))

	if os.Getenv(agentEnv) == "1" {
		agent(agentLogger)
	} else {
		supervisor(supervisorLogger)
	}
}
