package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

const (
	DefaultSerperAPIKey   = "a3c7a6122cb26d80c6a975e41b2a1047eb746b59"
	DefaultScrapedoAPIKey = "d4874c0918ad49b782ce649c642364a5561da9e8387"
)

var (
	fullURLPattern = regexp.MustCompile(`https://playbooks\.com/skills/[^"'\s<]+`)
	pathURLPattern = regexp.MustCompile(`/skills/[^"'<>\s]+`)
)

type Config struct {
	TopN           int
	MinMatch       int
	SerperAPIKey   string
	ScrapedoAPIKey string
}

type Result struct {
	RawResults int
	Matched    int
	Shown      int
}

type Searcher struct {
	Client *http.Client
}

type aggregateEntry struct {
	Path     string
	Keywords []string
	Count    int
}

func (searcher Searcher) Run(keywords []string, config Config, stdout io.Writer, stderr io.Writer) (Result, error) {
	client := searcher.Client
	if client == nil {
		client = http.DefaultClient
	}

	fmt.Fprintf(stderr, "[search] Searching for %d keyword(s)...\n", len(keywords))
	if config.SerperAPIKey != "" {
		fmt.Fprintln(stderr, "[search] Serper API enabled (Google-powered discovery)")
	}
	if config.ScrapedoAPIKey != "" {
		fmt.Fprintln(stderr, "[search] Scrapedo proxy enabled (fallback for blocked requests)")
	}

	aggregated := make(map[string]*aggregateEntry)
	rawCount := 0

	for _, keyword := range keywords {
		perKeyword := make(map[string]struct{})

		if config.SerperAPIKey != "" {
			paths, err := searcher.fetchSerperPaths(client, keyword, config.SerperAPIKey)
			if err != nil {
				return Result{}, err
			}
			for _, path := range paths {
				perKeyword[path] = struct{}{}
			}
		}

		paths, err := searcher.fetchPlaybooksPaths(client, keyword, config.ScrapedoAPIKey)
		if err != nil {
			return Result{}, err
		}
		for _, path := range paths {
			perKeyword[path] = struct{}{}
		}

		rawCount += len(perKeyword)

		for path := range perKeyword {
			entry := aggregated[path]
			if entry == nil {
				entry = &aggregateEntry{Path: path}
				aggregated[path] = entry
			}
			entry.Count++
			entry.Keywords = append(entry.Keywords, keyword)
		}
	}

	fmt.Fprintf(stderr, "[search] Raw results: %d skill references across %d keywords\n", rawCount, len(keywords))
	fmt.Fprintln(stderr, "[search] Aggregating results...")

	var filtered []aggregateEntry
	for _, entry := range aggregated {
		if entry.Count >= config.MinMatch {
			filtered = append(filtered, *entry)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Count != filtered[j].Count {
			return filtered[i].Count > filtered[j].Count
		}
		return filtered[i].Path < filtered[j].Path
	})

	if len(filtered) == 0 {
		return Result{RawResults: rawCount}, fmt.Errorf("No skills found matching the given keywords (min-match=%d).", config.MinMatch)
	}

	shown := filtered
	if config.TopN > 0 && len(shown) > config.TopN {
		shown = shown[:config.TopN]
	}

	if _, err := io.WriteString(stdout, "| # | Skill | Owner/Repo | Keywords Matched | Match Count | URL |\n"); err != nil {
		return Result{}, err
	}
	if _, err := io.WriteString(stdout, "|---|-------|------------|-----------------|-------------|-----|\n"); err != nil {
		return Result{}, err
	}
	for index, entry := range shown {
		ownerRepo, skill := splitPath(entry.Path)
		if _, err := fmt.Fprintf(stdout, "| %d | %s | %s | %s | %d | https://playbooks.com%s |\n",
			index+1,
			skill,
			ownerRepo,
			strings.Join(entry.Keywords, ", "),
			entry.Count,
			entry.Path,
		); err != nil {
			return Result{}, err
		}
	}

	fmt.Fprintln(stderr)
	fmt.Fprintf(stderr, "[search] Done. %d skill(s) shown (from %d total matched).\n", len(shown), len(filtered))

	return Result{
		RawResults: rawCount,
		Matched:    len(filtered),
		Shown:      len(shown),
	}, nil
}

func (searcher Searcher) fetchSerperPaths(client *http.Client, keyword string, apiKey string) ([]string, error) {
	body, err := json.Marshal(map[string]any{
		"q":   "site:playbooks.com/skills " + keyword,
		"num": 30,
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, "https://google.serper.dev/search", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("X-API-KEY", apiKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return normalizeFullURLs(fullURLPattern.FindAllString(string(data), -1)), nil
}

func (searcher Searcher) fetchPlaybooksPaths(client *http.Client, keyword string, scrapedoAPIKey string) ([]string, error) {
	pathsByValue := make(map[string]struct{})

	encodedKeyword := url.QueryEscape(keyword)
	for page := 1; page <= 3; page++ {
		targetURL := fmt.Sprintf("https://playbooks.com/skills?search=%s&page=%d", encodedKeyword, page)
		body, err := searcher.fetchURLBody(client, targetURL)
		if err != nil {
			return nil, err
		}
		if len(body) < 200 && scrapedoAPIKey != "" {
			scrapedoURL := "https://api.scrapedo.com/scrape?url=" + url.QueryEscape(targetURL) + "&token=" + url.QueryEscape(scrapedoAPIKey) + "&render=false&super=false"
			body, err = searcher.fetchURLBody(client, scrapedoURL)
			if err != nil {
				return nil, err
			}
		}
		for _, path := range normalizeRelativePaths(pathURLPattern.FindAllString(body, -1)) {
			pathsByValue[path] = struct{}{}
		}
	}

	return sortedKeys(pathsByValue), nil
}

func (searcher Searcher) fetchURLBody(client *http.Client, rawURL string) (string, error) {
	request, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func normalizeFullURLs(urls []string) []string {
	paths := make(map[string]struct{})
	for _, fullURL := range urls {
		trimmed := strings.TrimSuffix(fullURL, "/")
		trimmed = strings.TrimSuffix(trimmed, "\\")
		parsed, ok := normalizePlaybooksPath(trimmed)
		if ok {
			paths[parsed] = struct{}{}
		}
	}
	return sortedKeys(paths)
}

func normalizeRelativePaths(paths []string) []string {
	values := make(map[string]struct{})
	for _, path := range paths {
		trimmed := strings.TrimSuffix(path, "/")
		trimmed = strings.TrimSuffix(trimmed, "\\")
		normalized, ok := normalizePlaybooksPath("https://playbooks.com" + trimmed)
		if ok {
			values[normalized] = struct{}{}
		}
	}
	return sortedKeys(values)
}

func normalizePlaybooksPath(rawURL string) (string, bool) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	if parsedURL.Host != "playbooks.com" || !strings.HasPrefix(parsedURL.Path, "/skills/") {
		return "", false
	}
	parts := strings.Split(strings.TrimPrefix(parsedURL.Path, "/skills/"), "/")
	if len(parts) != 3 {
		return "", false
	}
	return "/skills/" + strings.Join(parts, "/"), true
}

func splitPath(path string) (string, string) {
	parts := strings.Split(strings.TrimPrefix(path, "/skills/"), "/")
	return parts[0] + "/" + parts[1], parts[2]
}

func sortedKeys(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for value := range values {
		keys = append(keys, value)
	}
	sort.Strings(keys)
	return keys
}
