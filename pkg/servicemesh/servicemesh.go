package servicemesh

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Manager manages service mesh for containr
type Manager struct {
	config       *Config
	services     map[string]*Service
	sidecars     map[string]*EnvoySidecar
	policies     map[string]*TrafficPolicy
	mtlsConfig   *MTLSConfig
	mu           sync.RWMutex
	storePath    string
	eventChan    chan Event
	stopChan     chan struct{}
}

// Config holds service mesh configuration
type Config struct {
	Enabled           bool                       `yaml:"enabled"`
	AutoInject        bool                       `yaml:"auto_inject"`
	MTLSMode          string                     `yaml:"mtls_mode"` // strict, permissive, disabled
	EnvoyImage        string                     `yaml:"envoy_image"`
	EnvoyVersion      string                     `yaml:"envoy_version"`
	DefaultPolicies   map[string]*TrafficPolicy  `yaml:"default_policies"`
	SidecarConfig     *SidecarConfig             `yaml:"sidecar_config"`
	ObservabilityConfig *ObservabilityConfig     `yaml:"observability_config"`
}

// SidecarConfig holds sidecar configuration
type SidecarConfig struct {
	AdminPort      int               `yaml:"admin_port"`
	InboundPort    int               `yaml:"inbound_port"`
	OutboundPort   int               `yaml:"outbound_port"`
	CPULimit       string            `yaml:"cpu_limit"`
	MemoryLimit    string            `yaml:"memory_limit"`
	LogLevel       string            `yaml:"log_level"`
	ExtraArgs      []string          `yaml:"extra_args"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	EnableTracing    bool   `yaml:"enable_tracing"`
	EnableMetrics    bool   `yaml:"enable_metrics"`
	TracingEndpoint  string `yaml:"tracing_endpoint"`
	MetricsEndpoint  string `yaml:"metrics_endpoint"`
}

// Service represents a service in the mesh
type Service struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	ContainerID     string            `json:"container_id"`
	Port            int               `json:"port"`
	Protocol        string            `json:"protocol"`
	SidecarInjected bool              `json:"sidecar_injected"`
	Labels          map[string]string `json:"labels"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// Event represents a service mesh event
type Event struct {
	Type      string    `json:"type"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// NewManager creates a new service mesh manager
func NewManager(configPath string, storePath string) (*Manager, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	mgr := &Manager{
		config:     config,
		services:   make(map[string]*Service),
		sidecars:   make(map[string]*EnvoySidecar),
		policies:   config.DefaultPolicies,
		storePath:  storePath,
		eventChan:  make(chan Event, 100),
		stopChan:   make(chan struct{}),
	}

	// Initialize mTLS config
	if config.MTLSMode != "disabled" {
		mtlsConfig, err := NewMTLSConfig(filepath.Join(storePath, "certs"))
		if err != nil {
			return nil, fmt.Errorf("failed to initialize mTLS: %w", err)
		}
		mgr.mtlsConfig = mtlsConfig
	}

	// Load existing services
	if err := mgr.loadServices(); err != nil {
		return nil, fmt.Errorf("failed to load services: %w", err)
	}

	// Load existing policies
	if err := mgr.loadPolicies(); err != nil {
		return nil, fmt.Errorf("failed to load policies: %w", err)
	}

	// Start event processor
	go mgr.processEvents()

	return mgr, nil
}

// RegisterService registers a service in the mesh
func (m *Manager) RegisterService(ctx context.Context, service *Service) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.services[service.ID]; exists {
		return fmt.Errorf("service %s already registered", service.ID)
	}

	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()

	// Auto-inject sidecar if enabled
	if m.config.AutoInject && !service.SidecarInjected {
		sidecar, err := m.createEnvoySidecar(service)
		if err != nil {
			return fmt.Errorf("failed to inject sidecar: %w", err)
		}
		m.sidecars[service.ID] = sidecar
		service.SidecarInjected = true
	}

	m.services[service.ID] = service

	// Emit event
	m.emitEvent(Event{
		Type:      "service.registered",
		Service:   service.Name,
		Timestamp: time.Now(),
		Data:      service,
	})

	return m.saveServices()
}

// UnregisterService unregisters a service from the mesh
func (m *Manager) UnregisterService(ctx context.Context, serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// Remove sidecar if injected
	if sidecar, ok := m.sidecars[serviceID]; ok {
		if err := sidecar.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop sidecar: %w", err)
		}
		delete(m.sidecars, serviceID)
	}

	delete(m.services, serviceID)

	// Emit event
	m.emitEvent(Event{
		Type:      "service.unregistered",
		Service:   service.Name,
		Timestamp: time.Now(),
		Data:      service,
	})

	return m.saveServices()
}

// GetService retrieves a service by ID
func (m *Manager) GetService(serviceID string) (*Service, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	service, ok := m.services[serviceID]
	if !ok {
		return nil, fmt.Errorf("service %s not found", serviceID)
	}

	return service, nil
}

// ListServices returns all registered services
func (m *Manager) ListServices() []*Service {
	m.mu.RLock()
	defer m.mu.RUnlock()

	services := make([]*Service, 0, len(m.services))
	for _, service := range m.services {
		services = append(services, service)
	}

	return services
}

// InjectSidecar injects an Envoy sidecar into a service
func (m *Manager) InjectSidecar(ctx context.Context, serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	if service.SidecarInjected {
		return fmt.Errorf("sidecar already injected for service %s", serviceID)
	}

	sidecar, err := m.createEnvoySidecar(service)
	if err != nil {
		return fmt.Errorf("failed to create sidecar: %w", err)
	}

	m.sidecars[serviceID] = sidecar
	service.SidecarInjected = true
	service.UpdatedAt = time.Now()

	// Emit event
	m.emitEvent(Event{
		Type:      "sidecar.injected",
		Service:   service.Name,
		Timestamp: time.Now(),
		Data:      sidecar,
	})

	return m.saveServices()
}

// RemoveSidecar removes the Envoy sidecar from a service
func (m *Manager) RemoveSidecar(ctx context.Context, serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	sidecar, ok := m.sidecars[serviceID]
	if !ok {
		return fmt.Errorf("no sidecar found for service %s", serviceID)
	}

	if err := sidecar.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop sidecar: %w", err)
	}

	delete(m.sidecars, serviceID)
	service.SidecarInjected = false
	service.UpdatedAt = time.Now()

	// Emit event
	m.emitEvent(Event{
		Type:      "sidecar.removed",
		Service:   service.Name,
		Timestamp: time.Now(),
	})

	return m.saveServices()
}

// ApplyPolicy applies a traffic policy to a service
func (m *Manager) ApplyPolicy(ctx context.Context, serviceID string, policy *TrafficPolicy) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	service, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// Store the policy
	policyKey := fmt.Sprintf("%s/%s", serviceID, policy.Name)
	m.policies[policyKey] = policy

	// Apply policy to sidecar if injected
	if sidecar, ok := m.sidecars[serviceID]; ok {
		if err := sidecar.ApplyPolicy(ctx, policy); err != nil {
			return fmt.Errorf("failed to apply policy to sidecar: %w", err)
		}
	}

	// Emit event
	m.emitEvent(Event{
		Type:      "policy.applied",
		Service:   service.Name,
		Timestamp: time.Now(),
		Data:      policy,
	})

	return m.savePolicies()
}

// GetPolicy retrieves a traffic policy
func (m *Manager) GetPolicy(serviceID string, policyName string) (*TrafficPolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	policyKey := fmt.Sprintf("%s/%s", serviceID, policyName)
	policy, ok := m.policies[policyKey]
	if !ok {
		return nil, fmt.Errorf("policy %s not found for service %s", policyName, serviceID)
	}

	return policy, nil
}

// EnableMTLS enables mTLS for a service
func (m *Manager) EnableMTLS(ctx context.Context, serviceID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.mtlsConfig == nil {
		return fmt.Errorf("mTLS not configured")
	}

	service, exists := m.services[serviceID]
	if !exists {
		return fmt.Errorf("service %s not found", serviceID)
	}

	// Generate certificates for the service
	cert, err := m.mtlsConfig.GenerateServiceCertificate(service.Name, service.Namespace)
	if err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}

	// Apply to sidecar if injected
	if sidecar, ok := m.sidecars[serviceID]; ok {
		if err := sidecar.ConfigureMTLS(ctx, cert); err != nil {
			return fmt.Errorf("failed to configure mTLS on sidecar: %w", err)
		}
	}

	// Emit event
	m.emitEvent(Event{
		Type:      "mtls.enabled",
		Service:   service.Name,
		Timestamp: time.Now(),
	})

	return nil
}

// GetMetrics retrieves metrics for a service
func (m *Manager) GetMetrics(ctx context.Context, serviceID string) (*ServiceMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sidecar, ok := m.sidecars[serviceID]
	if !ok {
		return nil, fmt.Errorf("no sidecar found for service %s", serviceID)
	}

	return sidecar.GetMetrics(ctx)
}

// Close closes the service mesh manager
func (m *Manager) Close() error {
	close(m.stopChan)
	close(m.eventChan)

	// Stop all sidecars
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx := context.Background()
	for _, sidecar := range m.sidecars {
		sidecar.Stop(ctx)
	}

	return nil
}

// createEnvoySidecar creates an Envoy sidecar for a service
func (m *Manager) createEnvoySidecar(service *Service) (*EnvoySidecar, error) {
	config := &EnvoyConfig{
		ServiceName:      service.Name,
		ServiceNamespace: service.Namespace,
		ServicePort:      service.Port,
		AdminPort:        m.config.SidecarConfig.AdminPort,
		InboundPort:      m.config.SidecarConfig.InboundPort,
		OutboundPort:     m.config.SidecarConfig.OutboundPort,
		Image:            m.config.EnvoyImage,
		Version:          m.config.EnvoyVersion,
		LogLevel:         m.config.SidecarConfig.LogLevel,
		MTLSEnabled:      m.config.MTLSMode != "disabled",
	}

	sidecar, err := NewEnvoySidecar(config)
	if err != nil {
		return nil, err
	}

	// Start the sidecar
	if err := sidecar.Start(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to start sidecar: %w", err)
	}

	return sidecar, nil
}

// emitEvent emits an event
func (m *Manager) emitEvent(event Event) {
	select {
	case m.eventChan <- event:
	default:
		// Event channel full, drop event
	}
}

// processEvents processes events
func (m *Manager) processEvents() {
	for {
		select {
		case event := <-m.eventChan:
			// Log the event (in production, this would be sent to observability system)
			fmt.Printf("[ServiceMesh] Event: %s for service %s\n", event.Type, event.Service)
		case <-m.stopChan:
			return
		}
	}
}

// loadConfig loads service mesh configuration from file
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Return default config if file doesn't exist
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

// defaultConfig returns default service mesh configuration
func defaultConfig() *Config {
	return &Config{
		Enabled:      true,
		AutoInject:   true,
		MTLSMode:     "permissive",
		EnvoyImage:   "envoyproxy/envoy",
		EnvoyVersion: "v1.28.0",
		SidecarConfig: &SidecarConfig{
			AdminPort:    15000,
			InboundPort:  15006,
			OutboundPort: 15001,
			CPULimit:     "500m",
			MemoryLimit:  "512Mi",
			LogLevel:     "info",
		},
		ObservabilityConfig: &ObservabilityConfig{
			EnableTracing:   true,
			EnableMetrics:   true,
			TracingEndpoint: "localhost:9411",
			MetricsEndpoint: "localhost:9090",
		},
		DefaultPolicies: make(map[string]*TrafficPolicy),
	}
}

// saveServices persists services to disk
func (m *Manager) saveServices() error {
	servicesFile := filepath.Join(m.storePath, "services.json")

	data, err := json.MarshalIndent(m.services, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal services: %w", err)
	}

	if err := os.MkdirAll(m.storePath, 0755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	if err := os.WriteFile(servicesFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write services file: %w", err)
	}

	return nil
}

// loadServices loads services from disk
func (m *Manager) loadServices() error {
	servicesFile := filepath.Join(m.storePath, "services.json")

	data, err := os.ReadFile(servicesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No services file yet
		}
		return fmt.Errorf("failed to read services file: %w", err)
	}

	if err := json.Unmarshal(data, &m.services); err != nil {
		return fmt.Errorf("failed to unmarshal services: %w", err)
	}

	return nil
}

// savePolicies persists policies to disk
func (m *Manager) savePolicies() error {
	policiesFile := filepath.Join(m.storePath, "policies.json")

	data, err := json.MarshalIndent(m.policies, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal policies: %w", err)
	}

	if err := os.MkdirAll(m.storePath, 0755); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}

	if err := os.WriteFile(policiesFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write policies file: %w", err)
	}

	return nil
}

// loadPolicies loads policies from disk
func (m *Manager) loadPolicies() error {
	policiesFile := filepath.Join(m.storePath, "policies.json")

	data, err := os.ReadFile(policiesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No policies file yet
		}
		return fmt.Errorf("failed to read policies file: %w", err)
	}

	var policies map[string]*TrafficPolicy
	if err := json.Unmarshal(data, &policies); err != nil {
		return fmt.Errorf("failed to unmarshal policies: %w", err)
	}

	// Merge with default policies
	for k, v := range policies {
		m.policies[k] = v
	}

	return nil
}
