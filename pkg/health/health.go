package health

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/events"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// HealthStatus represents the health status of a container
type HealthStatus string

const (
	// HealthStatusHealthy indicates container is healthy
	HealthStatusHealthy HealthStatus = "healthy"
	// HealthStatusUnhealthy indicates container is unhealthy
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	// HealthStatusStarting indicates health check hasn't completed yet
	HealthStatusStarting HealthStatus = "starting"
	// HealthStatusNone indicates no health check configured
	HealthStatusNone HealthStatus = "none"
)

// HealthCheck configuration
type HealthCheck struct {
	Test        []string      // Command to run (e.g., ["CMD", "/bin/check"])
	Interval    time.Duration // Time between checks (default: 30s)
	Timeout     time.Duration // Time before check is considered failed (default: 30s)
	StartPeriod time.Duration // Grace period before checks start (default: 0s)
	Retries     int           // Consecutive failures needed to report unhealthy (default: 3)
}

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status    HealthStatus
	Output    string
	ExitCode  int
	Timestamp time.Time
}

// HealthMonitor monitors container health
type HealthMonitor struct {
	containerID  string
	config       *HealthCheck
	status       HealthStatus
	results      []*HealthCheckResult
	failures     int
	stopChan     chan struct{}
	eventManager *events.EventManager
	mu           sync.RWMutex
	log          *logger.Logger
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(containerID string, config *HealthCheck, eventManager *events.EventManager) *HealthMonitor {
	if config == nil {
		return &HealthMonitor{
			containerID: containerID,
			status:      HealthStatusNone,
			log:         logger.New("health"),
		}
	}

	// Set defaults
	if config.Interval == 0 {
		config.Interval = 30 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.Retries == 0 {
		config.Retries = 3
	}

	return &HealthMonitor{
		containerID:  containerID,
		config:       config,
		status:       HealthStatusStarting,
		results:      make([]*HealthCheckResult, 0),
		stopChan:     make(chan struct{}),
		eventManager: eventManager,
		log:          logger.New("health"),
	}
}

// Start starts the health monitoring
func (hm *HealthMonitor) Start() {
	if hm.config == nil {
		return
	}

	hm.log.Infof("Starting health monitoring for container %s", hm.containerID)

	// Wait for start period
	if hm.config.StartPeriod > 0 {
		hm.log.Debugf("Waiting %v before starting health checks", hm.config.StartPeriod)
		time.Sleep(hm.config.StartPeriod)
	}

	// Start monitoring loop
	go hm.monitorLoop()
}

// Stop stops the health monitoring
func (hm *HealthMonitor) Stop() {
	if hm.config == nil {
		return
	}

	close(hm.stopChan)
	hm.log.Infof("Stopped health monitoring for container %s", hm.containerID)
}

// GetStatus returns the current health status
func (hm *HealthMonitor) GetStatus() HealthStatus {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.status
}

// GetLastResult returns the most recent health check result
func (hm *HealthMonitor) GetLastResult() *HealthCheckResult {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if len(hm.results) == 0 {
		return nil
	}

	return hm.results[len(hm.results)-1]
}

// monitorLoop runs health checks periodically
func (hm *HealthMonitor) monitorLoop() {
	ticker := time.NewTicker(hm.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-hm.stopChan:
			return
		case <-ticker.C:
			hm.runHealthCheck()
		}
	}
}

// runHealthCheck executes a single health check
func (hm *HealthMonitor) runHealthCheck() {
	hm.log.Debugf("Running health check for container %s", hm.containerID)

	result := &HealthCheckResult{
		Timestamp: time.Now(),
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), hm.config.Timeout)
	defer cancel()

	// Parse and execute command
	if len(hm.config.Test) == 0 {
		hm.log.Warn("No health check command configured")
		return
	}

	var cmd *exec.Cmd
	if hm.config.Test[0] == "CMD" {
		// Execute command directly
		if len(hm.config.Test) < 2 {
			hm.log.Warn("Invalid health check command")
			return
		}
		cmd = exec.CommandContext(ctx, hm.config.Test[1], hm.config.Test[2:]...)
	} else if hm.config.Test[0] == "CMD-SHELL" {
		// Execute via shell
		if len(hm.config.Test) < 2 {
			hm.log.Warn("Invalid health check command")
			return
		}
		cmd = exec.CommandContext(ctx, "/bin/sh", "-c", hm.config.Test[1])
	} else {
		hm.log.Warnf("Unknown health check type: %s", hm.config.Test[0])
		return
	}

	// Run command
	output, err := cmd.CombinedOutput()
	result.Output = strings.TrimSpace(string(output))

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
		}
		result.Status = HealthStatusUnhealthy
		hm.failures++
	} else {
		result.ExitCode = 0
		result.Status = HealthStatusHealthy
		hm.failures = 0
	}

	// Store result
	hm.mu.Lock()
	hm.results = append(hm.results, result)
	// Keep only last 10 results
	if len(hm.results) > 10 {
		hm.results = hm.results[len(hm.results)-10:]
	}
	hm.mu.Unlock()

	// Update overall status
	hm.updateStatus(result)

	hm.log.Debugf("Health check result: status=%s, exit=%d, failures=%d",
		result.Status, result.ExitCode, hm.failures)
}

// updateStatus updates the overall health status based on check results
func (hm *HealthMonitor) updateStatus(result *HealthCheckResult) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	oldStatus := hm.status

	// Determine new status based on consecutive failures
	if result.Status == HealthStatusHealthy {
		hm.status = HealthStatusHealthy
	} else if hm.failures >= hm.config.Retries {
		hm.status = HealthStatusUnhealthy
	}

	// Emit event if status changed
	if oldStatus != hm.status && hm.eventManager != nil {
		var eventType events.EventType
		if hm.status == HealthStatusHealthy {
			eventType = events.EventContainerHealthy
		} else {
			eventType = events.EventContainerUnhealthy
		}

		event := events.CreateEvent(eventType, hm.containerID, "", "", map[string]string{
			"previous_status": string(oldStatus),
			"current_status":  string(hm.status),
		})
		hm.eventManager.Emit(event)

		hm.log.Infof("Health status changed: %s -> %s", oldStatus, hm.status)
	}
}

// ParseHealthCheck parses a health check configuration from command flags
func ParseHealthCheck(cmd []string, interval, timeout, startPeriod time.Duration, retries int) *HealthCheck {
	if len(cmd) == 0 {
		return nil
	}

	return &HealthCheck{
		Test:        cmd,
		Interval:    interval,
		Timeout:     timeout,
		StartPeriod: startPeriod,
		Retries:     retries,
	}
}
