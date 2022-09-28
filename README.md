# music-match
Music Match HTTP Backend


## Installation
1. [Docker](https://www.docker.com/products/docker-desktop/)
2. [GoLang](https://go.dev/doc/install)

## Getting started
We use Make to simplify all common operations throughout the software development lifecycle.

### Building Executable
    make build

### Unit Testing
Unit tests are the first line of defense against breaking changes when devloping features, bugfixes, etc. It is imperative that unit tests are added for every PR created against the repo. The following command will trigger golang unit tests to execute:

    make test


### E2E Testing
E2E testing is equally as important as unit testing as it ensures every module within this repo works as once cohesive piece of software. Follow these steps in order to run an instance of music-match which connects to a containerized Neo4j DB. The following command will run a neo4j instance locally with a music-match backend server:

    make deploy

### Neo4j
Neo4j exposes a web UI for interacting with the DB directly. You can access this at localhost:7474 once Neo4j container is up and running!