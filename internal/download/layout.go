package download

import (
	"errors"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type ParsedURL struct {
	Owner string
	Repo  string
	Skill string
}

type SkillLocation struct {
	Path      string
	RootLevel bool
}

var (
	errFoundSkill = errors.New("skill found")

	rootExcludeDirs = map[string]struct{}{
		".git":         {},
		".github":      {},
		".gitlab":      {},
		"node_modules": {},
		".vscode":      {},
		".idea":        {},
	}

	rootExcludeFiles = map[string]struct{}{
		"README.md":          {},
		"LICENSE":            {},
		"LICENSE.md":         {},
		"CHANGELOG.md":       {},
		"CONTRIBUTING.md":    {},
		"CODE_OF_CONDUCT.md": {},
		"SECURITY.md":        {},
		".gitignore":         {},
		".gitattributes":     {},
		"package.json":       {},
		"package-lock.json":  {},
		"yarn.lock":          {},
		"pnpm-lock.yaml":     {},
		"bun.lockb":          {},
	}
)

func ParsePlaybooksURL(rawURL string) (ParsedURL, bool) {
	trimmed := strings.TrimSuffix(rawURL, "/")
	parsedURL, err := url.Parse(trimmed)
	if err != nil || parsedURL.Host != "playbooks.com" {
		return ParsedURL{}, false
	}

	const prefix = "/skills/"
	if !strings.HasPrefix(parsedURL.Path, prefix) {
		return ParsedURL{}, false
	}

	segments := strings.Split(strings.TrimPrefix(parsedURL.Path, prefix), "/")
	if len(segments) < 3 {
		return ParsedURL{}, false
	}

	return ParsedURL{
		Owner: segments[0],
		Repo:  segments[1],
		Skill: strings.Join(segments[2:], "/"),
	}, true
}

func CategorizeSkill(skill string, forcedCategory string, autoCategorize bool) string {
	if forcedCategory != "" {
		return forcedCategory
	}
	if !autoCategorize {
		return ""
	}

	name := strings.ToLower(skill)

	switch {
	case containsAny(name, "react-typescript", "react-ts", "nextjs-typescript", "next-ts"):
		return "react-typescript"
	case containsAny(name, "strict", "advanced-type", "type-expert", "guardian", "detector", "advanced-pattern"):
		return "strict-and-types"
	case containsAny(name, "sdk", "ts-library", "typescript-v", "library"):
		return "sdk-and-libraries"
	case containsAny(name, "review", "reviewer", "pro-skill", "magician", "code-review"):
		return "pro-and-review"
	case containsAny(name, "generator", "setup", "init", "tdd-", "circular", "tooling", "lang-typescript", "bun-"):
		return "tooling-and-setup"
	case containsAny(name, "coding-standard", "code-standard", "clean-code", "code-style", "best-practice", "check-try"):
		return "code-quality"
	case containsAny(name, "typescript", "expert", "clean-typescript", "javascript-typescript"):
		return "best-practices"
	default:
		return "general"
	}
}

func FindSkillDir(repoDir string, skill string) (SkillLocation, bool, error) {
	knownPaths := []string{
		filepath.Join("skills", skill),
		skill,
		filepath.Join(".skills", skill),
		filepath.Join(".claude", "skills", skill),
		filepath.Join(".agent", "skills", skill),
		filepath.Join(".opencode", "skills", skill),
		filepath.Join(".cursor", "skills", skill),
		filepath.Join(".agents", "skills", skill),
		filepath.Join("src", "skills", skill),
	}

	for _, relativePath := range knownPaths {
		candidate := filepath.Join(repoDir, relativePath)
		if fileExists(filepath.Join(candidate, "SKILL.md")) {
			return SkillLocation{Path: candidate}, true, nil
		}
	}

	var foundPath string
	err := filepath.WalkDir(repoDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}

		parentDir := filepath.Dir(path)
		if filepath.Base(parentDir) == skill {
			foundPath = parentDir
			return errFoundSkill
		}
		return nil
	})
	if err != nil && !errors.Is(err, errFoundSkill) {
		return SkillLocation{}, false, err
	}
	if foundPath != "" {
		return SkillLocation{Path: foundPath}, true, nil
	}

	if fileExists(filepath.Join(repoDir, "SKILL.md")) {
		return SkillLocation{Path: repoDir, RootLevel: true}, true, nil
	}

	return SkillLocation{}, false, nil
}

func CopySkill(source SkillLocation, dest string) (int, error) {
	if err := os.RemoveAll(dest); err != nil {
		return 0, err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return 0, err
	}

	if source.RootLevel {
		entries, err := os.ReadDir(source.Path)
		if err != nil {
			return 0, err
		}
		for _, entry := range entries {
			name := entry.Name()
			if shouldExcludeRootEntry(name, entry.IsDir()) {
				continue
			}
			if err := copyEntry(filepath.Join(source.Path, name), filepath.Join(dest, name)); err != nil {
				return 0, err
			}
		}
	} else {
		entries, err := os.ReadDir(source.Path)
		if err != nil {
			return 0, err
		}
		for _, entry := range entries {
			if err := copyEntry(filepath.Join(source.Path, entry.Name()), filepath.Join(dest, entry.Name())); err != nil {
				return 0, err
			}
		}
	}

	if err := os.RemoveAll(filepath.Join(dest, ".git")); err != nil {
		return 0, err
	}

	return countFiles(dest)
}

func shouldExcludeRootEntry(name string, isDir bool) bool {
	if isDir {
		_, excluded := rootExcludeDirs[name]
		return excluded
	}
	_, excluded := rootExcludeFiles[name]
	return excluded
}

func copyEntry(sourcePath string, destPath string) error {
	info, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if err := os.MkdirAll(destPath, info.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(sourcePath)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := copyEntry(filepath.Join(sourcePath, entry.Name()), filepath.Join(destPath, entry.Name())); err != nil {
				return err
			}
		}
		return nil
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}

	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func countFiles(root string) (int, error) {
	var count int

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func containsAny(value string, patterns ...string) bool {
	for _, pattern := range patterns {
		if strings.Contains(value, pattern) {
			return true
		}
	}
	return false
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
