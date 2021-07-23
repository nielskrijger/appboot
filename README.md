# goboot

*WARNING: This is not intended for public use, nor is any versioning applied. So be warned: things can suddenly break without warning.*

`goboot` instantiates an application context for web services. Its main purpose is to create an opinionated base for REST and gRPC services which I can easily upgrade and improve upon over time.

Goals:

- Easy and consistent service bootstrapping of services.
- Panic when bootstrapping a service failed. We never to start a broken server.
- Good logging and error reporting while bootstrapping. Debugging failed bootstrapping processes can be a pain.
- Avoid higher-level dependencies in `goboot` such as web frameworks, routers, query-builder/ORM or similar.

Non-goals:

- The utils and services in this package are not an abstraction of underlying libraries but only aid in bootstrapping or simplify using them.
- No need for flexibility of underlying drivers, being tied to one specific version of a lib and/or datastore is OK.

Given these goals & non-goals you'll find this codebase is strongly tied to:

- [Viper](https://github.com/spf13/viper) for configuration management;
- [Zerolog](https://github.com/rs/zerolog) for logging;
- all packages (elasticsearch, grpc, postgres, pubsub, redis) depend on libraries and may only work for a specific version of db/protocol.

It is not very likely the set of chosen libraries here would fit your project's needs or preferences. It's designed to fit mine for the type of projects I'm currently working on.

## Development

The repo contains a combination of integration and unit tests.

To run all of them in a human-readable way use:

```bash
$ make humantest
```

This requires [richgo](https://github.com/kyoh86/richgo) to be installed.

See the project's `Makefile` for other (more standard) commands.