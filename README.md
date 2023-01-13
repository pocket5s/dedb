# DeDB Domain Eventsource DataBase
Pronounced 'DeeDeeBee', this project is an event source style service, using a pluggable persistence layer
for the actual events. It uses gRPC as the communication protocol and a generic event definition which accepts
any payload for the domain events themselves. 

## Building and Testing
There are a few scripts in the bin directory for development purposes, they are:
* dev.sh launches a docker container for dev use
* shell.sh shells into the docker container
* shutdown.sh shuts down the docker container
* rcli.sh launches the redis-cli

Once the dev container is launced (dev.sh), shell into it (shell.sh) and from there you can use
the Makefile to generate the protos and run the unit tests. Note that the unit tests run against
a running redis instance, which is launched when the dev.sh script is run. 

### Makefile targets
* genproto - runs the protobuf compiler and generates the grpc code. 
* test - runs the unit tests via the gotest library. Installs the library if it doesn't find it.
* cover - runs the test coverage suite.
* build - build the executable

### Unit Tests
Unit testing is on this project is designed to be primary api based, so whenever possible
tests are run against the 'public' apis. However some aspects, such as the database implementations
have specific unit tests for them. The idea here is limit the amount of low level tests to avoid
fragile unit tests and facilitate refactoring as needed with minimal unit test changes.

