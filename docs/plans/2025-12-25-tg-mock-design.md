# tg-mock Design

A mock Telegram Bot API server for testing bots and bot libraries.

## Overview

**tg-mock** validates requests against the Telegram Bot API spec, returns realistic responses, and supports configurable scenarios for error/edge case testing.

### Design Principles

1. **Stateless by default** - No persistence between requests unless scenarios require it
2. **Spec-driven** - All validation and types generated from official spec
3. **Test-friendly** - Easy to set up scenarios, inject updates, and verify behavior
4. **Minimal config** - Works out of the box, customize when needed

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      tg-mock                            │
├─────────────────────────────────────────────────────────┤
│  HTTP Server (single port)                              │
│  ├── /bot<token>/*  → Bot API Handler                  │
│  └── /__control/*   → Control API Handler              │
├─────────────────────────────────────────────────────────┤
│  Core Components:                                       │
│  ├── Request Validator (generated from spec)           │
│  ├── Response Generator (fixtures + reflection)        │
│  ├── Scenario Engine (errors, rate limits, custom)     │
│  ├── Update Queue (for getUpdates polling)             │
│  ├── Webhook Pusher (sends updates to bot URL)         │
│  └── File Store (in-memory, optional disk)             │
├─────────────────────────────────────────────────────────┤
│  Codegen (build-time):                                  │
│  └── telegram-bot-api-spec → Go types + validators     │
└─────────────────────────────────────────────────────────┘
```

## Bot API Handler

Processes requests to `/bot<token>/<method>` endpoints.

### Request Flow

```
Request → Token Validation → Scenario Check → Request Validation → Response Generation
```

1. **Token Validation**
   - Validate format matches `<bot_id>:<secret>` pattern
   - If token registry enabled, verify token is registered
   - Check token status (active, banned, deactivated)
   - Return appropriate error if validation fails

2. **Scenario Check**
   - Check for `X-TG-Mock-Scenario` header
   - Check control API for queued scenarios matching this method
   - If scenario matches, return configured response

3. **Request Validation**
   - Validate method exists in spec
   - Validate required parameters present
   - Validate parameter types match spec
   - Return 400 with descriptive error if validation fails

4. **Response Generation**
   - Generate response with correct structure from spec
   - Reflect input values where applicable
   - Assign generated IDs for created resources
   - For file uploads, store file and return `file_id`

### Token Registry

```yaml
tokens:
  "123456789:ABC-xyz":
    status: active        # active | banned | deactivated
    bot_name: "TestBot"

  "987654321:XYZ-abc":
    status: banned        # Returns 403 "bot was banned"
```

## Control API

Located at `/__control/*` for managing test scenarios and injecting updates.

### Endpoints

**Scenarios:**
```
POST   /__control/scenarios          # Queue a scenario
GET    /__control/scenarios          # List active scenarios
DELETE /__control/scenarios          # Clear all scenarios
DELETE /__control/scenarios/:id      # Remove specific scenario
```

**Updates:**
```
POST   /__control/updates            # Inject an update
GET    /__control/updates            # View pending updates queue
DELETE /__control/updates            # Clear update queue
```

**State:**
```
POST   /__control/reset              # Reset all state
GET    /__control/state              # Debug: view internal state
```

**Tokens:**
```
POST   /__control/tokens             # Register a token
DELETE /__control/tokens/:token      # Remove a token
PATCH  /__control/tokens/:token      # Update token status
```

### Scenario Object

```json
{
  "id": "auto-generated",
  "method": "sendMessage",
  "match": { "chat_id": 123 },
  "times": 1,
  "response": {
    "error_code": 429,
    "description": "Too Many Requests: retry after 30",
    "retry_after": 30
  }
}
```

## Updates & Webhooks

### Update Injection

```json
POST /__control/updates
{
  "update_id": 123,
  "message": {
    "message_id": 1,
    "from": { "id": 111, "first_name": "Test", "is_bot": false },
    "chat": { "id": 111, "type": "private" },
    "date": 1703500000,
    "text": "Hello bot!"
  }
}
```

### Polling Mode (getUpdates)

- Injected updates queue up in memory
- `getUpdates` returns queued updates respecting `offset` and `limit`
- Updates removed from queue once acknowledged
- Long polling supported via `timeout` parameter

### Webhook Mode

Enabled when bot calls `setWebhook`:
- Injected updates POSTed to webhook URL immediately
- `getUpdates` returns error (matches real Telegram behavior)
- `deleteWebhook` switches back to polling mode

### Scripted Conversations

```yaml
conversations:
  - name: "greeting_flow"
    updates:
      - delay: 0
        message: { text: "/start", from: { id: 111 } }
      - delay: 500ms
        message: { text: "What can you do?", from: { id: 111 } }
```

Triggered via: `POST /__control/conversations/greeting_flow/start`

## Error Scenarios

### Pre-built Errors

**400 Bad Request:**
- `button_url_invalid`, `chat_not_found`, `entities_too_long`, `file_too_big`
- `group_deactivated`, `group_migrated`, `invalid_file_id`, `member_not_found`
- `message_cant_be_deleted`, `message_cant_be_edited`, `message_not_modified`
- `message_text_empty`, `message_to_delete_not_found`, `message_to_edit_not_found`
- `not_enough_rights_photos`, `not_enough_rights_text`, `peer_id_invalid`
- `reply_message_not_found`, `user_not_found`, `wrong_parameter_action`

**403 Forbidden:**
- `bot_blocked`, `cant_initiate_conversation`, `cant_send_to_bots`
- `not_member_channel`, `not_member_supergroup`, `bot_kicked`, `user_deactivated`

**409 Conflict:**
- `webhook_active`, `terminated_by_other_long_poll`

**429 Rate Limit:**
- `too_many_requests` (with `retry_after`)

**401 Unauthorized:**
- `unauthorized` (bad token)

Source: https://github.com/TelegramBotAPI/errors

### Header-Based Triggering

```bash
curl -H "X-TG-Mock-Scenario: rate_limit" \
     -H "X-TG-Mock-Retry-After: 60" \
     http://localhost:8081/bot123:abc/sendMessage
```

## File Storage

### Interface

```go
type FileStore interface {
    Store(data []byte, filename string, mimeType string) (fileID string, err error)
    Get(fileID string) (data []byte, metadata FileMetadata, err error)
    GetPath(fileID string) (filePath string, err error)
    Delete(fileID string) error
    Clear() error
}
```

### Implementations

- **MemoryStore (default):** Fast, no cleanup needed
- **DiskStore (--storage-dir):** Persists across restarts

Files downloadable at: `http://localhost:8081/file/bot<token>/<file_path>`

## Configuration

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8081` | HTTP server port |
| `--config` | none | Path to YAML config file |
| `--storage-dir` | none | Directory for file storage |
| `--verbose` | `false` | Enable detailed logging |
| `--strict` | `false` | Reject unknown parameters |

### Config File

```yaml
server:
  port: 8081
  verbose: true
  strict: false

storage:
  dir: ./uploads

tokens:
  "123456789:ABC-xyz":
    status: active
    bot_name: "MyTestBot"

scenarios:
  - method: sendMessage
    match: { chat_id: 999 }
    response:
      error_code: 400
      description: "Bad Request: chat not found"

conversations:
  greeting_flow:
    - delay: 0
      message:
        text: "/start"
        from: { id: 111, first_name: "Tester" }
        chat: { id: 111, type: "private" }
```

## Code Generation

### Input

JSON spec from https://github.com/PaulSonOfLars/telegram-bot-api-spec

### Generated Output

- `gen/types.go` - Telegram API types as Go structs
- `gen/methods.go` - Method registry & specs
- `gen/fixtures.go` - Response fixtures
- `gen/errors.go` - Pre-built error responses
- `gen/validators.go` - Request validators

### Build Process

```bash
go generate ./...
```

## Project Structure

```
tg-mock/
├── cmd/
│   ├── tg-mock/           # Main server binary
│   └── codegen/           # Code generation tool
├── gen/                   # Generated code (do not edit)
├── internal/
│   ├── server/            # HTTP server & routing
│   ├── scenario/          # Scenario engine
│   ├── updates/           # Update queue & webhook pusher
│   ├── storage/           # File storage
│   ├── tokens/            # Token registry
│   └── config/            # Config loading
├── spec/                  # Telegram API spec
├── errors/                # Error definitions
├── go.mod
├── Makefile
└── README.md
```

## Dependencies

- `github.com/go-chi/chi/v5` - HTTP router
- `gopkg.in/yaml.v3` - YAML config parsing
- Standard library for everything else

## Distribution

Initial release: `go install` only

Future: Binaries, Docker, Homebrew as needed
