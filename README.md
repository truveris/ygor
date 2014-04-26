# ygor - Truveris' Go Office Butler

Ygor provides access to devices such as TVs from our internal IRC channel and
mobile phones, it makes it easy to share a picture or sound with everyone and
permits a large amount of trolling.

All the process in this system communicate via AWS SQS. Here is the break down
of the processes:

 * **ygord**: central part of the system it takes instructions from various
   sources (IRC, web/API) and sends commands to its minions accordingly. It
   keeps a registration of all of them and knows where each minion belong (e.g.
   which room and IRC channel).
 * **ygor-minion**: process receiving commands from ygord, through SQS. We run one
   minion per controlled device and uses raspberry pis for that purpose.  You
   can then connect this little machine to a TV, speakers, large monitor, alarm
   system, etc. It can drive the playback of audio, video via mplayer/omxplayer
   and use the local "say" or "espeak" to communicate with humans around.

ygord does not connect to IRC by itself, since it could become a rather large
part of the system, it could crash and restart the bot. Instead ygord receive
all its IRC traffic through SQS as well, you can use the following project to
route your IRC traffic properly:

	https://github.com/truveris/sqs-irc-gateway

This allows you to modify the personality/setting of the bot and restart it
without dissconnecting/reconnecting your bot.


## Requirements
You need:

 * Go 1.2 to compile it
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
