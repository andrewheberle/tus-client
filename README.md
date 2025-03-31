# tus-client

[![Go Report Card](https://goreportcard.com/badge/github.com/andrewheberle/tus-client?logo=go&style=flat-square)](https://goreportcard.com/report/github.com/andrewheberle/tus-client)

A command-line tus client.

See https://tus.io for more information.

## Command Line Options

* `--chunksize`: Chunks size (in MB) for uploads
* `--disable-resume`: Disable the resumption of uploads (disables the use of the database)
* `-H`, `--header`: Extra HTTP header(s) to add to request (eg "Authorization: Bearer TOKEN"). Specify more than once to add multiple headers.
* `-h`, `--help`: help for tus-client
* `-i`, `--input`: File to upload
* `--no-progress`: Disable progress bar
* `-q`, `--quiet`:  Disable all output except for errors
* `--storepath`: Path of database or JSON file to allow resumable uploads
* `--url`: tus upload URL

## Resumption of Uploads

The tus protocol supports resuming uploads, which is implemented using the `tus.Store` interface.

This repo includes two implementations of this interface, one supporting SQLite and the the other writing to a JSON file.

Depending on the extension provided for the `--storepath` option, either the SQLite or JSON store type will be used.

### jsonstore

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/andrewheberle/tus-client/pkg/jsonstore)

This package is used to implement a JSON file-based version of the `tus.Store` interface.

### sqlitestore

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/andrewheberle/tus-client/pkg/sqlitestore)

This package is used to implement a SQLite version of the `tus.Store` interface.
