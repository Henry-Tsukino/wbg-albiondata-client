# Contributing

We welcome contributions from the community. This document provides guidelines for contributing to the Albion Data Client project.

## Getting Started

1. Fork the repository
2. Clone your fork locally
3. Create a new branch for your feature or bugfix
4. Make your changes
5. Submit a pull request

## Code Standards

- Follow Go conventions and best practices
- Run `scripts/validate-fmt.sh` to check code formatting
- Run `scripts/fmt.sh` to automatically format code
- Ensure your code is well-commented and documented

## Building the Project

### Prerequisites

- Go 1.16 or later
- Platform-specific tools (see scripts for details)

### Build Scripts

Build scripts are provided for each supported platform:

- **Windows**: `scripts/build-windows.sh`
- **macOS**: `scripts/build-darwin.sh`
- **Linux**: `scripts/build-linux.sh`

For custom Windows builds: `.\build-windows-custom.ps1`

## Testing

Before submitting a pull request, ensure your changes:

- Do not break existing functionality
- Follow the existing code style
- Include appropriate error handling
- Are thoroughly tested

## Pull Request Process

1. Update the README.md with any new features or changes
2. Ensure your code follows project conventions
3. Reference any related issues in your PR description
4. One or more maintainers will review your contribution

## Reporting Issues

When reporting issues, please include:

- Detailed description of the problem
- Steps to reproduce the issue
- Your environment (Windows/macOS/Linux version, Go version, etc.)
- Any relevant logs or error messages

## Questions

For questions about the project, visit:
- Official Project: https://www.albion-online-data.com/
- Original Repository: https://github.com/ao-data/albiondata-client

## License

By contributing to this project, you agree that your contributions will be licensed under the MIT License.
