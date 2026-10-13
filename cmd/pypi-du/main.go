// Command pypi-du reports PyPI project storage usage (per-project, per-version)
// against PyPI's per-project storage limit (default 10 GB).
//
// Usage:
//
//	pypi-du <project> [<project>...]      audit specific projects
//	pypi-du --user <pypi-user>            audit every project owned by a user
//	pypi-du --user pypa --limit 8GB       flag projects over 8 GB
//
// Read-only: it only issues GET requests to pypi.org. It never deletes anything.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/michaelact/pypi-du/internal/pypi"
)

// options holds parsed CLI configuration.
type options struct {
	user     string
	limitStr string
	asJSON   bool
	asBytes  bool
	perVer   bool
	quiet    bool
	projects []string
}

// parseArgs parses flags that may be interspersed with positional arguments.
//
// Go's flag package stops at the first non-flag token, so "pkg --versions"
// would otherwise treat "--versions" as a positional. We repeatedly parse the
// remaining args, collecting positionals, until none are left — the standard
// idiom for interspersed flags.
func parseArgs(name string, args []string) (*options, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.Usage = usage
	opt := &options{}
	fs.StringVar(&opt.user, "user", "", "audit every project owned by this PyPI user")
	fs.StringVar(&opt.limitStr, "limit", "10GB", "flag projects whose total size meets or exceeds this (e.g. 10GB, 500MB)")
	fs.BoolVar(&opt.asJSON, "json", false, "output JSON")
	fs.BoolVar(&opt.asBytes, "bytes", false, "show raw byte counts instead of human-readable sizes")
	fs.BoolVar(&opt.perVer, "versions", false, "show per-version breakdown for each project")
	fs.BoolVar(&opt.quiet, "quiet", false, "only print the total size line per project")

	rest := args
	for len(rest) > 0 {
		if err := fs.Parse(rest); err != nil {
			return nil, err
		}
		rest = fs.Args()
		if len(rest) == 0 {
			break
		}
		// First leftover token is a genuine positional; consume it and
		// re-parse the remainder so flags after it are still recognized.
		opt.projects = append(opt.projects, rest[0])
		rest = rest[1:]
	}
	return opt, nil
}

func main() {
	opt, err := parseArgs(os.Args[0], os.Args[1:])
	if err != nil {
		// flag already printed the error + usage.
		os.Exit(2)
	}

	limit, err := pypi.ParseSize(opt.limitStr)
	if err != nil {
		fatalf("invalid --limit: %v", err)
	}

	projects := opt.projects
	if opt.user == "" && len(projects) == 0 {
		usage()
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	client := pypi.NewClient()

	if opt.user != "" {
		discovered, derr := client.DiscoverUserProjects(ctx, opt.user)
		if derr != nil {
			fatalf("discover projects for user %q: %v", opt.user, derr)
		}
		if len(discovered) == 0 {
			fatalf("no projects found for user %q", opt.user)
		}
		projects = append(projects, discovered...)
	}

	results := make([]*pypi.ProjectStorage, 0, len(projects))
	var failures int
	for _, p := range projects {
		ps, ferr := client.FetchProject(ctx, p)
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %v\n", p, ferr)
			failures++
			continue
		}
		results = append(results, ps)
	}

	// Largest projects first — the ones nearest the cap matter most.
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].TotalSize > results[j].TotalSize
	})

	if opt.asJSON {
		printJSON(results, limit)
	} else {
		printText(results, limit, opt.asBytes, opt.perVer, opt.quiet)
	}

	if failures > 0 {
		os.Exit(1)
	}
}

func fmtSize(b int64, asBytes bool) string {
	if asBytes {
		return fmt.Sprintf("%d", b)
	}
	return pypi.FormatSize(b)
}

func printText(results []*pypi.ProjectStorage, limit int64, asBytes, perVer, quiet bool) {
	var grand int64
	for _, ps := range results {
		grand += ps.TotalSize
		flag := ""
		if limit > 0 && ps.TotalSize >= limit {
			flag = "  ⚠️  OVER LIMIT"
		}

		if quiet {
			fmt.Printf("%s\t%s\n", ps.Name, fmtSize(ps.TotalSize, asBytes))
			continue
		}

		fmt.Printf("\n%s  —  %s  (%d releases, %d files)%s\n",
			ps.Name, fmtSize(ps.TotalSize, asBytes), ps.ReleaseCount, ps.FileCount, flag)

		if perVer {
			for _, v := range ps.Versions {
				fmt.Printf("    %-18s %s  (%d files)\n",
					v.Version, fmtSize(v.Size, asBytes), v.FileCount)
			}
		}
	}
	if !quiet && len(results) > 1 {
		fmt.Printf("\nTotal across %d projects: %s\n", len(results), fmtSize(grand, asBytes))
	}
}

func printJSON(results []*pypi.ProjectStorage, limit int64) {
	type out struct {
		Limit    int64                   `json:"limit_bytes"`
		Projects []*pypi.ProjectStorage  `json:"projects"`
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out{Limit: limit, Projects: results})
}

func usage() {
	fmt.Fprint(os.Stderr, `pypi-du — audit PyPI project storage usage against the per-project limit.

Usage:
  pypi-du <project> [<project>...]
  pypi-du --user <pypi-user> [flags]

Flags:
  --user <name>     audit every project owned by this PyPI user
  --limit <size>    flag projects at/over this size (default 10GB)
  --versions        show per-version breakdown
  --json            output JSON
  --bytes           raw byte counts
  --quiet           only the total line per project

Read-only: only issues GET requests to pypi.org; never deletes anything.
`)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "pypi-du: "+format+"\n", args...)
	os.Exit(1)
}
