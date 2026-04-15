package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	OutputDir      string
	Category       string
	AutoCategorize bool
	Verbose        bool
	DryRun         bool
	Force          bool
}

type MainArgs struct {
	Config      Config
	Sources     []string
	ShowHelp    bool
	ShowVersion bool
}

type SearchArgs struct {
	Keywords   []string
	TopN       int
	MinMatch   int
	ShowHelp   bool
	SearchOnly bool
}

type UserError struct {
	Message    string
	UsageHint  string
	PrintHelp  bool
	SearchHelp bool
}

func defaultConfig() Config {
	return Config{
		OutputDir:      "./skills-collection",
		AutoCategorize: true,
	}
}

func parseMainArgs(args []string) (MainArgs, *UserError) {
	parsed := MainArgs{
		Config: defaultConfig(),
	}

	for index := 0; index < len(args); index++ {
		arg := args[index]

		switch arg {
		case "-h", "--help":
			parsed.ShowHelp = true
			return parsed, nil
		case "--version":
			parsed.ShowVersion = true
			return parsed, nil
		case "-o", "--output":
			if index+1 >= len(args) {
				return MainArgs{}, &UserError{Message: fmt.Sprintf("%s requires a value", arg)}
			}
			index++
			parsed.Config.OutputDir = args[index]
		case "-c", "--category":
			if index+1 >= len(args) {
				return MainArgs{}, &UserError{Message: fmt.Sprintf("%s requires a value", arg)}
			}
			index++
			parsed.Config.Category = args[index]
			parsed.Config.AutoCategorize = false
		case "--no-auto-category":
			parsed.Config.AutoCategorize = false
		case "-f", "--force":
			parsed.Config.Force = true
		case "--dry-run":
			parsed.Config.DryRun = true
		case "-v", "--verbose":
			parsed.Config.Verbose = true
		default:
			if strings.HasPrefix(arg, "-") && arg != "-" {
				return MainArgs{}, &UserError{
					Message:   fmt.Sprintf("Unknown option: %s", arg),
					UsageHint: "Run 'skill-dl --help' for usage.",
				}
			}
			parsed.Sources = append(parsed.Sources, arg)
		}
	}

	if len(parsed.Sources) == 0 {
		return MainArgs{}, &UserError{
			Message:   "No source provided.",
			PrintHelp: true,
		}
	}

	return parsed, nil
}

func parseSearchArgs(args []string) (SearchArgs, *UserError) {
	parsed := SearchArgs{
		MinMatch:   1,
		SearchOnly: true,
	}

	for index := 0; index < len(args); index++ {
		arg := args[index]

		switch arg {
		case "-h", "--help":
			parsed.ShowHelp = true
			return parsed, nil
		case "--top":
			if index+1 >= len(args) {
				return SearchArgs{}, &UserError{Message: "--top requires a positive integer"}
			}
			index++
			value, err := strconv.Atoi(args[index])
			if err != nil || value < 0 {
				return SearchArgs{}, &UserError{Message: "--top requires a positive integer"}
			}
			parsed.TopN = value
		case "--min-match":
			if index+1 >= len(args) {
				return SearchArgs{}, &UserError{Message: "--min-match requires a positive integer"}
			}
			index++
			value, err := strconv.Atoi(args[index])
			if err != nil || value < 0 {
				return SearchArgs{}, &UserError{Message: "--min-match requires a positive integer"}
			}
			parsed.MinMatch = value
		default:
			if strings.HasPrefix(arg, "-") {
				return SearchArgs{}, &UserError{
					Message:   fmt.Sprintf("Unknown search option: %s", arg),
					UsageHint: "Run 'skill-dl search --help' for usage.",
				}
			}
			parsed.Keywords = append(parsed.Keywords, arg)
		}
	}

	if len(parsed.Keywords) < 3 {
		return SearchArgs{}, &UserError{
			Message:   fmt.Sprintf("search requires at least 3 keywords (got %d).", len(parsed.Keywords)),
			UsageHint: "Run 'skill-dl search --help' for usage.",
		}
	}
	if len(parsed.Keywords) > 20 {
		return SearchArgs{}, &UserError{
			Message:   fmt.Sprintf("search accepts at most 20 keywords (got %d).", len(parsed.Keywords)),
			UsageHint: "Run 'skill-dl search --help' for usage.",
		}
	}

	return parsed, nil
}

func collectURLs(sources []string, stdin io.Reader) ([]string, error) {
	var urls []string

	for _, source := range sources {
		switch {
		case source == "-":
			loaded, err := collectURLsFromReader(stdin)
			if err != nil {
				return nil, err
			}
			urls = append(urls, loaded...)
		case fileExists(source):
			loaded, err := collectURLsFromFile(source)
			if err != nil {
				return nil, err
			}
			urls = append(urls, loaded...)
		case isURL(source):
			urls = append(urls, source)
		default:
			return nil, fmt.Errorf("Source not recognized: %s (expected URL, file, or '-')", source)
		}
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("No URLs found in the provided source(s).")
	}

	return urls, nil
}

func collectURLsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return collectURLsFromReader(file)
}

func collectURLsFromReader(reader io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(reader)
	var urls []string

	for scanner.Scan() {
		line := cleanLine(scanner.Text())
		if line != "" {
			urls = append(urls, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func cleanLine(line string) string {
	if hashIndex := strings.Index(line, "#"); hashIndex >= 0 {
		line = line[:hashIndex]
	}
	return strings.TrimSpace(line)
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var deduped []string

	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		deduped = append(deduped, value)
	}

	return deduped
}

func isURL(value string) bool {
	return strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://")
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
