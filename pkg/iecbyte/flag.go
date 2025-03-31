package iecbyte

import (
	"fmt"
	"strconv"
	"strings"
)

// Flag satisfies the pflag.Value interface
type Flag struct {
	n int64
}

type multiplier struct {
	Suffix string
	Value  int64
}

var multipliers = []multiplier{
	{"Ei", 1024 * 1024 * 1024 * 1024 * 1024 * 1024},
	{"Pi", 1024 * 1024 * 1024 * 1024 * 1024},
	{"Ti", 1024 * 1024 * 1024 * 1024},
	{"Gi", 1024 * 1024 * 1024},
	{"Mi", 1024 * 1024},
	{"Ki", 1024},
}

// NewFlag is used to initialise a new iecbyte.Flag with a default value
//
// A value of n that is < 0 will be set to 0
func NewFlag(n int64) Flag {
	if n < 0 {
		n = 0
	}
	return Flag{n}
}

func (f Flag) String() string {
	for _, m := range multipliers {
		if f.n >= m.Value && f.n%m.Value == 0 {
			return fmt.Sprintf("%d%s", f.n/m.Value, m.Suffix)
		}
	}

	return fmt.Sprintf("%d", f.n)
}

func (f *Flag) Set(value string) error {
	for _, m := range multipliers {
		if strings.HasSuffix(value, m.Suffix) {
			v := strings.TrimSuffix(value, m.Suffix)
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}

			if n < 0 {
				return fmt.Errorf("cannot be negative")
			}

			f.n = n * m.Value

			return nil
		}
	}

	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}

	if n < 0 {
		return fmt.Errorf("cannot be negative")
	}

	f.n = n

	return nil
}

func (f Flag) Type() string {
	return "bytes (IEC)"
}

func (f Flag) Get() int64 {
	return f.n
}
