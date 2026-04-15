package cli

const Version = "1.3.0"

const mainHelpText = `skill-dl v1.3.0 — Download skills from playbooks.com (with Serper + Scrapedo)

USAGE
  skill-dl <source> [options]

SOURCE (required, one or more)
  <url>           Single playbooks.com URL
  <file>          Text file with one URL per line (# comments allowed)
  -               Read URLs from stdin (pipe-friendly)

OPTIONS
  -o, --output <dir>       Output directory (default: ./skills-collection)
  -c, --category <name>    Force all skills into this category folder
  --no-auto-category       Flat structure, no category folders
  -f, --force              Overwrite existing skill directories
  --dry-run                Preview without downloading
  -v, --verbose            Debug output
  -h, --help               Show this help
  --version                Show version

ENVIRONMENT
  SERPER_API_KEY           Google search via serper.dev (built-in default)
  SCRAPEDO_API_KEY         Scraping proxy via scrapedo.com (built-in default)

URL FORMAT
  https://playbooks.com/skills/{owner}/{repo}/{skill-name}

WHAT GETS DOWNLOADED
  Subdirectory skills: the entire skill directory is copied.
  Root-level skills:   everything except repo metadata is copied.

  Excluded repo metadata:
    .git, .github, .gitlab, node_modules, .vscode, .idea
    README.md, LICENSE, CHANGELOG.md, CONTRIBUTING.md,
    .gitignore, package.json, lockfiles

  Everything else (SKILL.md, references/, rules/, examples/,
  or any custom folder) is included automatically.

EXAMPLES
  # Single skill
  skill-dl https://playbooks.com/skills/mcollina/skills/typescript-magician

  # Multiple skills inline
  skill-dl \
    https://playbooks.com/skills/mcollina/skills/typescript-magician \
    https://playbooks.com/skills/nickcrew/claude-cortex/typescript-advanced-patterns \
    https://playbooks.com/skills/greyhaven-ai/claude-code-config/tdd-typescript

  # From a file (one URL per line, # comments supported)
  skill-dl urls.txt -o ./my-skills

  # Mix URLs and files
  skill-dl urls.txt https://playbooks.com/skills/inkeep/skills/typescript-sdk

  # Pipe from another command
  cat urls.txt | skill-dl -

  # Force all into one category folder
  skill-dl urls.txt -c my-typescript-skills

  # Dry run to preview what would be downloaded
  skill-dl urls.txt --dry-run

  # Re-download everything (overwrite existing)
  skill-dl urls.txt --force
`

const searchHelpText = `skill-dl search — Discover skills via Google (Serper) + playbooks.com

USAGE
  skill-dl search <keyword1> <keyword2> ... [options]

  Requires 3-20 keywords. Searches using all available backends,
  then ranks skills by how many keywords they match.

SEARCH BACKENDS (auto-detected)
  Serper API    Google-powered, finds skills across all of playbooks.com
                Set SERPER_API_KEY env var or uses built-in key
  Playbooks     Direct HTML scraping (always active, 3 pages per keyword)
  Scrapedo      Proxy fallback when direct requests are blocked
                Set SCRAPEDO_API_KEY env var or uses built-in key

OPTIONS
  --top <N>         Show only top N results (default: all)
  --min-match <N>   Only show skills matching N+ keywords (default: 1)
  -h, --help        Show this help

EXAMPLES
  skill-dl search typescript react nextjs testing vitest
  skill-dl search "browser automation" playwright puppeteer --top 20
  skill-dl search mcp server sdk client tools --min-match 2
  SERPER_API_KEY=xxx skill-dl search openclaw agent skill --top 30
`
