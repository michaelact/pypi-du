# pypi-du

Audit **PyPI project storage usage** — per project and per version — against PyPI's
per-project storage limit (default **10 GB**).

PyPI counts the size of *every file in every release* toward a project's limit. When a
project hits the cap, uploads fail:

```
400 Project size too large. Limit for '<project>' total size is 10 GB.
```

`pypi-du` tells you which projects are near the cap and which versions are the biggest,
so you know what to prune. It is **read-only**: it only issues GET requests to pypi.org
and never deletes anything.

> Think of it as `du`, but for your PyPI projects.

## Install

```bash
go install github.com/michaelact/pypi-du/cmd/pypi-du@latest
```

Or build from source:

```bash
go build -o pypi-du ./cmd/pypi-du
```

## Usage

Audit specific projects:

```bash
pypi-du requests boto3
```

Audit every project owned by a PyPI user:

```bash
pypi-du --user pypa
```

Show per-version breakdown (what to delete first):

```bash
pypi-du --versions numpy
```

Flag everything at/over a threshold:

```bash
pypi-du --user pypa --limit 8GB
```

JSON output (for scripting / dashboards):

```bash
pypi-du --user pypa --json
```

### Flags

| Flag | Description |
|------|-------------|
| `--user <name>` | Audit every project owned by this PyPI user (scrapes the public profile page). |
| `--limit <size>` | Flag projects at/over this size. Accepts `10GB`, `500MB`, `1.5G`, or raw bytes. Default `10GB`. |
| `--versions` | Show per-version size breakdown (newest first). |
| `--json` | Emit JSON. |
| `--bytes` | Raw byte counts instead of human-readable units. |
| `--quiet` | Only the total line per project. |

## How it works

- Single project: `GET https://pypi.org/pypi/<project>/json`, then sums
  `releases[*][*].size` across all files in all releases.
- `--user`: PyPI has no JSON endpoint listing a user's projects, so the profile page
  (`/user/<name>/`) is parsed for `/project/<name>/` links.

Projects are reported largest-first. Nothing is cached, authenticated, or deleted.

## Development

```bash
go test ./...        # run tests
go vet ./...         # static checks
go build ./...       # compile
```

## License

Apache 2.0 — see [LICENSE](LICENSE).
