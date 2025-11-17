# Tutorial: Making Your First Contribution to Containr

**Difficulty:** Beginner
**Time:** 30-45 minutes
**Prerequisites:** Basic Git and GitHub knowledge

## Overview

This tutorial will guide you through making your first contribution to the Containr project. You'll learn about the contribution process, community guidelines, and best practices.

## Learning Objectives

By the end of this tutorial, you will:
- Understand Containr's contribution workflow
- Know how to set up your development environment
- Make a meaningful contribution
- Participate in code review
- Become part of the Containr community

## Prerequisites

Before starting, ensure you have:
- A GitHub account
- Git installed on your machine
- Go 1.16+ installed
- Basic understanding of containers
- Linux environment (or WSL on Windows)

## Step 1: Understanding the Project

### Read the Documentation

Start by familiarizing yourself with the project:

1. **Read the README:**
   ```bash
   # Visit https://github.com/therealutkarshpriyadarshi/containr
   # Read README.md for project overview
   ```

2. **Review Contribution Guidelines:**
   ```bash
   # Read CONTRIBUTING.md
   # Understand code style and testing requirements
   ```

3. **Check the Code of Conduct:**
   ```bash
   # Read CODE_OF_CONDUCT.md
   # Understand community expectations
   ```

### Explore the Codebase

```bash
# Clone the repository
git clone https://github.com/therealutkarshpriyadarshi/containr.git
cd containr

# Explore the structure
tree -L 2 .

# Expected output:
# containr/
# ├── cmd/           # Command-line interface
# ├── pkg/           # Core packages
# ├── docs/          # Documentation
# ├── examples/      # Example code
# └── test/          # Integration tests
```

## Step 2: Setting Up Your Environment

### Fork and Clone

1. **Fork the repository:**
   - Go to https://github.com/therealutkarshpriyadarshi/containr
   - Click "Fork" button
   - This creates your copy of the repository

2. **Clone your fork:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/containr.git
   cd containr
   ```

3. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/therealutkarshpriyadarshi/containr.git

   # Verify remotes
   git remote -v
   # origin    https://github.com/YOUR_USERNAME/containr.git (fetch)
   # origin    https://github.com/YOUR_USERNAME/containr.git (push)
   # upstream  https://github.com/therealutkarshpriyadarshi/containr.git (fetch)
   # upstream  https://github.com/therealutkarshpriyadarshi/containr.git (push)
   ```

### Install Dependencies

```bash
# Install Go dependencies
make deps

# Install development tools
make install-tools

# This installs:
# - golangci-lint (linter)
# - staticcheck (static analysis)
# - gosec (security scanner)
```

### Build and Test

```bash
# Build the project
make build

# Run tests
make test-unit

# Run linters
make lint

# Verify everything works
./bin/containr version
```

## Step 3: Finding Something to Work On

### Browse Issues

Look for issues labeled `good first issue`:

1. **Visit the Issues page:**
   https://github.com/therealutkarshpriyadarshi/containr/issues

2. **Filter by labels:**
   - Click "Labels" dropdown
   - Select "good first issue"
   - Select "help wanted"

3. **Read issue descriptions:**
   - Understand the problem
   - Check if it's still available
   - Ask questions if unclear

### Choose Your Contribution

Good first contributions:
- 📝 Fix typos in documentation
- 🐛 Fix small bugs
- ✅ Add missing tests
- 📖 Improve code comments
- 🎨 Code cleanup/refactoring

Example: Let's improve documentation!

## Step 4: Making Your Changes

### Create a Branch

```bash
# Update main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b docs/improve-readme

# Branch naming conventions:
# - feat/feature-name (new features)
# - fix/bug-description (bug fixes)
# - docs/doc-update (documentation)
# - test/test-addition (tests)
```

### Make Changes

For this tutorial, let's improve the README:

```bash
# Open README.md in your editor
nano README.md

# Make your improvements, for example:
# - Fix typos
# - Clarify instructions
# - Add missing information
# - Improve formatting
```

### Test Your Changes

```bash
# If code changes, run tests
make test-unit

# Run linters
make lint

# Format code
make fmt

# Build to ensure no errors
make build
```

## Step 5: Committing Your Changes

### Follow Commit Conventions

Containr uses [Conventional Commits](https://www.conventionalcommits.org/):

```bash
# Stage your changes
git add README.md

# Commit with conventional format
git commit -m "docs: improve quick start instructions

- Add prerequisite section
- Clarify installation steps
- Fix formatting issues
- Add troubleshooting tips

Closes #123"
```

**Commit message format:**
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance

### Push to Your Fork

```bash
# Push branch to your fork
git push origin docs/improve-readme
```

## Step 6: Creating a Pull Request

### Open Pull Request

1. **Go to your fork on GitHub:**
   https://github.com/YOUR_USERNAME/containr

2. **Click "Compare & pull request"**

3. **Fill out the PR template:**

```markdown
## Description
Improved quick start instructions in README.md

## Related Issues
Closes #123

## Type of Change
- [x] Documentation update

## Changes Made
### Changed
- Improved quick start section
- Added prerequisites
- Fixed formatting issues

## Checklist
- [x] I have read CONTRIBUTING.md
- [x] I have performed a self-review
- [x] Documentation is clear and accurate
- [x] No new warnings generated
```

4. **Submit the pull request**

### Participate in Review

**Responding to feedback:**

1. **Read all comments carefully**
2. **Ask questions if unclear**
3. **Make requested changes:**
   ```bash
   # Make changes in your branch
   nano README.md

   # Commit changes
   git add README.md
   git commit -m "docs: address review feedback"

   # Push to update PR
   git push origin docs/improve-readme
   ```
4. **Thank reviewers for their time**

## Step 7: After Your PR is Merged

### Celebrate! 🎉

Your contribution is now part of Containr!

### Update Your Fork

```bash
# Switch to main branch
git checkout main

# Pull latest changes
git pull upstream main

# Update your fork
git push origin main

# Delete your feature branch
git branch -d docs/improve-readme
git push origin --delete docs/improve-readme
```

### Next Steps

**Continue Contributing:**
- Look for more issues
- Help review other PRs
- Answer questions in Discussions
- Share your experience

**Level Up:**
- Tackle larger issues
- Add new features
- Improve test coverage
- Write tutorials

## Common Pitfalls and Solutions

### Issue: Fork is Out of Date

```bash
# Sync your fork
git checkout main
git pull upstream main
git push origin main
```

### Issue: Merge Conflicts

```bash
# Update your branch
git checkout your-feature-branch
git rebase main

# Resolve conflicts in editor
# Then:
git add .
git rebase --continue
git push --force-with-lease origin your-feature-branch
```

### Issue: Failed CI Checks

```bash
# Run checks locally
make pre-commit

# This runs:
# - make fmt (formatting)
# - make lint (linting)
# - make test (tests)
```

### Issue: Not Sure What to Work On

**Ask for help!**
- Comment on issues
- Ask in GitHub Discussions
- Reach out to maintainers
- Start with documentation

## Best Practices

### Code Quality

1. **Write tests:**
   ```go
   func TestMyFunction(t *testing.T) {
       result := MyFunction("input")
       expected := "output"
       if result != expected {
           t.Errorf("expected %s, got %s", expected, result)
       }
   }
   ```

2. **Add documentation:**
   ```go
   // MyFunction processes input and returns output.
   // It validates the input and handles errors gracefully.
   func MyFunction(input string) string {
       // Implementation
   }
   ```

3. **Follow code style:**
   ```bash
   # Use gofmt
   make fmt

   # Run linters
   make lint
   ```

### Communication

1. **Be respectful and professional**
2. **Ask questions when unsure**
3. **Provide context in comments**
4. **Thank maintainers for their time**
5. **Be patient with review process**

### Time Management

- **Start small:** Don't take on too much at once
- **Set realistic goals:** 1-2 hours per week is great
- **Communicate delays:** Let maintainers know if you need more time
- **It's okay to step back:** Life happens, no pressure

## Resources

### Documentation

- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Full contribution guide
- [CODE_OF_CONDUCT.md](../../CODE_OF_CONDUCT.md) - Community guidelines
- [GOVERNANCE.md](../../GOVERNANCE.md) - Project governance

### Community

- [GitHub Discussions](https://github.com/therealutkarshpriyadarshi/containr/discussions) - Ask questions
- [Issue Tracker](https://github.com/therealutkarshpriyadarshi/containr/issues) - Find work
- [Pull Requests](https://github.com/therealutkarshpriyadarshi/containr/pulls) - Review code

### Learning

- [Git Book](https://git-scm.com/book/) - Learn Git
- [Go Documentation](https://go.dev/doc/) - Learn Go
- [Linux Namespaces](https://man7.org/linux/man-pages/man7/namespaces.7.html) - Learn containers

## Conclusion

Congratulations! You now know how to contribute to Containr. Remember:

- **Start small** and build confidence
- **Ask questions** when you need help
- **Be patient** with yourself and others
- **Have fun** and learn!

**Welcome to the Containr community! 🎉**

## Next Tutorials

- [02-namespace-deep-dive.md](02-namespace-deep-dive.md) - Understanding Linux namespaces
- [03-building-a-feature.md](03-building-a-feature.md) - Building your first feature
- [04-testing-guide.md](04-testing-guide.md) - Writing comprehensive tests

---

**Questions?** Ask in [GitHub Discussions](https://github.com/therealutkarshpriyadarshi/discussions)!
