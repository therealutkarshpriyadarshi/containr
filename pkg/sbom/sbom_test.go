package sbom

import (
	"context"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator(FormatSPDX)
	if gen == nil {
		t.Fatal("Expected generator to be created")
	}

	if gen.format != FormatSPDX {
		t.Errorf("Expected format %s, got %s", FormatSPDX, gen.format)
	}
}

func TestGenerate(t *testing.T) {
	gen := NewGenerator(FormatSPDX)
	ctx := context.Background()

	sbom, err := gen.Generate(ctx, "alpine:latest")
	if err != nil {
		t.Fatalf("Failed to generate SBOM: %v", err)
	}

	if sbom == nil {
		t.Fatal("Expected SBOM to be generated")
	}

	if sbom.Format != FormatSPDX {
		t.Errorf("Expected format %s, got %s", FormatSPDX, sbom.Format)
	}
}

func TestNewScanner(t *testing.T) {
	scanner := NewScanner(BackendTrivy)
	if scanner == nil {
		t.Fatal("Expected scanner to be created")
	}

	if scanner.backend != BackendTrivy {
		t.Errorf("Expected backend %s, got %s", BackendTrivy, scanner.backend)
	}
}

func TestScan(t *testing.T) {
	scanner := NewScanner(BackendTrivy)
	ctx := context.Background()

	result, err := scanner.Scan(ctx, "alpine:latest")
	if err != nil {
		t.Fatalf("Failed to scan: %v", err)
	}

	if result == nil {
		t.Fatal("Expected scan result")
	}

	if result.Summary == nil {
		t.Fatal("Expected scan summary")
	}
}

func TestFilterBySeverity(t *testing.T) {
	scanner := NewScanner(BackendTrivy)

	result := &ScanResult{
		Vulnerabilities: []*Vulnerability{
			{ID: "CVE-1", Severity: "critical"},
			{ID: "CVE-2", Severity: "high"},
			{ID: "CVE-3", Severity: "medium"},
			{ID: "CVE-4", Severity: "low"},
		},
	}

	filtered := scanner.FilterBySeverity(result, "high")
	if len(filtered) != 2 {
		t.Errorf("Expected 2 vulnerabilities, got %d", len(filtered))
	}
}

func TestComplianceChecker(t *testing.T) {
	checker := NewComplianceChecker(
		[]string{"MIT", "Apache-2.0"},
		[]string{"GPL"},
		"high",
	)

	if checker == nil {
		t.Fatal("Expected compliance checker to be created")
	}
}

func TestComplianceCheck(t *testing.T) {
	checker := NewComplianceChecker(
		[]string{"MIT"},
		[]string{"GPL"},
		"medium",
	)

	sbom := &SBOM{
		Packages: []*Package{
			{Name: "pkg1", Licenses: []string{"MIT"}},
			{Name: "pkg2", Licenses: []string{"GPL"}},
		},
	}

	scanResult := &ScanResult{
		Vulnerabilities: []*Vulnerability{
			{ID: "CVE-1", Package: "pkg1", Severity: "critical"},
		},
	}

	result, err := checker.Check(sbom, scanResult)
	if err != nil {
		t.Fatalf("Compliance check failed: %v", err)
	}

	if result.Compliant {
		t.Error("Expected non-compliant result")
	}

	if len(result.Violations) == 0 {
		t.Error("Expected violations")
	}
}
