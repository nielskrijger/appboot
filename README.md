# goboot

*WARNING: This is not intended for public use, things can suddenly break without warning. There is no versioning at the moment either so best to clone this and pick whatever you need.*

The functions and services in this package do not provide an abstraction from underlying libraries but only aid in bootstrapping them.

We make liberal use of `panic` during service bootstrap to avoid running a broken web server.