package servicemesh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// EnvoySidecar represents an Envoy proxy sidecar
type EnvoySidecar struct {
	config      *EnvoyConfig
	containerID string
	status      SidecarStatus
	metrics     *ServiceMetrics
	mu          sync.RWMutex
	stopChan    chan struct{}
}

// EnvoyConfig holds Envoy sidecar configuration
type EnvoyConfig struct {
	ServiceName      string            `json:"service_name"`
	ServiceNamespace string            `json:"service_namespace"`
	ServicePort      int               `json:"service_port"`
	AdminPort        int               `json:"admin_port"`
	InboundPort      int               `json:"inbound_port"`
	OutboundPort     int               `json:"outbound_port"`
	Image            string            `json:"image"`
	Version          string            `json:"version"`
	LogLevel         string            `json:"log_level"`
	MTLSEnabled      bool              `json:"mtls_enabled"`
	ConfigPath       string            `json:"config_path"`
	Clusters         []*EnvoyCluster   `json:"clusters"`
	Listeners        []*EnvoyListener  `json:"listeners"`
}

// EnvoyCluster represents an Envoy cluster configuration
type EnvoyCluster struct {
	Name              string              `json:"name"`
	Type              string              `json:"type"` // STATIC, STRICT_DNS, LOGICAL_DNS, EDS
	ConnectTimeout    string              `json:"connect_timeout"`
	LBPolicy          string              `json:"lb_policy"` // ROUND_ROBIN, LEAST_REQUEST, RING_HASH, RANDOM
	Endpoints         []*ClusterEndpoint  `json:"endpoints"`
	HealthCheck       *HealthCheckConfig  `json:"health_check,omitempty"`
	CircuitBreaker    *CircuitBreaker     `json:"circuit_breaker,omitempty"`
}

// ClusterEndpoint represents a cluster endpoint
type ClusterEndpoint struct {
	Address  string `json:"address"`
	Port     int    `json:"port"`
	Weight   int    `json:"weight,omitempty"`
	Priority int    `json:"priority,omitempty"`
}

// EnvoyListener represents an Envoy listener configuration
type EnvoyListener struct {
	Name          string               `json:"name"`
	Address       string               `json:"address"`
	Port          int                  `json:"port"`
	FilterChains  []*FilterChain       `json:"filter_chains"`
}

// FilterChain represents a filter chain
type FilterChain struct {
	Filters       []*Filter            `json:"filters"`
	TransportSocket *TransportSocket   `json:"transport_socket,omitempty"`
}

// Filter represents an Envoy filter
type Filter struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

// TransportSocket represents transport socket configuration for mTLS
type TransportSocket struct {
	Name       string                 `json:"name"`
	TypedConfig map[string]interface{} `json:"typed_config"`
}

// HealthCheckConfig represents health check configuration
type HealthCheckConfig struct {
	Timeout            string `json:"timeout"`
	Interval           string `json:"interval"`
	UnhealthyThreshold int    `json:"unhealthy_threshold"`
	HealthyThreshold   int    `json:"healthy_threshold"`
	Path               string `json:"path,omitempty"`
}

// SidecarStatus represents the status of a sidecar
type SidecarStatus string

const (
	SidecarStatusStarting SidecarStatus = "starting"
	SidecarStatusRunning  SidecarStatus = "running"
	SidecarStatusStopping SidecarStatus = "stopping"
	SidecarStatusStopped  SidecarStatus = "stopped"
	SidecarStatusFailed   SidecarStatus = "failed"
)

// ServiceMetrics represents metrics for a service
type ServiceMetrics struct {
	RequestCount      int64              `json:"request_count"`
	RequestRate       float64            `json:"request_rate"`
	ErrorCount        int64              `json:"error_count"`
	ErrorRate         float64            `json:"error_rate"`
	LatencyP50        float64            `json:"latency_p50"`
	LatencyP95        float64            `json:"latency_p95"`
	LatencyP99        float64            `json:"latency_p99"`
	ActiveConnections int                `json:"active_connections"`
	CircuitBreakerOpen bool              `json:"circuit_breaker_open"`
	LastUpdated       time.Time          `json:"last_updated"`
}

// NewEnvoySidecar creates a new Envoy sidecar
func NewEnvoySidecar(config *EnvoyConfig) (*EnvoySidecar, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Validate configuration
	if config.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}

	// Set defaults
	if config.AdminPort == 0 {
		config.AdminPort = 15000
	}
	if config.InboundPort == 0 {
		config.InboundPort = 15006
	}
	if config.OutboundPort == 0 {
		config.OutboundPort = 15001
	}
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	return &EnvoySidecar{
		config:   config,
		status:   SidecarStatusStopped,
		metrics:  &ServiceMetrics{},
		stopChan: make(chan struct{}),
	}, nil
}

// Start starts the Envoy sidecar
func (e *EnvoySidecar) Start(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status == SidecarStatusRunning {
		return fmt.Errorf("sidecar already running")
	}

	e.status = SidecarStatusStarting

	// Generate Envoy configuration
	envoyConfig, err := e.generateEnvoyConfig()
	if err != nil {
		e.status = SidecarStatusFailed
		return fmt.Errorf("failed to generate Envoy config: %w", err)
	}

	// In a real implementation, this would:
	// 1. Write the Envoy config to a file
	// 2. Start the Envoy container/process
	// 3. Wait for it to be healthy

	// For this implementation, we'll simulate the startup
	e.containerID = fmt.Sprintf("envoy-%s-%d", e.config.ServiceName, time.Now().Unix())

	// Simulate startup delay
	time.Sleep(100 * time.Millisecond)

	e.status = SidecarStatusRunning

	// Start metrics collection
	go e.collectMetrics()

	// Log the configuration (in production, write to file)
	_ = envoyConfig

	return nil
}

// Stop stops the Envoy sidecar
func (e *EnvoySidecar) Stop(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status == SidecarStatusStopped {
		return nil
	}

	e.status = SidecarStatusStopping

	// Signal metrics collection to stop
	close(e.stopChan)

	// In a real implementation, this would stop the Envoy container
	// For now, we just update the status
	time.Sleep(50 * time.Millisecond)

	e.status = SidecarStatusStopped
	e.containerID = ""

	return nil
}

// Reload reloads the Envoy configuration
func (e *EnvoySidecar) Reload(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.reloadLocked()
}

// reloadLocked performs the reload without acquiring the lock (caller must hold lock)
func (e *EnvoySidecar) reloadLocked() error {
	if e.status != SidecarStatusRunning {
		return fmt.Errorf("sidecar not running")
	}

	// Generate new configuration
	envoyConfig, err := e.generateEnvoyConfig()
	if err != nil {
		return fmt.Errorf("failed to generate Envoy config: %w", err)
	}

	// In a real implementation, this would:
	// 1. Write the new config to a file
	// 2. Send a hot restart signal to Envoy
	// 3. Wait for the new instance to be healthy

	_ = envoyConfig

	return nil
}

// ApplyPolicy applies a traffic policy to the sidecar
func (e *EnvoySidecar) ApplyPolicy(ctx context.Context, policy *TrafficPolicy) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Update cluster configuration based on policy
	for _, cluster := range e.config.Clusters {
		// Apply load balancing policy
		if policy.LoadBalancing != nil {
			cluster.LBPolicy = policy.LoadBalancing.Algorithm
		}

		// Apply circuit breaker
		if policy.CircuitBreaker != nil {
			cluster.CircuitBreaker = policy.CircuitBreaker
		}

		// Apply health check
		if policy.HealthCheck != nil {
			cluster.HealthCheck = &HealthCheckConfig{
				Timeout:            policy.HealthCheck.Timeout,
				Interval:           policy.HealthCheck.Interval,
				UnhealthyThreshold: policy.HealthCheck.UnhealthyThreshold,
				HealthyThreshold:   policy.HealthCheck.HealthyThreshold,
				Path:               policy.HealthCheck.Path,
			}
		}
	}

	// Reload configuration (using internal method to avoid deadlock)
	return e.reloadLocked()
}

// ConfigureMTLS configures mTLS for the sidecar
func (e *EnvoySidecar) ConfigureMTLS(ctx context.Context, cert *Certificate) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.config.MTLSEnabled {
		return fmt.Errorf("mTLS not enabled for this sidecar")
	}

	// In a real implementation, this would:
	// 1. Write the certificate and key to files
	// 2. Update Envoy configuration with mTLS settings
	// 3. Reload Envoy

	_ = cert

	e.config.MTLSEnabled = true

	return e.reloadLocked()
}

// GetStatus returns the current status of the sidecar
func (e *EnvoySidecar) GetStatus() SidecarStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.status
}

// GetMetrics retrieves current metrics from the sidecar
func (e *EnvoySidecar) GetMetrics(ctx context.Context) (*ServiceMetrics, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.status != SidecarStatusRunning {
		return nil, fmt.Errorf("sidecar not running")
	}

	// Return a copy of the metrics
	metricsCopy := *e.metrics
	return &metricsCopy, nil
}

// GetAdminEndpoint returns the admin endpoint URL
func (e *EnvoySidecar) GetAdminEndpoint() string {
	return fmt.Sprintf("http://localhost:%d", e.config.AdminPort)
}

// GetConfig returns the sidecar configuration
func (e *EnvoySidecar) GetConfig() *EnvoyConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.config
}

// AddCluster adds a new cluster to the Envoy configuration
func (e *EnvoySidecar) AddCluster(ctx context.Context, cluster *EnvoyCluster) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Check if cluster already exists
	for _, c := range e.config.Clusters {
		if c.Name == cluster.Name {
			return fmt.Errorf("cluster %s already exists", cluster.Name)
		}
	}

	e.config.Clusters = append(e.config.Clusters, cluster)

	if e.status == SidecarStatusRunning {
		return e.reloadLocked()
	}

	return nil
}

// RemoveCluster removes a cluster from the Envoy configuration
func (e *EnvoySidecar) RemoveCluster(ctx context.Context, clusterName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	clusters := make([]*EnvoyCluster, 0)
	found := false

	for _, c := range e.config.Clusters {
		if c.Name != clusterName {
			clusters = append(clusters, c)
		} else {
			found = true
		}
	}

	if !found {
		return fmt.Errorf("cluster %s not found", clusterName)
	}

	e.config.Clusters = clusters

	if e.status == SidecarStatusRunning {
		return e.reloadLocked()
	}

	return nil
}

// generateEnvoyConfig generates the Envoy proxy configuration
func (e *EnvoySidecar) generateEnvoyConfig() (map[string]interface{}, error) {
	config := map[string]interface{}{
		"node": map[string]interface{}{
			"id":      e.config.ServiceName,
			"cluster": e.config.ServiceNamespace,
		},
		"admin": map[string]interface{}{
			"access_log_path": "/dev/stdout",
			"address": map[string]interface{}{
				"socket_address": map[string]interface{}{
					"address":    "0.0.0.0",
					"port_value": e.config.AdminPort,
				},
			},
		},
		"static_resources": map[string]interface{}{
			"listeners": e.generateListeners(),
			"clusters":  e.generateClusters(),
		},
	}

	return config, nil
}

// generateListeners generates listener configurations
func (e *EnvoySidecar) generateListeners() []map[string]interface{} {
	listeners := make([]map[string]interface{}, 0)

	// Inbound listener
	inboundListener := map[string]interface{}{
		"name": "inbound_listener",
		"address": map[string]interface{}{
			"socket_address": map[string]interface{}{
				"address":    "0.0.0.0",
				"port_value": e.config.InboundPort,
			},
		},
		"filter_chains": []map[string]interface{}{
			{
				"filters": []map[string]interface{}{
					{
						"name": "envoy.filters.network.http_connection_manager",
						"typed_config": map[string]interface{}{
							"@type":      "type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager",
							"stat_prefix": "ingress_http",
							"route_config": map[string]interface{}{
								"name": "local_route",
								"virtual_hosts": []map[string]interface{}{
									{
										"name":    "backend",
										"domains": []string{"*"},
										"routes": []map[string]interface{}{
											{
												"match": map[string]interface{}{
													"prefix": "/",
												},
												"route": map[string]interface{}{
													"cluster": "local_service",
												},
											},
										},
									},
								},
							},
							"http_filters": []map[string]interface{}{
								{
									"name": "envoy.filters.http.router",
								},
							},
						},
					},
				},
			},
		},
	}

	listeners = append(listeners, inboundListener)

	// Add custom listeners
	for _, listener := range e.config.Listeners {
		listeners = append(listeners, map[string]interface{}{
			"name": listener.Name,
			"address": map[string]interface{}{
				"socket_address": map[string]interface{}{
					"address":    listener.Address,
					"port_value": listener.Port,
				},
			},
		})
	}

	return listeners
}

// generateClusters generates cluster configurations
func (e *EnvoySidecar) generateClusters() []map[string]interface{} {
	clusters := make([]map[string]interface{}, 0)

	// Local service cluster
	localCluster := map[string]interface{}{
		"name":            "local_service",
		"connect_timeout": "5s",
		"type":            "STATIC",
		"lb_policy":       "ROUND_ROBIN",
		"load_assignment": map[string]interface{}{
			"cluster_name": "local_service",
			"endpoints": []map[string]interface{}{
				{
					"lb_endpoints": []map[string]interface{}{
						{
							"endpoint": map[string]interface{}{
								"address": map[string]interface{}{
									"socket_address": map[string]interface{}{
										"address":    "127.0.0.1",
										"port_value": e.config.ServicePort,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	clusters = append(clusters, localCluster)

	// Add custom clusters
	for _, cluster := range e.config.Clusters {
		clusterConfig := map[string]interface{}{
			"name":            cluster.Name,
			"connect_timeout": cluster.ConnectTimeout,
			"type":            cluster.Type,
			"lb_policy":       cluster.LBPolicy,
		}

		// Add endpoints
		endpoints := make([]map[string]interface{}, 0)
		for _, ep := range cluster.Endpoints {
			endpoints = append(endpoints, map[string]interface{}{
				"endpoint": map[string]interface{}{
					"address": map[string]interface{}{
						"socket_address": map[string]interface{}{
							"address":    ep.Address,
							"port_value": ep.Port,
						},
					},
				},
				"load_balancing_weight": ep.Weight,
			})
		}

		clusterConfig["load_assignment"] = map[string]interface{}{
			"cluster_name": cluster.Name,
			"endpoints": []map[string]interface{}{
				{
					"lb_endpoints": endpoints,
				},
			},
		}

		clusters = append(clusters, clusterConfig)
	}

	return clusters
}

// collectMetrics collects metrics from the Envoy admin endpoint
func (e *EnvoySidecar) collectMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			e.updateMetrics()
		case <-e.stopChan:
			return
		}
	}
}

// updateMetrics updates the metrics from Envoy
func (e *EnvoySidecar) updateMetrics() {
	// In a real implementation, this would query the Envoy admin endpoint
	// For now, we'll simulate some metrics

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != SidecarStatusRunning {
		return
	}

	// Simulate fetching metrics from admin endpoint
	adminURL := fmt.Sprintf("http://localhost:%d/stats", e.config.AdminPort)

	// In production, you would actually make this HTTP call:
	// resp, err := http.Get(adminURL)
	// For now, we'll just simulate
	_ = adminURL

	// Update with simulated values
	e.metrics.LastUpdated = time.Now()
	e.metrics.RequestCount++
	e.metrics.ActiveConnections = 5
}

// queryAdminEndpoint queries the Envoy admin endpoint
func (e *EnvoySidecar) queryAdminEndpoint(path string) ([]byte, error) {
	url := fmt.Sprintf("http://localhost:%d%s", e.config.AdminPort, path)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to query admin endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("admin endpoint returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// DrainConnections drains active connections from the sidecar
func (e *EnvoySidecar) DrainConnections(ctx context.Context, timeout time.Duration) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != SidecarStatusRunning {
		return fmt.Errorf("sidecar not running")
	}

	// In a real implementation, this would:
	// 1. POST to /healthcheck/fail to mark unhealthy
	// 2. Wait for connections to drain
	// 3. POST to /drain_listeners to stop accepting new connections

	time.Sleep(timeout)

	return nil
}

// GetConfigDump retrieves the current Envoy configuration dump
func (e *EnvoySidecar) GetConfigDump(ctx context.Context) (map[string]interface{}, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.status != SidecarStatusRunning {
		return nil, fmt.Errorf("sidecar not running")
	}

	// Query the config_dump endpoint
	data, err := e.queryAdminEndpoint("/config_dump")
	if err != nil {
		// Return simulated config if query fails
		return e.generateEnvoyConfig()
	}

	var configDump map[string]interface{}
	if err := json.Unmarshal(data, &configDump); err != nil {
		return nil, fmt.Errorf("failed to parse config dump: %w", err)
	}

	return configDump, nil
}
