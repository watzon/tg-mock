# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.1] - 2025-12-26

### Added

- Docker support with multi-arch builds (linux/amd64, linux/arm64)
- Multi-stage Dockerfile with scratch base (~6.7MB image)
- `docker-build`, `docker-run`, `docker-push` Makefile targets

### Fixed

- CI: Docker images only tagged with version on releases (manual runs get SHA only)

### Documentation

- Add Docker installation instructions to README

## [0.2.0] - 2025-12-25

### Added

- Smart faker system for realistic response generation with field name heuristics
- Type-specific generators for 40+ Telegram API types
- Request parameter reflection back into responses (e.g., chat_id → chat.id)
- Configurable seed for deterministic/reproducible tests (`--faker-seed` CLI flag, `faker_seed` config)
- Scenario response_data overrides for success response customization

### Changed

- Replace limited responder (5 types) with comprehensive faker system

## [0.1.0] - 2025-12-25

### Added

- Initial release of tg-mock, a mock Telegram Bot API server
- Token validation with format `<bot_id>:<secret>`
- Token registry with status tracking (active/banned/deactivated)
- Scenario engine for configurable error responses
- Update queue for getUpdates polling simulation
- Request validation against official Telegram Bot API spec
- Realistic response generation with proper fixtures
- Code generator for types, methods, and fixtures from API spec
- YAML configuration support
- CLI with `--port`, `--verbose`, `--config`, and `--storage-dir` flags
- Control API at `/__control/*` for test configuration
- CI/CD pipeline with GitHub Actions
- Multi-platform releases (Linux, macOS, Windows)
