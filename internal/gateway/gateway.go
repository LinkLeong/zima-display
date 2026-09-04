package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultRuntimePath = "/var/run/casaos"

type route struct {
	Path   string `json:"path"`
	Target string `json:"target"`
}

func RegisterAsync(ctx context.Context, logger *log.Logger, runtimePath, path, target string) {
	if runtimePath == "" {
		runtimePath = defaultRuntimePath
	}
	go func() {
		for attempt := 1; attempt <= 60; attempt++ {
			if err := register(ctx, runtimePath, path, target); err == nil {
				logger.Printf("gateway route registered: %s -> %s", path, target)
				return
			} else if attempt == 1 || attempt%10 == 0 {
				logger.Printf("gateway route not ready (attempt %d/60): %v", attempt, err)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
		logger.Printf("gateway registration gave up after 60 attempts")
	}()
}

func register(ctx context.Context, runtimePath, path, target string) error {
	addressData, err := os.ReadFile(strings.TrimRight(runtimePath, "/") + "/management.url")
	if err != nil {
		return err
	}
	address := strings.TrimRight(strings.TrimSpace(string(addressData)), "/")
	if address == "" {
		return fmt.Errorf("management.url is empty")
	}
	body, err := json.Marshal(route{Path: path, Target: target})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, address+"/v1/gateway/routes", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("gateway returned %s: %s", resp.Status, strings.TrimSpace(string(message)))
	}
	return nil
}
