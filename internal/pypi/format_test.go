package pypi

import "testing"

func TestFormatSize(t *testing.T) {
	cases := map[int64]string{
		0:                 "0 B",
		512:               "512 B",
		1024:              "1.0 KB",
		1536:              "1.5 KB",
		1 << 20:           "1.0 MB",
		10 * (1 << 30):    "10.0 GB",
	}
	for in, want := range cases {
		if got := FormatSize(in); got != want {
			t.Errorf("FormatSize(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestParseSize(t *testing.T) {
	cases := map[string]int64{
		"10GB":    10 * (1 << 30),
		"500 MB":  500 * (1 << 20),
		"1.5G":    int64(1.5 * float64(1<<30)),
		"1048576": 1 << 20,
		"2TB":     2 * (1 << 40),
	}
	for in, want := range cases {
		got, err := ParseSize(in)
		if err != nil {
			t.Errorf("ParseSize(%q) error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseSize(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestParseSizeInvalid(t *testing.T) {
	for _, in := range []string{"", "abc", "10XB"} {
		if _, err := ParseSize(in); err == nil {
			t.Errorf("ParseSize(%q) expected error, got nil", in)
		}
	}
}
