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

var (
	multipliers = map[string]int64{
		"Ei": 1024 * 1024 * 1024 * 1024 * 1024 * 1024,
		"Pi": 1024 * 1024 * 1024 * 1024 * 1024,
		"Ti": 1024 * 1024 * 1024 * 1024,
		"Gi": 1024 * 1024 * 1024,
		"Mi": 1024 * 1024,
		"Ki": 1024,
	}
	order = []string{"Ei", "Pi", "Ti", "Gi", "Mi", "Ki"}
)

func NewFlag(n int64) Flag {
	return Flag{n}
}

func (f *Flag) String() string {
	for _, suffix := range order {
		m := multipliers[suffix]
		if f.n >= m && f.n%m == 0 {
			return fmt.Sprintf("%d%s", f.n/m, suffix)
		}
	}

	return fmt.Sprintf("%d", f.n)
}

func (f *Flag) Set(value string) error {
	for _, suffix := range order {
		if strings.HasSuffix(value, suffix) {
			v := strings.TrimSuffix(value, suffix)
			n, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}

			if n < 0 {
				return fmt.Errorf("cannot be negative")
			}

			f.n = n * multipliers[suffix]

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

func (f *Flag) Type() string {
	return "bytes (IEC)"
}

func (f *Flag) Get() int64 {
	return f.n
}
