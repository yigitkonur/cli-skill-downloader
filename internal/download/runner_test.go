package download

import (
	"strings"
	"testing"
)

type fakeCloner struct {
	repos map[string]string
	calls []string
	errs  map[string]error
}

func (f *fakeCloner) Clone(repoKey string, dest string) error {
	f.calls = append(f.calls, repoKey)
	if err := f.errs[repoKey]; err != nil {
		return err
	}
	return copyEntry(f.repos[repoKey], dest)
}

func TestProcessorRunClonesEachRepoOnceAndCopiesMultipleSkills(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	mustWriteFile(t, repoDir+"/skills/alpha/SKILL.md", "alpha")
	mustWriteFile(t, repoDir+"/skills/beta/SKILL.md", "beta")

	cloner := &fakeCloner{
		repos: map[string]string{"owner/repo": repoDir},
	}
	outputDir := t.TempDir()
	processor := Processor{Git: cloner}

	result := processor.Run([]string{
		"https://playbooks.com/skills/owner/repo/alpha",
		"https://playbooks.com/skills/owner/repo/beta",
	}, Options{OutputDir: outputDir, AutoCategorize: true}, &strings.Builder{})

	if result.OK != 2 || result.Failed != 0 || result.Skipped != 0 || result.Total != 2 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if len(cloner.calls) != 1 || cloner.calls[0] != "owner/repo" {
		t.Fatalf("expected one clone call for owner/repo, got %#v", cloner.calls)
	}
}

func TestProcessorRunSkipsExistingDestinationWithoutForce(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	mustWriteFile(t, repoDir+"/skills/alpha/SKILL.md", "alpha")

	outputDir := t.TempDir()
	mustWriteFile(t, outputDir+"/general/owner--repo--alpha/existing.txt", "keep")

	cloner := &fakeCloner{
		repos: map[string]string{"owner/repo": repoDir},
	}
	processor := Processor{Git: cloner}

	result := processor.Run([]string{
		"https://playbooks.com/skills/owner/repo/alpha",
	}, Options{OutputDir: outputDir, AutoCategorize: true}, &strings.Builder{})

	if result.OK != 0 || result.Skipped != 1 || result.Failed != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestProcessorRunMarksFailuresForMissingSkillsAndCloneErrors(t *testing.T) {
	t.Parallel()

	repoDir := t.TempDir()
	mustWriteFile(t, repoDir+"/skills/existing/SKILL.md", "existing")

	cloner := &fakeCloner{
		repos: map[string]string{
			"owner/repo":  repoDir,
			"owner/bad":   repoDir,
			"owner/empty": t.TempDir(),
		},
		errs: map[string]error{
			"owner/bad": assertiveError("clone failed"),
		},
	}
	processor := Processor{Git: cloner}

	result := processor.Run([]string{
		"https://playbooks.com/skills/owner/repo/missing",
		"https://playbooks.com/skills/owner/bad/alpha",
	}, Options{OutputDir: t.TempDir(), AutoCategorize: true}, &strings.Builder{})

	if result.Failed != 2 || result.OK != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
	if len(result.FailedURLs) != 2 {
		t.Fatalf("expected failed URLs to be recorded, got %#v", result.FailedURLs)
	}
}

type assertiveError string

func (e assertiveError) Error() string { return string(e) }
