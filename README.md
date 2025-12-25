# tg-mock

A mock Telegram Bot API server for testing bots and bot libraries. Inspired by [stripe/stripe-mock](https://github.com/stripe/stripe-mock).

## Table of Contents

- [tg-mock](#tg-mock)
  - [Table of Contents](#table-of-contents)
  - [Background](#background)
  - [Install](#install)
    - [Using Go](#using-go)
    - [From Source](#from-source)
  - [Usage](#usage)
    - [CLI Flags](#cli-flags)
    - [Connecting Your Bot](#connecting-your-bot)
    - [Configuration](#configuration)
  - [Control API](#control-api)
    - [Scenarios](#scenarios)
    - [Updates](#updates)
    - [Header-based Errors](#header-based-errors)
      - [Available Built-in Scenarios](#available-built-in-scenarios)
  - [Examples](#examples)
    - [Testing Error Handling](#testing-error-handling)
    - [Simulating Incoming Messages](#simulating-incoming-messages)
  - [Contributing](#contributing)
  - [License](#license)

## Background

Testing Telegram bots is challenging because:
- The real Telegram API requires network connectivity
- Simulating errors and edge cases is difficult
- You can't control the timing and content of incoming updates

tg-mock solves these problems by providing a drop-in replacement for `api.telegram.org` that:
- Validates requests against the official Bot API spec
- Generates realistic responses for all Bot API methods
- Supports scenario-based error simulation
- Provides a control API for injecting updates and managing test scenarios
- Includes built-in error responses for common Telegram API errors
- Handles file uploads with configurable storage

## Install

### Using Go

```bash
go install github.com/watzon/tg-mock/cmd/tg-mock@latest
```

### From Source

```bash
git clone https://github.com/watzon/tg-mock.git
cd tg-mock
go build -o tg-mock ./cmd/tg-mock
```

## Usage

```bash
# Start with defaults (port 8081)
tg-mock

# Custom port
tg-mock --port 9090

# With config file
tg-mock --config config.yaml

# Verbose logging
tg-mock --verbose

# Custom file storage directory
tg-mock --storage-dir /tmp/tg-mock-files
```

### CLI Flags

| Flag            | Description                | Default    |
| --------------- | -------------------------- | ---------- |
| `--port`        | HTTP server port           | 8081       |
| `--config`      | Path to YAML config file   | (none)     |
| `--verbose`     | Enable verbose logging     | false      |
| `--storage-dir` | Directory for file storage | (temp dir) |

### Connecting Your Bot

Point your bot library to the mock server:

```
http://localhost:8081/bot<TOKEN>/<METHOD>
```

For example:

```bash
curl http://localhost:8081/bot123456789:ABC-xyz/getMe
```

### Configuration

Create a YAML configuration file for persistent settings:

```yaml
server:
  port: 8081
  verbose: true

storage:
  dir: /tmp/tg-mock-files

tokens:
  "123456789:ABC-xyz":
    status: active
    bot_name: MyTestBot
  "987654321:XYZ-abc":
    status: revoked

scenarios:
  - method: sendMessage
    match:
      chat_id: 999
    response:
      error_code: 400
      description: "Bad Request: chat not found"
```

## Control API

The control API allows you to manage scenarios and inject updates during tests.

### Scenarios

Add test scenarios to simulate specific responses:

```bash
# Add a scenario that returns an error once
curl -X POST http://localhost:8081/__control/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "method": "sendMessage",
    "times": 1,
    "response": {
      "error_code": 429,
      "description": "Too Many Requests: retry after 30",
      "retry_after": 30
    }
  }'

# Add a scenario with request matching
curl -X POST http://localhost:8081/__control/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "method": "sendMessage",
    "match": {"chat_id": 999},
    "times": 1,
    "response": {
      "error_code": 400,
      "description": "Bad Request: chat not found"
    }
  }'

# List all active scenarios
curl http://localhost:8081/__control/scenarios

# Clear all scenarios
curl -X DELETE http://localhost:8081/__control/scenarios
```

### Updates

Inject updates to simulate incoming messages, callbacks, etc.:

```bash
# Inject a message update
curl -X POST http://localhost:8081/__control/updates \
  -H "Content-Type: application/json" \
  -d '{
    "message": {
      "message_id": 1,
      "text": "Hello from test!",
      "chat": {"id": 123, "type": "private"},
      "from": {"id": 456, "is_bot": false, "first_name": "Test"}
    }
  }'

# View pending updates
curl http://localhost:8081/__control/updates
```

### Header-based Errors

Use the `X-TG-Mock-Scenario` header to trigger built-in error responses:

```bash
# Trigger rate limiting
curl -H "X-TG-Mock-Scenario: rate_limit" \
  http://localhost:8081/bot123:abc/sendMessage

# Trigger bot blocked error
curl -H "X-TG-Mock-Scenario: bot_blocked" \
  http://localhost:8081/bot123:abc/sendMessage

# Trigger chat not found
curl -H "X-TG-Mock-Scenario: chat_not_found" \
  http://localhost:8081/bot123:abc/sendMessage
```

#### Available Built-in Scenarios

| Scenario               | Error Code | Description                                                   |
| ---------------------- | ---------- | ------------------------------------------------------------- |
| `bad_request`          | 400        | Bad Request                                                   |
| `chat_not_found`       | 400        | Bad Request: chat not found                                   |
| `user_not_found`       | 400        | Bad Request: user not found                                   |
| `message_not_found`    | 400        | Bad Request: message to edit not found                        |
| `message_not_modified` | 400        | Bad Request: message is not modified                          |
| `message_too_long`     | 400        | Bad Request: message is too long                              |
| `file_too_big`         | 400        | Bad Request: file is too big                                  |
| `invalid_file_id`      | 400        | Bad Request: invalid file id                                  |
| `unauthorized`         | 401        | Unauthorized                                                  |
| `forbidden`            | 403        | Forbidden                                                     |
| `bot_blocked`          | 403        | Forbidden: bot was blocked by the user                        |
| `bot_kicked`           | 403        | Forbidden: bot was kicked from the chat                       |
| `cant_initiate`        | 403        | Forbidden: bot can't initiate conversation with a user        |
| `webhook_active`       | 409        | Conflict: can't use getUpdates method while webhook is active |
| `rate_limit`           | 429        | Too Many Requests: retry after 30                             |
| `flood_wait`           | 429        | Flood control exceeded. Retry in 60 seconds                   |

## Examples

### Testing Error Handling

```bash
# Start the mock server
tg-mock --verbose &

# Test that your bot handles rate limiting correctly
curl -X POST http://localhost:8081/__control/scenarios \
  -d '{"method":"sendMessage","times":3,"response":{"error_code":429,"description":"Too Many Requests","retry_after":5}}'

# Your bot should now get rate limited on the next 3 sendMessage calls
```

### Simulating Incoming Messages

```bash
# Inject a /start command
curl -X POST http://localhost:8081/__control/updates \
  -H "Content-Type: application/json" \
  -d '{
    "message": {
      "message_id": 1,
      "text": "/start",
      "chat": {"id": 123, "type": "private"},
      "from": {"id": 456, "is_bot": false, "first_name": "User"},
      "entities": [{"type": "bot_command", "offset": 0, "length": 6}]
    }
  }'

# Your bot can now receive this via getUpdates
```

## Contributing

PRs are welcome! Please open an issue first to discuss any major changes.

If you find a bug or have a feature request, please [open an issue](https://github.com/watzon/tg-mock/issues).

## License

MIT © watzon
