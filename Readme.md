# Home IP Notifier

[![pipeline status](https://git.windmaker.net/a-castellano/home-ip-notifier/badges/master/pipeline.svg)](https://git.windmaker.net/a-castellano/home-ip-notifier/pipelines)
[![coverage report](https://git.windmaker.net/a-castellano/home-ip-notifier/badges/master/coverage.svg)](https://a-castellano.gitpages.windmaker.net/home-ip-notifier/coverage.html)
[![Quality Gate Status](https://sonarqube.windmaker.net/api/project_badges/measure?project=a-castellano_home-ip-notifier_c4da9a70-dcc5-4ef5-8425-3f91b0d7526d&metric=alert_status&token=sqb_efd83d3e4b6a20b336f469385f469e63fdab1fc3)](https://sonarqube.windmaker.net/dashboard?id=a-castellano_home-ip-notifier_c4da9a70-dcc5-4ef5-8425-3f91b0d7526d)

Go microservice that monitors IP changes and sends email notifications. This service is part of the [home-ip-monitor](https://git.windmaker.net/a-castellano/home-ip-monitor) ecosystem and subscribes to RabbitMQ queues to receive IP change notifications.

## What This Program Does

The Home IP Notifier is designed to:

- **Subscribe** to a RabbitMQ queue for IP change notifications
- **Process** incoming IP change messages
- **Send** email notifications when IP changes are detected
- **Provide** reliable, production-ready IP monitoring notifications

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  home-ip-monitor│───▶│  RabbitMQ Queue  │───▶│ home-ip-notifier│
│  (IP Detector)  │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                                          │
                                                          ▼
                                               ┌─────────────────┐
                                               │   SMTP Server   │
                                               │   (Email)       │
                                               └─────────────────┘
```

## Features

- **Reliable Message Processing**: Handles RabbitMQ messages with error recovery
- **Graceful Shutdown**: Proper signal handling for clean service termination
- **System Integration**: Syslog logging for production environments
- **Configurable**: Environment-based configuration management
- **Production Ready**: Systemd service with security hardening

## Prerequisites

- **Go 1.24+** for building and development
- **RabbitMQ Server** for message queuing
- **SMTP Server** for email delivery
- **Linux/Unix** system for production deployment

## Configuration

### Environment Variables

#### Required Variables

| Variable       | Description                            | Example               |
| -------------- | -------------------------------------- | --------------------- |
| `MAILFROM`     | Sender email username (without domain) | `"no-reply"`          |
| `MAILDOMAIN`   | Sender email domain                    | `"example.com"`       |
| `SMTPHOST`     | SMTP server hostname                   | `"smtp.gmail.com"`    |
| `SMTPPORT`     | SMTP server port                       | `"587"`               |
| `SMTPNAME`     | SMTP authentication username           | `"user@example.com"`  |
| `SMTPPASSWORD` | SMTP authentication password           | `"your-password"`     |
| `DESTINATION`  | Recipient email address                | `"admin@example.com"` |

#### Optional Variables

| Variable            | Description                               | Default                           |
| ------------------- | ----------------------------------------- | --------------------------------- |
| `NOTIFY_QUEUE_NAME` | RabbitMQ queue name for notifications     | `"home-ip-monitor-notifications"` |
| `SMTPTLSVALIDATION` | Enable/disable TLS certificate validation | `"true"`                          |

#### RabbitMQ Configuration

The following RabbitMQ environment variables are required (see [go-types documentation](https://git.windmaker.net/a-castellano/go-types/-/tree/master/rabbitmq?ref_type=heads)):

| Variable            | Description              | Default       |
| ------------------- | ------------------------ | ------------- |
| `RABBITMQ_HOST`     | RabbitMQ server hostname | `"localhost"` |
| `RABBITMQ_PORT`     | RabbitMQ server port     | `"5672"`      |
| `RABBITMQ_USER`     | RabbitMQ username        | `"guest"`     |
| `RABBITMQ_PASSWORD` | RabbitMQ password        | `"guest"`     |

## Development

### Running Tests

```bash
# Run unit tests
make test

# Run integration tests
make test_integration

# Generate coverage report
make coverage
```

### Development Environment

The project includes a Docker Compose setup for development:

```bash
# Start development environment
cd development
docker-compose up -d

# Access services:
# - RabbitMQ Management: http://localhost:15672
# - MailHog Web UI: http://localhost:8025
# - MailHog SMTP: localhost:6465
```

### Testing Email Functionality

Use `swaks` to test email functionality in the development environment (inside golang iimage):

```bash
docker-compose -f development/docker-compose.yml exec  golang /bin/bash
# Install swaks (if not already installed)
apt-get install swaks

# Test email sending
swaks --to test@example.com \
      --server mailhog:6465 \
      --tls \
      --header "Subject: Test IP Change Notification" \
      --body "Your home IP has changed to 192.168.1.100"
```

## Deployment

### Systemd Service Installation

The application includes a systemd service file for production deployment:

```bash
# Install the service
sudo systemctl enable windmaker-home-ip-notifier.service
sudo systemctl start windmaker-home-ip-notifier.service

# Check service status
sudo systemctl status windmaker-home-ip-notifier.service

# View logs
sudo journalctl -u windmaker-home-ip-notifier.service -f
```

## License

This project is licensed under the GPL v3 License - see the [LICENSE](LICENSE) file for details.

## Authors

- **Álvaro Castellano Vela** - _Initial work_ - [a-castellano](https://git.windmaker.net/a-castellano)

## Related Projects

- [home-ip-monitor](https://git.windmaker.net/a-castellano/home-ip-monitor) - The IP monitoring service that feeds this notifier
- [go-services](https://git.windmaker.net/a-castellano/go-services) - Shared Go service utilities
- [go-types](https://git.windmaker.net/a-castellano/go-types) - Shared Go type definitions
