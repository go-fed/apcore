# apcore

*Under Construction*

`go get github.com/go-fed/apcore`

apcore is a powerful single server
[ActivityPub](https://www.w3.org/TR/activitypub)
framework for performant Fediverse applications.

It is built on top of the
[go-fed/activity](https://github.com/go-fed/activity)
suite of libraries, which means it can readily allow application developers to
iterate and leverage new
[ActivityStreams](https://www.w3.org/TR/activitystreams-core)
or RDF vocabularies.

## Features

*This list is a work in progress.*

* Uses `go-fed/activity`
  * ActivityPub S2S (Server-to-Server) Protocol supported
  * ActivityPub C2S (Client-to-Server) Protocol supported
  * Both S2S and C2S can be used at the same time
  * Comes with the Core & Extended ActivityStreams types
  * Readily expands to support new ActivityStreams types and/or RDF vocabularies
* Federation & Moderation Policy System
  * Administrators and/or users can create policies to customize their federation experience
  * Auditable results of applying policies on incoming federated data
* Supports common out-of-the-box command-line commands for:
  * Initializing a database with the appropriate `apcore` tables as well as your application-specific tables
  * Initializing a new administrator account
  * Creating a server configuration file in a guided flow
  * Comprehensive help command
  * Guided command line flow for administrators for all the above tasks, featuring Clarke the Cow
* Configuration file support
  * Add your configuration options to the existing `apcore` configuration options
  * Administrators can customize their ActivityPub and your app's experience
* Database support
  * Currently, only PostgreSQL supported
  * Others can be added with a some SQL work, in the future
  * No ORM overhead
  * Your custom application has access to `apcore` tables, and more
* OAuth2 support
  * Easy API to build authorization grant and validation flows
  * Handles server side state for you
* Webfinger & Host-Meta support

## How To Use This Framework

*This guide is a work in progress.*

Building an application is not an easy thing to do, but following these steps
reduces the cost of building a *federated* application:

0. Implement the `apcore.Application` interface.
0. Call `apcore.Run` with your implementation in `main`.

The most work is in the first step, as your application logic is able to live as
functional closures as the `Application` is used within the `apcore` framework.
See the documentation on the `Application` interface for specific details.
