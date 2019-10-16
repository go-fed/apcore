# apcore example

This shows a very bare-bones server that has both S2S (Federation API) and C2S
(Social API) enabled using the `apcore` framework.

## Requirements

* A platform that can build `go-fed/activity` and `go-fed/apcore`.
* A local Postgres server.

## Guide

Once the example has been obtained, you will go through setting up the example
like any administrator would when using the apcore framework. This lets all
administrators benefit from a common workflow; improvements to this workflow are
then shared across all applications, including this example one.

The steps are:

0. Configuring the software (`configure`)
0. Initializing the database tables (`init-db`)
0. Creating an administrator's account (`init-admin`) 
0. Launching the server (`serve`)

### Building the Binary

Build the example's binary:

`go install github.com/go-fed/apcore/example`

This binary will be referred to as `example`.

### Preparing

Note: These steps may require you to have a certificate (and key) to run a local
server that serves `https` connections. A self-signed certificate is sufficient
for development and non-public purposes. It can be generated using `openssl`:

`openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365`

If you have a domain name and certificate for that domain name, you can run this
example application using that domain name to test federation with other
instances.

This application is meant to be a proof-of-concept, exmaple, or a base for other
applications, and *not* a permanent production server. **To enforce this, it will
terminate itself after being up for one hour.**

#### Configuring

First, we must configure the software. This configuration can be changed later,
but to take effect requires restarting the server. It mainly sets parameters
for tuning resource usage, but applications can include other specific
configuration parameters.

To launch the guided flow, run:

`example configure`

which will write a `config.ini` file.

#### Database Initialization

Second, the configuration will be used to create the tables in a database. It
requires connecting to the database.

To do so, run:

`example init-db`

which will run `CREATE TABLE IF NOT EXISTS` commands.

#### Administrator Account Creation

Third, the first account needs to be created: an administrator account! It will
be done in a guided flow, and requires connecting to the database.

To launch the guided flow, run:

`example init-admin`

which will run several queries to create the administrator user.

### Launching The Example App Server

The time has come! To launch the server, run the following command:

`example serve`

which can be exited with `ctl+c`.

Once the server is up, it can be visited by going to `https://localhost/`.

## Features

TODO
