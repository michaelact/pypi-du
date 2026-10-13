// Package pypi computes PyPI project storage usage from the public JSON API.
//
// PyPI enforces a per-project total storage limit (default 10 GB) counting the
// size of every file across every release. This package sums exactly that, so
// you can see which projects are near the cap and which versions to prune.
package pypi

import (
	"encoding/json"
	"sort"
	"strings"
)

// ProjectJSON mirrors the subset of https://pypi.org/pypi/<project>/json we use.
type ProjectJSON struct {
	Info struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"info"`
	// releases maps a version string to the list of distribution files.
	Releases map[string][]ReleaseFile `json:"releases"`
}

// ReleaseFile is one uploaded artifact (wheel or sdist) within a release.
type ReleaseFile struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

// VersionStorage is the aggregated size of a single version.
type VersionStorage struct {
	Version   string `json:"version"`
	Size      int64  `json:"size"`
	FileCount int    `json:"file_count"`
}

// ProjectStorage is the full per-project storage summary.
type ProjectStorage struct {
	Name         string           `json:"name"`
	TotalSize    int64            `json:"total_size"`
	ReleaseCount int              `json:"release_count"`
	FileCount    int              `json:"file_count"`
	Versions     []VersionStorage `json:"versions"`
}

// Summarize sums the size of every file in every release of the project.
//
// It is deliberately total: yanked/empty releases count as zero but still
// appear, and files missing a size are treated as zero rather than failing.
func Summarize(p *ProjectJSON) *ProjectStorage {
	out := &ProjectStorage{Name: p.Info.Name}

	for version, files := range p.Releases {
		vs := VersionStorage{Version: version, FileCount: len(files)}
		for _, f := range files {
			vs.Size += f.Size
		}
		out.Versions = append(out.Versions, vs)
		out.TotalSize += vs.Size
		out.FileCount += vs.FileCount
	}
	out.ReleaseCount = len(p.Releases)

	sortVersionsDescending(out.Versions)
	return out
}

// SummarizeBytes parses raw PyPI JSON bytes and summarizes them.
func SummarizeBytes(raw []byte) (*ProjectStorage, error) {
	var pj ProjectJSON
	if err := json.Unmarshal(raw, &pj); err != nil {
		return nil, err
	}
	return Summarize(&pj), nil
}

// sortVersionsDescending orders versions newest-first using a best-effort
// PEP 440-ish comparison, falling back to string order for unparseable ones.
func sortVersionsDescending(vs []VersionStorage) {
	sort.SliceStable(vs, func(i, j int) bool {
		return compareVersions(vs[i].Version, vs[j].Version) > 0
	})
}

// compareVersions returns >0 if a is newer than b, <0 if older, 0 if equal.
// It compares dot-separated numeric segments; non-numeric or ragged inputs
// fall back to lexicographic comparison so nothing crashes.
func compareVersions(a, b string) int {
	as := splitNumeric(a)
	bs := splitNumeric(b)
	// If either failed to parse as numeric segments, use string order.
	if as == nil || bs == nil {
		return strings.Compare(a, b)
	}
	for i := 0; i < len(as) && i < len(bs); i++ {
		if as[i] != bs[i] {
			if as[i] > bs[i] {
				return 1
			}
			return -1
		}
	}
	switch {
	case len(as) > len(bs):
		return 1
	case len(as) < len(bs):
		return -1
	default:
		return 0
	}
}

// splitNumeric splits "1.2.3" into [1,2,3]. Returns nil if any leading
// release segment is non-numeric (e.g. "not-a-version"), signalling the
// caller to fall back to string comparison.
func splitNumeric(v string) []int {
	// Only consider the release segment before any pre/post/dev suffix.
	v = strings.FieldsFunc(v, func(r rune) bool {
		return r == '-' || r == '+' || r == 'a' || r == 'b' || r == 'r' || r == 'c' || r == 'd' || r == 'p'
	})[0]
	parts := strings.Split(v, ".")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		n, ok := atoi(p)
		if !ok {
			return nil
		}
		nums = append(nums, n)
	}
	return nums
}

func atoi(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
		n = n*10 + int(r-'0')
	}
	return n, true
}
