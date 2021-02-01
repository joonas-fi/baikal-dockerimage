![Build status](https://github.com/joonas-fi/baikal-dockerimage/workflows/Build/badge.svg)
[![DockerHub](https://img.shields.io/docker/pulls/joonas/baikal.svg?style=for-the-badge)](https://hub.docker.com/r/joonas/baikal/)

Docker image for [Baïkal](https://github.com/sabre-io/Baikal/).

([Their own Docker image is in planning stages](https://sabre.io/baikal/docker-ready/).)


How to run
----------

```console
$ docker run --rm -it \
	-v "/tmp/baikal-db:/data" \
	--label traefik.frontend.rule=Host:baikal.example.com \
	joonas/baikal:SEE_TAG_FROM_DOCKERHUB
```

It is assumed that you're using TLS termination proxy in front of Baikal.


TODO
----

Calendar invitation / any emailing things are not set up! I'm pretty sure any of it doesn't work.

What works is:

- Baïkal's web UI
- syncing to Home Assistant
- syncing to Android app


State
-----

All state is in what you mount to `/data` inside the container. The state looks like this:

```console
$ sudo tree /tmp/baikal-db/
/tmp/baikal-db/
├── INSTALL_DISABLED
├── baikal.yaml
└── db
    └── db.sqlite

1 directory, 3 files
```


Adding to OneCalendar in Android
--------------------------------


Adding to Home Assistant
------------------------

https://www.home-assistant.io/integrations/caldav/

NOTE: despite URL looking like `https://baikal.my-server.net/cal.php/calendars/john.doe@test.com/default`
it's not actually by email, rather username and it's better to not include `/default` as that refers to only one calendar,
but actually use `https://baikal.my-server.net/cal.php/calendars/john.doe@test.com` - your calendars will be discovered!


Includes Home Assistant patch
-----------------------------

https://github.com/sabre-io/dav/issues/1318#issuecomment-757380175
