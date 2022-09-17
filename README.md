# goboot

**WARNING: This repo not intended for public use and has no versioning applied. So be warned: things can break between commits. This project will have long periods of non-activity with short bursts of high activity based on my schedule. Meaning: don't count on my support.**

`goboot` instantiates an application environment for web services. Its main purpose is to create an opinionated base for REST and gRPC apps.

Goals:

- Easy and consistent service bootstrapping of common services (databases, queues, etc).
- Panic if bootstrapping a service failed. We don't want to start a broken server.
- Good logging and error reporting while bootstrapping. Debugging broken boot procedures on infra can be a pain...
- Avoid higher-level dependencies in `goboot` such as web frameworks or routers.

Non-goals:

- The services in this package are not an abstraction of underlying libraries but only aid in bootstrapping or simplify using them.
- No need for flexibility of underlying drivers, being tied to one specific version of a lib and/or datastore is OK.

Given these goals & non-goals you'll find this codebase is strongly tied to:

- [Viper](https://github.com/spf13/viper) for configuration management;
- [Zerolog](https://github.com/rs/zerolog) for logging;
- all packages (elasticsearch, postgres, pubsub, redis) depend on libraries and may only work for a specific version of db/protocol.

It's quite likely the set of chosen libraries here would not fit your project's needs or personal preferences.

## Development

This codebase contains integration tests that depend on real databases.

Ensure databases are running:

```bash
$ docker compose up
```

And in a different tab:

```bash
$ make test

# or if you have richgo installed run:
$ make humantest
```

This requires [richgo](https://github.com/kyoh86/richgo) to be installed.

See the project's `Makefile` for other (more standard) commands.