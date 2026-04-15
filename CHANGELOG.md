# Changelog

All notable changes to `skill-dl` will be documented in this file.

The format is based on Keep a Changelog and the project uses SemVer tags for
GitHub Releases.

## [1.3.0] - 2026-04-15

### Added

- Go rewrite of `skill-dl` with parity-focused tests against the preserved bash
  reference implementation.
- Stable GitHub Release packaging via GoReleaser for macOS, Linux, and Windows
  on both `amd64` and `arm64`.
- Release asset verification scripts:
  - `scripts/build-release-artifacts.sh`
  - `scripts/verify-release-layout.sh`
  - `scripts/test-install-local-release.sh`
  - `scripts/compare-reference.sh`
- Curl-based installer that downloads prebuilt release assets, verifies
  checksums, and installs atomically.

### Changed

- Distribution now centers on GitHub Releases instead of a raw shell-script
  install path or local source builds.
- Release assets now include Windows downloads in `.zip` format alongside the
  existing `.tar.gz` archives for macOS and Linux.
- README installation guidance now covers direct per-platform release downloads.

### Fixed

- Direct reference comparison now runs the preserved bash implementation under
  its original command name so deterministic help/version output stays stable.
- Local installer verification now covers both explicit-version installs and the
  `/releases/latest` redirect path.
