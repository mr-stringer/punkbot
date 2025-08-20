# Creating a configuration file

Punkbot supports both yaml and json configuration files. By default, punkbot
will search the current directory for either `botcnf.yml` or `botcnf.json`.

The config consists of three elements

| Element | Type | Required | Description |
|:--------|:-----|:---------|:------------|
| Identifier | string | yes | The name of the bluesky account to post from |
| Terms | array of strings | yes | The terms the bot will search for |
| JetStreamServer | string | no | The name or IP address of the JetStream server to use |

The identifier can be a bluesky handle such as `@iampunkbot.bsky.social`.
However, it is possible to change handles and if this happens the bot will no
longer be able to authenticate. Therefore its recommended to use the account's
DID, which is a long-term identifier. An easy way to find a DID for an account
is with [this service](https://rmdes.github.io/).

The terms part is a simple a list of strings that the bot should search posts
for. This may be hashtags but can be any string. Don't worry about case, the
term text and post text are both converted to lowercase for comparison.

 **Be careful**, using terms that are too common will likely cause you problems.
Ensure you read [BlueSky's rate limit
policy](https://docs.bsky.app/docs/advanced-guides/rate-limits). Don't become a
spammer!

The JetStreamServer argument is optional. It can be used if you have your own
instance of JetStream or if you want to connect to a specific JetStream
instance. However, if don't know what a JetSteams server does and just want the
bot to work, simply leave this option out and the bot will select a public
JetStream server with the lowest latency.

Below, is a simple  YAML configuration for a bot.

```yaml
---
Identifier: iamapunkbot@bsky.app
Terms:
  - "#runningpunks"
  - "punkbot"
```

The above configuration will post as the account with the handle
`iamapunkbot@bsky.app` and will re-post and like any content that includes
"#runningpunks" or "punkbot". No JetStream server was present in the
configuration, so punkbot will connect to a public instance.

A JSON version of the same configuration can be found [here](botcnf.json)