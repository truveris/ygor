# ygor - IRC-controlled TV broadcast

Ygor provides access to devices such as TVs from our internal IRC channel and
mobile phones, it makes it easy to share a picture or sound with everyone in a
physical room or channel.

**ygord** is central part of the system it takes instructions from various
sources (IRC, web/API) and sends commands to its clients.  The clients are web
browsers

## Requirements
You need:

 * Go 1.2+ to compile it
 * All the dependencies downloaded:
    - go get github.com/jessevdk/go-flags
    - go get github.com/mikedewar/aws4
    - go get github.com/tamentis/go-mplayer
    - go get github.com/truveris/ygor
    - go get github.com/truveris/sqs
    - go get github.com/truveris/sqs/sqschan

## Installation
You can run ygord and ygor-minion however you want, we use supervisor, you
could just run it from tmux or build an init script. At @truveris, we copy all
the binaries to /usr/bin and run them with supervisor. Here is an example
supervisor configuration:

```ini
[program:ygor-minion]
user=ygor
directory=/tmp
command=ygor-minion
```

By default, both program look for their configuration in /etc.


## Why use SQS?
Using SQS means we don't have to worry about having all the components on the
same network, we only need to make sure they have access to SQS. Additionally,
SQS allows us to manage auth easily from IAM.

ygord could in theory accept TCP connections and provide authentication to all
the minions and IRC feeds, patches/forks are welcome.


## sample ygord policy
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Stmt01",
      "Effect": "Allow",
      "Action": [ "sqs:CreateQueue", "sqs:DeleteMessage", "sqs:ReceiveMessage" ],
      "Resource": [ "arn:aws:sqs:us-east-1:000000000000:ygor-irc-incoming" ]
    },
    {
      "Sid": "Stmt02",
      "Effect": "Allow",
      "Action": [ "sqs:CreateQueue", "sqs:SendMessage" ],
      "Resource": [ "arn:aws:sqs:us-east-1:000000000000:ygor-irc-outgoing" ]
    },
    {
      "Sid": "Stmt03",
      "Effect": "Allow",
      "Action": [ "sqs:CreateQueue", "sqs:DeleteMessage", "sqs:DeleteQueue", "sqs:ReceiveMessage" ],
      "Resource": [ "arn:aws:sqs:us-east-1:000000000000:ygord" ]
    },
    {
      "Sid": "Stmt04",
      "Effect": "Allow",
      "Action": [ "sqs:CreateQueue", "sqs:SendMessage" ],
      "Resource": [ "arn:aws:sqs:us-east-1:000000000000:ygor-minion-*" ]
    }
  ]
}
```


## sample ygor-minion policy
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Stmt01",
      "Effect": "Allow",
      "Action": [ "sqs:SendMessage", "sqs:GetQueueUrl" ],
      "Resource": [ "arn:aws:sqs:us-east-1:000000000000:ygord" ]
    },
    {
      "Sid": "Stmt02",
      "Effect": "Allow",
      "Action": [ "sqs:CreateQueue", "sqs:DeleteMessage", "sqs:DeleteQueue", "sqs:ReceiveMessage" ],
      "Resource": [ "arn:aws:sqs:us-east-1:000000000000:ygor-minion-yourhost" ]
    }
  ]
}
```
