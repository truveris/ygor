# ygor - Truveris' Go Office Butler

Ygor provides access to devices such as TVs from our internal IRC channel and
mobile phones, it makes it easy to share a picture or sound with everyone and
permits a large amount of trolling.

All the process in this system communicate via AWS SQS. Here is the break down
of the processes:

  ygord: central part of the system it takes instructions from various sources
         (IRC, web/API) and sends commands to its minions accordingly. It keeps
	 a registration of all of them and knows where each minion belong (e.g.
	 which room and IRC channel).

  ygorlet: process receiving commands from ygord, through SQS. We run one
	 ygorlet per controlled device and uses raspberry pis for that purpose.
	 You can then connect this little machine to a TV, speakers, large
	 monitor, alarm system, etc. It can drive the playback of audio, video
	 via mplayer/omxplayer and use the local "say" or "espeak" to
	 communicate with humans around.

ygord does not connect to IRC by itself, since it could become a rather large
part of the system, it could crash and restart the bot. Instead ygord receive
all its IRC traffic through SQS as well, you can use the following project to
route your IRC traffic properly:

	https://github.com/truveris/sqs-irc-gateway

This allows you to modify the personality/setting of the bot and restart it
without dissconnecting/reconnecting your bot.


## How to start/restart
We run ygor via supervisord, just copy our example.conf files and start your
processes with:

```sh
ygord -c /etc/ygord.conf
ygorlet -c /etc/ygorlet.conf
```


## How to restart it
If you use supervisord, you can just run the following command:

```sh
sudo supervisorctl restart ygord
# -or-
sudo supervisorctl restart ygorlet
```


## Why use SQS?
Using SQS means we don't have to worry about having all the components on the
same network, we only need to make sure they have access to SQS. Additionally,
SQS allows us to manage auth easily from IAM.

ygord could in theory accept TCP connections and provide authentication to all
the ygorlets and IRC feeds, patches/forks are welcome.


## Why use such a plain-text line-based protocol instead of XML, JSON, you name it?
It's great to be able to debug your minion and control it from the command
line, for example:

    ygorlet -i
    > play stuff
    *plays stuff*


## Soul <-> Minion(s) Communication
The soul and the minions communicate with each others via SQS. On the soul
side, everything from IRC and from minion(s) goes through the inputQueue
channel. All the minion messages start with MINIONMSG.

A command that expects responses from the minions should define a non-nil
MinionMsgFunction. The MinionMsgs will be routed to it upon receival based
on the command name. Here is an example:

    ygord->ygorlet   play tunes/something.ogg
    ygorlet->ygord   play error file not found


## ygord policy
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
      "Resource": [ "arn:aws:sqs:us-east-1:000000000000:ygorlet-*" ]
    }
  ]
}
```


## ygorlet policy
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
      "Resource": [ "arn:aws:sqs:us-east-1:000000000000:ygorlet-yourhost" ]
    }
  ]
}
```
