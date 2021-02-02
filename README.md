DANGER
======

DO NOT USE THIS CODE YET. I found a [serious security vulnerability](https://github.com/yookoala/gofast/issues/60)
during my vetting process of a dependency of this project. For now I have pulled my built images from Docker Hub.
I will remove this warning and re-publish once the issue is resolved.


![Build status](https://github.com/joonas-fi/baikal-dockerimage/workflows/Build/badge.svg)
[![DockerHub](https://img.shields.io/docker/pulls/joonas/baikal.svg?style=for-the-badge)](https://hub.docker.com/r/joonas/baikal/)

Docker image for [Baïkal](https://github.com/sabre-io/Baikal/), with an eye for security:

- Uses a memory safe HTTP server (other images seem to be using Apache / nginx)
- Runs as non-root user
- I have [vetted the security](https://github.com/yookoala/gofast/issues/60) of most important dependencies

([Baïkal's own Docker image is in planning stages](https://sabre.io/baikal/docker-ready/).)


How to run
----------

```console
$ docker run --rm -it \
	-v "/home/joonas/baikal-db:/data" \
	--user "$(id -u)" \
	--label traefik.frontend.rule=Host:baikal.example.com \
	joonas/baikal:SEE_TAG_FROM_DOCKERHUB
```

It is assumed that you're using TLS termination proxy in front of Baïkal like
[Traefik](https://github.com/traefik/traefik) or [Edgerouter](https://github.com/function61/edgerouter)
(has traefik-compatible labels for HTTP ingress). Use of HTTPS is currently hardcoded into our headers.


Setting up
----------

There's still interactive setup to do (I didn't bother skipping setup via specifying container vars
because this is done only once)..

Notes:

- In Baïkal Settings set server time zone to UTC (inside container the time is UTC unless you explicitly
  defined something else via Docker command line).

- I suggest you set "Email invite sender address" as empty, because for now we don't support email

- I don't know what all "WebDAV authentication type" affects, but I set it to `Basic` and everything
  works for me. UPDATE: I tested with digest and it affects all clients so best set it to Basic!

- For "Baïkal Database Settings" the defaults work fine. (The container doesn't support MySQL. SQLite is fine.)


TODO
----

Calendar invitation / any emailing things are not set up! I'm pretty sure any of it doesn't work.

What works is:

- Baïkal's web UI
- syncing to Home Assistant
- syncing to Android app


State, backups
--------------

All state is in what you mount to `/data` inside the container. The state looks like this:

```console
$ tree /home/joonas/baikal-db
/home/joonas/baikal-db
├── INSTALL_DISABLED
├── baikal.yaml
└── db
    └── db.sqlite

1 directory, 3 files
```

You should backup this file tree. If you use [µbackup](https://github.com/function61/ubackup), you can
specify `BACKUP_COMMAND=tar -cC /data -f - .`


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
