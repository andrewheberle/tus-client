# tus-client

[![Go Report Card](https://goreportcard.com/badge/github.com/andrewheberle/tus-client?logo=go&style=flat-square)](https://goreportcard.com/report/github.com/andrewheberle/tus-client)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/andrewheberle/tus-client/pkg/sqlitestore)

A command-line tus client.

See https://tus.io for more information.

## Command Line Options

* `--chunksize`: Chunks size (in MB) for uploads
* `--db`: Path of database to allow resumable uploads
* `--disable-resume`: Disable the resumption of uploads (disables the use of the database)
* `-h`, `--help`: help for tus-client
* `-i`, `--input`: File to upload
* `--no-progress`: Disable progress bar
* `-q`, `--quiet`:  Disable all output except for errors
* `--token`: Authorization Bearer token
* `--url`: tus upload URL
