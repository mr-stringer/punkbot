# punkbot 🤖

<img src="docs/punkbot.png" alt="Sid the punkbot, hard at work" style="width:500px;">

A simple bot for liking and re-posting posts on bluesky. Written in go-lang.

This bot is used by the account
[@iampunkbot](https://bsky.app/profile/iampunkbot.bsky.social) on bluesky.

## Concept

Punkbot is used re-post and like original posts (not replies to posts) that
match a one or more defined strings or hashtags.

In order to do this the bot connects to
[JetStream](https://docs.bsky.app/blog/jetstream), Bluesky's JSON firehose.
The bot scans all original posts and when a match if found it likes and re-posts
it.

The bot doesn't currently persist any data from Bluesky, but may record stats in
the future.

## Getting punkbot

The best way to run punkbot is with Docker. For other methods of running
punkbot, [click here](docs/running.md)

The punkbot docker image supports amd64 and arm64 architectures. You can pull
the image with:

```shell
docker pull mrstringer/punkbot
```

## Running punkbot

In order to run `punkbot`, you'll need:

* A Bluesky account
* An app password for your bluesky account
* A completed configuration file

### Create a Bluesky account

Head to [bsky.app](https://bsky.app/) and create a new account.

### Create an app password for your account

A simple guide for creating an app password can be found [here](docs/apppass.md)

### Make a configuration file

Follow [this guide](docs/config.md) to learn how to write a punkbot config file.

### Running punkbot in docker

With the docker image downloaded, the configuration file prepared and the
application password created, punkbot can now be run.

The config files needs to be mounted in the container and the password needs to
be passed via a command line variable. Due to the design of punkbot, it is
recommended to run with the docker restart option 'unless-stopped'. The command
to run is as follows:

```shell
docker container run -d --restart=unless-stopped -v <YOUR CONFIG FILE>:/app/botcnf.yml -e PUNKBOT_PASSWORD='<YOUR APP PASSWORD>' mrstringer/punkbot
```

With the above command, punkbot will run until stopped and it's log output can
be monitored with command `docker logs <container id>`

### Command line arguments

Logging level, logging location and some other options are controlled by passing
specific command line arguments. These are documented [here](docs/cmd.md)

## Feedback

I hope you find this project useful. Report any issues [the usual
way](https://github.com/mr-stringer/punkbot/issues). 