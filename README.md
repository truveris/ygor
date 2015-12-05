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
    - go get github.com/truveris/ygor

## Installation
You can run ygord however you want, we use supervisor, you could just run it
from tmux or build an init script. At @truveris, we copy these binaries to
/usr/bin and run them with supervisor. Here is an example supervisor
configuration:

```ini
[program:ygord]
user=ygor
directory=/tmp
command=ygord
```

By default, ygord look for its configuration in /etc/ygord.conf.
