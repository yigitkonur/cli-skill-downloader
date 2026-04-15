# Bash Parity Contract

This file freezes the observable behavior of the original bash implementation
before the Go rewrite replaces it.

## Deterministic command fixtures

The fixtures under `testdata/parity/` were generated from the original bash
implementation preserved at `testdata/reference/skill-dl-bash` with:

```bash
./scripts/capture-parity-fixtures.sh
```

Each case stores three files:

- `*.stdout`
- `*.stderr`
- `*.exit`

The current baseline cases are:

- `help`
- `version`
- `no-args`
- `invalid-source`
- `search-help`
- `search-too-few-keywords`
- `dry-run-single-url`
- `dry-run-mixed-sources`

## Non-deterministic behaviors that still need parity

These paths depend on live network, temp paths, repo contents, or local install
permissions, so they are frozen here as behavioral expectations rather than
exact text goldens.

### Download mode

- Groups requested skills by `owner/repo` and clones each GitHub repo at most
  once.
- Uses `git clone --depth 1 --quiet https://github.com/{owner}/{repo}.git`.
- Resolves the skill directory using known paths first, then a recursive
  `SKILL.md` search by matching the parent directory name, then root-level
  `SKILL.md`.
- Copies subdirectory skills wholesale.
- Copies root-level skills with an exclusion filter that drops repo metadata
  such as `.git`, `.github`, `README.md`, `LICENSE`, and JS lockfiles.
- Writes progress and summary lines to `stderr`.
- Returns exit code `0` only when every requested skill succeeds or is skipped
  without failure; any failed URL makes the command exit non-zero.

### Search mode

- Requires 3 to 20 keywords.
- Aggregates results from Serper, direct playbooks scraping, and Scrapedo
  fallback when keys are present.
- Writes progress logs to `stderr`.
- Writes only the markdown result table to `stdout`.
- Ranks by descending keyword match count, then ascending skill path.

### Install mode

- `install.sh` requires macOS or Linux, Bash 4+, and `git`.
- It downloads a single executable named `skill-dl` from the raw GitHub URL.
- It prefers `/usr/local/bin`, otherwise falls back to `~/.local/bin`.
- It verifies the downloaded file looks like a bash script before installing.
- It reports the installed binary version by invoking `skill-dl --version`.
