# Home IP Notifier

[![pipeline status](https://git.windmaker.net/a-castellano/home-ip-notifier/badges/master/pipeline.svg)](https://git.windmaker.net/a-castellano/home-ip-notifier/pipelines)[![coverage report](https://git.windmaker.net/a-castellano/home-ip-notifier/badges/master/coverage.svg)](https://a-castellano.gitpages.windmaker.net/home-ip-notifier/coverage.html)[![Quality Gate Status](https://sonarqube.windmaker.net/api/project_badges/measure?project=a-castellano_home-ip-notifier_c4da9a70-dcc5-4ef5-8425-3f91b0d7526d&metric=alert_status&token=sqb_efd83d3e4b6a20b336f469385f469e63fdab1fc3)](https://sonarqube.windmaker.net/dashboard?id=a-castellano_home-ip-notifier_c4da9a70-dcc5-4ef5-8425-3f91b0d7526d)

This program is suscribed to [home-ip-monitor](https://git.windmaker.net/a-castellano/home-ip-monitor) notify queue, it will notify about IP changes by e-mail.

# What this progam does?

Reads IP's from configured queue and nofies it by e-mail.

# Required variables

## Queue names

**NOTIFY_QUEUE_NAME**: Queue name where IP will be sended.

## Mail Config

The following mail config env variables are required:

- **MailFrom**
- **MailDomain**
- **SMTPHost**
- **SMTPPort**
- **SMTPName**
- **SMTPPassword**
- **Destination**
- **NotifyQueue**

## RabbitMQ Config

RabbitMQ required config can be found in its [go types](https://git.windmaker.net/a-castellano/go-types/-/tree/master/rabbitmq?ref_type=heads) Readme.
