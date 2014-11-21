dagr
====

runs programs every day (in Norse mythology, Dagr is day personified)

## Prerequisites

### Build time prerequisites

* go
* zip
* nrsc

To install nrsc (packaging tool):

    $ go get -u bitbucket.org/tebeka/nrsc/nrsc

### Run time prerequisites

* git

## Build

    $ make

## Install

    $ cp dagr somewhere

## Run

    $ dagr --port :8080 --repo git@github.com:company/dagr-programs --work /tmp/dagr-work
