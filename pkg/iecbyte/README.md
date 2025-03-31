# iecbyte [![Go Report Card](https://goreportcard.com/badge/github.com/andrewheberle/tus-client/pkg/iecbyte?logo=go&style=flat-square)](https://goreportcard.com/report/github.com/andrewheberle/tus-client/pkg/iecbyte) [![GoDoc](https://godoc.org/github.com/andrewheberle/tus-client/pkg/iecbyte?status.svg)](http://godoc.org/github.com/andrewheberle/tus-client/pkg/iecbyte)

A package that can be used as a custom flag type for `flag` and `github.com/sp13/pflag`.

## Example

```go
package main

import (
    "flag"

    "github.com/andrewheberle/tus-client/pkg/iecbyte"
)

func main() {
	size := iecbyte.NewFlag(1024 * 1024)

	flag.Var(&size, "size", "Size in IEC bytes")
	flag.Parse()

	fmt.Printf("Size is %s\n", size)
	// Output: Size is 1Mi
}
```