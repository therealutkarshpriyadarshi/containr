package servicemesh

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestManager_RegisterService tests service registration
func TestManager_RegisterService(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)
	defer mgr.Close()

	ctx := context.Background()

	service := &Service{
		ID:        "svc-1",
		Name:      "test-service",
		Namespace: "default",
		Port:      8080,
		Protocol:  "http",
		Labels:    map[string]string{"app": "test"},
	}

	err := mgr.RegisterService(ctx, service)
	if err != nil {
		t.Fatalf("RegisterService failed: %v", err)
	}

	// Verify service was registered
	retrieved, err := mgr.GetService("svc-1")
	if err != nil {
		t.Fatalf("GetService failed: %v", err)
	}

	if retrieved.Name != "test-service" {
		t.Errorf("expected service name test-service, got %s", retrieved.Name)
	}

	if !retrieved.SidecarInjected {
		t.Error("expected sidecar to be injected")
	}
}

// TestManager_UnregisterService tests service unregistration
func TestManager_UnregisterService(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)
	defer mgr.Close()

	ctx := context.Background()

	service := &Service{
		ID:        "svc-1",
		Name:      "test-service",
		Namespace: "default",
		Port:      8080,
		Protocol:  "http",
	}

	mgr.RegisterService(ctx, service)

	// Unregister service
	err := mgr.UnregisterService(ctx, "svc-1")
	if err != nil {
		t.Fatalf("UnregisterService failed: %v", err)
	}

	// Verify service was removed
	_, err = mgr.GetService("svc-1")
	if err == nil {
		t.Error("expected error when getting unregistered service")
	}
}

// TestManager_InjectSidecar tests sidecar injection
func TestManager_InjectSidecar(t *testing.T) {
	tmpDir := t.TempDir()

	// Create manager with auto-inject disabled
	config := defaultConfig()
	config.AutoInject = false
	mgr := setupTestManagerWithConfig(t, tmpDir, config)
	defer mgr.Close()

	ctx := context.Background()

	service := &Service{
		ID:        "svc-1",
		Name:      "test-service",
		Namespace: "default",
		Port:      8080,
		Protocol:  "http",
	}

	mgr.RegisterService(ctx, service)

	// Service should not have sidecar initially
	retrieved, _ := mgr.GetService("svc-1")
	if retrieved.SidecarInjected {
		t.Error("sidecar should not be injected initially")
	}

	// Inject sidecar
	err := mgr.InjectSidecar(ctx, "svc-1")
	if err != nil {
		t.Fatalf("InjectSidecar failed: %v", err)
	}

	// Verify sidecar was injected
	retrieved, _ = mgr.GetService("svc-1")
	if !retrieved.SidecarInjected {
		t.Error("expected sidecar to be injected")
	}
}

// TestManager_ApplyPolicy tests applying traffic policies
func TestManager_ApplyPolicy(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)
	defer mgr.Close()

	ctx := context.Background()

	service := &Service{
		ID:        "svc-1",
		Name:      "test-service",
		Namespace: "default",
		Port:      8080,
		Protocol:  "http",
	}

	mgr.RegisterService(ctx, service)

	// Apply policy
	policy := DefaultTrafficPolicy()
	policy.Name = "test-policy"

	err := mgr.ApplyPolicy(ctx, "svc-1", policy)
	if err != nil {
		t.Fatalf("ApplyPolicy failed: %v", err)
	}

	// Verify policy was applied
	retrieved, err := mgr.GetPolicy("svc-1", "test-policy")
	if err != nil {
		t.Fatalf("GetPolicy failed: %v", err)
	}

	if retrieved.Name != "test-policy" {
		t.Errorf("expected policy name test-policy, got %s", retrieved.Name)
	}
}

// TestManager_EnableMTLS tests enabling mTLS
func TestManager_EnableMTLS(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)
	defer mgr.Close()

	ctx := context.Background()

	service := &Service{
		ID:        "svc-1",
		Name:      "test-service",
		Namespace: "default",
		Port:      8080,
		Protocol:  "http",
	}

	mgr.RegisterService(ctx, service)

	// Enable mTLS
	err := mgr.EnableMTLS(ctx, "svc-1")
	if err != nil {
		t.Fatalf("EnableMTLS failed: %v", err)
	}

	// Verify certificate was generated
	cert, err := mgr.mtlsConfig.GetCertificate("test-service", "default")
	if err != nil {
		t.Fatalf("GetCertificate failed: %v", err)
	}

	if cert.ServiceName != "test-service" {
		t.Errorf("expected service name test-service, got %s", cert.ServiceName)
	}
}

// TestEnvoySidecar_StartStop tests starting and stopping an Envoy sidecar
func TestEnvoySidecar_StartStop(t *testing.T) {
	config := &EnvoyConfig{
		ServiceName:      "test-service",
		ServiceNamespace: "default",
		ServicePort:      8080,
		AdminPort:        15000,
		InboundPort:      15006,
		OutboundPort:     15001,
		Image:            "envoyproxy/envoy",
		Version:          "v1.28.0",
		LogLevel:         "info",
	}

	sidecar, err := NewEnvoySidecar(config)
	if err != nil {
		t.Fatalf("NewEnvoySidecar failed: %v", err)
	}

	ctx := context.Background()

	// Start sidecar
	err = sidecar.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if sidecar.GetStatus() != SidecarStatusRunning {
		t.Errorf("expected status running, got %s", sidecar.GetStatus())
	}

	// Stop sidecar
	err = sidecar.Stop(ctx)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if sidecar.GetStatus() != SidecarStatusStopped {
		t.Errorf("expected status stopped, got %s", sidecar.GetStatus())
	}
}

// TestEnvoySidecar_ApplyPolicy tests applying a policy to a sidecar
func TestEnvoySidecar_ApplyPolicy(t *testing.T) {
	config := &EnvoyConfig{
		ServiceName:      "test-service",
		ServiceNamespace: "default",
		ServicePort:      8080,
	}

	sidecar, err := NewEnvoySidecar(config)
	if err != nil {
		t.Fatalf("NewEnvoySidecar failed: %v", err)
	}

	ctx := context.Background()
	sidecar.Start(ctx)
	defer sidecar.Stop(ctx)

	// Apply policy
	policy := DefaultTrafficPolicy()
	err = sidecar.ApplyPolicy(ctx, policy)
	if err != nil {
		t.Fatalf("ApplyPolicy failed: %v", err)
	}
}

// TestEnvoySidecar_AddRemoveCluster tests adding and removing clusters
func TestEnvoySidecar_AddRemoveCluster(t *testing.T) {
	config := &EnvoyConfig{
		ServiceName:      "test-service",
		ServiceNamespace: "default",
		ServicePort:      8080,
	}

	sidecar, err := NewEnvoySidecar(config)
	if err != nil {
		t.Fatalf("NewEnvoySidecar failed: %v", err)
	}

	ctx := context.Background()

	// Add cluster
	cluster := &EnvoyCluster{
		Name:           "backend-cluster",
		Type:           "STATIC",
		ConnectTimeout: "5s",
		LBPolicy:       "ROUND_ROBIN",
		Endpoints: []*ClusterEndpoint{
			{Address: "192.168.1.100", Port: 8080},
			{Address: "192.168.1.101", Port: 8080},
		},
	}

	err = sidecar.AddCluster(ctx, cluster)
	if err != nil {
		t.Fatalf("AddCluster failed: %v", err)
	}

	// Verify cluster was added
	if len(sidecar.config.Clusters) != 1 {
		t.Errorf("expected 1 cluster, got %d", len(sidecar.config.Clusters))
	}

	// Remove cluster
	err = sidecar.RemoveCluster(ctx, "backend-cluster")
	if err != nil {
		t.Fatalf("RemoveCluster failed: %v", err)
	}

	// Verify cluster was removed
	if len(sidecar.config.Clusters) != 0 {
		t.Errorf("expected 0 clusters, got %d", len(sidecar.config.Clusters))
	}
}

// TestTrafficPolicy_Validate tests policy validation
func TestTrafficPolicy_Validate(t *testing.T) {
	tests := []struct {
		name      string
		policy    *TrafficPolicy
		expectErr bool
	}{
		{
			name:      "valid default policy",
			policy:    DefaultTrafficPolicy(),
			expectErr: false,
		},
		{
			name: "invalid policy - empty name",
			policy: &TrafficPolicy{
				Name: "",
			},
			expectErr: true,
		},
		{
			name: "invalid retry policy - too many attempts",
			policy: &TrafficPolicy{
				Name: "test",
				Retry: &RetryPolicy{
					Attempts:      15,
					PerTryTimeout: "5s",
					RetryOn:       []string{"5xx"},
				},
			},
			expectErr: true,
		},
		{
			name: "valid high availability policy",
			policy: HighAvailabilityPolicy(),
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}
		})
	}
}

// TestLoadBalancingPolicy_Validate tests load balancing policy validation
func TestLoadBalancingPolicy_Validate(t *testing.T) {
	tests := []struct {
		name      string
		policy    *LoadBalancingPolicy
		expectErr bool
	}{
		{
			name: "valid round robin",
			policy: &LoadBalancingPolicy{
				Algorithm: "ROUND_ROBIN",
			},
			expectErr: false,
		},
		{
			name: "valid least request",
			policy: &LoadBalancingPolicy{
				Algorithm: "LEAST_REQUEST",
			},
			expectErr: false,
		},
		{
			name: "invalid algorithm",
			policy: &LoadBalancingPolicy{
				Algorithm: "INVALID",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}
		})
	}
}

// TestCircuitBreaker_Validate tests circuit breaker validation
func TestCircuitBreaker_Validate(t *testing.T) {
	tests := []struct {
		name      string
		cb        *CircuitBreaker
		expectErr bool
	}{
		{
			name: "valid circuit breaker",
			cb: &CircuitBreaker{
				MaxConnections:        1024,
				MaxPendingRequests:    1024,
				MaxRequests:           1024,
				MaxRetries:            3,
				ErrorThresholdPercent: 50.0,
				SleepWindow:           "30s",
			},
			expectErr: false,
		},
		{
			name: "invalid - negative max connections",
			cb: &CircuitBreaker{
				MaxConnections:        -1,
				MaxPendingRequests:    1024,
				MaxRequests:           1024,
				MaxRetries:            3,
				ErrorThresholdPercent: 50.0,
				SleepWindow:           "30s",
			},
			expectErr: true,
		},
		{
			name: "invalid - error threshold > 100",
			cb: &CircuitBreaker{
				MaxConnections:        1024,
				MaxPendingRequests:    1024,
				MaxRequests:           1024,
				MaxRetries:            3,
				ErrorThresholdPercent: 150.0,
				SleepWindow:           "30s",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cb.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}
		})
	}
}

// TestCircuitBreaker_IsOpen tests circuit breaker open logic
func TestCircuitBreaker_IsOpen(t *testing.T) {
	cb := &CircuitBreaker{
		MaxConnections:        1024,
		MaxPendingRequests:    1024,
		MaxRequests:           1024,
		MaxRetries:            3,
		ErrorThresholdPercent: 50.0,
		SleepWindow:           "30s",
	}

	tests := []struct {
		name     string
		metrics  *ServiceMetrics
		expected bool
	}{
		{
			name: "below threshold",
			metrics: &ServiceMetrics{
				RequestCount: 100,
				ErrorCount:   10,
			},
			expected: false,
		},
		{
			name: "at threshold",
			metrics: &ServiceMetrics{
				RequestCount: 100,
				ErrorCount:   50,
			},
			expected: true,
		},
		{
			name: "above threshold",
			metrics: &ServiceMetrics{
				RequestCount: 100,
				ErrorCount:   75,
			},
			expected: true,
		},
		{
			name: "circuit breaker already open",
			metrics: &ServiceMetrics{
				RequestCount:       100,
				ErrorCount:         10,
				CircuitBreakerOpen: true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cb.IsOpen(tt.metrics)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestMTLSConfig_GenerateCertificate tests certificate generation
func TestMTLSConfig_GenerateCertificate(t *testing.T) {
	tmpDir := t.TempDir()
	certsPath := filepath.Join(tmpDir, "certs")

	mtlsConfig, err := NewMTLSConfig(certsPath)
	if err != nil {
		t.Fatalf("NewMTLSConfig failed: %v", err)
	}

	// Generate service certificate
	cert, err := mtlsConfig.GenerateServiceCertificate("test-service", "default")
	if err != nil {
		t.Fatalf("GenerateServiceCertificate failed: %v", err)
	}

	if cert.ServiceName != "test-service" {
		t.Errorf("expected service name test-service, got %s", cert.ServiceName)
	}

	if cert.Namespace != "default" {
		t.Errorf("expected namespace default, got %s", cert.Namespace)
	}

	if len(cert.CertPEM) == 0 {
		t.Error("expected certificate PEM to be populated")
	}

	if len(cert.KeyPEM) == 0 {
		t.Error("expected key PEM to be populated")
	}

	// Verify certificate was saved to disk
	certFile := filepath.Join(certsPath, "default", "test-service", "tls.crt")
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		t.Error("certificate file was not created")
	}
}

// TestMTLSConfig_VerifyCertificate tests certificate verification
func TestMTLSConfig_VerifyCertificate(t *testing.T) {
	tmpDir := t.TempDir()
	certsPath := filepath.Join(tmpDir, "certs")

	mtlsConfig, err := NewMTLSConfig(certsPath)
	if err != nil {
		t.Fatalf("NewMTLSConfig failed: %v", err)
	}

	// Generate certificate
	cert, err := mtlsConfig.GenerateServiceCertificate("test-service", "default")
	if err != nil {
		t.Fatalf("GenerateServiceCertificate failed: %v", err)
	}

	// Verify certificate
	err = mtlsConfig.VerifyCertificate(cert.CertPEM)
	if err != nil {
		t.Fatalf("VerifyCertificate failed: %v", err)
	}
}

// TestMTLSConfig_RenewCertificate tests certificate renewal
func TestMTLSConfig_RenewCertificate(t *testing.T) {
	tmpDir := t.TempDir()
	certsPath := filepath.Join(tmpDir, "certs")

	mtlsConfig, err := NewMTLSConfig(certsPath)
	if err != nil {
		t.Fatalf("NewMTLSConfig failed: %v", err)
	}

	// Generate initial certificate
	oldCert, err := mtlsConfig.GenerateServiceCertificate("test-service", "default")
	if err != nil {
		t.Fatalf("GenerateServiceCertificate failed: %v", err)
	}

	oldSerial := oldCert.SerialNumber

	// Wait a bit to ensure different timestamp
	time.Sleep(10 * time.Millisecond)

	// Renew certificate
	newCert, err := mtlsConfig.RenewCertificate("test-service", "default")
	if err != nil {
		t.Fatalf("RenewCertificate failed: %v", err)
	}

	// Verify new certificate has different serial number
	if oldSerial.Cmp(newCert.SerialNumber) == 0 {
		t.Error("expected different serial number after renewal")
	}
}

// TestMTLSConfig_RevokeCertificate tests certificate revocation
func TestMTLSConfig_RevokeCertificate(t *testing.T) {
	tmpDir := t.TempDir()
	certsPath := filepath.Join(tmpDir, "certs")

	mtlsConfig, err := NewMTLSConfig(certsPath)
	if err != nil {
		t.Fatalf("NewMTLSConfig failed: %v", err)
	}

	// Generate certificate
	_, err = mtlsConfig.GenerateServiceCertificate("test-service", "default")
	if err != nil {
		t.Fatalf("GenerateServiceCertificate failed: %v", err)
	}

	// Revoke certificate
	err = mtlsConfig.RevokeCertificate("test-service", "default")
	if err != nil {
		t.Fatalf("RevokeCertificate failed: %v", err)
	}

	// Verify certificate was removed
	_, err = mtlsConfig.GetCertificate("test-service", "default")
	if err == nil {
		t.Error("expected error when getting revoked certificate")
	}
}

// TestCertificate_IsValid tests certificate validity check
func TestCertificate_IsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		cert      *Certificate
		expected  bool
	}{
		{
			name: "valid certificate",
			cert: &Certificate{
				NotBefore: now.Add(-1 * time.Hour),
				NotAfter:  now.Add(1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "not yet valid",
			cert: &Certificate{
				NotBefore: now.Add(1 * time.Hour),
				NotAfter:  now.Add(2 * time.Hour),
			},
			expected: false,
		},
		{
			name: "expired",
			cert: &Certificate{
				NotBefore: now.Add(-2 * time.Hour),
				NotAfter:  now.Add(-1 * time.Hour),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cert.IsValid()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestCertificate_IsExpiringSoon tests expiration check
func TestCertificate_IsExpiringSoon(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		cert      *Certificate
		within    time.Duration
		expected  bool
	}{
		{
			name: "expiring soon",
			cert: &Certificate{
				NotAfter: now.Add(6 * time.Hour),
			},
			within:   12 * time.Hour,
			expected: true,
		},
		{
			name: "not expiring soon",
			cert: &Certificate{
				NotAfter: now.Add(48 * time.Hour),
			},
			within:   12 * time.Hour,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cert.IsExpiringSoon(tt.within)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestRetryPolicy_Validate tests retry policy validation
func TestRetryPolicy_Validate(t *testing.T) {
	tests := []struct {
		name      string
		policy    *RetryPolicy
		expectErr bool
	}{
		{
			name: "valid policy",
			policy: &RetryPolicy{
				Attempts:      3,
				PerTryTimeout: "5s",
				RetryOn:       []string{"5xx"},
			},
			expectErr: false,
		},
		{
			name: "invalid - zero attempts",
			policy: &RetryPolicy{
				Attempts:      0,
				PerTryTimeout: "5s",
				RetryOn:       []string{"5xx"},
			},
			expectErr: true,
		},
		{
			name: "invalid - too many attempts",
			policy: &RetryPolicy{
				Attempts:      15,
				PerTryTimeout: "5s",
				RetryOn:       []string{"5xx"},
			},
			expectErr: true,
		},
		{
			name: "invalid - no retry conditions",
			policy: &RetryPolicy{
				Attempts:      3,
				PerTryTimeout: "5s",
				RetryOn:       []string{},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}
		})
	}
}

// setupTestManager creates a test service mesh manager
func setupTestManager(t *testing.T, tmpDir string) *Manager {
	config := defaultConfig()
	return setupTestManagerWithConfig(t, tmpDir, config)
}

// setupTestManagerWithConfig creates a test manager with custom config
func setupTestManagerWithConfig(t *testing.T, tmpDir string, config *Config) *Manager {
	configPath := filepath.Join(tmpDir, "servicemesh.yaml")
	storePath := filepath.Join(tmpDir, "store")

	// Write config file
	data, _ := yaml.Marshal(config)
	os.WriteFile(configPath, data, 0644)

	mgr, err := NewManager(configPath, storePath)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}

	return mgr
}

// TestEnvoyConfig_Validation tests Envoy configuration validation
func TestEnvoyConfig_Validation(t *testing.T) {
	tests := []struct {
		name      string
		config    *EnvoyConfig
		expectErr bool
	}{
		{
			name: "valid config",
			config: &EnvoyConfig{
				ServiceName:      "test",
				ServiceNamespace: "default",
				ServicePort:      8080,
			},
			expectErr: false,
		},
		{
			name: "missing service name",
			config: &EnvoyConfig{
				ServiceName:      "",
				ServiceNamespace: "default",
				ServicePort:      8080,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEnvoySidecar(tt.config)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error %v, got %v", tt.expectErr, err)
			}
		})
	}
}

// TestManager_ListServices tests listing services
func TestManager_ListServices(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := setupTestManager(t, tmpDir)
	defer mgr.Close()

	ctx := context.Background()

	// Register multiple services
	services := []*Service{
		{ID: "svc-1", Name: "service-1", Namespace: "default", Port: 8080},
		{ID: "svc-2", Name: "service-2", Namespace: "default", Port: 8081},
		{ID: "svc-3", Name: "service-3", Namespace: "production", Port: 8082},
	}

	for _, svc := range services {
		mgr.RegisterService(ctx, svc)
	}

	// List all services
	allServices := mgr.ListServices()
	if len(allServices) != 3 {
		t.Errorf("expected 3 services, got %d", len(allServices))
	}
}

// TestDefaultPolicies tests that default policies are created correctly
func TestDefaultPolicies(t *testing.T) {
	policies := []struct {
		name   string
		policy *TrafficPolicy
	}{
		{"default", DefaultTrafficPolicy()},
		{"high-availability", HighAvailabilityPolicy()},
		{"performance", PerformancePolicy()},
	}

	for _, p := range policies {
		t.Run(p.name, func(t *testing.T) {
			if err := p.policy.Validate(); err != nil {
				t.Errorf("default policy %s is invalid: %v", p.name, err)
			}
		})
	}
}
