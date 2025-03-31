package iecbyte

import (
	"fmt"
	"testing"
)

type test struct {
	name      string
	f         *Flag
	value     string
	want      string
	wantInt64 int64
	wantErr   bool
}

func tests() []test {
	return []test{
		// valid
		{"0", &Flag{0}, "0", "0", 0, false},
		{"1", &Flag{1}, "1", "1", 1, false},
		{"1023", &Flag{1023}, "1023", "1023", 1023, false},
		{"1024", &Flag{1024}, "1024", "1Ki", 1024, false},
		{"1Ki", &Flag{1024}, "1Ki", "1Ki", 1024, false},
		{"1025", &Flag{1025}, "1025", "1025", 1025, false},
		{"2048", &Flag{2048}, "2048", "2Ki", 2048, false},
		{"2Ki", &Flag{2048}, "2Ki", "2Ki", 2048, false},
		{"1048576", &Flag{1048576}, "1048576", "1Mi", 1048576, false},
		{"1Mi", &Flag{1048576}, "1Mi", "1Mi", 1048576, false},
		{"1049600", &Flag{1049600}, "1049600", "1025Ki", 1049600, false},
		{"1025Ki", &Flag{1049600}, "1025Ki", "1025Ki", 1049600, false},
		{"1050623", &Flag{1050623}, "1050623", "1050623", 1050623, false},
		{"1050624", &Flag{1050624}, "1050624", "1026Ki", 1050624, false},
		{"1026Ki", &Flag{1050624}, "1026Ki", "1026Ki", 1050624, false},
		{"1073741824", &Flag{1073741824}, "1073741824", "1Gi", 1073741824, false},
		{"1Gi", &Flag{1073741824}, "1Gi", "1Gi", 1073741824, false},
		{"1073741825", &Flag{1073741825}, "1073741825", "1073741825", 1073741825, false},
		{"1073742848", &Flag{1073742848}, "1073742848", "1048577Ki", 1073742848, false},
		{"1048577Ki", &Flag{1073742848}, "1048577Ki", "1048577Ki", 1073742848, false},

		// invalid
		{"-1", &Flag{}, "-1", "", 0, true},
		{"-1Ki", &Flag{}, "-1Ki", "", 0, true},
		{"1a", &Flag{}, "1a", "", 0, true},
		{"a", &Flag{}, "a", "", 0, true},
		{"1024mi", &Flag{}, "1024mi", "", 0, true},
		{"", &Flag{}, "", "", 0, true},
		{"Mi", &Flag{}, "Mi", "", 0, true},
	}
}

func TestFlag_Set(t *testing.T) {
	fmt.Printf("%s\n", &Flag{1073742848})
	for _, tt := range tests() {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.f.Set(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("Test: %s, Flag.Set() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestFlag_String(t *testing.T) {
	for _, tt := range tests() {
		if tt.wantErr {
			continue
		}

		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("Test: %s, Flag.String() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestFlag_Get(t *testing.T) {
	for _, tt := range tests() {
		if tt.wantErr {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Get(); got != tt.wantInt64 {
				t.Errorf("Test: %s, Flag.Get() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestFlag_Misc(t *testing.T) {
	f := NewFlag(100)
	t.Run("Test NewFlag()", func(t *testing.T) {
		if f.String() != "100" {
			t.Errorf("NewFlag().String() = %v, want %v", f.String(), "100")
		}
	})
	t.Run("Test Get()", func(t *testing.T) {
		if f.Get() != 100 {
			t.Errorf("NewFlag().Get() = %v, want %v", f.Get(), 100)
		}
	})
	t.Run("Test Type()", func(t *testing.T) {
		if f.Type() != "bytes (IEC)" {
			t.Errorf("NewFlag().Type() = %v, want %v", f.Type(), "bytes (IEC)")
		}
	})

	f = NewFlag(-100)
	t.Run("Test NewFlag()", func(t *testing.T) {
		if f.String() != "0" {
			t.Errorf("NewFlag().String() = %v, want %v", f.String(), "0")
		}
	})
	t.Run("Test Get()", func(t *testing.T) {
		if f.Get() != 0 {
			t.Errorf("NewFlag().Get() = %v, want %v", f.Get(), 0)
		}
	})
}
