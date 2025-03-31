package iecbyte

import (
	"fmt"
	"testing"
)

type test struct {
	name    string
	f       *Flag
	value   string
	want    string
	wantErr bool
}

func tests() []test {
	return []test{
		// valid
		{"0", &Flag{0}, "0", "0", false},
		{"1", &Flag{1}, "1", "1", false},
		{"1023", &Flag{1023}, "1023", "1023", false},
		{"1024", &Flag{1024}, "1024", "1Ki", false},
		{"1Ki", &Flag{1024}, "1Ki", "1Ki", false},
		{"1025", &Flag{1025}, "1025", "1025", false},
		{"2048", &Flag{2048}, "2048", "2Ki", false},
		{"2Ki", &Flag{2048}, "2Ki", "2Ki", false},
		{"1048576", &Flag{1048576}, "1048576", "1Mi", false},
		{"1Mi", &Flag{1048576}, "1Mi", "1Mi", false},
		{"1049600", &Flag{1049600}, "1049600", "1025Ki", false},
		{"1025Ki", &Flag{1049600}, "1025Ki", "1025Ki", false},
		{"1050623", &Flag{1050623}, "1050623", "1050623", false},
		{"1050624", &Flag{1050624}, "1050624", "1026Ki", false},
		{"1026Ki", &Flag{1050624}, "1026Ki", "1026Ki", false},
		{"1073741824", &Flag{1073741824}, "1073741824", "1Mi", false},
		{"1Mi", &Flag{1073741824}, "1Mi", "1Mi", false},
		{"1073741825", &Flag{1073741825}, "1073741825", "1073741825", false},
		{"1Mi + 1Ki", &Flag{1073742848}, "1073742848", "1048577Ki", false},
		{"1Mi + 1Ki", &Flag{1073742848}, "1048577Ki", "1048577Ki", false},

		// invalid
		{"-1", &Flag{}, "-1", "", true},
		{"1a", &Flag{}, "1a", "", true},
		{"a", &Flag{}, "a", "", true},
		{"1024mi", &Flag{}, "1024mi", "", true},
		{"", &Flag{}, "", "", true},
		{"Mi", &Flag{}, "Mi", "", true},
	}
}

func TestFlag_Set(t *testing.T) {
	fmt.Printf("%s\n", &Flag{1073742848})
	for _, tt := range tests() {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.f.Set(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("Flag.Set() error = %v, wantErr %v", err, tt.wantErr)
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
				t.Errorf("Flag.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
