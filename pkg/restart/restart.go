package restart

import (
	"fmt"
	"sync"
	"time"

	"github.com/therealutkarshpriyadarshi/containr/pkg/events"
	"github.com/therealutkarshpriyadarshi/containr/pkg/logger"
)

// Policy represents a restart policy
type Policy string

const (
	// PolicyNo - Never restart container
	PolicyNo Policy = "no"
	// PolicyAlways - Always restart container
	PolicyAlways Policy = "always"
	// PolicyOnFailure - Restart only on non-zero exit
	PolicyOnFailure Policy = "on-failure"
	// PolicyUnlessStopped - Always restart unless explicitly stopped
	PolicyUnlessStopped Policy = "unless-stopped"
)

// Config holds restart policy configuration
type Config struct {
	Policy            Policy
	MaxRetries        int           // Maximum restart attempts (0 = unlimited)
	RestartDelay      time.Duration // Delay between restarts
	BackoffMultiplier float64       // Backoff multiplier for exponential backoff
	MaxDelay          time.Duration // Maximum delay between restarts
}

// Manager manages container restart policies
type Manager struct {
	config       *Config
	containerID  string
	restartCount int
	stopChan     chan struct{}
	eventManager *events.EventManager
	restartFunc  func() error // Function to call to restart the container
	mu           sync.RWMutex
	log          *logger.Logger
}

// NewManager creates a new restart policy manager
func NewManager(containerID string, config *Config, eventManager *events.EventManager) *Manager {
	if config == nil {
		config = &Config{
			Policy: PolicyNo,
		}
	}

	// Set defaults
	if config.RestartDelay == 0 {
		config.RestartDelay = 100 * time.Millisecond
	}
	if config.BackoffMultiplier == 0 {
		config.BackoffMultiplier = 2.0
	}
	if config.MaxDelay == 0 {
		config.MaxDelay = 1 * time.Minute
	}

	return &Manager{
		config:       config,
		containerID:  containerID,
		stopChan:     make(chan struct{}),
		eventManager: eventManager,
		log:          logger.New("restart"),
	}
}

// SetRestartFunc sets the function to call to restart the container
func (m *Manager) SetRestartFunc(fn func() error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.restartFunc = fn
}

// ShouldRestart determines if a container should be restarted based on exit code
func (m *Manager) ShouldRestart(exitCode int, manuallyStopped bool) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch m.config.Policy {
	case PolicyNo:
		return false

	case PolicyAlways:
		return true

	case PolicyOnFailure:
		// Only restart on non-zero exit
		if exitCode == 0 {
			return false
		}
		// Check max retries
		if m.config.MaxRetries > 0 && m.restartCount >= m.config.MaxRetries {
			m.log.Infof("Max restart attempts (%d) reached", m.config.MaxRetries)
			return false
		}
		return true

	case PolicyUnlessStopped:
		// Don't restart if manually stopped
		return !manuallyStopped

	default:
		return false
	}
}

// HandleExit handles container exit and performs restart if needed
func (m *Manager) HandleExit(exitCode int, manuallyStopped bool) {
	if !m.ShouldRestart(exitCode, manuallyStopped) {
		m.log.Debugf("Container %s will not be restarted (policy=%s, exit=%d, manual=%v)",
			m.containerID, m.config.Policy, exitCode, manuallyStopped)
		return
	}

	m.log.Infof("Container %s will be restarted (attempt %d, policy=%s, exit=%d)",
		m.containerID, m.restartCount+1, m.config.Policy, exitCode)

	// Calculate delay with exponential backoff
	delay := m.calculateDelay()

	m.log.Debugf("Waiting %v before restart", delay)
	time.Sleep(delay)

	// Perform restart
	if err := m.performRestart(); err != nil {
		m.log.Errorf("Failed to restart container: %v", err)

		// Emit error event
		if m.eventManager != nil {
			event := events.CreateErrorEvent(events.EventContainerRestart, m.containerID, err)
			m.eventManager.Emit(event)
		}
	}
}

// performRestart performs the actual container restart
func (m *Manager) performRestart() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.restartFunc == nil {
		return fmt.Errorf("restart function not set")
	}

	m.restartCount++

	m.log.Infof("Restarting container %s (attempt %d)", m.containerID, m.restartCount)

	// Emit restart event
	if m.eventManager != nil {
		event := events.CreateEvent(events.EventContainerRestart, m.containerID, "", "", map[string]string{
			"restart_count": fmt.Sprintf("%d", m.restartCount),
			"policy":        string(m.config.Policy),
		})
		m.eventManager.Emit(event)
	}

	// Call restart function
	if err := m.restartFunc(); err != nil {
		return fmt.Errorf("restart failed: %w", err)
	}

	m.log.Infof("Container %s restarted successfully", m.containerID)
	return nil
}

// calculateDelay calculates the restart delay with exponential backoff
func (m *Manager) calculateDelay() time.Duration {
	delay := m.config.RestartDelay

	// Apply exponential backoff
	for i := 0; i < m.restartCount; i++ {
		delay = time.Duration(float64(delay) * m.config.BackoffMultiplier)
		if delay > m.config.MaxDelay {
			delay = m.config.MaxDelay
			break
		}
	}

	return delay
}

// GetRestartCount returns the current restart count
func (m *Manager) GetRestartCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.restartCount
}

// ResetRestartCount resets the restart count
func (m *Manager) ResetRestartCount() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.restartCount = 0
	m.log.Debugf("Reset restart count for container %s", m.containerID)
}

// Stop stops the restart manager
func (m *Manager) Stop() {
	close(m.stopChan)
}

// ParsePolicy parses a restart policy string
func ParsePolicy(policyStr string) (Policy, error) {
	switch policyStr {
	case "no", "":
		return PolicyNo, nil
	case "always":
		return PolicyAlways, nil
	case "on-failure":
		return PolicyOnFailure, nil
	case "unless-stopped":
		return PolicyUnlessStopped, nil
	default:
		return PolicyNo, fmt.Errorf("invalid restart policy: %s (valid: no, always, on-failure, unless-stopped)", policyStr)
	}
}

// DefaultConfig returns a default restart configuration
func DefaultConfig() *Config {
	return &Config{
		Policy:            PolicyNo,
		MaxRetries:        0,
		RestartDelay:      100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		MaxDelay:          1 * time.Minute,
	}
}
