# Home IP Notifier

[![pipeline status](https://git.windmaker.net/a-castellano/home-ip-notifier/badges/master/pipeline.svg)](https://git.windmaker.net/a-castellano/home-ip-notifier/pipelines)
[![coverage report](https://git.windmaker.net/a-castellano/home-ip-notifier/badges/master/coverage.svg)](https://a-castellano.gitpages.windmaker.net/home-ip-notifier/coverage.html)
[![Quality Gate Status](https://sonarqube.windmaker.net/api/project_badges/measure?project=a-castellano_home-ip-notifier_c4da9a70-dcc5-4ef5-8425-3f91b0d7526d&metric=alert_status&token=sqb_efd83d3e4b6a20b336f469385f469e63fdab1fc3)](https://sonarqube.windmaker.net/dashboard?id=a-castellano_home-ip-notifier_c4da9a70-dcc5-4ef5-8425-3f91b0d7526d)

A Go-based service that consumes the IP-change notifications published by
[home-ip-monitor](https://git.windmaker.net/a-castellano/home-ip-monitor) on
RabbitMQ and delivers them by email. It is designed to run as a long-lived
systemd service.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Development](#development)
- [Testing](#testing)
- [License](#license)

## Overview

Home IP Notifier is a lightweight service that:

1. **Subscribes** to the RabbitMQ notification queue fed by home-ip-monitor
2. **Unwraps** each message envelope and continues the distributed trace it carries
3. **Builds** an email notification for the configured destination address
4. **Delivers** it through SMTP; failed messages are logged and dropped
   (consumption is auto-ack, so exiting would not requeue them)

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ home-ip-monitor │───▶│  RabbitMQ Queue  │───▶│ home-ip-notifier│
│  (IP Detector)  │    │   (envelopes)    │    │                 │
└─────────────────┘    └──────────────────┘    └────────┬────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │   SMTP Server   │
                                               │    (Email)      │
                                               └─────────────────┘
```

## Features

- **Distributed tracing**: messages arrive as envelopes carrying OpenTelemetry
  context, so the trace started by the monitor continues through the email
  delivery
- **Structured logging** through [`log/slog`](https://pkg.go.dev/log/slog)
  (via go-types `slog`), captured by systemd into the journal
- **Resilient consumption**: malformed or empty messages are logged and
  dropped instead of crash-looping the service
- **Graceful shutdown** on SIGINT/SIGTERM, flushing pending telemetry before
  exiting
- **Systemd service** with environment-based configuration

## Architecture

The project follows a Clean Architecture layout: an inner `domain` layer that
depends on nothing, an `app` layer that holds the use case, and an `infra`
layer with the adapters that talk to the outside world (RabbitMQ, SMTP).
Dependencies always point inwards — the use case only knows about the domain
port (interface), never about concrete infrastructure.

```
                      cmd/home-ip-notifier (composition root)
                                    │ wires adapters into the use case
                                    ▼
┌──────────────────────────────────────────────────────────────────────┐
│ infra/consume: inbound adapter                                       │
│   unwraps envelope → continues trace → calls the use case            │
└──────────────────────────────────────────────────────────────────────┘
        │ calls through its own Processor interface
        ▼
┌──────────────────────────────────────────────────────────────────────┐
│ app: Announcer use case                                              │
│   fixed subject "Home IP has changed" + received message             │
└──────────────────────────────────────────────────────────────────────┘
        │ depends only on the domain port (interface)
        ▼
┌──────────────────────────────────────────────────────────────────────┐
│ domain: Notifier                                                     │
└──────────────────────────────────────────────────────────────────────┘
        ▲ implemented by
        │
┌──────────────────────────────────────────────────────────────────────┐
│ infra/notify → go-services notificator → SMTP                        │
└──────────────────────────────────────────────────────────────────────┘
```

### Layers and components

- **`internal/domain`**: the `Notifier` port (interface), with no external
  dependencies. It is defined at the level this application needs — title and
  message; the destination is deployment configuration and lives in the
  adapter.
- **`internal/app`**: the `Announcer` use case. It receives the domain port
  via `NewAnnouncer`, so it has zero knowledge of SMTP or RabbitMQ. The
  notification subject is a business constant of this layer.
- **`internal/infra/consume`**: inbound adapter. Unmarshals the envelope,
  extracts the OpenTelemetry context and calls the use case under a CONSUMER
  span that joins the monitor's trace.
- **`internal/infra/notify`**: outbound adapter implementing `domain.Notifier`
  over the go-services `notificator` backed by an SMTP `MailSender`.
- **`internal/infra/config`**: environment-based configuration, composing the
  SMTP and RabbitMQ configs provided by go-types.
- **`cmd/home-ip-notifier`**: the composition root (`main`) that builds every
  adapter, wires them into the use case and runs the message loop.

## Installation

### Prerequisites

- **Arch Linux** or compatible distribution
- **RabbitMQ** server for message queuing
- **SMTP** server for email delivery
- **Systemd** for service management (Linux)

### Package Installation (Recommended)

The project provides pre-built packages for Arch Linux. Download the latest package from the CI artifacts:

1. **Download the package** from the latest successful pipeline:

   - Go to [GitLab CI/CD Pipelines](https://git.windmaker.net/a-castellano/home-ip-notifier/pipelines)
   - Find the latest successful pipeline
   - Download the `arch_package` artifact

2. **Install the package**:

   ```bash
   # Install the downloaded package
   sudo pacman -U windmaker-home-ip-notifier-*.pkg.tar.zst
   ```

3. **Configure the service**:

   ```bash
   # Copy the sample file and edit your configuration
   sudo cp /etc/default/windmaker-home-ip-notifier-example /etc/default/windmaker-home-ip-notifier
   sudo vim /etc/default/windmaker-home-ip-notifier
   ```

4. **Enable and start the service**:
   ```bash
   sudo systemctl enable windmaker-home-ip-notifier.service
   sudo systemctl start windmaker-home-ip-notifier.service
   ```

## Configuration

### Environment Variables

All configuration is done through environment variables.

#### Required Variables

| Variable      | Description                             | Example               |
| ------------- | --------------------------------------- | --------------------- |
| `DESTINATION` | Recipient email address (validated)     | `"admin@example.com"` |

#### Optional Variables

| Variable            | Description                     | Default                           |
| ------------------- | ------------------------------- | --------------------------------- |
| `NOTIFY_QUEUE_NAME` | Queue to consume envelopes from | `"home-ip-monitor-notifications"` |

#### Application and Logging

Logging is handled through [go-types `slog`](https://git.windmaker.net/a-castellano/go-types/-/tree/master/slog). `APP_NAME` is required by that type; the rest fall back to sane defaults.

| Variable          | Description                                   | Default      |
| ----------------- | --------------------------------------------- | ------------ |
| `APP_NAME`        | Application name attached to every log entry  | _(required)_ |
| `SLOG_LEVEL`      | Log level: `Debug`, `Info`, `Warn` or `Error` | `Info`       |
| `SLOG_FORMAT`     | Log format: `JSON` or `plain`                 | `JSON`       |
| `SLOG_ADD_SOURCE` | Whether to add `file:line` to log entries     | `true`       |

#### Telemetry

OpenTelemetry is opt-in through [go-types `opentelemetry`](https://git.windmaker.net/a-castellano/go-types/-/tree/master/opentelemetry). `APP_NAME` doubles as the telemetry `service.name`, so `OTEL_SERVICE_NAME` and `OTEL_RESOURCE_ATTRIBUTES` must **not** be set — the config rejects them.

| Variable           | Description                                | Default |
| ------------------ | ------------------------------------------ | ------- |
| `ENABLE_TELEMETRY` | Enable tracing: only `true` or `false`     | `false` |

#### SMTP Configuration

See [go-types SMTP documentation](https://git.windmaker.net/a-castellano/go-types/-/tree/master/smtp) for complete SMTP configuration options.

| Variable            | Description                                    | Default      |
| ------------------- | ---------------------------------------------- | ------------ |
| `SMTP_FROM`         | Sender address (full, validated email)         | _(required)_ |
| `SMTP_HOST`         | SMTP server hostname                           | _(required)_ |
| `SMTP_PORT`         | SMTP server port (integer)                     | _(required)_ |
| `SMTP_USERNAME`     | SMTP authentication username                   | _(required)_ |
| `SMTP_PASSWORD`     | SMTP authentication password                   | _(required)_ |
| `SMTP_VALIDATE_TLS` | Enable/disable TLS certificate validation      | `"true"`     |

#### RabbitMQ Configuration

See [go-types RabbitMQ documentation](https://git.windmaker.net/a-castellano/go-types/-/tree/master/rabbitmq) for complete RabbitMQ configuration options.

| Variable            | Description              | Default       |
| ------------------- | ------------------------ | ------------- |
| `RABBITMQ_HOST`     | RabbitMQ server hostname | `"localhost"` |
| `RABBITMQ_PORT`     | RabbitMQ server port     | `5672`        |
| `RABBITMQ_USER`     | RabbitMQ username        | `"guest"`     |
| `RABBITMQ_PASSWORD` | RabbitMQ password        | `"guest"`     |

### Configuration File

The package installs a sample file at `/etc/default/windmaker-home-ip-notifier-example`. Copy it to `/etc/default/windmaker-home-ip-notifier` (the path read by the systemd unit) and edit it to configure the service:

```bash
sudo cp /etc/default/windmaker-home-ip-notifier-example /etc/default/windmaker-home-ip-notifier
sudo vim /etc/default/windmaker-home-ip-notifier
```

```bash
# Application and logging
APP_NAME="home-ip-notifier"
SLOG_LEVEL="Info"
SLOG_FORMAT="JSON"

# Required configuration
DESTINATION="admin@example.com"

# Queue configuration
NOTIFY_QUEUE_NAME="home-ip-monitor-notifications"

# SMTP configuration
SMTP_FROM="no-reply@example.com"
SMTP_HOST="smtp.example.com"
SMTP_PORT=587
SMTP_USERNAME="user@example.com"
SMTP_PASSWORD="your-password"
SMTP_VALIDATE_TLS="true"

# RabbitMQ configuration
RABBITMQ_HOST="localhost"
RABBITMQ_PORT=5672
RABBITMQ_USER="guest"
RABBITMQ_PASSWORD="guest"
```

## Usage

### Service Management

```bash
# Enable and start the service
sudo systemctl enable windmaker-home-ip-notifier.service
sudo systemctl start windmaker-home-ip-notifier.service

# Check service status
sudo systemctl status windmaker-home-ip-notifier.service

# View logs
sudo journalctl -u windmaker-home-ip-notifier.service -f
```

### Message Format

The service consumes the envelopes home-ip-monitor publishes on the
notification queue: a JSON document carrying the OpenTelemetry propagation
context (`carrier`) and the human-readable message (`body`):

```json
{
  "carrier": { "traceparent": "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01" },
  "body": "SG9tZSBJUCBoYXMgY2hhbmdlZCB0byAxOTIuMTY4LjEuMTAwLg=="
}
```

The body is delivered as the email content, under the subject
`Home IP has changed` (prefixed with the severity by the notificator library,
e.g. `[Info] - Home IP has changed`). Messages that are not valid envelopes,
or whose body is empty, are logged at Error level and discarded.

### Monitoring and Logging

The service uses structured logging through [`log/slog`](https://pkg.go.dev/log/slog) (via go-types `slog`). Output goes to standard streams, which systemd captures into the journal. The format (`JSON` or `plain`) and verbosity are controlled by `SLOG_FORMAT` and `SLOG_LEVEL`.

- **DEBUG**: configuration loading, wiring and per-message processing steps
- **INFO**: high-level milestones — service startup, waiting for messages,
  notification delivery
- **ERROR**: connection failures, configuration errors, invalid messages,
  delivery failures

Each entry carries an `operation` attribute (e.g. `NewConfig`, `consume`,
`main.run`) plus structured fields. With telemetry enabled, delivery failures
also show up as failed spans in the trace the monitor started.

## Development

### Prerequisites

- Go 1.26+
- Docker (or Podman) and Docker (or Podman) Compose
- Make

### Setup Development Environment

Every Go task runs inside the development container defined in
`development/docker-compose.yml` — the same image used in CI and production:

```bash
# Start development services (Go container + RabbitMQ + MailHog)
docker-compose -f development/docker-compose.yml up -d

# Run tests inside the container
docker-compose -f development/docker-compose.yml exec golang make test
docker-compose -f development/docker-compose.yml exec golang make test_integration

# Generate coverage report
docker-compose -f development/docker-compose.yml exec golang make coverage
docker-compose -f development/docker-compose.yml exec golang make coverhtml
```

Available services:

- **RabbitMQ Management**: http://localhost:15672
- **MailHog Web UI**: http://localhost:8025 (emails sent by the tests land here)
- **MailHog SMTP (TLS)**: localhost:6465

### Testing Email Functionality

Use `swaks` to test email delivery in the development environment (inside the golang container):

```bash
docker-compose -f development/docker-compose.yml exec golang /bin/bash
# Install swaks (if not already installed)
apt-get install swaks

# Test email sending
swaks --to test@example.com \
      --server mailhog:6465 \
      --tls \
      --header "Subject: Test IP Change Notification" \
      --body "Your home IP has changed to 192.168.1.100"
```

### Project Structure

```
home-ip-notifier/
├── cmd/
│   └── home-ip-notifier/   # main package: composition root / wiring
├── internal/
│   ├── domain/             # Notifier port (no external deps)
│   ├── app/                # Announcer use case (depends only on domain)
│   └── infra/              # adapters around the use case
│       ├── config/         # environment-based configuration
│       ├── consume/        # envelope unwrapping + trace continuation
│       └── notify/         # SMTP delivery via go-services notificator
├── development/            # Docker/Podman dev setup and coverage script
└── packaging/              # nfpm spec, systemd unit and defaults
```

## Testing

### Test Types

- **Unit Tests**: individual component testing with fakes for the ports
- **Integration Tests**: full delivery testing against the RabbitMQ and
  MailHog containers

### Running Tests

```bash
# Unit tests only
make test

# Integration tests
make test_integration

# All tests with coverage
make coverage

# HTML coverage report
make coverhtml
```

## License

This project is licensed under the GPLv3 License - see the [LICENSE](LICENSE) file for details.

## Authors

- **Álvaro Castellano Vela** - _Initial work_ - [a-castellano](https://git.windmaker.net/a-castellano)

## Related Projects

- [home-ip-monitor](https://git.windmaker.net/a-castellano/home-ip-monitor) - The IP monitoring service that feeds this notifier
- [go-services](https://git.windmaker.net/a-castellano/go-services) - Shared Go service utilities
- [go-types](https://git.windmaker.net/a-castellano/go-types) - Shared Go type definitions
