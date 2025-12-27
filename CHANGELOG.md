# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.2] - 2025-12-26

### Added

- 32 new built-in error scenarios for comprehensive Telegram API error testing
  - Chat errors: `chat_admin_required`, `chat_not_modified`, `chat_restricted`, `chat_write_forbidden`, `channel_private`, `group_upgraded`, `supergroup_channel_only`, `not_in_chat`, `topic_not_modified`
  - User errors: `user_id_invalid`, `user_is_admin`, `participant_id_invalid`, `cant_remove_owner`
  - Message errors: `message_to_delete_not_found`, `message_id_invalid`, `message_thread_not_found`
  - Permission errors: `no_rights_to_send`, `not_enough_rights`, `not_enough_rights_pin`, `not_enough_rights_restrict`, `not_enough_rights_send_text`
  - Admin/other errors: `admin_rank_emoji_not_allowed`, `inline_button_url_invalid`, `hide_requester_missing`
  - Forbidden variants: `bot_kicked_channel`, `bot_kicked_group`, `bot_kicked_supergroup`

### Changed

- README: Reorganized built-in scenarios into collapsible sections by category

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
