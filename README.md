# tg-mock

A mock Telegram Bot API server for testing bots and bot libraries. Inspired by [stripe/stripe-mock](https://github.com/stripe/stripe-mock).

## Table of Contents

- [tg-mock](#tg-mock)
  - [Table of Contents](#table-of-contents)
  - [Background](#background)
  - [Install](#install)
    - [Docker](#docker)
    - [Using Go](#using-go)
    - [From Source](#from-source)
  - [Usage](#usage)
    - [CLI Flags](#cli-flags)
    - [Connecting Your Bot](#connecting-your-bot)
    - [Configuration](#configuration)
  - [Response Generation](#response-generation)
    - [Smart Faker](#smart-faker)
    - [Deterministic Mode](#deterministic-mode)
  - [Control API](#control-api)
    - [Scenarios](#scenarios)
    - [Response Data Overrides](#response-data-overrides)
    - [Updates](#updates)
    - [Header-based Errors](#header-based-errors)
      - [Available Built-in Scenarios](#available-built-in-scenarios)
  - [Examples](#examples)
    - [Testing Error Handling](#testing-error-handling)
    - [Simulating Incoming Messages](#simulating-incoming-messages)
    - [Custom Response Data](#custom-response-data)
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

### Docker

```bash
docker pull ghcr.io/watzon/tg-mock:latest

# Run with defaults
docker run -p 8081:8081 ghcr.io/watzon/tg-mock

# With custom config and persistent storage
docker run -p 8081:8081 \
  -v ./config.yaml:/config.yaml \
  -v ./data:/data \
  ghcr.io/watzon/tg-mock --config /config.yaml
```

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

| Flag            | Description                                     | Default    |
| --------------- | ----------------------------------------------- | ---------- |
| `--port`        | HTTP server port                                | 8081       |
| `--config`      | Path to YAML config file                        | (none)     |
| `--verbose`     | Enable verbose logging                          | false      |
| `--storage-dir` | Directory for file storage                      | (temp dir) |
| `--faker-seed`  | Seed for faker (0 = random, >0 = deterministic) | 0          |

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
  faker_seed: 12345  # Fixed seed for reproducible tests (0 = random)

storage:
  dir: /tmp/tg-mock-files

tokens:
  "123456789:ABC-xyz":
    status: active
    bot_name: MyTestBot
  "987654321:XYZ-abc":
    status: revoked

scenarios:
  # Error scenario
  - method: sendMessage
    match:
      chat_id: 999
    response:
      error_code: 400
      description: "Bad Request: chat not found"

  # Success override scenario
  - method: getMe
    response_data:
      id: 123456789
      first_name: "MyTestBot"
      username: "my_test_bot"
```

## Response Generation

tg-mock generates realistic mock responses for all Telegram Bot API methods using a smart faker system.

### Smart Faker

The faker uses field name heuristics to generate appropriate values:

| Field Pattern             | Generated Value             |
| ------------------------- | --------------------------- |
| `*_id`, `*Id`             | Large random int64          |
| `date`, `*_date`          | Recent Unix timestamp       |
| `username`                | `@user_xxxx` format         |
| `first_name`, `last_name` | Realistic names             |
| `title`, `name`           | Title-case strings          |
| `text`, `caption`         | Lorem-like text             |
| `url`, `*_url`            | `https://example.com/...`   |
| `latitude`                | -90 to 90                   |
| `longitude`               | -180 to 180                 |
| `file_path`               | `files/document.ext`        |
| `is_*`, `can_*`, `has_*`  | Boolean                     |
| `language_code`           | IETF tag (e.g., `en`, `ru`) |

The faker also reflects request parameters back into responses. For example, when you call `sendMessage` with `chat_id: 12345`, the response `Message.chat.id` will be `12345`.

### Deterministic Mode

For reproducible tests, use a fixed faker seed:

```bash
# CLI flag
tg-mock --faker-seed 12345

# Or in config file
server:
  faker_seed: 12345
```

With a fixed seed, the same sequence of API calls will always produce identical responses. This is essential for snapshot testing and debugging flaky tests.

When `faker_seed` is 0 (the default), responses are randomized on each server start.

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

### Response Data Overrides

Scenarios can also override specific fields in successful responses without triggering errors. This is useful for testing specific data conditions:

```bash
# Override specific fields in the response
curl -X POST http://localhost:8081/__control/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "method": "getChat",
    "match": {"chat_id": 12345},
    "times": 1,
    "response_data": {
      "id": 12345,
      "type": "supergroup",
      "title": "My Custom Group",
      "username": "mycustomgroup"
    }
  }'

# The next getChat call for chat_id 12345 will return a supergroup
# with the specified title and username, while other fields are faker-generated
```

You can combine `response_data` with `match` to create conditional responses:

```bash
# Different response based on chat_id
curl -X POST http://localhost:8081/__control/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "method": "sendMessage",
    "match": {"chat_id": 999},
    "response_data": {
      "message_id": 42,
      "text": "Custom reply for chat 999"
    }
  }'
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

### Custom Response Data

```bash
# Start with deterministic mode for reproducible tests
tg-mock --faker-seed 12345 &

# Set up a scenario that returns a specific user for getMe
curl -X POST http://localhost:8081/__control/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "method": "getMe",
    "response_data": {
      "id": 123456789,
      "is_bot": true,
      "first_name": "TestBot",
      "username": "my_test_bot",
      "can_join_groups": true,
      "can_read_all_group_messages": false,
      "supports_inline_queries": true
    }
  }'

# Now getMe returns your custom bot info
curl http://localhost:8081/bot123:abc/getMe
# Returns: {"ok":true,"result":{"id":123456789,"is_bot":true,"first_name":"TestBot",...}}

# Set up photo responses with specific dimensions
curl -X POST http://localhost:8081/__control/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "method": "sendPhoto",
    "response_data": {
      "photo": [
        {"file_id": "small_photo_id", "width": 90, "height": 90},
        {"file_id": "medium_photo_id", "width": 320, "height": 320},
        {"file_id": "large_photo_id", "width": 800, "height": 800}
      ]
    }
  }'
```

## Contributing

PRs are welcome! Please open an issue first to discuss any major changes.

If you find a bug or have a feature request, please [open an issue](https://github.com/watzon/tg-mock/issues).

## License

MIT © watzon
