package cli

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type parityCase struct {
	name string
	args []string
}

func TestParityDeterministicFixtures(t *testing.T) {
	t.Parallel()

	cases := []parityCase{
		{name: "help", args: []string{"--help"}},
		{name: "version", args: []string{"--version"}},
		{name: "no-args"},
		{name: "invalid-source", args: []string{"definitely-not-a-real-source"}},
		{name: "search-help", args: []string{"search", "--help"}},
		{name: "search-too-few-keywords", args: []string{"search", "typescript", "react"}},
		{name: "dry-run-single-url", args: []string{"https://playbooks.com/skills/mcollina/skills/typescript-magician", "--dry-run"}},
		{name: "dry-run-mixed-sources", args: []string{"examples/typescript-skills.txt", "https://playbooks.com/skills/nickcrew/claude-cortex/typescript-advanced-patterns", "--dry-run", "--no-auto-category"}},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			wantStdout := readFixture(t, tc.name+".stdout")
			wantStderr := readFixture(t, tc.name+".stderr")
			wantExit := strings.TrimSpace(readFixture(t, tc.name+".exit"))

			var stdout strings.Builder
			var stderr strings.Builder

			exitCode := Run(resolveArgs(t, tc.args), strings.NewReader(""), &stdout, &stderr)

			if got := stdout.String(); got != wantStdout {
				t.Fatalf("stdout mismatch for %s\nwant:\n%s\ngot:\n%s", tc.name, wantStdout, got)
			}
			if got := stderr.String(); got != wantStderr {
				t.Fatalf("stderr mismatch for %s\nwant:\n%s\ngot:\n%s", tc.name, wantStderr, got)
			}
			if got := strconv.Itoa(exitCode); got != wantExit {
				t.Fatalf("exit mismatch for %s: want %s got %d", tc.name, wantExit, exitCode)
			}
		})
	}
}

func resolveArgs(t *testing.T, args []string) []string {
	t.Helper()

	resolved := make([]string, len(args))
	copy(resolved, args)

	for index, arg := range resolved {
		if strings.HasPrefix(arg, "examples/") {
			resolved[index] = filepath.Join("..", "..", arg)
		}
	}

	return resolved
}

func readFixture(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", "parity", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}

	return string(data)
}
