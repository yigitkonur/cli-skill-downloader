# skill-dl

Download AI coding skills from [playbooks.com](https://playbooks.com) with a single command.

Skills are structured markdown files (`SKILL.md` + references) hosted on GitHub that provide context to AI coding agents like Claude Code, Cursor, and OpenCode. This tool bulk-downloads them by resolving playbooks.com URLs to their GitHub source, cloning only what's needed, and organizing them into categorized folders.

## Install

```bash
git clone https://github.com/yigitkonur/cli-skill-downloader.git
cd cli-skill-downloader
chmod +x skill-dl
sudo ln -sf "$(pwd)/skill-dl" /usr/local/bin/skill-dl
```

**Requirements:** `git`, `bash` 4+, `find` (macOS/Linux)

## Quick Start

```bash
# Download a single skill
skill-dl https://playbooks.com/skills/mcollina/skills/typescript-magician

# Download from a file of URLs
skill-dl urls.txt

# Pipe URLs
cat urls.txt | skill-dl -

# With options
skill-dl urls.txt --output ./my-skills --force --verbose
```

## Usage

```
skill-dl <source> [options]
```

### Source (required)

| Source | Description |
|--------|------------|
| `<url>` | Single playbooks.com skill URL |
| `<file>` | Text file with one URL per line (`#` comments supported) |
| `-` | Read URLs from stdin |

Multiple sources can be combined: `skill-dl url1 url2 urls.txt`

### Options

| Flag | Description | Default |
|------|-------------|---------|
| `-o, --output <dir>` | Output directory | `./skills-collection` |
| `-c, --category <name>` | Force all skills into one category folder | auto |
| `--no-auto-category` | Flat structure, no category folders | off |
| `-f, --force` | Overwrite existing skill directories | off |
| `--dry-run` | Preview what would be downloaded | off |
| `-v, --verbose` | Show debug output | off |
| `-h, --help` | Show help | — |
| `--version` | Show version | — |

## How It Works

```
playbooks.com/skills/{owner}/{repo}/{skill}
                        │       │       │
                        ▼       ▼       ▼
              github.com/{owner}/{repo} → git clone --depth 1
                                          │
                                          ▼
                              Find the skill directory containing SKILL.md
                              (9 known paths + recursive search by dir name)
                                          │
                                          ▼
                              Copy SKILL.md + references/ + rules/ + examples/ etc.
                                          │
                                          ▼
                              skills-collection/{category}/{owner}--{repo}--{skill}/
```

**Key details:**

- **No API rate limits** — uses `git clone --depth 1` (git protocol), not the GitHub API
- **Clone once** — groups skills by repo so each repo is cloned only once
- **Precise extraction** — only copies the specific skill directory, not the whole repo
- **Smart discovery** — finds skills across 9+ common locations (`.claude/skills/`, `.agent/skills/`, `.cursor/skills/`, `plugins/*/skills/`, deeply nested paths, etc.)
- **Root-level repos** — for repos that ARE a single skill, selectively copies `SKILL.md` + known subdirs, skipping repo metadata (`README.md`, `LICENSE`, `.github/`)
- **Deduplicates** — same URL listed twice is only downloaded once

## What Gets Downloaded

Each skill extracts **only** the files that belong to it:

| File/Dir | Description |
|----------|-------------|
| `SKILL.md` | Main skill file (always present) |
| `references/` | Reference documentation, guides |
| `rules/` | Rule files for the skill |
| `examples/` | Example code and documents |
| `assets/` | Templates, configs, presets |
| `scripts/` | Automation scripts |
| `templates/` | Template files |
| `checklists/` | Checklist documents |
| `patterns/` | Pattern documentation |
| `build/`, `setup/`, `workflows/` | Build and setup docs |

**Not included:** `README.md`, `LICENSE`, `.github/`, `.git/`, other skills in the same repo.

## Auto-Categorization

Skills are auto-sorted by name into folders:

| Category | Matches |
|----------|---------|
| `strict-and-types/` | `*strict*`, `*advanced-type*`, `*type-expert*`, `*guardian*`, `*advanced-pattern*` |
| `best-practices/` | `*typescript*`, `*expert*`, `*clean-typescript*` |
| `code-quality/` | `*coding-standard*`, `*clean-code*`, `*code-style*`, `*best-practice*` |
| `react-typescript/` | `*react-typescript*`, `*react-ts*`, `*nextjs-*` |
| `pro-and-review/` | `*pro-skill*`, `*reviewer*`, `*code-review*`, `*magician*` |
| `sdk-and-libraries/` | `*sdk*`, `*ts-library*`, `*typescript-v*` |
| `tooling-and-setup/` | `*generator*`, `*setup*`, `*init*`, `*tdd-*`, `*tooling*` |
| `general/` | Everything else |

Override: `--category <name>` or `--no-auto-category`

## URL File Format

```bash
# my-skills.txt
# Comments and blank lines are ignored

https://playbooks.com/skills/mcollina/skills/typescript-magician
https://playbooks.com/skills/lobehub/lobehub/typescript

# React skills
https://playbooks.com/skills/madappgang/claude-code/react-typescript
```

## Output Structure

```
skills-collection/
├── best-practices/
│   └── lobehub--lobehub--typescript/
│       └── SKILL.md
├── pro-and-review/
│   └── mcollina--skills--typescript-magician/
│       ├── SKILL.md
│       └── rules/
│           ├── generics-basics.md
│           ├── conditional-types.md
│           └── ...
├── tooling-and-setup/
│   └── greyhaven-ai--claude-code-config--tdd-typescript/
│       ├── SKILL.md
│       ├── checklists/
│       ├── examples/
│       ├── reference/
│       └── templates/
└── ...
```

## Examples

```bash
# Dry run to see what would be downloaded
skill-dl urls.txt --dry-run

# Download into a custom directory
skill-dl urls.txt -o ~/projects/ai-skills

# Force re-download everything
skill-dl urls.txt --force

# Put all skills in a single folder
skill-dl urls.txt -c all-skills

# Verbose debug output
skill-dl https://playbooks.com/skills/inkeep/skills/typescript-sdk -v

# Combine multiple sources
skill-dl skill1-url skill2-url more-urls.txt
```

## License

MIT
