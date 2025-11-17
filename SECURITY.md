# Security Policy

## Supported Versions

We release patches for security vulnerabilities in the following versions:

| Version | Supported          | Status      |
| ------- | ------------------ | ----------- |
| 1.x.x   | :white_check_mark: | Current     |
| < 1.0   | :x:                | Unsupported |

**Note:** As an educational project, we prioritize fixing security issues promptly but recommend using established production runtimes (Docker, containerd, Podman) for production workloads.

## Our Commitment

Containr takes security seriously despite being primarily an educational project. We are committed to:

- **Prompt Response:** Acknowledge security reports within 48 hours
- **Transparent Communication:** Keep reporters informed of progress
- **Coordinated Disclosure:** Work with security researchers on responsible disclosure
- **Learning Opportunity:** Use security issues as educational opportunities

## Reporting a Vulnerability

### How to Report

**Please DO NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them privately using one of these methods:

#### 1. GitHub Security Advisories (Preferred)

1. Go to the [Security tab](https://github.com/therealutkarshpriyadarshi/containr/security)
2. Click "Report a vulnerability"
3. Fill out the form with details
4. Submit privately

#### 2. Email

Send an email to **security@containr-project.org** (replace with actual email) with:

- **Subject:** `[SECURITY] Brief description`
- **Content:**
  - Description of the vulnerability
  - Steps to reproduce
  - Potential impact
  - Suggested fix (if any)
  - Your name and contact info (for credit)

#### 3. Encrypted Email (for sensitive issues)

For highly sensitive vulnerabilities, use PGP encryption:

- **PGP Key:** [Available at keybase.io/containr](https://keybase.io/containr) (example)
- **Fingerprint:** `XXXX XXXX XXXX XXXX XXXX XXXX XXXX XXXX`

### What to Include

A good security report should include:

```markdown
## Vulnerability Description
Brief description of the vulnerability

## Affected Components
- Package: pkg/namespace
- Version: 1.0.0
- Function: SetupNamespaces()

## Attack Vector
How can this vulnerability be exploited?

## Impact Assessment
- Confidentiality: High/Medium/Low
- Integrity: High/Medium/Low
- Availability: High/Medium/Low
- Scope: System/Container/Limited

## Steps to Reproduce
1. Step one
2. Step two
3. Expected result vs actual result

## Proof of Concept
```go
// Code demonstrating the vulnerability
```

## Suggested Fix
Possible solution or mitigation

## Additional Context
- Environment details
- Related issues
- References
```

### What to Expect

After you submit a vulnerability report:

1. **Acknowledgment (48 hours):**
   - We'll confirm receipt of your report
   - Assign a tracking number
   - Provide expected timeline

2. **Initial Assessment (1 week):**
   - Verify the vulnerability
   - Assess severity using CVSS
   - Determine affected versions
   - Estimate fix timeline

3. **Fix Development (2-4 weeks):**
   - Develop and test fix
   - Create security advisory
   - Prepare release notes
   - Keep you updated on progress

4. **Coordinated Disclosure:**
   - Agree on disclosure date
   - Publish fix and advisory
   - Credit reporter (if desired)
   - Announce to community

### Disclosure Timeline

We follow a **90-day disclosure deadline** from initial report:

- **Day 0:** Vulnerability reported
- **Day 7:** Severity assessment complete
- **Day 30:** Fix developed and tested (target)
- **Day 45:** Security release published (target)
- **Day 90:** Public disclosure (maximum)

If we need more time, we'll discuss extension with reporter.

## Security Vulnerability Response

### Severity Levels

We use [CVSS v3.1](https://www.first.org/cvss/calculator/3.1) to assess severity:

| CVSS Score | Severity | Response Time | Priority |
|------------|----------|---------------|----------|
| 9.0 - 10.0 | Critical | 24-48 hours  | P0       |
| 7.0 - 8.9  | High     | 1 week       | P1       |
| 4.0 - 6.9  | Medium   | 2-4 weeks    | P2       |
| 0.1 - 3.9  | Low      | 4-8 weeks    | P3       |

**Critical (9.0-10.0):** Container escape, privilege escalation, remote code execution

**High (7.0-8.9):** Denial of service, information disclosure, authentication bypass

**Medium (4.0-6.9):** Limited DoS, minor information leaks, input validation issues

**Low (0.1-3.9):** Minor issues with limited impact

### Response Team

Security issues are handled by:
- **Security Lead:** [Name] (email)
- **BDFL:** Project founder
- **Core Maintainers:** See [MAINTAINERS.md](MAINTAINERS.md)

### Fix and Release Process

1. **Private Development:**
   - Fix developed in private repository
   - Tested thoroughly
   - Reviewed by security team

2. **Security Release:**
   - New version with fix
   - Security advisory published
   - CVE assigned (if applicable)
   - Announcement to community

3. **Backporting:**
   - Critical/High: Backport to all supported versions
   - Medium: Backport to current version only
   - Low: Fix in next regular release

## Known Security Considerations

### Educational Project Disclaimer

⚠️ **Important:** Containr is primarily an educational project. While we take security seriously, it is **not recommended for production use**. For production workloads, use:

- [Docker](https://www.docker.com/)
- [Podman](https://podman.io/)
- [containerd](https://containerd.io/)
- [CRI-O](https://cri-o.io/)

### Inherent Limitations

As an educational container runtime, Containr has limitations:

1. **Root Required:** Most operations require root privileges
2. **Limited Security Auditing:** Not professionally audited
3. **Developing Features:** Some features still in development
4. **Testing Coverage:** Not as extensive as production runtimes

### Security Features

Containr implements several security features:

- **Namespaces:** Process, network, mount, UTS, IPC isolation
- **Cgroups:** Resource limits and accounting
- **Capabilities:** Linux capability management
- **Seccomp:** Syscall filtering (Phase 1.2)
- **AppArmor/SELinux:** MAC support (Phase 1.2)
- **User Namespaces:** UID/GID remapping (Phase 2.4)

See [docs/SECURITY.md](docs/SECURITY.md) for detailed security documentation.

## Security Best Practices

### For Users

1. **Run with Minimal Privileges:**
   ```bash
   # Use user namespaces when possible
   containr run --userns-remap alpine /bin/sh
   ```

2. **Apply Resource Limits:**
   ```bash
   # Limit container resources
   containr run --memory 100m --cpus 0.5 alpine /bin/sh
   ```

3. **Drop Unnecessary Capabilities:**
   ```bash
   # Drop dangerous capabilities
   containr run --cap-drop=NET_RAW alpine /bin/sh
   ```

4. **Use Seccomp Profiles:**
   ```bash
   # Apply restrictive seccomp profile
   containr run --security-opt seccomp=default.json alpine /bin/sh
   ```

5. **Keep Updated:**
   ```bash
   # Check for updates regularly
   containr version
   ```

### For Developers

1. **Input Validation:** Always validate user input
2. **Error Handling:** Don't expose sensitive info in errors
3. **Resource Cleanup:** Ensure proper cleanup on all paths
4. **Least Privilege:** Run with minimum required privileges
5. **Security Testing:** Include security tests in PR

## Security Checklist for Contributors

Before submitting code that touches security-sensitive areas:

- [ ] Input validation is comprehensive
- [ ] Error messages don't leak sensitive information
- [ ] Resource cleanup happens in all paths (including errors)
- [ ] No hardcoded credentials or secrets
- [ ] Privilege escalation is properly controlled
- [ ] Race conditions are considered and handled
- [ ] Tests include security scenarios
- [ ] Documentation mentions security implications
- [ ] Follows secure coding guidelines

## CVE Assignment

For confirmed vulnerabilities, we may request a CVE (Common Vulnerabilities and Exposures) ID:

1. Severity is High or Critical
2. Affects released versions
3. Has security impact on users

**CVE Numbering Authority:** We'll work with GitHub Security Advisories or MITRE for CVE assignment.

## Security Hall of Fame

We recognize security researchers who help improve Containr:

### 2025

- *No vulnerabilities reported yet*

### Credit Policy

- Reporters will be credited in:
  - Security advisory
  - Release notes
  - SECURITY.md (this file)
  - GitHub Security Advisory
- Anonymous reporting is supported
- We respect embargo agreements

## Past Security Advisories

### 2025

- *No security advisories published yet*

All security advisories are available at:
[https://github.com/therealutkarshpriyadarshi/containr/security/advisories](https://github.com/therealutkarshpriyadarshi/containr/security/advisories)

## Hardening Guides

See our security documentation:

- [Security Architecture](docs/SECURITY.md)
- [Capabilities Guide](docs/SECURITY.md#capabilities)
- [Seccomp Profiles](docs/SECURITY.md#seccomp)
- [LSM Support](docs/SECURITY.md#lsm)
- [User Namespaces](docs/PHASE2.md#user-namespaces)

## Security Audit History

### External Audits

- *No external security audits conducted yet*

### Planned Audits

- Internal security review (Q2 2026)
- Consider professional audit when reaching 1000+ users

## Dependencies Security

### Dependency Scanning

We use automated tools to scan dependencies:

- **Dependabot:** Enabled for automatic updates
- **Snyk:** Vulnerability scanning (planned)
- **Go vulnerability database:** `govulncheck`

### Updating Dependencies

```bash
# Check for vulnerable dependencies
go list -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy
```

## Contact

### Security Team

- **Email:** security@containr-project.org (replace with actual)
- **GitHub:** [@containr-security](https://github.com/containr-security) (example)
- **PGP Key:** Available at keybase.io/containr

### General Security Questions

For non-vulnerability security questions:
- GitHub Discussions: [Discussions](https://github.com/therealutkarshpriyadarshi/containr/discussions)
- Documentation: [docs/SECURITY.md](docs/SECURITY.md)

## Additional Resources

- [OWASP Container Security](https://owasp.org/www-project-docker-security/)
- [CIS Docker Benchmark](https://www.cisecurity.org/benchmark/docker)
- [NIST Container Security](https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-190.pdf)
- [Linux Kernel Security](https://www.kernel.org/doc/html/latest/admin-guide/security-bugs.html)

## Updates to This Policy

This security policy is reviewed and updated:
- Quarterly
- After major security incidents
- When new security features are added
- Based on community feedback

**Last Updated:** November 17, 2025
**Version:** 1.0
**Next Review:** February 17, 2026

---

Thank you for helping keep Containr and its users safe!
