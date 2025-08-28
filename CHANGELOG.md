# Changelog

All notable changes to Osiris Lite will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.1] - 2025-08-04

### Added
- Initial release of Osiris Lite CLI tool
- SSH-based remote Docker orchestration
- Flexible configuration management (YAML/JSON/TOML + env vars)
- File synchronization via rsync
- Real-time job monitoring and status checking
- Support for ProxyJump and complex SSH configurations
- Command-line interface with subcommands (run, status, kill, pull, logs)

### Features
- Remote fuzzing workflow management
- Docker container lifecycle management
- Efficient file transfer with rsync
- Multi-source configuration (config files, environment variables, CLI flags)
- SSH key and password authentication support

### Technical
- Built with Go 1.21+
- Uses Cobra for CLI framework
- Uses Viper for configuration management
- Native SSH client implementation
- Battle-tested underlying tools (SSH, rsync, Docker) 