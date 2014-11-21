dagr
====

runs programs every day (in Norse mythology, Dagr is day personified)

## Prerequisites

### Build time prerequisites

* go
* zip

### Run time prerequisites

* git

## Installation

Install dagr

    go get -u github.com/uswitch/dagr

Install resource packaging tool

    go get -u bitbucket.org/tebeka/nrsc/nrsc

Package dagr with its resources

    nrsc $GOPATH/bin/dagr $GOPATH/src/github.com/uswitch/dagr/resources

Now, $GOPATH/bin/dagr can be copied anywhere and run.
