package iecbyte_test

import (
	"flag"
	"fmt"

	"github.com/andrewheberle/tus-client/pkg/iecbyte"
)

func ExampleNewFlag() {
	size := iecbyte.NewFlag(1024 * 1024)

	flag.Var(&size, "size", "Size in IEC bytes")
	flag.Parse()

	fmt.Printf("Size is %s\n", size)
	// Output: Size is 1Mi
}
