package cli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseMainArgsSupportsInterleavedFlagsAndSources(t *testing.T) {
	t.Parallel()

	parsed, parseErr := parseMainArgs([]string{
		"https://playbooks.com/skills/owner/repo/alpha",
		"--dry-run",
		"-o", "./custom-out",
		"--category", "forced",
		"-",
		"https://playbooks.com/skills/owner/repo/beta",
	})
	if parseErr != nil {
		t.Fatalf("parseMainArgs returned error: %v", parseErr)
	}

	if !reflect.DeepEqual(parsed.Sources, []string{
		"https://playbooks.com/skills/owner/repo/alpha",
		"-",
		"https://playbooks.com/skills/owner/repo/beta",
	}) {
		t.Fatalf("unexpected sources: %#v", parsed.Sources)
	}
	if !parsed.Config.DryRun {
		t.Fatalf("expected dry-run to be enabled")
	}
	if parsed.Config.OutputDir != "./custom-out" {
		t.Fatalf("unexpected output dir: %q", parsed.Config.OutputDir)
	}
	if parsed.Config.Category != "forced" {
		t.Fatalf("unexpected category: %q", parsed.Config.Category)
	}
	if parsed.Config.AutoCategorize {
		t.Fatalf("expected auto-categorize to be disabled when category is forced")
	}
}

func TestCollectURLsReadsURLsFromURLFileAndStdinInOrder(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "urls.txt")
	fileContents := strings.Join([]string{
		"# comment",
		"",
		"https://playbooks.com/skills/owner/repo/from-file",
		"  https://playbooks.com/skills/owner/repo/from-file-2  # trailing comment",
		"",
	}, "\n")
	if err := os.WriteFile(sourceFile, []byte(fileContents), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	got, err := collectURLs(
		[]string{
			"https://playbooks.com/skills/owner/repo/inline",
			sourceFile,
			"-",
		},
		strings.NewReader(strings.Join([]string{
			"",
			"# stdin comment",
			"https://playbooks.com/skills/owner/repo/from-stdin",
			"  https://playbooks.com/skills/owner/repo/from-stdin-2  ",
			"",
		}, "\n")),
	)
	if err != nil {
		t.Fatalf("collectURLs returned error: %v", err)
	}

	want := []string{
		"https://playbooks.com/skills/owner/repo/inline",
		"https://playbooks.com/skills/owner/repo/from-file",
		"https://playbooks.com/skills/owner/repo/from-file-2",
		"https://playbooks.com/skills/owner/repo/from-stdin",
		"https://playbooks.com/skills/owner/repo/from-stdin-2",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("collectURLs mismatch\nwant: %#v\ngot:  %#v", want, got)
	}
}

func TestDedupeStringsPreservesFirstOccurrenceOrder(t *testing.T) {
	t.Parallel()

	got := dedupeStrings([]string{"a", "b", "a", "c", "b", "d"})
	want := []string{"a", "b", "c", "d"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dedupeStrings mismatch\nwant: %#v\ngot:  %#v", want, got)
	}
}
