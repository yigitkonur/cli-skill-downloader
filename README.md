# skill-dl

> Search, discover, and bulk-download AI coding skills from [playbooks.com](https://playbooks.com) in one command.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey.svg)]()
[![Shell](https://img.shields.io/badge/shell-bash%204%2B-green.svg)]()

Skills are structured markdown files (`SKILL.md` + references) that inject expert context into AI coding agents like **Claude Code**, **Cursor**, and **OpenCode**. `skill-dl` resolves playbooks.com URLs to their GitHub source, clones only what's needed, and organizes everything into categorized folders. v1.3.0 adds **Serper API** (Google-powered search) and **Scrapedo** proxy for enhanced discovery.

---

## Install

### One-liner (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/yigitkonur/cli-skill-downloader/main/install.sh | bash
```

Installs `skill-dl` to `/usr/local/bin` (or `~/.local/bin` as a fallback). Requires **Bash 4+** and **git**.

> **macOS users:** The system ships Bash 3. Upgrade with `brew install bash` first.

### Manual

```bash
git clone https://github.com/yigitkonur/cli-skill-downloader.git
cd cli-skill-downloader
chmod +x skill-dl
sudo ln -sf "$(pwd)/skill-dl" /usr/local/bin/skill-dl
```

### Requirements

| Requirement | Version |
|-------------|---------|
| `bash` | 4.0+ |
| `git` | any |
| macOS or Linux | — |

---

## Quick Start

```bash
# Download a single skill
skill-dl https://playbooks.com/skills/mcollina/skills/typescript-magician

# Download multiple skills
skill-dl \
  https://playbooks.com/skills/mcollina/skills/typescript-magician \
  https://playbooks.com/skills/nickcrew/claude-cortex/typescript-advanced-patterns \
  https://playbooks.com/skills/greyhaven-ai/claude-code-config/tdd-typescript

# Download from a file of URLs
skill-dl examples/typescript-skills.txt

# Pipe URLs from stdin
cat urls.txt | skill-dl -

# Preview without downloading
skill-dl urls.txt --dry-run

# Custom output directory
skill-dl urls.txt --output ./my-skills
```

---

## Usage

```
skill-dl <source...> [options]
```

### Sources (one or more, combinable)

| Source | Description |
|--------|-------------|
| `<url>` | `https://playbooks.com/skills/{owner}/{repo}/{skill}` |
| `<file>` | Text file with one URL per line (`#` comments supported) |
| `-` | Read URLs from stdin |

Mix freely: `skill-dl urls.txt https://playbooks.com/skills/...`

### Options

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output <dir>` | `./skills-collection` | Output directory |
| `-c, --category <name>` | auto | Force all skills into one category folder |
| `--no-auto-category` | off | Flat output, no category subfolders |
| `-f, --force` | off | Overwrite existing skill directories |
| `--dry-run` | off | Preview what would be downloaded |
| `-v, --verbose` | off | Show debug output |
| `-h, --help` | — | Show help |
| `--version` | — | Show version |

---

## Search & Discovery

```bash
# Search for skills by keyword (3-20 keywords required)
skill-dl search typescript react nextjs testing vitest

# Limit results
skill-dl search openclaw agent plugin workflow --top 20

# Only show skills matching 2+ keywords
skill-dl search mcp server sdk tools --min-match 2
```

### Search backends (auto-detected)

| Backend | How it works | When it activates |
|---------|-------------|-------------------|
| **Serper API** | Google search for `site:playbooks.com/skills` — broadest coverage | `SERPER_API_KEY` is set (built-in default included) |
| **Playbooks scrape** | Direct HTML scraping, 3 pages per keyword | Always active |
| **Scrapedo proxy** | Proxy fallback when direct requests are blocked | `SCRAPEDO_API_KEY` is set (built-in default included) |

Results are ranked by how many of your keywords each skill matched. Built-in API keys are included — override with environment variables if needed.

### Pipe search → download

```bash
# Search, extract URLs, download top 20
skill-dl search openclaw workflow cron --top 20 \
  | grep -oE 'https://playbooks\.com/skills/[^ |]+' \
  | skill-dl - -o ./my-skills --no-auto-category -f
```

---

## How It Works

```
playbooks.com/skills/{owner}/{repo}/{skill}
                        │       │       │
                        ▼       ▼       ▼
              github.com/{owner}/{repo}  →  git clone --depth 1
                                            │
                                            ▼
                                Find skill directory containing SKILL.md
                                (9 known paths + recursive name search)
                                            │
                                            ▼
                                Copy SKILL.md + references/ + rules/ etc.
                                            │
                                            ▼
                                skills-collection/{category}/{owner}--{repo}--{skill}/
```

**Key properties:**

| Property | Detail |
|----------|--------|
| No API rate limits | Uses `git clone --depth 1` (git protocol), not GitHub API |
| Clone once | Groups skills by repo — each repo is cloned only once |
| Precise extraction | Copies only the specific skill directory, not the whole repo |
| Smart discovery | 9 known paths + recursive SKILL.md search by directory name |
| Root-level repos | Single-skill repos are copied wholesale (exclusion-based) |
| Deduplication | Same URL listed twice is downloaded only once |

---

## What Gets Copied

**Subdirectory skills** (most common): the entire skill folder is copied as-is.

**Root-level skills** (repo = the skill): everything is copied **except** standard repo metadata:

| Excluded | Items |
|----------|-------|
| Dirs | `.git`, `.github`, `.gitlab`, `node_modules`, `.vscode`, `.idea` |
| Files | `README.md`, `LICENSE`, `CHANGELOG.md`, `CONTRIBUTING.md`, `.gitignore`, `package.json`, lockfiles |

Everything else — `SKILL.md`, `references/`, `rules/`, `examples/`, or any custom folder — is included automatically.

---

## Auto-Categorization

Skills are sorted into subfolders by name pattern:

| Folder | Matches |
|--------|---------|
| `react-typescript/` | `*react-typescript*`, `*react-ts*`, `*nextjs-*` |
| `strict-and-types/` | `*strict*`, `*advanced-type*`, `*type-expert*`, `*guardian*` |
| `sdk-and-libraries/` | `*sdk*`, `*ts-library*`, `*library*` |
| `pro-and-review/` | `*review*`, `*reviewer*`, `*magician*`, `*pro-skill*` |
| `tooling-and-setup/` | `*generator*`, `*setup*`, `*init*`, `*tdd-*`, `*tooling*` |
| `code-quality/` | `*coding-standard*`, `*clean-code*`, `*best-practice*` |
| `best-practices/` | `*typescript*`, `*expert*`, `*clean-typescript*` |
| `general/` | Everything else |

Override with `--category <name>` or disable with `--no-auto-category`.

---

## URL File Format

```bash
# my-skills.txt — comments and blank lines are ignored

https://playbooks.com/skills/mcollina/skills/typescript-magician
https://playbooks.com/skills/lobehub/lobehub/typescript

# React skills
https://playbooks.com/skills/madappgang/claude-code/react-typescript
```

See [`examples/typescript-skills.txt`](examples/typescript-skills.txt) for a ready-to-use collection.

---

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
│           └── conditional-types.md
├── tooling-and-setup/
│   └── greyhaven-ai--claude-code-config--tdd-typescript/
│       ├── SKILL.md
│       ├── checklists/
│       ├── examples/
│       └── templates/
└── strict-and-types/
    └── nickcrew--claude-cortex--typescript-advanced-patterns/
        └── SKILL.md
```

---

## Examples

```bash
# Download the full TypeScript collection from the examples file
skill-dl examples/typescript-skills.txt

# Multiple skills at once with verbose output
skill-dl \
  https://playbooks.com/skills/mcollina/skills/typescript-magician \
  https://playbooks.com/skills/jwynia/agent-skills/typescript-best-practices \
  --verbose

# All into a single folder (no auto-categorization)
skill-dl examples/typescript-skills.txt -c my-skills

# Custom output path + force overwrite
skill-dl urls.txt -o ~/ai-skills --force

# Dry run to preview before committing
skill-dl examples/typescript-skills.txt --dry-run

# Pipe from grep/awk/etc.
grep "react" my-urls.txt | skill-dl -
```

---

## License

MIT — see [LICENSE](LICENSE).
