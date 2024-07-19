# Home IP Notifier

This program is suscribed to [home-ip-monitor](https://git.windmaker.net/a-castellano/home-ip-monitor) notify queue, it will update required DNS record with reases IP's from queue.

# What this progam does?

Reads IP's from configured queue and updated reqired DNS record.

# Required variables

## Queue names

**UPDATE_QUEUE_NAME**: Queue name where new IP's will be sended.

## AWS Config

DNS service is hosted in AWS Route53 service, AWS auth and Route53 keys are required:

**AWS_ACCESS_KEY_ID**: AWS account key
**AWS_SECRET_ACCESS_KEY**: AWS account secret key
**AWS_REGION**: AWS region, default value is "us-west-2"
**AWS_ZONE_ID**: Route53 zone ID
**SUBDOMAIN**: Subdomain record to update

In this case, an user with limited permissions is created with the following policy:
```
"Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": [
                "route53:ChangeResourceRecordSets",
                "route53:ListResourceRecordSets"
            ],
            "Resource": "arn:aws:route53:::hostedzone/AWS_ZONE_ID"
        }
    ]
}
```

## RabbitMQ Config

RabbitMQ required config can be found in its [go types](https://git.windmaker.net/a-castellano/go-types/-/tree/master/rabbitmq?ref_type=heads) Readme.
