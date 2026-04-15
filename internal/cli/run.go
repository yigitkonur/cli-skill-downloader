package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/yigitkonur/cli-skill-downloader/internal/download"
	skillsearch "github.com/yigitkonur/cli-skill-downloader/internal/search"
)

func Run(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "search" {
		return runSearch(args[1:], stdout, stderr)
	}

	parsed, userErr := parseMainArgs(args)
	if userErr != nil {
		writeUserError(stderr, userErr)
		return 1
	}
	if parsed.ShowHelp {
		_, _ = io.WriteString(stdout, mainHelpText)
		return 0
	}
	if parsed.ShowVersion {
		_, _ = fmt.Fprintf(stdout, "skill-dl v%s\n", Version)
		return 0
	}

	urls, err := collectURLs(parsed.Sources, stdin)
	if err != nil {
		writeUserError(stderr, &UserError{Message: err.Error()})
		return 1
	}

	result := download.Processor{}.Run(urls, download.Options{
		OutputDir:      parsed.Config.OutputDir,
		Category:       parsed.Config.Category,
		AutoCategorize: parsed.Config.AutoCategorize,
		Verbose:        parsed.Config.Verbose,
		DryRun:         parsed.Config.DryRun,
		Force:          parsed.Config.Force,
		Version:        Version,
	}, stderr)
	if result.Failed > 0 {
		return 1
	}
	return 0
}

func runSearch(args []string, stdout io.Writer, stderr io.Writer) int {
	parsed, userErr := parseSearchArgs(args)
	if userErr != nil {
		writeUserError(stderr, userErr)
		return 1
	}
	if parsed.ShowHelp {
		_, _ = io.WriteString(stdout, searchHelpText)
		return 0
	}

	searcher := skillsearch.Searcher{}
	_, err := searcher.Run(parsed.Keywords, skillsearch.Config{
		TopN:           parsed.TopN,
		MinMatch:       parsed.MinMatch,
		SerperAPIKey:   envOrDefault("SERPER_API_KEY", skillsearch.DefaultSerperAPIKey),
		ScrapedoAPIKey: envOrDefault("SCRAPEDO_API_KEY", skillsearch.DefaultScrapedoAPIKey),
	}, stdout, stderr)
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err.Error())
		return 1
	}
	return 0
}

func writeUserError(stderr io.Writer, userErr *UserError) {
	if userErr.Message != "" {
		_, _ = fmt.Fprintf(stderr, "[ERR]  %s\n", userErr.Message)
	}
	if userErr.UsageHint != "" {
		_, _ = fmt.Fprintf(stderr, "%s\n", userErr.UsageHint)
	}
	if userErr.PrintHelp {
		_, _ = fmt.Fprintf(stderr, "\n%s", mainHelpText)
	}
	if userErr.SearchHelp {
		_, _ = fmt.Fprintf(stderr, "\n%s", searchHelpText)
	}
}

func envOrDefault(name string, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}
