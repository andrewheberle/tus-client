# tus-client

[![Go Report Card](https://goreportcard.com/badge/github.com/andrewheberle/tus-client?logo=go&style=flat-square)](https://goreportcard.com/report/github.com/andrewheberle/tus-client)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/andrewheberle/tus-client/pkg/sqlitestore)

A command-line tus client.

See https://tus.io for more information.

## Command Line Options

* `--chunksize`: Chunks size (in MB) for uploads
* `--db`: Path of database to allow resumable uploads
* `--disable-resume`: Disable the resumption of uploads (disables the use of the database)
* `-H`, `--header`: Extra HTTP header(s) to add to request (eg "Authorization: Bearer TOKEN"). Specify more than once to add multiple headers.
* `-h`, `--help`: help for tus-client
* `-i`, `--input`: File to upload
* `--no-progress`: Disable progress bar
* `-q`, `--quiet`:  Disable all output except for errors
* `--url`: tus upload URL
