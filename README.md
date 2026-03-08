# skill-dl

Download AI coding skills from [playbooks.com](https://playbooks.com) with a single command.

Skills are structured markdown files (`SKILL.md` + references) hosted on GitHub that provide context to AI coding agents like Claude Code, Cursor, and OpenCode. This tool bulk-downloads them by resolving playbooks.com URLs to their GitHub source, cloning only what's needed, and organizing them into categorized folders.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/yigitkonur/cli-skill-downloader/main/skill-dl \
  -o /usr/local/bin/skill-dl && chmod +x /usr/local/bin/skill-dl
```

Or clone the repo:

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

# Download multiple skills inline
skill-dl \
  https://playbooks.com/skills/mcollina/skills/typescript-magician \
  https://playbooks.com/skills/nickcrew/claude-cortex/typescript-advanced-patterns \
  https://playbooks.com/skills/greyhaven-ai/claude-code-config/tdd-typescript

# Download from a file of URLs
skill-dl urls.txt

# Mix URLs and files
skill-dl urls.txt https://playbooks.com/skills/inkeep/skills/typescript-sdk

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
- **Root-level repos** — for single-skill repos, copies everything except repo metadata (exclusion-based, not a hardcoded whitelist — any folder name works)
- **Deduplicates** — same URL listed twice is only downloaded once

## What Gets Downloaded

**Subdirectory skills** (most common): the entire skill directory is copied as-is.

**Root-level skills** (repo IS the skill): everything is copied **except** known repo metadata:

| Excluded | Items |
|----------|-------|
| Directories | `.git`, `.github`, `.gitlab`, `node_modules`, `.vscode`, `.idea` |
| Files | `README.md`, `LICENSE`, `CHANGELOG.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `.gitignore`, `.gitattributes`, `package.json`, lockfiles |

Everything else — `SKILL.md`, `references/`, `rules/`, `examples/`, or any custom folder the skill author creates — is included automatically.

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
# Download a single skill
skill-dl https://playbooks.com/skills/mcollina/skills/typescript-magician

# Download multiple skills at once
skill-dl \
  https://playbooks.com/skills/mcollina/skills/typescript-magician \
  https://playbooks.com/skills/jwynia/agent-skills/typescript-best-practices \
  https://playbooks.com/skills/onmax/nuxt-skills/ts-library \
  https://playbooks.com/skills/exploration-labs/typescript-code-review/typescript-code-review

# From a file of URLs
skill-dl urls.txt -o ~/projects/ai-skills

# Mix files and inline URLs
skill-dl urls.txt \
  https://playbooks.com/skills/inkeep/skills/typescript-sdk \
  https://playbooks.com/skills/greyhaven-ai/claude-code-config/tdd-typescript

# Pipe from another command
cat urls.txt | skill-dl -

# Force all into one category folder
skill-dl urls.txt -c my-typescript-skills

# Dry run to preview
skill-dl urls.txt --dry-run

# Re-download everything (overwrite existing)
skill-dl urls.txt --force

# Verbose debug output
skill-dl https://playbooks.com/skills/inkeep/skills/typescript-sdk -v
```

## License

MIT
