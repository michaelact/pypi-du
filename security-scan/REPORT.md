# Security Scan — pypi-du

- Target: `/home/michaelact/GDPLabs/pypi-du`
- Commit: `9356105` (initial)
- Scanned (UTC): 2026-10-13 ~18:19
- Tooling: Trivy 0.61.1 (`fs` — vuln + secret + misconfig), `go vet`, manual pattern review

## Verdict: PASS — clean

No vulnerabilities, no secrets, no misconfigurations, no dangerous code patterns.

## 1. Trivy filesystem scan

```
┌────────┬───────┬─────────────────┬─────────┬───────────────────┐
│ Target │ Type  │ Vulnerabilities │ Secrets │ Misconfigurations │
├────────┼───────┼─────────────────┼─────────┼───────────────────┤
│ go.mod │ gomod │        0        │    -    │         -         │
└────────┴───────┴─────────────────┴─────────┴───────────────────┘
```

- Vulnerabilities: 0 (go.mod has no third-party dependencies — stdlib only)
- Secrets: none detected
- Misconfigurations: 0 config files, none flagged

Raw artifacts: `trivy.txt`, `trivy.json`.

## 2. go vet

Clean — no issues.

## 3. Manual pattern review (Go source)

| Check | Result |
|-------|--------|
| `os/exec` / `exec.Command` / `syscall` / `unsafe` / `reflect` | NONE |
| Outbound network destinations | ONLY `https://pypi.org` (project JSON + user profile). The `github.com/...` string is the User-Agent URL, not a request target. |
| File writes / deletes (`os.Remove`, `WriteFile`, `Create`, `OpenFile`) | NONE — the tool only reads from the network and writes to stdout |
| Env / credential access (`os.Getenv`, `os.Environ`) | NONE |
| Third-party dependencies | NONE — stdlib only (`net/http`, `encoding/json`, `regexp`, etc.) |

## Notes

- The tool is read-only by design: it issues HTTP GETs to `pypi.org` and prints results.
  It performs no authentication, no deletion, and no local file writes.
- Zero external dependencies means no transitive supply-chain surface to audit.
- Tests use `net/http/httptest` (in-process mock server); running `go test` makes no real
  network calls.

## Raw outputs in this directory

- `trivy.txt`  — human-readable Trivy report
- `trivy.json` — machine-readable Trivy report
