# Containr Testing & Validation Guide

**Last Updated:** November 18, 2025

This comprehensive guide covers all aspects of testing Containr, from unit tests to production readiness validation.

## Table of Contents

1. [Testing Overview](#testing-overview)
2. [Unit & Integration Tests](#unit--integration-tests)
3. [Performance Benchmarking](#performance-benchmarking)
4. [Stress Testing](#stress-testing)
5. [Security Auditing](#security-auditing)
6. [Penetration Testing](#penetration-testing)
7. [Real-World Validation](#real-world-validation)
8. [Test Results & Coverage](#test-results--coverage)

---

## Testing Overview

Containr employs a multi-layered testing strategy:

```
┌─────────────────────────────────────┐
│         Unit Tests (70%+ coverage)  │
├─────────────────────────────────────┤
│         Integration Tests            │
├─────────────────────────────────────┤
│         Performance Benchmarks       │
├─────────────────────────────────────┤
│         Stress Tests (100+ containers)│
├─────────────────────────────────────┤
│         Security Audit               │
├─────────────────────────────────────┤
│         Penetration Testing          │
├─────────────────────────────────────┤
│         Real-World Validation        │
└─────────────────────────────────────┘
```

---

## Unit & Integration Tests

### Running Tests

```bash
# Run all unit tests
make test-unit

# Run integration tests (requires root)
sudo make test-integration

# Run with coverage
make test-coverage

# Run all tests
make test-all
```

### Test Coverage

Current test coverage by package:

| Package         | Coverage | Status |
|----------------|----------|--------|
| benchmark      | 80.6%    | ✅      |
| build          | 80.9%    | ✅      |
| capabilities   | 64.5%    | ✅      |
| cgroup         | 70.5%    | ✅      |
| container      | 65%+     | ✅      |
| network        | 60%+     | ✅      |

**Overall Coverage Target:** >70%

### Viewing Coverage Reports

```bash
# Generate coverage report
make test-coverage

# View HTML report
open coverage/coverage.html
```

---

## Performance Benchmarking

### Running Benchmarks

```bash
# Run all benchmarks
make bench

# Run with CPU profiling
make bench-cpu

# Run with memory profiling
make bench-mem
```

### Benchmark Metrics

Key performance indicators:

- **Container Creation Time:** <2s for cached images
- **Memory Overhead:** <50MB per container
- **Network Latency:** <1ms for bridge networking
- **Build Cache Hit Rate:** >80% for unchanged layers

### Profiling

```bash
# Generate CPU profile
make profile-cpu

# Generate memory profile
make profile-mem

# View profiles in browser
make profile-view-cpu
make profile-view-mem
```

---

## Stress Testing

### Overview

The stress test script validates Containr's stability under heavy load by creating and managing 100+ containers simultaneously.

### Running Stress Tests

```bash
# Default (100 containers)
sudo ./scripts/stress-test.sh

# Custom number of containers
sudo ./scripts/stress-test.sh 200

# With custom image
TEST_IMAGE=ubuntu sudo ./scripts/stress-test.sh 150
```

### Test Phases

1. **Sequential Creation:** Create N containers one by one
2. **Resource Monitoring:** Monitor system resources under load
3. **Cleanup Test:** Verify proper resource cleanup
4. **Rapid Creation/Deletion:** Test rapid start/stop cycles

### Success Criteria

- ✅ >90% container creation success rate
- ✅ <10 leftover cgroups after cleanup
- ✅ <10 leftover veth interfaces after cleanup
- ✅ No memory leaks detected
- ✅ System remains responsive

---

## Security Auditing

### Overview

The security audit script performs comprehensive security analysis including:

- Static code analysis
- Dependency vulnerability scanning
- Capability and privilege checks
- Seccomp profile validation
- Network security assessment
- Input validation verification

### Running Security Audit

```bash
# Run full security audit
sudo ./scripts/security-audit.sh
```

### Prerequisites

Install recommended tools for comprehensive analysis:

```bash
# Security scanners
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Dependency scanner (optional)
brew install sonatype-nexus-community/nancy-tap/nancy
```

### Security Report

Results are saved in `security-audit-[timestamp]/`:

- `security-report.md` - Executive summary
- `gosec.json` - Security scanner results
- `staticcheck.log` - Static analysis issues
- `dependencies.txt` - Dependency list

---

## Penetration Testing

### Overview

Penetration testing validates security controls by attempting common container escape techniques and exploits.

### Running Penetration Tests

```bash
# Run all penetration tests
sudo ./scripts/pentest.sh
```

### Test Categories

1. Namespace Escape Tests
2. Capability Tests
3. Filesystem Escape Tests
4. Device Access Tests
5. Proc Filesystem Tests
6. Seccomp Bypass Tests
7. Resource Limit Bypass
8. Network Security
9. User Namespace Tests
10. Docker Socket Escape

**Exit Codes:**
- `0` - No critical vulnerabilities
- `1` - Vulnerabilities detected (requires immediate attention)

---

## Real-World Validation

### Scenario-Based Testing

Test Containr with real-world use cases:

#### 1. Web Application Stack

```bash
# Run nginx web server
sudo ./bin/containr pull nginx
sudo ./bin/containr run -p 8080:80 --name web nginx
```

#### 2. Database Container

```bash
# Run with persistent storage
sudo ./bin/containr volume create db-data
sudo ./bin/containr run -v db-data:/var/lib/mysql --name db mysql
```

#### 3. Build Environment

```bash
# Multi-stage build
sudo ./bin/containr build -f Dockerfile -t myapp:latest .
```

---

## Test Results & Coverage

### Latest Test Run Summary

**Date:** November 18, 2025

#### Unit Tests
- **Total Packages:** 35+
- **Coverage:** 70.5% (target: >70%)
- **Status:** ✅ PASSING

#### Stress Tests
- **Containers Tested:** 100
- **Success Rate:** 98%
- **Status:** ✅ PASSING

#### Security Audit
- **Tests Run:** 40+
- **Critical Issues:** 0
- **Status:** ✅ PASSING

#### Penetration Tests
- **Tests Run:** 25
- **Vulnerabilities:** 0
- **Status:** ✅ PASSING

---

## Best Practices

### 1. Test Before Commit

```bash
make pre-commit
```

### 2. Monitor Coverage

```bash
make test-coverage
```

### 3. Regular Security Audits

```bash
./scripts/security-audit.sh
```

### 4. Stress Test Before Release

```bash
sudo ./scripts/stress-test.sh 200
```

---

## Resources

- [Architecture Documentation](ARCHITECTURE.md)
- [Security Guide](SECURITY.md)
- [Contributing Guide](../CONTRIBUTING.md)

---

**Happy Testing! 🧪**
