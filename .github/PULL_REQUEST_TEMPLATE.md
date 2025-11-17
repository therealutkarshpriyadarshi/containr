# Pull Request

## Description

<!-- Provide a clear and concise description of your changes -->

## Related Issues

<!-- Link to related issues using keywords like "Closes #123" or "Fixes #456" -->

Closes #
Related to #

## Type of Change

<!-- Check all that apply -->

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📝 Documentation update
- [ ] 🎨 Code style/formatting
- [ ] ♻️ Refactoring (no functional changes)
- [ ] ⚡ Performance improvement
- [ ] ✅ Test addition/update
- [ ] 🔧 Build/CI configuration
- [ ] 🔒 Security fix

## Changes Made

<!-- Describe the changes in detail. Use bullet points for clarity -->

### Added
-

### Changed
-

### Removed
-

### Fixed
-

## Testing

<!-- Describe the tests you ran and how to reproduce them -->

### Test Configuration

- **OS:** [e.g., Ubuntu 22.04]
- **Kernel:** [e.g., 5.15.0]
- **Go Version:** [e.g., 1.21.0]

### Tests Performed

- [ ] Unit tests pass (`make test-unit`)
- [ ] Integration tests pass (`sudo make test-integration`)
- [ ] Linters pass (`make lint`)
- [ ] Code formatted (`make fmt`)
- [ ] Benchmarks run (if performance-related)
- [ ] Manual testing performed

### Manual Testing Steps

1. Step 1
2. Step 2
3. Expected result

### Test Results

```bash
# Paste relevant test output here
```

## Performance Impact

<!-- For performance-related changes, include benchmark results -->

### Before
```
BenchmarkOperation-8    1000    1234567 ns/op
```

### After
```
BenchmarkOperation-8    2000    654321 ns/op
```

## Breaking Changes

<!-- If this PR includes breaking changes, describe them here -->

- [ ] This PR includes breaking changes
- [ ] Migration guide included/updated
- [ ] CHANGELOG.md updated

### Migration Guide

<!-- If breaking changes, explain how users should migrate -->

## Documentation

<!-- Check all that apply -->

- [ ] Updated README.md
- [ ] Updated relevant documentation in `docs/`
- [ ] Added/updated code comments
- [ ] Updated CHANGELOG.md
- [ ] Updated API documentation (GoDoc)
- [ ] Added examples (if new feature)

## Code Quality

<!-- Ensure your code meets quality standards -->

- [ ] Code follows project style guidelines
- [ ] Self-review performed
- [ ] Comments added for complex code
- [ ] No new compiler warnings
- [ ] No security vulnerabilities introduced
- [ ] Error handling is appropriate
- [ ] Resource cleanup is handled properly

## Checklist

<!-- Complete this checklist before submitting -->

### Before Submission

- [ ] I have read [CONTRIBUTING.md](../CONTRIBUTING.md)
- [ ] My code follows the project's coding standards
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published

### Git Hygiene

- [ ] Commits are logical and atomic
- [ ] Commit messages follow [conventional commits](https://www.conventionalcommits.org/)
- [ ] No merge commits (rebased on latest main)
- [ ] No unnecessary whitespace changes
- [ ] No commented-out code

### Security

- [ ] No hardcoded credentials or secrets
- [ ] Input validation is comprehensive
- [ ] Error messages don't leak sensitive information
- [ ] Security implications considered and documented

## Screenshots/Demos

<!-- If applicable, add screenshots or demo output -->

### Before
```
# Old behavior
```

### After
```
# New behavior
```

## Additional Context

<!-- Add any other context about the pull request here -->

### Dependencies

<!-- List any new dependencies or dependency updates -->

- Added: `package@version` - Reason
- Updated: `package@version` → `package@newversion` - Reason

### Future Work

<!-- Optional: Note any follow-up work that should be done -->

- [ ] Future improvement 1
- [ ] Future improvement 2

## Reviewer Notes

<!-- Any specific notes for reviewers? Areas you want feedback on? -->

### Areas for Review Focus

-
-
-

### Questions for Reviewers

-
-

---

## For Maintainers

<!-- Maintainers: Complete before merging -->

### Review Checklist

- [ ] Code quality is satisfactory
- [ ] Tests are comprehensive
- [ ] Documentation is complete
- [ ] No security concerns
- [ ] Breaking changes are justified and documented
- [ ] CHANGELOG.md updated (if needed)
- [ ] Ready to merge

### Merge Strategy

- [ ] Squash and merge (default)
- [ ] Rebase and merge
- [ ] Create merge commit

---

**Thank you for contributing to Containr! 🎉**
