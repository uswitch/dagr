<img src="http://upload.wikimedia.org/wikipedia/commons/7/7d/Dagr_by_Arbo.jpg" alt="Dagr by Arbo" width="400px">

dagr
====

runs programs every day (in Norse mythology, Dagr is day personified)

## Prerequisites

### Build time prerequisites

* go
* zip

### Run time prerequisites

* git

## Build

    $ make deps
    $ export PATH=$GOPATH/bin:$PATH
    $ make

## Run

    $ dagr --http :8080 --repo git@github.com:company/dagr-programs --work /tmp/dagr-work
