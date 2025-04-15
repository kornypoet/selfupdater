package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func agent(log *zap.SugaredLogger) {
	log.Infow("Starting agent",
		"version", strings.TrimSpace(version),
		"pid", os.Getpid())

	supervisorPid := os.Getppid()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	for {
		select {
		case s := <-sig:
			if s == syscall.SIGHUP {
				// SIGHUP will come from the supervisor, exit cleanly
				log.Info("Received signal from supervisor, shutting down")
				os.Exit(0)
			}
			// All other signals should be interpreted as fatal
			log.Errorf("Received %s, shutting down", s)
			os.Exit(1)
		default:
			// Exit if we ever lose the supervisor
			if os.Getppid() != supervisorPid {
				log.Fatal("Supervisor pid is gone, shutting down")
			}
			log.Info("Working")
			time.Sleep(1 * time.Second)
		}
	}
}
