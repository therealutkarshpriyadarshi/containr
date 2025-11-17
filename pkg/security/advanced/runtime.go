package advanced

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RuntimeMonitor monitors container runtime security
type RuntimeMonitor struct {
	config      *RuntimeConfig
	detectors   map[string]ThreatDetector
	alerts      chan *SecurityAlert
	incidents   []*SecurityIncident
	mu          sync.RWMutex
	stopCh      chan struct{}
	running     bool
	subscribers []AlertSubscriber
}

// RuntimeConfig holds runtime security monitoring configuration
type RuntimeConfig struct {
	// EnableBehaviorMonitoring enables container behavior monitoring
	EnableBehaviorMonitoring bool
	// EnableFileIntegrity enables file integrity monitoring
	EnableFileIntegrity bool
	// EnableNetworkMonitoring enables network activity monitoring
	EnableNetworkMonitoring bool
	// EnableProcessMonitoring enables process execution monitoring
	EnableProcessMonitoring bool
	// AlertThreshold is the threshold for raising alerts
	AlertThreshold int
	// MonitoringInterval is the interval for monitoring checks
	MonitoringInterval time.Duration
	// MaxIncidents is the maximum number of incidents to keep in memory
	MaxIncidents int
}

// ThreatDetector is an interface for threat detection modules
type ThreatDetector interface {
	// Name returns the detector name
	Name() string
	// Detect performs threat detection
	Detect(ctx context.Context, event *RuntimeEvent) (*ThreatDetection, error)
	// Start starts the detector
	Start(ctx context.Context) error
	// Stop stops the detector
	Stop() error
}

// RuntimeEvent represents a runtime security event
type RuntimeEvent struct {
	Type        EventType              `json:"type"`
	Timestamp   time.Time              `json:"timestamp"`
	ContainerID string                 `json:"container_id"`
	ProcessID   int                    `json:"process_id,omitempty"`
	User        string                 `json:"user,omitempty"`
	Command     string                 `json:"command,omitempty"`
	FilePath    string                 `json:"file_path,omitempty"`
	Network     *NetworkEvent          `json:"network,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// EventType represents the type of runtime event
type EventType string

const (
	EventTypeProcess      EventType = "process"
	EventTypeFile         EventType = "file"
	EventTypeNetwork      EventType = "network"
	EventTypeSystemCall   EventType = "syscall"
	EventTypeCapability   EventType = "capability"
	EventTypePrivilege    EventType = "privilege"
)

// NetworkEvent represents a network-related event
type NetworkEvent struct {
	Protocol    string `json:"protocol"`
	SourceIP    string `json:"source_ip"`
	SourcePort  int    `json:"source_port"`
	DestIP      string `json:"dest_ip"`
	DestPort    int    `json:"dest_port"`
	BytesSent   int64  `json:"bytes_sent"`
	BytesRecv   int64  `json:"bytes_recv"`
}

// ThreatDetection represents a detected threat
type ThreatDetection struct {
	Severity    Severity               `json:"severity"`
	ThreatType  ThreatType             `json:"threat_type"`
	Description string                 `json:"description"`
	Confidence  float64                `json:"confidence"`
	Evidence    []string               `json:"evidence"`
	Mitigations []string               `json:"mitigations"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Severity represents the severity level of a threat
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// ThreatType represents the type of threat
type ThreatType string

const (
	ThreatTypePrivilegeEscalation ThreatType = "privilege_escalation"
	ThreatTypeMaliciousProcess    ThreatType = "malicious_process"
	ThreatTypeAnomalousNetwork    ThreatType = "anomalous_network"
	ThreatTypeFileModification    ThreatType = "file_modification"
	ThreatTypeCryptoMining        ThreatType = "crypto_mining"
	ThreatTypeDataExfiltration    ThreatType = "data_exfiltration"
	ThreatTypeRootkit             ThreatType = "rootkit"
)

// SecurityAlert represents a security alert
type SecurityAlert struct {
	ID          string                 `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    Severity               `json:"severity"`
	Type        ThreatType             `json:"type"`
	ContainerID string                 `json:"container_id"`
	Event       *RuntimeEvent          `json:"event"`
	Detection   *ThreatDetection       `json:"detection"`
	Status      AlertStatus            `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AlertStatus represents the status of an alert
type AlertStatus string

const (
	AlertStatusNew          AlertStatus = "new"
	AlertStatusInvestigating AlertStatus = "investigating"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusIgnored      AlertStatus = "ignored"
)

// SecurityIncident represents a security incident
type SecurityIncident struct {
	ID          string           `json:"id"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     time.Time        `json:"end_time,omitempty"`
	Severity    Severity         `json:"severity"`
	ContainerID string           `json:"container_id"`
	Alerts      []*SecurityAlert `json:"alerts"`
	Description string           `json:"description"`
	Status      IncidentStatus   `json:"status"`
	Response    string           `json:"response,omitempty"`
}

// IncidentStatus represents the status of an incident
type IncidentStatus string

const (
	IncidentStatusOpen       IncidentStatus = "open"
	IncidentStatusContained  IncidentStatus = "contained"
	IncidentStatusMitigated  IncidentStatus = "mitigated"
	IncidentStatusResolved   IncidentStatus = "resolved"
)

// AlertSubscriber receives security alerts
type AlertSubscriber interface {
	OnAlert(alert *SecurityAlert) error
}

// NewRuntimeMonitor creates a new runtime security monitor
func NewRuntimeMonitor(config *RuntimeConfig) (*RuntimeMonitor, error) {
	if config == nil {
		config = defaultRuntimeConfig()
	}

	monitor := &RuntimeMonitor{
		config:      config,
		detectors:   make(map[string]ThreatDetector),
		alerts:      make(chan *SecurityAlert, 100),
		incidents:   make([]*SecurityIncident, 0),
		stopCh:      make(chan struct{}),
		subscribers: make([]AlertSubscriber, 0),
	}

	// Initialize default detectors
	if config.EnableProcessMonitoring {
		monitor.AddDetector(&ProcessThreatDetector{})
	}
	if config.EnableFileIntegrity {
		monitor.AddDetector(&FileIntegrityDetector{})
	}
	if config.EnableNetworkMonitoring {
		monitor.AddDetector(&NetworkThreatDetector{})
	}

	return monitor, nil
}

// defaultRuntimeConfig returns default runtime configuration
func defaultRuntimeConfig() *RuntimeConfig {
	return &RuntimeConfig{
		EnableBehaviorMonitoring: true,
		EnableFileIntegrity:      true,
		EnableNetworkMonitoring:  true,
		EnableProcessMonitoring:  true,
		AlertThreshold:           3,
		MonitoringInterval:       10 * time.Second,
		MaxIncidents:             1000,
	}
}

// Start starts the runtime monitor
func (m *RuntimeMonitor) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("monitor already running")
	}
	m.running = true
	m.mu.Unlock()

	// Start all detectors
	for _, detector := range m.detectors {
		if err := detector.Start(ctx); err != nil {
			return fmt.Errorf("failed to start detector %s: %w", detector.Name(), err)
		}
	}

	// Start alert processor
	go m.processAlerts(ctx)

	return nil
}

// Stop stops the runtime monitor
func (m *RuntimeMonitor) Stop() error {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return fmt.Errorf("monitor not running")
	}
	m.running = false
	m.mu.Unlock()

	// Stop all detectors
	for _, detector := range m.detectors {
		if err := detector.Stop(); err != nil {
			return fmt.Errorf("failed to stop detector %s: %w", detector.Name(), err)
		}
	}

	close(m.stopCh)
	close(m.alerts)

	return nil
}

// ProcessEvent processes a runtime event
func (m *RuntimeMonitor) ProcessEvent(ctx context.Context, event *RuntimeEvent) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Run event through all detectors
	for _, detector := range m.detectors {
		detection, err := detector.Detect(ctx, event)
		if err != nil {
			// Log error but continue with other detectors
			continue
		}

		if detection != nil {
			// Create alert
			alert := &SecurityAlert{
				ID:          generateAlertID(),
				Timestamp:   time.Now(),
				Severity:    detection.Severity,
				Type:        detection.ThreatType,
				ContainerID: event.ContainerID,
				Event:       event,
				Detection:   detection,
				Status:      AlertStatusNew,
			}

			// Send alert
			select {
			case m.alerts <- alert:
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Alert channel full, log warning
			}
		}
	}

	return nil
}

// processAlerts processes security alerts
func (m *RuntimeMonitor) processAlerts(ctx context.Context) {
	for {
		select {
		case alert, ok := <-m.alerts:
			if !ok {
				return
			}
			m.handleAlert(ctx, alert)
		case <-m.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// handleAlert handles a security alert
func (m *RuntimeMonitor) handleAlert(ctx context.Context, alert *SecurityAlert) {
	// Notify subscribers
	for _, subscriber := range m.subscribers {
		if err := subscriber.OnAlert(alert); err != nil {
			// Log error but continue with other subscribers
		}
	}

	// Check if alert should create an incident
	if alert.Severity == SeverityCritical || alert.Severity == SeverityHigh {
		m.createIncident(alert)
	}
}

// createIncident creates a security incident from an alert
func (m *RuntimeMonitor) createIncident(alert *SecurityAlert) {
	m.mu.Lock()
	defer m.mu.Unlock()

	description := fmt.Sprintf("%s detected", alert.Type)
	if alert.Detection != nil {
		description = fmt.Sprintf("%s detected: %s", alert.Type, alert.Detection.Description)
	}

	incident := &SecurityIncident{
		ID:          generateIncidentID(),
		StartTime:   alert.Timestamp,
		Severity:    alert.Severity,
		ContainerID: alert.ContainerID,
		Alerts:      []*SecurityAlert{alert},
		Description: description,
		Status:      IncidentStatusOpen,
	}

	m.incidents = append(m.incidents, incident)

	// Trim incidents if exceeding max
	if len(m.incidents) > m.config.MaxIncidents {
		m.incidents = m.incidents[len(m.incidents)-m.config.MaxIncidents:]
	}
}

// AddDetector adds a threat detector
func (m *RuntimeMonitor) AddDetector(detector ThreatDetector) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.detectors[detector.Name()] = detector
}

// RemoveDetector removes a threat detector
func (m *RuntimeMonitor) RemoveDetector(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.detectors, name)
}

// Subscribe adds an alert subscriber
func (m *RuntimeMonitor) Subscribe(subscriber AlertSubscriber) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.subscribers = append(m.subscribers, subscriber)
}

// GetIncidents returns recent security incidents
func (m *RuntimeMonitor) GetIncidents(limit int) []*SecurityIncident {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.incidents) {
		limit = len(m.incidents)
	}

	result := make([]*SecurityIncident, limit)
	copy(result, m.incidents[len(m.incidents)-limit:])

	return result
}

// GetMetrics returns runtime monitoring metrics
func (m *RuntimeMonitor) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make(map[string]interface{})
	metrics["running"] = m.running
	metrics["total_detectors"] = len(m.detectors)
	metrics["total_incidents"] = len(m.incidents)
	metrics["alert_queue_size"] = len(m.alerts)

	return metrics
}

// ProcessThreatDetector detects process-based threats
type ProcessThreatDetector struct{}

func (d *ProcessThreatDetector) Name() string {
	return "process_threat_detector"
}

func (d *ProcessThreatDetector) Detect(ctx context.Context, event *RuntimeEvent) (*ThreatDetection, error) {
	if event.Type != EventTypeProcess {
		return nil, nil
	}

	// Detect suspicious processes
	suspiciousCommands := []string{
		"/bin/nc",         // netcat
		"/usr/bin/nmap",   // nmap
		"/usr/bin/curl",   // curl (suspicious if connecting to unknown IPs)
		"cryptominer",     // crypto miners
		"/tmp/",           // execution from tmp
	}

	for _, cmd := range suspiciousCommands {
		if contains(event.Command, cmd) {
			return &ThreatDetection{
				Severity:    SeverityHigh,
				ThreatType:  ThreatTypeMaliciousProcess,
				Description: fmt.Sprintf("Suspicious process detected: %s", event.Command),
				Confidence:  0.8,
				Evidence:    []string{fmt.Sprintf("Command: %s", event.Command)},
				Mitigations: []string{"Terminate process", "Review container image"},
			}, nil
		}
	}

	return nil, nil
}

func (d *ProcessThreatDetector) Start(ctx context.Context) error {
	return nil
}

func (d *ProcessThreatDetector) Stop() error {
	return nil
}

// FileIntegrityDetector detects file integrity violations
type FileIntegrityDetector struct{}

func (d *FileIntegrityDetector) Name() string {
	return "file_integrity_detector"
}

func (d *FileIntegrityDetector) Detect(ctx context.Context, event *RuntimeEvent) (*ThreatDetection, error) {
	if event.Type != EventTypeFile {
		return nil, nil
	}

	// Detect modifications to critical files
	criticalPaths := []string{
		"/etc/passwd",
		"/etc/shadow",
		"/etc/sudoers",
		"/root/.ssh/",
		"/bin/",
		"/sbin/",
	}

	for _, path := range criticalPaths {
		if contains(event.FilePath, path) {
			return &ThreatDetection{
				Severity:    SeverityCritical,
				ThreatType:  ThreatTypeFileModification,
				Description: fmt.Sprintf("Critical file modified: %s", event.FilePath),
				Confidence:  0.95,
				Evidence:    []string{fmt.Sprintf("File: %s", event.FilePath)},
				Mitigations: []string{"Restore file from backup", "Investigate container"},
			}, nil
		}
	}

	return nil, nil
}

func (d *FileIntegrityDetector) Start(ctx context.Context) error {
	return nil
}

func (d *FileIntegrityDetector) Stop() error {
	return nil
}

// NetworkThreatDetector detects network-based threats
type NetworkThreatDetector struct{}

func (d *NetworkThreatDetector) Name() string {
	return "network_threat_detector"
}

func (d *NetworkThreatDetector) Detect(ctx context.Context, event *RuntimeEvent) (*ThreatDetection, error) {
	if event.Type != EventTypeNetwork || event.Network == nil {
		return nil, nil
	}

	// Detect suspicious network activity
	// Example: connections to crypto mining pools, data exfiltration

	// Detect unusual ports
	suspiciousPorts := []int{4444, 1337, 31337} // Common backdoor ports

	for _, port := range suspiciousPorts {
		if event.Network.DestPort == port {
			return &ThreatDetection{
				Severity:    SeverityHigh,
				ThreatType:  ThreatTypeAnomalousNetwork,
				Description: fmt.Sprintf("Connection to suspicious port: %d", port),
				Confidence:  0.7,
				Evidence:    []string{fmt.Sprintf("Destination: %s:%d", event.Network.DestIP, event.Network.DestPort)},
				Mitigations: []string{"Block network connection", "Review container network policy"},
			}, nil
		}
	}

	// Detect large data transfers (potential exfiltration)
	if event.Network.BytesSent > 100*1024*1024 { // 100MB
		return &ThreatDetection{
			Severity:    SeverityMedium,
			ThreatType:  ThreatTypeDataExfiltration,
			Description: fmt.Sprintf("Large data transfer detected: %d bytes", event.Network.BytesSent),
			Confidence:  0.6,
			Evidence:    []string{fmt.Sprintf("Bytes sent: %d", event.Network.BytesSent)},
			Mitigations: []string{"Monitor network traffic", "Review data access patterns"},
		}, nil
	}

	return nil, nil
}

func (d *NetworkThreatDetector) Start(ctx context.Context) error {
	return nil
}

func (d *NetworkThreatDetector) Stop() error {
	return nil
}

// Helper functions

func generateAlertID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}

func generateIncidentID() string {
	return fmt.Sprintf("incident-%d", time.Now().UnixNano())
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}
