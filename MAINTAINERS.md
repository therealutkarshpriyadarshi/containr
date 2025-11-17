# Containr Maintainers

This document lists the current maintainers of the Containr project, their roles, and areas of responsibility.

## Current Maintainers

### BDFL (Benevolent Dictator For Life)

#### Utkarsh Priyadarshi (@therealutkarshpriyadarshi)
- **Role:** Project Lead & BDFL
- **Responsibilities:**
  - Project vision and strategic direction
  - Final decision authority on contentious issues
  - Major architectural decisions
  - Maintainer appointments
- **Focus Areas:** Overall project, architecture, roadmap
- **Contact:** GitHub: [@therealutkarshpriyadarshi](https://github.com/therealutkarshpriyadarshi)

### Core Maintainers

*Core maintainer positions are open for qualified contributors. See "Becoming a Maintainer" section below.*

### Area Maintainers

Area maintainers focus on specific subsystems. Currently seeking maintainers for:

#### Container & Runtime (`pkg/container`, `pkg/runtime`)
- **Seeking Maintainer**
- **Responsibilities:** Container lifecycle, OCI compliance, state management
- **Skills Needed:** Go, Linux containers, OCI spec

#### Networking (`pkg/network`)
- **Seeking Maintainer**
- **Responsibilities:** Network isolation, bridge networking, CNI support
- **Skills Needed:** Go, Linux networking, iptables, network namespaces

#### Storage (`pkg/rootfs`, `pkg/volume`)
- **Seeking Maintainer**
- **Responsibilities:** Filesystem isolation, volumes, storage drivers
- **Skills Needed:** Go, overlay filesystems, mount operations

#### Security (`pkg/capabilities`, `pkg/seccomp`, `pkg/security`)
- **Seeking Maintainer**
- **Responsibilities:** Security features, capabilities, seccomp, LSM
- **Skills Needed:** Go, Linux security, capabilities, seccomp

#### Image Management (`pkg/image`, `pkg/registry`)
- **Seeking Maintainer**
- **Responsibilities:** Image handling, registry client, OCI images
- **Skills Needed:** Go, OCI image format, HTTP/REST APIs

#### Resource Management (`pkg/cgroup`, `pkg/metrics`)
- **Seeking Maintainer**
- **Responsibilities:** Cgroups, resource limits, metrics collection
- **Skills Needed:** Go, Linux cgroups v1/v2, monitoring

#### Documentation
- **Seeking Maintainer**
- **Responsibilities:** Documentation quality, tutorials, examples
- **Skills Needed:** Technical writing, container knowledge

#### Testing & CI/CD (`.github/workflows`, `test/`)
- **Seeking Maintainer**
- **Responsibilities:** Test coverage, CI/CD pipelines, quality assurance
- **Skills Needed:** Go testing, GitHub Actions, CI/CD

## Emeritus Maintainers

Former maintainers who have stepped down but may still advise:

*None yet - project is new!*

## Maintainer Responsibilities

All maintainers are expected to:

### Code Review
- Review pull requests in their areas
- Provide constructive feedback
- Ensure code quality and test coverage
- Merge PRs when ready

### Issue Triage
- Monitor issues in their areas
- Label and prioritize issues
- Respond to questions
- Close resolved issues

### Community Engagement
- Help new contributors
- Participate in discussions
- Be welcoming and inclusive
- Follow code of conduct

### Technical Leadership
- Guide technical decisions
- Maintain architectural consistency
- Document design decisions
- Share knowledge

### Availability
- Respond within 1 week for non-urgent matters
- Participate in monthly maintainer meetings
- Notify others when unavailable for extended periods

## Becoming a Maintainer

### Criteria

To become a maintainer, you should demonstrate:

1. **Sustained Contribution** (6+ months)
   - Regular, high-quality contributions
   - Multiple merged pull requests
   - Consistent engagement with project

2. **Technical Expertise**
   - Deep understanding of codebase
   - Strong Go programming skills
   - Knowledge of container technologies
   - Good architectural sense

3. **Community Involvement**
   - Help other contributors
   - Participate in discussions
   - Review pull requests
   - Answer questions

4. **Alignment with Project Values**
   - Educational focus
   - Code quality commitment
   - Inclusive behavior
   - Professional conduct

5. **Time Commitment**
   - Able to dedicate regular time
   - Responsive to issues/PRs
   - Participate in meetings
   - Long-term interest

### Process

1. **Nomination**
   - Current maintainer nominates candidate
   - OR candidate expresses interest via email/discussion
   - Nomination includes justification

2. **Discussion**
   - Maintainers discuss privately
   - Review contribution history
   - Assess readiness
   - Consensus or BDFL decision

3. **Invitation**
   - BDFL extends invitation
   - Candidate accepts
   - Onboarding begins

4. **Onboarding**
   - Add to MAINTAINERS.md
   - Grant repository access
   - Pair with mentor
   - Gradual responsibility increase

5. **Announcement**
   - Public announcement in discussions
   - Welcome to the team!

### How to Express Interest

If you're interested in becoming a maintainer:

1. **Build Track Record**
   - Contribute regularly for 6+ months
   - Focus on quality over quantity
   - Show technical expertise

2. **Get Involved**
   - Review PRs
   - Help in discussions
   - Answer questions
   - Write documentation

3. **Express Interest**
   - Email BDFL or current maintainer
   - Mention area of interest
   - Describe your contributions
   - Explain your motivation

4. **Be Patient**
   - Maintainer positions open based on need
   - Continue contributing regardless
   - Your contributions are valued!

## Stepping Down

Maintainers can step down at any time:

### Process

1. **Notify BDFL**
   - Send private email/message
   - Explain reason (optional)
   - Propose transition plan

2. **Transition**
   - Transfer ongoing work
   - Document decisions in progress
   - Help find replacement (if willing)

3. **Update Records**
   - Move to "Emeritus" section
   - Update MAINTAINERS.md
   - Adjust repository permissions

4. **Announcement**
   - Thank you message
   - Recognize contributions
   - Note emeritus status

### Emeritus Status

Former maintainers become "Emeritus Maintainers":
- Recognized for past contributions
- Can participate in major decisions
- May advise on technical matters
- Can return to active status
- No ongoing obligations

## Maintainer Expectations

### Time Commitment

Approximate time expectations:

- **Core Maintainer:** 5-10 hours/week
- **Area Maintainer:** 3-5 hours/week
- **Release Manager:** 10-15 hours during release weeks

### Communication

Maintainers should:
- Check GitHub notifications daily
- Respond to @mentions within 48 hours
- Attend monthly meetings when possible
- Notify team if unavailable >1 week

### Decision Making

Maintainers participate in:
- Technical decisions in their areas
- PR approvals and merges
- Issue prioritization
- Release planning
- Policy discussions

See [GOVERNANCE.md](GOVERNANCE.md) for detailed decision-making process.

## Maintainer Resources

### Internal Documentation

- [GOVERNANCE.md](GOVERNANCE.md) - Governance model
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guide
- [ROADMAP.md](ROADMAP.md) - Project roadmap
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - Architecture overview

### Tools & Access

Maintainers have access to:
- Repository write permissions
- GitHub Actions secrets (as needed)
- Release management tools
- Private security repository (for security issues)

### Meetings

- **Monthly Maintainer Meetings**
  - First Monday of each month
  - 1 hour duration
  - Virtual (Google Meet/Zoom)
  - Notes published in GitHub Discussions

- **Quarterly Community Calls**
  - Open to all contributors
  - Project updates
  - Q&A session
  - Recorded and published

### Communication Channels

- **GitHub Issues:** Public discussions
- **GitHub Discussions:** Q&A, announcements
- **Email:** Private maintainer discussions
- **Slack/Discord:** (Future) Real-time chat

## Contact

### Reach Maintainers

- **General Questions:** GitHub Discussions
- **Private Matters:** Email BDFL
- **Security Issues:** security@containr-project.org (see [SECURITY.md](SECURITY.md))

### Propose Changes

To suggest changes to maintainer structure:
1. Create GitHub Discussion
2. Explain proposal
3. Allow 2 weeks for feedback
4. BDFL makes final decision

## Recognition

Maintainers are recognized through:

- Listed in MAINTAINERS.md (this file)
- GitHub repository permissions
- Mentioned in release notes
- Special thanks in blog posts
- Conference opportunities (when available)
- Swag and recognition (when available)

## Acknowledgments

Thank you to all maintainers past, present, and future for your dedication to making Containr a premier educational container runtime!

---

**Last Updated:** November 17, 2025
**For Changes:** Contact BDFL or create GitHub Discussion
**See Also:** [GOVERNANCE.md](GOVERNANCE.md), [CONTRIBUTING.md](CONTRIBUTING.md)
