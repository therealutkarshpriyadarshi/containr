package advanced

import (
	"context"
	"testing"
	"time"
)

// TestOPAEngine tests OPA policy engine functionality
func TestOPAEngine_AddPolicy(t *testing.T) {
	engine, err := NewOPAEngine(nil)
	if err != nil {
		t.Fatalf("NewOPAEngine failed: %v", err)
	}

	ctx := context.Background()

	policy := &Policy{
		ID:          "test-policy-1",
		Name:        "Test Policy",
		Description: "Test policy description",
		Package:     "test.policies",
		Rules:       []string{"deny_privileged"},
	}

	err = engine.AddPolicy(ctx, policy)
	if err != nil {
		t.Fatalf("AddPolicy failed: %v", err)
	}

	// Verify policy was added
	retrieved, err := engine.GetPolicy("test-policy-1")
	if err != nil {
		t.Fatalf("GetPolicy failed: %v", err)
	}

	if retrieved.Name != "Test Policy" {
		t.Errorf("expected policy name 'Test Policy', got %s", retrieved.Name)
	}
}

func TestOPAEngine_Evaluate(t *testing.T) {
	engine, err := NewOPAEngine(nil)
	if err != nil {
		t.Fatalf("NewOPAEngine failed: %v", err)
	}

	ctx := context.Background()

	// Add test policy
	policy := &Policy{
		ID:      "security-policy",
		Name:    "Security Policy",
		Package: "containr.security",
		Rules:   []string{"deny_privileged", "deny_host_network"},
	}
	engine.AddPolicy(ctx, policy)

	tests := []struct {
		name           string
		input          *PolicyInput
		expectAllow    bool
		expectViolation bool
	}{
		{
			name: "allow non-privileged container",
			input: &PolicyInput{
				Resource: "container",
				Action:   "create",
				Subject:  "user1",
				Attributes: map[string]interface{}{
					"privileged":    false,
					"host_network":  false,
				},
			},
			expectAllow:    true,
			expectViolation: false,
		},
		{
			name: "deny privileged container",
			input: &PolicyInput{
				Resource: "container",
				Action:   "create",
				Subject:  "user1",
				Attributes: map[string]interface{}{
					"privileged":    true,
					"host_network":  false,
				},
			},
			expectAllow:    false,
			expectViolation: true,
		},
		{
			name: "deny host network",
			input: &PolicyInput{
				Resource: "container",
				Action:   "create",
				Subject:  "user1",
				Attributes: map[string]interface{}{
					"privileged":    false,
					"host_network":  true,
				},
			},
			expectAllow:    false,
			expectViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := engine.Evaluate(ctx, tt.input)
			if err != nil {
				t.Fatalf("Evaluate failed: %v", err)
			}

			if decision.Allow != tt.expectAllow {
				t.Errorf("expected allow=%v, got %v", tt.expectAllow, decision.Allow)
			}

			hasViolations := len(decision.Violations) > 0
			if hasViolations != tt.expectViolation {
				t.Errorf("expected violations=%v, got %v (violations: %v)",
					tt.expectViolation, hasViolations, decision.Violations)
			}
		})
	}
}

func TestOPAEngine_PolicyCache(t *testing.T) {
	config := &OPAConfig{
		EnableCache: true,
		CacheTTL:    1 * time.Second,
	}

	engine, err := NewOPAEngine(config)
	if err != nil {
		t.Fatalf("NewOPAEngine failed: %v", err)
	}

	ctx := context.Background()

	// Add policy
	policy := &Policy{
		ID:      "cache-test",
		Name:    "Cache Test Policy",
		Package: "test",
		Rules:   []string{},
	}
	engine.AddPolicy(ctx, policy)

	input := &PolicyInput{
		Resource: "container",
		Action:   "create",
		Subject:  "user1",
	}

	// First evaluation
	decision1, err := engine.Evaluate(ctx, input)
	if err != nil {
		t.Fatalf("First evaluation failed: %v", err)
	}

	// Second evaluation should use cache
	decision2, err := engine.Evaluate(ctx, input)
	if err != nil {
		t.Fatalf("Second evaluation failed: %v", err)
	}

	if decision1.Allow != decision2.Allow {
		t.Errorf("cached decision differs from original")
	}

	// Wait for cache to expire
	time.Sleep(2 * time.Second)

	// Third evaluation should re-evaluate
	decision3, err := engine.Evaluate(ctx, input)
	if err != nil {
		t.Fatalf("Third evaluation failed: %v", err)
	}

	if decision3 == nil {
		t.Error("expected non-nil decision after cache expiry")
	}
}

func TestOPAEngine_DefaultPolicies(t *testing.T) {
	policies := DefaultSecurityPolicies()

	if len(policies) == 0 {
		t.Error("expected default policies, got none")
	}

	// Verify baseline policy exists
	found := false
	for _, policy := range policies {
		if policy.ID == "pod-security-baseline" {
			found = true
			if len(policy.Rules) == 0 {
				t.Error("baseline policy has no rules")
			}
		}
	}

	if !found {
		t.Error("pod-security-baseline policy not found in defaults")
	}
}

// TestCosignVerifier tests Cosign image verification
func TestCosignVerifier_VerifyImage(t *testing.T) {
	verifier, err := NewCosignVerifier(nil)
	if err != nil {
		t.Fatalf("NewCosignVerifier failed: %v", err)
	}

	ctx := context.Background()

	imageRef := "docker.io/library/nginx:latest"

	result, err := verifier.VerifyImage(ctx, imageRef)
	if err != nil {
		t.Fatalf("VerifyImage failed: %v", err)
	}

	if result.ImageRef != imageRef {
		t.Errorf("expected imageRef %s, got %s", imageRef, result.ImageRef)
	}

	if result.Digest == "" {
		t.Error("expected non-empty digest")
	}

	if len(result.Signatures) == 0 {
		t.Error("expected at least one signature")
	}
}

func TestCosignVerifier_SignImage(t *testing.T) {
	verifier, err := NewCosignVerifier(nil)
	if err != nil {
		t.Fatalf("NewCosignVerifier failed: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name    string
		req     *SigningRequest
		wantErr bool
	}{
		{
			name: "keyless signing",
			req: &SigningRequest{
				ImageRef: "docker.io/test/image:v1",
				Keyless:  true,
			},
			wantErr: false,
		},
		{
			name: "key-based signing",
			req: &SigningRequest{
				ImageRef:   "docker.io/test/image:v1",
				PrivateKey: []byte("test-private-key"),
				Keyless:    false,
			},
			wantErr: false,
		},
		{
			name: "missing image ref",
			req: &SigningRequest{
				Keyless: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig, err := verifier.SignImage(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SignImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && sig == nil {
				t.Error("expected signature, got nil")
			}

			if !tt.wantErr && sig.Signature == "" {
				t.Error("expected non-empty signature")
			}
		})
	}
}

func TestCosignVerifier_Keyring(t *testing.T) {
	verifier, err := NewCosignVerifier(nil)
	if err != nil {
		t.Fatalf("NewCosignVerifier failed: %v", err)
	}

	key := &PublicKey{
		ID:      "test-key-1",
		Name:    "Test Key",
		KeyData: []byte("test-key-data"),
	}

	// Add key
	err = verifier.AddPublicKey(key)
	if err != nil {
		t.Fatalf("AddPublicKey failed: %v", err)
	}

	// List keys
	keys := verifier.ListPublicKeys()
	if len(keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(keys))
	}

	// Remove key
	err = verifier.RemovePublicKey("test-key-1")
	if err != nil {
		t.Fatalf("RemovePublicKey failed: %v", err)
	}

	// Verify key was removed
	keys = verifier.ListPublicKeys()
	if len(keys) != 0 {
		t.Errorf("expected 0 keys after removal, got %d", len(keys))
	}
}

func TestKeyring_ExportImport(t *testing.T) {
	keyring := NewKeyring()

	// Add keys
	key1 := &PublicKey{
		ID:      "key1",
		Name:    "Key 1",
		KeyData: []byte("key-data-1"),
	}
	key2 := &PublicKey{
		ID:      "key2",
		Name:    "Key 2",
		KeyData: []byte("key-data-2"),
	}

	keyring.AddKey(key1)
	keyring.AddKey(key2)

	// Export keyring
	data, err := keyring.ExportKeyring()
	if err != nil {
		t.Fatalf("ExportKeyring failed: %v", err)
	}

	// Create new keyring and import
	newKeyring := NewKeyring()
	err = newKeyring.ImportKeyring(data)
	if err != nil {
		t.Fatalf("ImportKeyring failed: %v", err)
	}

	// Verify keys were imported
	keys := newKeyring.ListKeys()
	if len(keys) != 2 {
		t.Errorf("expected 2 keys after import, got %d", len(keys))
	}
}

// TestRuntimeMonitor tests runtime security monitoring
func TestRuntimeMonitor_ProcessEvent(t *testing.T) {
	monitor, err := NewRuntimeMonitor(nil)
	if err != nil {
		t.Fatalf("NewRuntimeMonitor failed: %v", err)
	}

	ctx := context.Background()

	// Start monitor
	err = monitor.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer monitor.Stop()

	tests := []struct {
		name        string
		event       *RuntimeEvent
		expectAlert bool
	}{
		{
			name: "benign process",
			event: &RuntimeEvent{
				Type:        EventTypeProcess,
				Timestamp:   time.Now(),
				ContainerID: "container-1",
				Command:     "/usr/bin/python app.py",
			},
			expectAlert: false,
		},
		{
			name: "suspicious process - netcat",
			event: &RuntimeEvent{
				Type:        EventTypeProcess,
				Timestamp:   time.Now(),
				ContainerID: "container-1",
				Command:     "/bin/nc -l -p 4444",
			},
			expectAlert: true,
		},
		{
			name: "critical file modification",
			event: &RuntimeEvent{
				Type:        EventTypeFile,
				Timestamp:   time.Now(),
				ContainerID: "container-1",
				FilePath:    "/etc/passwd",
			},
			expectAlert: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := monitor.ProcessEvent(ctx, tt.event)
			if err != nil {
				t.Fatalf("ProcessEvent failed: %v", err)
			}

			// Give time for alert processing
			time.Sleep(100 * time.Millisecond)

			// Check metrics
			metrics := monitor.GetMetrics()
			running, ok := metrics["running"].(bool)
			if !ok || !running {
				t.Error("monitor should be running")
			}
		})
	}
}

func TestRuntimeMonitor_Detectors(t *testing.T) {
	ctx := context.Background()

	// Test process detector
	processDetector := &ProcessThreatDetector{}
	event := &RuntimeEvent{
		Type:    EventTypeProcess,
		Command: "/usr/bin/nmap -sS target.com",
	}

	detection, err := processDetector.Detect(ctx, event)
	if err != nil {
		t.Fatalf("Process detector failed: %v", err)
	}

	if detection == nil {
		t.Error("expected detection for suspicious command")
	} else {
		if detection.ThreatType != ThreatTypeMaliciousProcess {
			t.Errorf("expected threat type %s, got %s", ThreatTypeMaliciousProcess, detection.ThreatType)
		}
	}

	// Test file integrity detector
	fileDetector := &FileIntegrityDetector{}
	fileEvent := &RuntimeEvent{
		Type:     EventTypeFile,
		FilePath: "/etc/shadow",
	}

	detection, err = fileDetector.Detect(ctx, fileEvent)
	if err != nil {
		t.Fatalf("File detector failed: %v", err)
	}

	if detection == nil {
		t.Error("expected detection for critical file modification")
	} else {
		if detection.Severity != SeverityCritical {
			t.Errorf("expected severity %s, got %s", SeverityCritical, detection.Severity)
		}
	}

	// Test network detector
	networkDetector := &NetworkThreatDetector{}
	networkEvent := &RuntimeEvent{
		Type: EventTypeNetwork,
		Network: &NetworkEvent{
			Protocol:   "tcp",
			DestIP:     "192.168.1.100",
			DestPort:   4444,
			BytesSent:  1024,
		},
	}

	detection, err = networkDetector.Detect(ctx, networkEvent)
	if err != nil {
		t.Fatalf("Network detector failed: %v", err)
	}

	if detection == nil {
		t.Error("expected detection for suspicious port")
	}
}

func TestRuntimeMonitor_Incidents(t *testing.T) {
	monitor, err := NewRuntimeMonitor(nil)
	if err != nil {
		t.Fatalf("NewRuntimeMonitor failed: %v", err)
	}

	ctx := context.Background()

	// Create high severity alert
	alert := &SecurityAlert{
		ID:          "alert-1",
		Timestamp:   time.Now(),
		Severity:    SeverityHigh,
		Type:        ThreatTypeMaliciousProcess,
		ContainerID: "container-1",
		Status:      AlertStatusNew,
		Detection: &ThreatDetection{
			Severity:    SeverityHigh,
			ThreatType:  ThreatTypeMaliciousProcess,
			Description: "Suspicious process detected",
			Confidence:  0.9,
		},
	}

	monitor.handleAlert(ctx, alert)

	// Get incidents
	incidents := monitor.GetIncidents(10)

	if len(incidents) == 0 {
		t.Error("expected incident to be created for high severity alert")
	} else {
		if incidents[0].Severity != SeverityHigh {
			t.Errorf("expected incident severity %s, got %s", SeverityHigh, incidents[0].Severity)
		}
	}
}

// TestScanner tests vulnerability and compliance scanning
func TestScanner_ScanImage(t *testing.T) {
	scanner, err := NewScanner(nil)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	ctx := context.Background()

	req := &ScanRequest{
		Target: "docker.io/library/nginx:latest",
		Type:   ScanTypeImage,
	}

	result, err := scanner.Scan(ctx, req)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.Status != ScanStatusCompleted {
		t.Errorf("expected status %s, got %s", ScanStatusCompleted, result.Status)
	}

	if result.Summary == nil {
		t.Error("expected non-nil summary")
	}

	if len(result.Vulnerabilities) == 0 {
		t.Log("warning: no vulnerabilities found (expected for test image)")
	}
}

func TestScanner_ScanCompliance(t *testing.T) {
	scanner, err := NewScanner(nil)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	ctx := context.Background()

	req := &ScanRequest{
		Target: "docker.io/library/nginx:latest",
		Type:   ScanTypeCompliance,
	}

	result, err := scanner.Scan(ctx, req)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}

	if result.Compliance == nil {
		t.Error("expected compliance result")
	} else {
		if result.Compliance.TotalControls == 0 {
			t.Error("expected compliance controls")
		}

		if result.Compliance.Score < 0 || result.Compliance.Score > 100 {
			t.Errorf("invalid compliance score: %.2f", result.Compliance.Score)
		}
	}
}

func TestScanner_FullScan(t *testing.T) {
	scanner, err := NewScanner(nil)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	ctx := context.Background()

	req := &ScanRequest{
		Target: "docker.io/library/alpine:latest",
		Type:   ScanTypeFull,
	}

	result, err := scanner.Scan(ctx, req)
	if err != nil {
		t.Fatalf("Full scan failed: %v", err)
	}

	// Verify both vulnerability and compliance results
	if result.Summary == nil {
		t.Error("expected scan summary")
	}

	if result.Compliance == nil {
		t.Error("expected compliance results in full scan")
	}

	if result.Duration == 0 {
		t.Error("expected non-zero scan duration")
	}
}

func TestScanner_ScanHistory(t *testing.T) {
	config := &ScannerConfig{
		MaxScanHistory: 5,
	}

	scanner, err := NewScanner(config)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	ctx := context.Background()

	// Perform multiple scans
	for i := 0; i < 10; i++ {
		req := &ScanRequest{
			Target: "test-image",
			Type:   ScanTypeImage,
		}
		scanner.Scan(ctx, req)
	}

	// Get history
	history := scanner.GetScanHistory(10)

	// Verify history is limited to MaxScanHistory
	if len(history) > config.MaxScanHistory {
		t.Errorf("expected at most %d history entries, got %d", config.MaxScanHistory, len(history))
	}
}

func TestVulnerabilityDatabase_FindVulnerabilities(t *testing.T) {
	db := NewVulnerabilityDatabase()

	pkg := &Package{
		Name:    "openssl",
		Version: "1.1.1",
		Type:    "os",
	}

	vulns := db.FindVulnerabilities(pkg)

	if len(vulns) == 0 {
		t.Error("expected to find vulnerabilities for openssl")
	}

	for _, vuln := range vulns {
		if vuln.Severity == "" {
			t.Error("vulnerability missing severity")
		}
		if vuln.Score == 0 {
			t.Error("vulnerability missing score")
		}
	}
}

func TestComplianceDatabase_Benchmarks(t *testing.T) {
	db := NewComplianceDatabase()

	// Add benchmark
	benchmark := &ComplianceBenchmark{
		ID:          "test-benchmark",
		Name:        "Test Benchmark",
		Version:     "1.0",
		Framework:   "TEST",
		Controls:    []*ComplianceControl{
			{
				ID:        "1.1",
				Title:     "Test Control",
				Automated: true,
			},
		},
	}

	db.AddBenchmark(benchmark)

	// Retrieve benchmark
	retrieved, err := db.GetBenchmark("TEST")
	if err != nil {
		t.Fatalf("GetBenchmark failed: %v", err)
	}

	if retrieved.Name != "Test Benchmark" {
		t.Errorf("expected benchmark name 'Test Benchmark', got %s", retrieved.Name)
	}

	if len(retrieved.Controls) != 1 {
		t.Errorf("expected 1 control, got %d", len(retrieved.Controls))
	}
}

func TestScanSummary_RiskScore(t *testing.T) {
	scanner, err := NewScanner(nil)
	if err != nil {
		t.Fatalf("NewScanner failed: %v", err)
	}

	result := &ScanResult{
		Vulnerabilities: []*VulnerabilityFinding{
			{
				Vulnerability: &Vulnerability{
					Severity: "Critical",
				},
			},
			{
				Vulnerability: &Vulnerability{
					Severity: "High",
				},
			},
			{
				Vulnerability: &Vulnerability{
					Severity: "Medium",
				},
			},
		},
	}

	summary := scanner.generateSummary(result)

	if summary.TotalVulnerabilities != 3 {
		t.Errorf("expected 3 total vulnerabilities, got %d", summary.TotalVulnerabilities)
	}

	if summary.CriticalCount != 1 {
		t.Errorf("expected 1 critical vulnerability, got %d", summary.CriticalCount)
	}

	if summary.RiskScore == 0 {
		t.Error("expected non-zero risk score")
	}

	// Risk score = 1*10 + 1*5 + 1*2 = 17
	expectedRisk := float64(17)
	if summary.RiskScore != expectedRisk {
		t.Errorf("expected risk score %.0f, got %.0f", expectedRisk, summary.RiskScore)
	}
}

// Benchmark tests
func BenchmarkOPAEngine_Evaluate(b *testing.B) {
	engine, _ := NewOPAEngine(nil)
	ctx := context.Background()

	policy := &Policy{
		ID:      "bench-policy",
		Name:    "Benchmark Policy",
		Package: "bench",
		Rules:   []string{"deny_privileged"},
	}
	engine.AddPolicy(ctx, policy)

	input := &PolicyInput{
		Resource: "container",
		Action:   "create",
		Subject:  "user1",
		Attributes: map[string]interface{}{
			"privileged": false,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Evaluate(ctx, input)
	}
}

func BenchmarkCosignVerifier_VerifyImage(b *testing.B) {
	verifier, _ := NewCosignVerifier(nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		verifier.VerifyImage(ctx, "test-image:latest")
	}
}

func BenchmarkScanner_Scan(b *testing.B) {
	scanner, _ := NewScanner(nil)
	ctx := context.Background()

	req := &ScanRequest{
		Target: "test-image:latest",
		Type:   ScanTypeImage,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner.Scan(ctx, req)
	}
}
