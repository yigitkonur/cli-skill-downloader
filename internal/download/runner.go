package download

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Options struct {
	OutputDir      string
	Category       string
	AutoCategorize bool
	Verbose        bool
	DryRun         bool
	Force          bool
	Version        string
}

type Result struct {
	Total      int
	OK         int
	Failed     int
	Skipped    int
	FailedURLs []string
}

type GitCloner interface {
	Clone(repoKey string, dest string) error
}

type Processor struct {
	Git       GitCloner
	MakeTemp  func() (string, error)
	RemoveAll func(string) error
}

type repoSkill struct {
	RepoKey    string
	Owner      string
	Repo       string
	Skill      string
	Category   string
	URL        string
	FolderName string
}

type gitCommandCloner struct{}

func (gitCommandCloner) Clone(repoKey string, dest string) error {
	cmd := exec.Command("git", "clone", "--depth", "1", "--quiet", "https://github.com/"+repoKey+".git", dest)
	return cmd.Run()
}

func (processor Processor) Run(urls []string, options Options, stderr io.Writer) Result {
	if processor.Git == nil {
		processor.Git = gitCommandCloner{}
	}
	if processor.MakeTemp == nil {
		processor.MakeTemp = func() (string, error) {
			return os.MkdirTemp("", "skill-dl-*")
		}
	}
	if processor.RemoveAll == nil {
		processor.RemoveAll = os.RemoveAll
	}

	fmt.Fprintln(stderr)
	fmt.Fprintf(stderr, "skill-dl v%s\n", options.Version)
	fmt.Fprintf(stderr, "Downloading %d skill(s) into: %s\n", len(urls), options.OutputDir)
	fmt.Fprintln(stderr)

	uniqueURLs := dedupeStrings(urls)
	if len(uniqueURLs) != len(urls) {
		info(stderr, "Deduplicated: %d → %d unique URLs", len(urls), len(uniqueURLs))
	}

	repoGroups, repoOrder := groupURLs(uniqueURLs, options, stderr)
	result := Result{Total: len(uniqueURLs)}

	current := 0
	for _, repoKey := range repoOrder {
		skills := repoGroups[repoKey]
		if len(skills) == 0 {
			continue
		}

		var cloneDir string
		cloneOK := true

		if !options.DryRun {
			var err error
			cloneDir, err = processor.MakeTemp()
			if err != nil {
				cloneOK = false
			} else if err := processor.Git.Clone(repoKey, cloneDir); err != nil {
				errorf(stderr, "Could not clone github.com/%s", repoKey)
				cloneOK = false
			}
		}

		for _, skill := range skills {
			current++
			dest := filesystemDestinationPath(options.OutputDir, skill.Category, skill.FolderName)
			displayDest := displayDestinationPath(options.OutputDir, skill.Category, skill.FolderName)

			if options.DryRun {
				info(stderr, "[%d/%d] %s/%s/%s → %s", current, result.Total, skill.Owner, skill.Repo, skill.Skill, displayDest)
				result.Skipped++
				continue
			}

			if !cloneOK {
				errorf(stderr, "[%d/%d] %s — repo unavailable", current, result.Total, skill.Skill)
				result.Failed++
				result.FailedURLs = append(result.FailedURLs, skill.URL)
				continue
			}

			if pathExists(dest) && !options.Force {
				warn(stderr, "[%d/%d] Exists: %s (--force to overwrite)", current, result.Total, skill.FolderName)
				result.Skipped++
				continue
			}

			info(stderr, "[%d/%d] %s", current, result.Total, skill.Skill)

			location, ok, err := FindSkillDir(cloneDir, skill.Skill)
			if err != nil || !ok {
				errorf(stderr, "  Not found in repo")
				result.Failed++
				result.FailedURLs = append(result.FailedURLs, skill.URL)
				continue
			}

			fileCount, err := CopySkill(location, dest)
			if err != nil {
				errorf(stderr, "  Copy failed")
				result.Failed++
				result.FailedURLs = append(result.FailedURLs, skill.URL)
				continue
			}

			success(stderr, "  %d files → %s", fileCount, successPath(skill.Category, skill.FolderName))
			result.OK++
		}

		if cloneDir != "" {
			_ = processor.RemoveAll(cloneDir)
		}
	}

	fmt.Fprintln(stderr)
	fmt.Fprintln(stderr, "━━━ Summary ━━━")
	fmt.Fprintf(stderr, "  Total:   %d\n", result.Total)
	if result.OK > 0 {
		fmt.Fprintf(stderr, "  Success: %d\n", result.OK)
	}
	if result.Skipped > 0 {
		fmt.Fprintf(stderr, "  Skipped: %d\n", result.Skipped)
	}
	if result.Failed > 0 {
		fmt.Fprintf(stderr, "  Failed:  %d\n", result.Failed)
	}

	if len(result.FailedURLs) > 0 {
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Failed URLs:")
		for _, failedURL := range result.FailedURLs {
			fmt.Fprintf(stderr, "  %s\n", failedURL)
		}
	}

	if !options.DryRun && result.OK > 0 {
		fmt.Fprintln(stderr)
		fmt.Fprintf(stderr, "Output: %s\n", options.OutputDir)
	}

	fmt.Fprintln(stderr)
	return result
}

func groupURLs(urls []string, options Options, stderr io.Writer) (map[string][]repoSkill, []string) {
	repoGroups := make(map[string][]repoSkill)
	var repoOrder []string

	for _, rawURL := range urls {
		parsed, ok := ParsePlaybooksURL(rawURL)
		if !ok {
			errorf(stderr, "Invalid URL (skipped): %s", rawURL)
			continue
		}

		repoKey := parsed.Owner + "/" + parsed.Repo
		entry := repoSkill{
			RepoKey:    repoKey,
			Owner:      parsed.Owner,
			Repo:       parsed.Repo,
			Skill:      parsed.Skill,
			Category:   CategorizeSkill(parsed.Skill, options.Category, options.AutoCategorize),
			URL:        rawURL,
			FolderName: parsed.Owner + "--" + parsed.Repo + "--" + parsed.Skill,
		}

		if _, exists := repoGroups[repoKey]; !exists {
			repoOrder = append(repoOrder, repoKey)
		}
		repoGroups[repoKey] = append(repoGroups[repoKey], entry)
	}

	return repoGroups, repoOrder
}

func filesystemDestinationPath(outputDir string, category string, folderName string) string {
	if category == "" {
		return filepath.Join(outputDir, folderName)
	}
	return filepath.Join(outputDir, category, folderName)
}

func displayDestinationPath(outputDir string, category string, folderName string) string {
	parts := []string{strings.TrimRight(outputDir, "/")}
	if category != "" {
		parts = append(parts, category)
	}
	parts = append(parts, folderName)
	return strings.Join(parts, "/")
}

func successPath(category string, folderName string) string {
	if category == "" {
		return folderName
	}
	return filepath.ToSlash(filepath.Join(category, folderName))
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var deduped []string

	for _, value := range values {
		value = strings.TrimSuffix(value, "/")
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		deduped = append(deduped, value)
	}

	return deduped
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func info(stderr io.Writer, format string, args ...any) {
	fmt.Fprintf(stderr, "[INFO] "+format+"\n", args...)
}

func warn(stderr io.Writer, format string, args ...any) {
	fmt.Fprintf(stderr, "[WARN] "+format+"\n", args...)
}

func errorf(stderr io.Writer, format string, args ...any) {
	fmt.Fprintf(stderr, "[ERR]  "+format+"\n", args...)
}

func success(stderr io.Writer, format string, args ...any) {
	fmt.Fprintf(stderr, "[OK]   "+format+"\n", args...)
}
