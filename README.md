# goboot

*WARNING: This is not intended for public use, it's just a collection of app bootstrapping services that I use across a few app. Please be warned: things can suddenly break without warning. There is no versioning either so best to clone this and pick whatever you need.*

The functions and services in this package do not provide an abstraction from underlying libraries but only aid in bootstrapping them.

The code makes liberal use of `panic` during service bootstrap to avoid running a broken server.

