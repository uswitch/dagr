dagr
====

runs programs every day (in Norse mythology, Dagr is day personified)

## Prerequisites

### Build time prerequisites

* go
* zip
* rice

To install rice (packaging tool):

    $ go get github.com/GeertJohan/go.rice
    $ go get github.com/GeertJohan/go.rice/rice

### Run time prerequisites

* git

## Build

    $ make

## Install

    $ cp dagr somewhere

## Run

    $ dagr --port :8080 --repo git@github.com:company/dagr-programs --work /tmp/dagr-work
