package search

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestSearcherRunAggregatesResultsAndRendersSortedMarkdown(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			switch {
			case request.URL.Host == "google.serper.dev":
				body := `{"organic":[{"link":"https://playbooks.com/skills/acme/skills/alpha"},{"link":"https://playbooks.com/skills/acme/skills/beta"}]}`
				return responseWithBody(body), nil
			case request.URL.Host == "playbooks.com":
				body := `<a href="/skills/acme/skills/beta">beta</a><a href="/skills/zed/repo/gamma">gamma</a>`
				return responseWithBody(body), nil
			default:
				t.Fatalf("unexpected request: %s", request.URL.String())
				return nil, nil
			}
		}),
	}

	searcher := Searcher{Client: client}
	var stdout strings.Builder
	var stderr strings.Builder

	result, err := searcher.Run([]string{"typescript", "react", "testing"}, Config{
		TopN:           2,
		MinMatch:       2,
		SerperAPIKey:   "serper-key",
		ScrapedoAPIKey: "",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.Shown != 2 || result.Matched != 3 {
		t.Fatalf("unexpected result counts: %#v", result)
	}

	wantStdout := strings.Join([]string{
		"| # | Skill | Owner/Repo | Keywords Matched | Match Count | URL |",
		"|---|-------|------------|-----------------|-------------|-----|",
		"| 1 | alpha | acme/skills | typescript, react, testing | 3 | https://playbooks.com/skills/acme/skills/alpha |",
		"| 2 | beta | acme/skills | typescript, react, testing | 3 | https://playbooks.com/skills/acme/skills/beta |",
		"",
	}, "\n")
	if stdout.String() != wantStdout {
		t.Fatalf("stdout mismatch\nwant:\n%s\ngot:\n%s", wantStdout, stdout.String())
	}

	if !strings.Contains(stderr.String(), "[search] Searching for 3 keyword(s)...") {
		t.Fatalf("expected search start log, got:\n%s", stderr.String())
	}
	if !strings.Contains(stderr.String(), "[search] Done. 2 skill(s) shown (from 3 total matched).") {
		t.Fatalf("expected completion log, got:\n%s", stderr.String())
	}
}

func TestSearcherRunFallsBackToScrapedoWhenDirectHTMLIsTooShort(t *testing.T) {
	t.Parallel()

	client := &http.Client{
		Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			switch {
			case request.URL.Host == "google.serper.dev":
				return responseWithBody(`{"organic":[]}`), nil
			case request.URL.Host == "playbooks.com":
				return responseWithBody("tiny"), nil
			case request.URL.Host == "api.scrapedo.com":
				return responseWithBody(`<a href="/skills/scraped/repo/delta">delta</a>`), nil
			default:
				t.Fatalf("unexpected request: %s", request.URL.String())
				return nil, nil
			}
		}),
	}

	searcher := Searcher{Client: client}
	var stdout strings.Builder
	var stderr strings.Builder

	result, err := searcher.Run([]string{"openclaw", "agent", "skill"}, Config{
		TopN:           0,
		MinMatch:       1,
		SerperAPIKey:   "serper-key",
		ScrapedoAPIKey: "scrapedo-key",
	}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if result.Shown != 1 || !strings.Contains(stdout.String(), "https://playbooks.com/skills/scraped/repo/delta") {
		t.Fatalf("expected scrapedo fallback result, got stdout:\n%s", stdout.String())
	}
}

func responseWithBody(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}
