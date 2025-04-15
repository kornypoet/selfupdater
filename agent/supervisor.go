package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/mod/semver"
)

const serverURL = "http://localhost:8080"

func supervisor(log *zap.SugaredLogger) {
	log.Infow("Starting supervisor",
		"version", strings.TrimSpace(version),
		"pid", os.Getpid())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	reloadChan := make(chan bool, 1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		updateHandler(context.WithValue(ctx, "log", log), reloadChan)
	}()

	for loop := true; loop; {
		log.Info("Starting agent")
		agentCtx, agentCancel := context.WithCancel(ctx)

		cmd := exec.CommandContext(agentCtx, os.Args[0])
		cmd.Env = append(os.Environ(), agentEnv+"=1")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Start(); err != nil {
			log.Errorw("Failed to start agent", "error", err)
			agentCancel()
			break
		}

		exitChan := make(chan error, 1)
		wg.Add(1)
		go func() {
			defer wg.Done()
			exitChan <- cmd.Wait()
		}()

		select {
		case sig := <-sigChan:
			log.Infof("Signal received: %s. Passing through to agent first", sig)
			cmd.Process.Signal(sig)
			<-exitChan
			agentCancel()
			loop = false
			break

		case err := <-exitChan:
			if err != nil {
				log.Infow("Agent exited with error", "error", err)
			}
			agentCancel()
			log.Info("Agent exited outside of the supervisor, restarting")

		case <-reloadChan:
			log.Info("Update completed, restarting agent")
			cmd.Process.Signal(syscall.SIGHUP)
			<-exitChan
			agentCancel()
		}
	}

	cancel()
	wg.Wait()
	log.Info("Exiting all")
}

func updateHandler(ctx context.Context, reload chan bool) {
	log, _ := ctx.Value("log").(*zap.SugaredLogger)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Update handler done")
			return
		case <-ticker.C:
			log.Info("Checking for update")
			if ver := updateAvailable(ctx); ver != "" {
				log.Info("Newer version available")
				err := downloadUpdate(ctx, ver)
				if err != nil {
					log.Errorw("Error downloading update", "error", err)
				} else {
					select {
					case reload <- true:
						log.Info("Reloading agent")
					default:
						log.Info("Reload already in progress")
					}
				}
			}
		}
	}
}

type ApiResponse struct {
	Latest string `json:"latest"`
}

func updateAvailable(ctx context.Context) (ver string) {
	log, _ := ctx.Value("log").(*zap.SugaredLogger)
	url, _ := url.JoinPath(serverURL, "latest")

	resp, err := http.Get(url)
	if err != nil {
		log.Errorw("Failed to GET /latest", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("Unexpected status code: %d", resp.StatusCode)
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var data ApiResponse
	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorw("Failed to parse server response", "error", err)
	}

	log.Infof("Latest version available: %s", data.Latest)
	if semver.Compare(strings.TrimSpace(version), data.Latest) == -1 {
		return data.Latest
	}
	return
}

func downloadUpdate(ctx context.Context, ver string) (err error) {
	log, _ := ctx.Value("log").(*zap.SugaredLogger)
	url, _ := url.JoinPath(serverURL, "download", ver)

	execPath, _ := os.Executable()
	outPath := filepath.Join(filepath.Dir(execPath), ver)

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return
	}

	err = os.Chmod(outPath, 0755)
	if err != nil {
		return
	}

	log.Info("Downloaded updated version")

	oldPath := filepath.Join(filepath.Dir(execPath), fmt.Sprintf("%s.old", filepath.Base(execPath)))
	err = os.Rename(execPath, oldPath)
	if err != nil {
		return
	}

	err = os.Rename(outPath, execPath)
	if err != nil {
		log.Fatal("Failed to replace executable with new download; manual intervention needed")
	}

	version = ver
	log.Infof("Updated running version to %s", version)
	return
}
