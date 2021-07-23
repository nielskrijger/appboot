# goboot

*WARNING: This is not intended for public use, nor is any versioning applied. So be warned: things can suddenly break without warning. There is no versioning since this is not catering for external users.*

`goboot` instantiates an application context for web services. Its main purpose is to create stable opinionated "base" for REST or gRPC apps that I can easily upgrade and improve upon across multiple apps.

Goals:

- Easy and consistent service bootstrapping of services.
- Panic when bootstrapping a service failed. We never want to run a broken server.
- Good logging and error reporting while bootstrapping. Debugging failed bootstrapping processes can be a pain otherwise.

Non-goals:

- The utils & services in this package should not be an abstraction from underlying libraries but only aid in bootstrapping them.
- There should be no loose-coupling with dependencies.

Given these goals & non-goals you'll find this codebase is strongly tied to:

- (Viper)[https://github.com/spf13/viper] for configuration management
- (Zerolog)[https://github.com/rs/zerolog] for logging 
- all packages (elasticsearch, grpc, postgres, pubsub, redis) depend on libraries and may only work for a specific version of db/protocol.

It is not very likely the set of chosen libraries here would fit your needs or preferences. It's designed to fit mine.

## development

The repo contains a combination of integration and unit tests.

To run all of them in a human-readable way run:

```bash
$ make humantest
```

This requires (richgo)[https://github.com/kyoh86/richgo] (see Makefile for details).

See the project's `Makefile` for other (more standard) commands.