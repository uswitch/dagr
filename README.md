<img src="http://upload.wikimedia.org/wikipedia/commons/7/7d/Dagr_by_Arbo.jpg" alt="Dagr by Arbo" width="400px">

Dagr
====

Runs programs every day (in Norse mythology, Dagr is day personified).

Dagr monitors a specified git repository for programs (directories
that contain a `main` executable) and ensures each of them is executed
every day (or at any other frequency expressible in the Cron
syntax). Program output (`stderr` and `stdout`) are captured and
showed on a monitoring page. A program's exit code can be used to
signal to Dagr whether a program succeeded (`0`) or failed (`2`), or
whether it should be retried after a delay (`1`).

## Running

    $ cd $GOPATH
    $ ./bin/dagr --http :8080 --repo git@github.com:uswitch/dagr-sample-programs --work /tmp/dagr-work --ui ./ui

### Configuration

The scheduling of dagr programs can be controlled by an optional
[TOML](http://github.com/toml-lang/toml) file called `dagr.toml` in
the same directory as `main`.

#### Scheduling programs

If `dagr.toml` contains a line in this format:

    schedule = "CRON EXPRESSION"

the Cron expression gives the schedule on which to run the program.

Valid Cron expressions are those accepted by the
[robfig/cron](https://godoc.org/github.com/robfig/cron) library.

e.g.

    schedule = "0 */5 * * * *"

would ensure the program would be run every five minutes.

If `dagr.toml` does not exist or the schedule line is not
present, the schedule defaults to `@daily` (i.e. every day at
midnight).

#### Running programs immediately on startup

If `dagr.toml` contains a line like this:

    immediate = BOOLEAN

the program will be run immediately when dagr starts (as well as its
usually scheduled times) if the boolean value is true.

i.e.

    immediate = true

In any other case, the program will *not* be run immediately (i.e. it
will run at scheduled times only)

#### Notify slack about job failures

Optional feature switched off by default. Can be configured with environment variables. To switch on set `DAGR_SLACK_WEBHOOK_URL` and `DAGR_SLACK_ALERTS_ROOM` to point to the slack incoming webhook and the slack room to be alerted. By default Dagr will notify the slack room for every failed job both with exit code 1 (retryable) or 2. By setting `DAGR_SLACK_SILENCE_RETRYABLE` slack will be only notifed if a job fails with exit code 2. The appearance of the message can be configured via setting `DAGR_SLACK_USER_NAME` (defaults to `Dagr`), `DAGR_SLACK_ICON_EMOJI` (defaults to `:exclamation:`) and `DAGR_SLACK_ICON_URL` (no default).

Naturally individual jobs can have their own warning logic, this is for an easy option to get basic monitoring on jobs for cheap.

### Examples
For examples please see our
[sample programs repository](https://github.com/uswitch/dagr-sample-programs).

### Dagr Dashboard
<img src="doc/dashboard.png" alt="Dagr dashboard" width="800px">

The dashboard provides an overview of which programs are available to run, their most recent status and three counters showing how many programs succeeded or failed.

### Execution Page
<img src="doc/execution.png" alt="Execution page" width="800px">

The execution page allows you to view `stderr` and `stdout` for a program- the state is updated via a websocket.

## Build

### Pre-requisites

* Go
* Zip
* Git

Dagr contains packages which are specified as `github.com/uswitch/dagr/foo` etc. When developing its helpful to ensure
you pull the code using `go get` and build using the provided `Makefile`.

    $ export PATH=$GOPATH/bin:$PATH
    $ cd $GOPATH
    $ go get github.com/uswitch/dagr
    $ make -C src/github.com/uswitch/dagr deps
    $ make -C src/github.com/uswitch/dagr

### Docker

The Dockerfile requires that a RSA key be present that has read-only access to the github repo to be watched.
The file should be named `id_rsa` and be copied into the same folder as the Dockerfile.

The command passed to the container would be the same as you would run from the command line.

Usage:

```bash
docker-compose build

#example usage
docker-compose ps
docker-compose run dagr --http :8080 --repo git@github.com:uswitch/dagr-sample-programs --work /tmp/dagr-work --ui ./ui
```
