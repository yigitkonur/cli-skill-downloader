package download

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestParsePlaybooksURLHandlesTrailingSlashAndNestedSkillPath(t *testing.T) {
	t.Parallel()

	parsed, ok := ParsePlaybooksURL("https://playbooks.com/skills/owner/repo/skill/name/")
	if !ok {
		t.Fatalf("expected URL to parse")
	}

	if parsed.Owner != "owner" || parsed.Repo != "repo" || parsed.Skill != "skill/name" {
		t.Fatalf("unexpected parsed URL: %#v", parsed)
	}
}

func TestParsePlaybooksURLRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	for _, rawURL := range []string{
		"https://example.com/skills/owner/repo/skill",
		"https://playbooks.com/skills/owner/repo",
		"not-a-url",
	} {
		rawURL := rawURL
		t.Run(rawURL, func(t *testing.T) {
			t.Parallel()

			if _, ok := ParsePlaybooksURL(rawURL); ok {
				t.Fatalf("expected URL to be rejected: %s", rawURL)
			}
		})
	}
}

func TestCategorizeSkillMatchesBashRulesAndOverrides(t *testing.T) {
	t.Parallel()

	if got := CategorizeSkill("typescript-magician", "", true); got != "pro-and-review" {
		t.Fatalf("unexpected category for typescript-magician: %q", got)
	}
	if got := CategorizeSkill("nextjs-typescript", "", true); got != "react-typescript" {
		t.Fatalf("unexpected category for nextjs-typescript: %q", got)
	}
	if got := CategorizeSkill("anything", "forced", true); got != "forced" {
		t.Fatalf("unexpected forced category: %q", got)
	}
	if got := CategorizeSkill("anything", "", false); got != "" {
		t.Fatalf("expected empty category when auto-categorize is disabled, got %q", got)
	}
}

func TestFindSkillDirUsesKnownPathsRecursiveSearchAndRootFallback(t *testing.T) {
	t.Parallel()

	t.Run("known path", func(t *testing.T) {
		t.Parallel()

		repoDir := t.TempDir()
		skillDir := filepath.Join(repoDir, "skills", "alpha")
		mustWriteFile(t, filepath.Join(skillDir, "SKILL.md"), "alpha")

		location, ok, err := FindSkillDir(repoDir, "alpha")
		if err != nil {
			t.Fatalf("FindSkillDir returned error: %v", err)
		}
		if !ok {
			t.Fatalf("expected known-path skill to be found")
		}
		if location.RootLevel {
			t.Fatalf("did not expect root-level location")
		}
		if location.Path != skillDir {
			t.Fatalf("unexpected skill dir: %q", location.Path)
		}
	})

	t.Run("recursive search", func(t *testing.T) {
		t.Parallel()

		repoDir := t.TempDir()
		skillDir := filepath.Join(repoDir, "nested", "deep", "beta")
		mustWriteFile(t, filepath.Join(skillDir, "SKILL.md"), "beta")

		location, ok, err := FindSkillDir(repoDir, "beta")
		if err != nil {
			t.Fatalf("FindSkillDir returned error: %v", err)
		}
		if !ok {
			t.Fatalf("expected recursive skill to be found")
		}
		if location.RootLevel {
			t.Fatalf("did not expect root-level location")
		}
		if location.Path != skillDir {
			t.Fatalf("unexpected recursive skill dir: %q", location.Path)
		}
	})

	t.Run("root-level fallback", func(t *testing.T) {
		t.Parallel()

		repoDir := t.TempDir()
		mustWriteFile(t, filepath.Join(repoDir, "SKILL.md"), "root")

		location, ok, err := FindSkillDir(repoDir, "repo-skill")
		if err != nil {
			t.Fatalf("FindSkillDir returned error: %v", err)
		}
		if !ok {
			t.Fatalf("expected root-level skill to be found")
		}
		if !location.RootLevel {
			t.Fatalf("expected root-level location")
		}
		if location.Path != repoDir {
			t.Fatalf("unexpected root-level location path: %q", location.Path)
		}
	})
}

func TestCopySkillRootLevelExcludesRepoMetadataAndKeepsSkillArtifacts(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	destDir := filepath.Join(t.TempDir(), "dest")

	mustWriteFile(t, filepath.Join(sourceDir, "SKILL.md"), "root skill")
	mustWriteFile(t, filepath.Join(sourceDir, "README.md"), "readme")
	mustWriteFile(t, filepath.Join(sourceDir, "LICENSE"), "license")
	mustWriteFile(t, filepath.Join(sourceDir, ".editorconfig"), "root=true")
	mustWriteFile(t, filepath.Join(sourceDir, ".git", "HEAD"), "ref: refs/heads/main")
	mustWriteFile(t, filepath.Join(sourceDir, ".github", "workflows", "ci.yml"), "name: ci")
	mustWriteFile(t, filepath.Join(sourceDir, "rules", "rule.md"), "rule")
	mustWriteFile(t, filepath.Join(sourceDir, ".custom", "note.txt"), "note")

	fileCount, err := CopySkill(SkillLocation{Path: sourceDir, RootLevel: true}, destDir)
	if err != nil {
		t.Fatalf("CopySkill returned error: %v", err)
	}
	if fileCount != 4 {
		t.Fatalf("unexpected file count: %d", fileCount)
	}

	gotFiles := listFiles(t, destDir)
	wantFiles := []string{
		".custom/note.txt",
		".editorconfig",
		"SKILL.md",
		"rules/rule.md",
	}
	if !reflect.DeepEqual(gotFiles, wantFiles) {
		t.Fatalf("copied files mismatch\nwant: %#v\ngot:  %#v", wantFiles, gotFiles)
	}
}

func TestCopySkillSubdirectoryCopiesWholeDirectoryAndRemovesGitArtifacts(t *testing.T) {
	t.Parallel()

	sourceDir := t.TempDir()
	destDir := filepath.Join(t.TempDir(), "dest")

	mustWriteFile(t, filepath.Join(sourceDir, "SKILL.md"), "subdir skill")
	mustWriteFile(t, filepath.Join(sourceDir, "rules", "rule.md"), "rule")
	mustWriteFile(t, filepath.Join(sourceDir, ".git", "HEAD"), "ref: refs/heads/main")

	fileCount, err := CopySkill(SkillLocation{Path: sourceDir}, destDir)
	if err != nil {
		t.Fatalf("CopySkill returned error: %v", err)
	}
	if fileCount != 2 {
		t.Fatalf("unexpected file count: %d", fileCount)
	}

	gotFiles := listFiles(t, destDir)
	wantFiles := []string{
		"SKILL.md",
		"rules/rule.md",
	}
	if !reflect.DeepEqual(gotFiles, wantFiles) {
		t.Fatalf("copied files mismatch\nwant: %#v\ngot:  %#v", wantFiles, gotFiles)
	}
}

func mustWriteFile(t *testing.T, path string, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func listFiles(t *testing.T, root string) []string {
	t.Helper()

	var files []string
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(relPath))
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}

	sort.Strings(files)
	return files
}
