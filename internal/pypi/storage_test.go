package pypi

import "testing"

const sampleJSON = `{
  "info": {"name": "demo-pkg", "version": "0.2.0"},
  "releases": {
    "0.1.0": [
      {"filename": "demo-0.1.0-py3-none-any.whl", "size": 1000},
      {"filename": "demo-0.1.0.tar.gz", "size": 500}
    ],
    "0.2.0": [
      {"filename": "demo-0.2.0-py3-none-any.whl", "size": 3000}
    ],
    "0.0.1": []
  }
}`

func TestSummarizeTotals(t *testing.T) {
	ps, err := SummarizeBytes([]byte(sampleJSON))
	if err != nil {
		t.Fatalf("SummarizeBytes: %v", err)
	}
	if ps.Name != "demo-pkg" {
		t.Errorf("Name = %q, want demo-pkg", ps.Name)
	}
	if ps.TotalSize != 4500 {
		t.Errorf("TotalSize = %d, want 4500", ps.TotalSize)
	}
	if ps.ReleaseCount != 3 {
		t.Errorf("ReleaseCount = %d, want 3", ps.ReleaseCount)
	}
	if ps.FileCount != 3 {
		t.Errorf("FileCount = %d, want 3", ps.FileCount)
	}
}

func TestPerVersionBreakdown(t *testing.T) {
	ps, _ := SummarizeBytes([]byte(sampleJSON))
	got := map[string]VersionStorage{}
	for _, v := range ps.Versions {
		got[v.Version] = v
	}
	if got["0.1.0"].Size != 1500 || got["0.1.0"].FileCount != 2 {
		t.Errorf("0.1.0 = %+v, want size 1500 files 2", got["0.1.0"])
	}
	if got["0.2.0"].Size != 3000 {
		t.Errorf("0.2.0 size = %d, want 3000", got["0.2.0"].Size)
	}
	if got["0.0.1"].Size != 0 {
		t.Errorf("0.0.1 size = %d, want 0", got["0.0.1"].Size)
	}
}

func TestVersionsSortedDescending(t *testing.T) {
	ps, _ := SummarizeBytes([]byte(sampleJSON))
	if ps.Versions[0].Version != "0.2.0" {
		t.Errorf("first version = %q, want 0.2.0 (newest)", ps.Versions[0].Version)
	}
}

func TestUnparseableVersionDoesNotCrash(t *testing.T) {
	raw := `{"info":{"name":"weird","version":"x"},"releases":{"not-a-version":[{"filename":"x","size":10}]}}`
	ps, err := SummarizeBytes([]byte(raw))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ps.TotalSize != 10 || ps.ReleaseCount != 1 {
		t.Errorf("got total=%d releases=%d, want 10/1", ps.TotalSize, ps.ReleaseCount)
	}
}

func TestMissingSizeTreatedAsZero(t *testing.T) {
	raw := `{"info":{"name":"nosize","version":"1.0"},"releases":{"1.0":[{"filename":"x.whl"}]}}`
	ps, _ := SummarizeBytes([]byte(raw))
	if ps.TotalSize != 0 {
		t.Errorf("TotalSize = %d, want 0", ps.TotalSize)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int // sign
	}{
		{"1.0.0", "0.9.9", 1},
		{"0.2.0", "0.10.0", -1}, // numeric, not lexical: 2 < 10
		{"1.2.3", "1.2.3", 0},
		{"2.0", "1.9.9", 1},
	}
	for _, c := range cases {
		got := compareVersions(c.a, c.b)
		if sign(got) != c.want {
			t.Errorf("compareVersions(%q,%q) sign = %d, want %d", c.a, c.b, sign(got), c.want)
		}
	}
}

func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	default:
		return 0
	}
}
