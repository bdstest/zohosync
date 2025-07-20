# CI/CD Workflows

## Status: Disabled

The CI/CD workflows have been temporarily disabled for this demonstration repository.

## Reason for Disabling

This is a portfolio/demonstration project showcasing technical capabilities. Continuous integration workflows are not necessary for:

- Portfolio showcase repositories
- Demonstration of architectural patterns
- Technical capability examples

## Re-enabling CI/CD

To re-enable automated testing and security scanning:

1. Rename `ci.yml.disabled` back to `ci.yml`
2. Ensure all dependencies are properly configured
3. Verify Go module dependencies are up to date
4. Test locally before enabling automated workflows

## Local Development

For local testing and validation:

```bash
# Run tests locally
make test

# Run security scans
make security-quick

# Build applications
make build
```

This approach maintains code quality while avoiding unnecessary CI/CD overhead for demonstration purposes.