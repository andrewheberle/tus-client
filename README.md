# tus-client

[![Go Report Card](https://goreportcard.com/badge/github.com/andrewheberle/tus-client?logo=go&style=flat-square)](https://goreportcard.com/report/github.com/andrewheberle/tus-client)

A command-line tus client.

See https://tus.io for more information.

## Command Line Options

* `--chunksize`: Chunks size for uploads (can be specified in bytes or with "Ki", "Mi", "Gi", "Ti", "Pi" or "Ei" suffix)
* `--disable-resume`: Disable the resumption of uploads (disables the use of the store)
* `-H`, `--header`: Extra HTTP header(s) to add to request (eg "Authorization: Bearer TOKEN"). Specify more than once to add multiple headers.
* `-h`, `--help`: help for tus-client
* `-i`, `--input`: File to upload
* `--no-progress`: Disable progress bar
* `-q`, `--quiet`:  Disable all output except for errors
* `--storepath`: Path of database or JSON file to allow resumable uploads (defaults to using SQLite store in a user specific OS dependent location)
* `--storetype` : Type of store (one of `none`, `auto`, `bolt`, `json` or `sqlite`. When `auto` is used the extension of file in ``storepath` is used to set the type)
* `--url`: tus upload URL

## Resumption of Uploads

The tus protocol supports resuming uploads, which is implemented using the `tus.Store` interface.

This repo includes three implementations of this interface, one supporting Bolt, SQLite and the the other writing to a JSON file.

Depending on the extension provided for the `--storepath` option, either the Bolt, SQLite or JSON store type will be used as follows:

* `.bdb`: Bolt
* `.db`: SQLite
* `.json`: JSON

The default is to use the `sqlitestore.Store` implemention below, with the default path to the database dependent on the OS in question:

| OS      | Location                                                                                   |
|---------|--------------------------------------------------------------------------------------------|
| Linux   | `$XDG_CONFIG_HOME/tus-client/resume.db` or `$HOME/.config/tus-client/resume.db`            |
| Windows | `%APPDATA%\tus-client\resume.db` or `C:\Users\%USER%\AppData\Roaming\tus-client\resume.db` |
| macOS   | `$HOME/Library/Application Support/tus-client/resume.db`                                   |

### boltstore

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/andrewheberle/tus-client/pkg/boltstore)

This package is used to implement a [bbolt](https://github.com/etcd-io/bbolt) version of the `tus.Store` interface.

### jsonstore

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/andrewheberle/tus-client/pkg/jsonstore)

This package is used to implement a JSON file-based version of the `tus.Store` interface.

Due to the simplisitic nature of this implementation, it is not safe to have multiple users of the same JSON file as a `tus.Store`.

### sqlitestore

[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/andrewheberle/tus-client/pkg/sqlitestore)

This package is used to implement a SQLite version of the `tus.Store` interface.

## Authentication

The tus protocol does not specify any Authentication/Authorization scheme, so this is up to the server implementation in question.

This client supports the addition of arbritrary HTTP headers to the request which allows various authentication options, as follows:

```sh
tus-client -H "Remote-User-Name: user" -H "Remote-User-Secret: SECRET-API-TOKEN" --input upload.mp4 --url https://some.tus.server/url
```

Or using "bearer" authentication:

```sh
tus-client -H "Authorization: Bearer SECRET-API-TOKEN" --input upload.mp4 --url https://some.tus.server/url
```
