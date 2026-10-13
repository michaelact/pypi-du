package pypi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"time"
)

// DefaultBaseURL is the public PyPI JSON API root. Overridable for testing.
const DefaultBaseURL = "https://pypi.org"

const userAgent = "pypi-du/0.1 (+https://github.com/michaelact/pypi-du)"

// Client fetches project metadata from a PyPI-compatible host.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// NewClient returns a Client with sensible defaults.
func NewClient() *Client {
	return &Client{
		BaseURL: DefaultBaseURL,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchProject retrieves and summarizes storage for a single project.
func (c *Client) FetchProject(ctx context.Context, project string) (*ProjectStorage, error) {
	raw, err := c.get(ctx, fmt.Sprintf("%s/pypi/%s/json", c.BaseURL, project))
	if err != nil {
		return nil, err
	}
	ps, err := SummarizeBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", project, err)
	}
	if ps.Name == "" {
		ps.Name = project
	}
	return ps, nil
}

// userProjectRe extracts project slugs from a PyPI user profile page.
// The profile lists each project as a link: <a href="/project/<name>/">.
var userProjectRe = regexp.MustCompile(`href="/project/([^/"]+)/"`)

// DiscoverUserProjects scrapes a PyPI user's profile page for their project
// names. PyPI has no JSON endpoint for this, so we parse the public HTML.
func (c *Client) DiscoverUserProjects(ctx context.Context, user string) ([]string, error) {
	raw, err := c.get(ctx, fmt.Sprintf("%s/user/%s/", c.BaseURL, user))
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	var projects []string
	for _, m := range userProjectRe.FindAllStringSubmatch(string(raw), -1) {
		name := m[1]
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		projects = append(projects, name)
	}
	sort.Strings(projects)
	return projects, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: HTTP %d %s", url, resp.StatusCode, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
