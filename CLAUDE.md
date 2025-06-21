# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

AWSID is a Go CLI tool that retrieves AWS account IDs from alias names using AWS Organizations API. It's a single-file Go application (`main.go`) that automatically fetches account information and provides multiple output formats (standard, JSON, table, CSV).

## Architecture

- **Single binary**: All functionality is contained in `main.go`
- **Data storage**: Account information is cached in `~/.aws/account_info` as CSV
- **AWS Integration**: Uses AWS SDK v2 with Organizations service (requires us-east-1 region)
- **CLI Framework**: Built with Cobra for command-line interface
- **Search Logic**: Supports both exact and partial matching with different output behaviors

## Key Components

- `AccountInfo` struct: Core data model for alias name and account ID pairs
- `readAccountInfo()`: CSV file parser with header detection
- `updateAccountInfoFromAWS()`: AWS Organizations API integration
- Output formatters: `outputJSON()`, `outputTable()`, `outputCSV()`
- Search logic: Exact match returns account ID only, partial matches show "alias: id" format

## Build and Development Commands

```bash
# Build for current platform
go build -o awsid

# Run without building
go run main.go [args]

# Test (no test files currently exist)
go test

# Format code
go fmt

# Vet code for issues
go vet

# Get dependencies
go mod tidy

# Cross-compile for multiple platforms (like existing dist/ binaries)
GOOS=darwin GOARCH=amd64 go build -o awsid_darwin_amd64
GOOS=darwin GOARCH=arm64 go build -o awsid_darwin_arm64
GOOS=linux GOARCH=amd64 go build -o awsid_linux_amd64
GOOS=windows GOARCH=amd64 go build -o awsid_windows_amd64.exe
```

## Dependencies

- `github.com/spf13/cobra`: CLI framework
- `github.com/olekukonko/tablewriter`: Table output formatting
- `github.com/aws/aws-sdk-go-v2/*`: AWS SDK for Organizations API

## Important Notes

- AWS Organizations API calls are hardcoded to use us-east-1 region
- Account info file location is fixed at `~/.aws/account_info`
- No test files exist currently - consider adding when implementing new features
- Version is hardcoded in main.go as a const (currently "0.4.0")

## AWS Organizations Access

**IMPORTANT**: When accessing AWS Organizations API, always switch to the org profile first:

```bash
# Switch to Organizations profile and run awsid
chprof org && ./awsid

# Test Organizations API access
chprof org && aws organizations list-accounts --region us-east-1
```

This is required because:
- Claude Code runs each bash command in separate shell sessions
- Environment variables (AWS_PROFILE) don't persist between commands
- AWS Organizations API requires proper IAM permissions in the management account