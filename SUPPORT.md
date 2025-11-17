# Support

Looking for help with Containr? You're in the right place!

## Table of Contents

- [Getting Help](#getting-help)
- [Documentation](#documentation)
- [Community Support](#community-support)
- [Bug Reports](#bug-reports)
- [Feature Requests](#feature-requests)
- [Security Issues](#security-issues)
- [Professional Support](#professional-support)
- [Contributing](#contributing)

## Getting Help

### Quick Links

- 📖 **Documentation:** [docs/](docs/)
- 💬 **Discussions:** [GitHub Discussions](https://github.com/therealutkarshpriyadarshi/containr/discussions)
- 🐛 **Bug Reports:** [GitHub Issues](https://github.com/therealutkarshpriyadarshi/containr/issues)
- 🔒 **Security:** [SECURITY.md](SECURITY.md)
- 🤝 **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md)

### Before Asking for Help

Please check these resources first:

1. **README** - [README.md](README.md) for quick start and overview
2. **Documentation** - [docs/](docs/) for detailed guides
3. **Existing Issues** - Search [closed](https://github.com/therealutkarshpriyadarshi/containr/issues?q=is%3Aissue+is%3Aclosed) and [open](https://github.com/therealutkarshpriyadarshi/containr/issues) issues
4. **Discussions** - Browse [discussions](https://github.com/therealutkarshpriyadarshi/containr/discussions)
5. **Troubleshooting** - See troubleshooting sections in docs

## Documentation

### Core Documentation

- **[README.md](README.md)** - Project overview and quick start
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** - System architecture
- **[GETTING_STARTED.md](docs/GETTING_STARTED.md)** - Detailed getting started guide

### Phase Documentation

- **[PHASE1.md](docs/PHASE1.md)** - Foundation features (namespaces, cgroups, security)
- **[PHASE2.md](docs/PHASE2.md)** - Feature completeness (CLI, volumes, registry)
- **[PHASE3.md](docs/PHASE3.md)** - Advanced features (networking, monitoring)
- **[PHASE4.md](docs/PHASE4.md)** - Production polish (performance, OCI)
- **[PHASE5.md](docs/PHASE5.md)** - Community & growth

### Specialized Documentation

- **[SECURITY.md](docs/SECURITY.md)** - Security guide
- **[TESTING.md](docs/TESTING.md)** - Testing guide
- **[LOGGING.md](docs/LOGGING.md)** - Logging and debugging
- **[ERROR_HANDLING.md](docs/ERROR_HANDLING.md)** - Error handling guide

### Development

- **[CONTRIBUTING.md](CONTRIBUTING.md)** - How to contribute
- **[GOVERNANCE.md](GOVERNANCE.md)** - Project governance
- **[ROADMAP.md](ROADMAP.md)** - Project roadmap

## Community Support

### GitHub Discussions (Recommended)

[GitHub Discussions](https://github.com/therealutkarshpriyadarshi/containr/discussions) is the best place for:

- ❓ **Questions** - "How do I...?", "What is...?"
- 💡 **Ideas** - Feature suggestions, improvements
- 🎉 **Show and Tell** - Share what you've built
- 📢 **Announcements** - Project updates
- 💬 **General** - Anything else!

**How to use:**
1. Go to [Discussions](https://github.com/therealutkarshpriyadarshi/containr/discussions)
2. Search existing discussions
3. Create new discussion if needed
4. Choose appropriate category
5. Provide clear title and description

### GitHub Issues

Use [GitHub Issues](https://github.com/therealutkarshpriyadarshi/containr/issues) only for:

- 🐛 **Bug Reports** - Something is broken
- 🚀 **Feature Requests** - Propose new features
- 📝 **Documentation Issues** - Docs are wrong or unclear

**Do NOT use issues for:**
- Questions (use Discussions instead)
- Support requests (use Discussions instead)
- Security vulnerabilities (use [SECURITY.md](SECURITY.md))

### Community Channels

#### Coming Soon
- **Discord/Slack** - Real-time chat (planned)
- **Mailing List** - Email announcements (planned)
- **Forums** - Long-form discussions (planned)

#### Current
- **GitHub Discussions** - Primary community forum
- **GitHub Issues** - Bug reports and features
- **Twitter/X** - Follow [@containr](https://twitter.com/containr) for updates (example)

## Bug Reports

### Reporting a Bug

Found a bug? Here's how to report it:

1. **Search First**
   - Check [open issues](https://github.com/therealutkarshpriyadarshi/containr/issues)
   - Check [closed issues](https://github.com/therealutkarshpriyadarshi/containr/issues?q=is%3Aissue+is%3Aclosed)

2. **Create Issue**
   - Use [Bug Report template](https://github.com/therealutkarshpriyadarshi/containr/issues/new?template=bug_report.md)
   - Fill in all sections
   - Provide complete information

3. **What to Include**
   ```markdown
   ## Description
   Clear description of the bug

   ## Steps to Reproduce
   1. Run command: containr run alpine /bin/sh
   2. Execute: xyz
   3. See error

   ## Expected Behavior
   Container should start

   ## Actual Behavior
   Container fails with error

   ## Environment
   - OS: Ubuntu 22.04
   - Containr Version: 1.0.0
   - Go Version: 1.21.0
   - Kernel: 5.15.0

   ## Logs
   ```
   paste logs here with --debug flag
   ```

   ## Additional Context
   Screenshots, configs, etc.
   ```

### Debug Mode

Always include debug logs:

```bash
# Run with debug mode
sudo containr run --debug alpine /bin/sh

# Save logs to file
sudo containr run --debug alpine /bin/sh 2>&1 | tee debug.log
```

## Feature Requests

### Requesting a Feature

Want a new feature? Here's how:

1. **Check Roadmap**
   - Review [ROADMAP.md](ROADMAP.md)
   - Check if already planned

2. **Search Existing**
   - Look for similar requests
   - Add to existing discussion

3. **Create Request**
   - Use [Feature Request template](https://github.com/therealutkarshpriyadarshi/containr/issues/new?template=feature_request.md)
   - Explain use case
   - Describe proposed solution

4. **What to Include**
   ```markdown
   ## Feature Description
   Clear description of proposed feature

   ## Use Case
   Why is this needed? What problem does it solve?

   ## Proposed Implementation
   How should this work?

   ## Alternatives Considered
   Other ways to solve this

   ## Additional Context
   Examples, references, etc.
   ```

### Feature Prioritization

Features are prioritized based on:
- Educational value (40%)
- User impact (30%)
- Implementation effort (20%)
- Strategic alignment (10%)

See [ROADMAP.md](ROADMAP.md) for details.

## Security Issues

### Reporting Security Vulnerabilities

**DO NOT report security issues publicly!**

Instead:
1. Read [SECURITY.md](SECURITY.md)
2. Report via [GitHub Security Advisories](https://github.com/therealutkarshpriyadarshi/containr/security)
3. Or email: security@containr-project.org

We respond to security reports within 48 hours.

### Security Questions

For non-vulnerability security questions:
- Read [docs/SECURITY.md](docs/SECURITY.md)
- Ask in [Discussions](https://github.com/therealutkarshpriyadarshi/containr/discussions)

## Professional Support

### Educational Use

Containr is designed for education. We support:

- **Students** - Learning container technology
- **Educators** - Teaching containerization
- **Developers** - Understanding containers
- **Researchers** - Container research

### Commercial Support

Currently, Containr does not offer commercial support.

**For production use, we recommend:**
- [Docker](https://www.docker.com/)
- [Podman](https://podman.io/)
- [containerd](https://containerd.io/)
- [CRI-O](https://cri-o.io/)

### Training & Workshops

Interested in Containr training or workshops?
- Contact: [email@example.com]
- Check [Discussions](https://github.com/therealutkarshpriyadarshi/containr/discussions) for announcements

## Contributing

### Want to Help?

Contributions are welcome! Here's how:

1. **Read Contributing Guide**
   - [CONTRIBUTING.md](CONTRIBUTING.md)

2. **Find Issues**
   - Look for [`good first issue`](https://github.com/therealutkarshpriyadarshi/containr/labels/good%20first%20issue) label
   - Look for [`help wanted`](https://github.com/therealutkarshpriyadarshi/containr/labels/help%20wanted) label

3. **Join Discussions**
   - Participate in [Discussions](https://github.com/therealutkarshpriyadarshi/containr/discussions)
   - Help answer questions
   - Review pull requests

4. **Improve Documentation**
   - Fix typos
   - Add examples
   - Write tutorials

### Ways to Contribute

- 💻 **Code** - Fix bugs, add features
- 📝 **Documentation** - Improve docs
- 🧪 **Testing** - Write tests, report bugs
- 💬 **Community** - Answer questions, help others
- 🎨 **Design** - UI/UX improvements
- 🌍 **Translation** - Translate docs (future)

## Response Times

### Expected Response Times

| Type | Response Time | Resolution Time |
|------|--------------|----------------|
| Security (Critical) | 24-48 hours | 1 week |
| Security (High) | 48 hours | 2 weeks |
| Bugs (Critical) | 1-2 days | 1 week |
| Bugs (Normal) | 1 week | 2-4 weeks |
| Questions | 1-3 days | Varies |
| Feature Requests | 1 week | Planned in roadmap |
| Documentation | 1 week | 2-4 weeks |

**Note:** These are targets, not guarantees. Response times depend on maintainer availability.

## Troubleshooting

### Common Issues

#### "Operation not permitted"
```bash
# Solution: Run with sudo
sudo containr run alpine /bin/sh
```

#### "Cannot create namespace"
```bash
# Check kernel version
uname -r  # Should be 3.8+

# Enable debug mode
sudo containr run --debug alpine /bin/sh
```

#### "Cgroup not supported"
```bash
# Check cgroup version
mount | grep cgroup

# Try cgroup v1 or v2 explicitly
```

### Getting More Help

If troubleshooting doesn't help:

1. **Enable debug mode:**
   ```bash
   sudo containr run --debug --log-level debug alpine /bin/sh
   ```

2. **Collect information:**
   - OS and version
   - Kernel version
   - Containr version
   - Full error logs

3. **Ask for help:**
   - Create [GitHub Discussion](https://github.com/therealutkarshpriyadarshi/containr/discussions)
   - Include all information above
   - Be patient and polite

## Code of Conduct

All community interactions are governed by our [Code of Conduct](CODE_OF_CONDUCT.md).

We are committed to providing a welcoming and inclusive environment for everyone.

**Expected behavior:**
- Be respectful and considerate
- Welcome newcomers
- Focus on constructive feedback
- Accept gracefully when others disagree

**Unacceptable behavior:**
- Harassment or discrimination
- Personal attacks
- Publishing others' private information
- Disruptive behavior

**To report violations:**
- Email: conduct@containr-project.org
- See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for details

## Additional Resources

### Learning Resources

- **Container Basics:**
  - [Linux Namespaces](https://man7.org/linux/man-pages/man7/namespaces.7.html)
  - [Cgroups Documentation](https://www.kernel.org/doc/Documentation/cgroup-v2.txt)

- **Container Technology:**
  - [OCI Runtime Spec](https://github.com/opencontainers/runtime-spec)
  - [Docker Documentation](https://docs.docker.com/)

- **Security:**
  - [Linux Capabilities](https://man7.org/linux/man-pages/man7/capabilities.7.html)
  - [Seccomp](https://www.kernel.org/doc/html/latest/userspace-api/seccomp_filter.html)

### External Communities

- **Container Community:**
  - [OCI Community](https://opencontainers.org/)
  - [CNCF Slack](https://slack.cncf.io/)

- **Go Community:**
  - [Go Forum](https://forum.golangbridge.org/)
  - [Gophers Slack](https://invite.slack.golangbridge.org/)

## Contact

### General Inquiries

- **GitHub Discussions:** [Ask here](https://github.com/therealutkarshpriyadarshi/containr/discussions)
- **Email:** containr@example.com (replace with actual)

### Maintainers

See [MAINTAINERS.md](MAINTAINERS.md) for maintainer contacts.

### Social Media

- **Twitter/X:** [@containr](https://twitter.com/containr) (example)
- **Blog:** [blog.containr.io](https://blog.containr.io) (planned)

---

**Thank you for using Containr!**

We're here to help you learn container technology. Don't hesitate to ask questions!

**Last Updated:** November 17, 2025
