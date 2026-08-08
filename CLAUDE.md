# CLAUDE.md

Project instructions for working on this repository. This is **not** a deployed
dotfile — it documents how to validate changes to the dotfiles themselves.

## What this repo is

Personal macOS configuration managed with [chezmoi](https://www.chezmoi.io/).
Source files map to targets in `$HOME`. There is no build step and no package
manager: the "code" is Go-template `.tmpl` files, a zsh rc, and YAML/TOML config.

## Layout

`.chezmoiroot` contains `home`, so chezmoi's source directory is `home/` and
**everything outside it is invisible to `chezmoi apply`**. That is why there is no
`.chezmoiignore`: repo-meta files (`README.md`, `CLAUDE.md`, `CONTEXT.md`,
`TODO.md`, `renovate.json`, `scripts/`, `docs/`, CI and lint config) sit at the
root and can never be deployed into `$HOME` by accident.

```
home/                  # the chezmoi source directory — everything here is deployed
home/.chezmoi.toml.tmpl  # the machine profile (prompts); must live inside home/
<repo root>            # docs, scripts, CI, lint config — never deployed
```

Adding a file to be deployed means putting it under `home/`. Adding a repo-meta
file means putting it at the root and doing nothing else.

## Validate (lint)

```bash
prek run --all-files
```

Runs the `.pre-commit-config.yaml` stack: end-of-file-fixer, trailing-whitespace,
check-added-large-files, check-merge-conflict, yamllint (`--strict -c .yamllint`),
gitleaks, and shellcheck. `prek` is the project's pre-commit runner.

shellcheck runs twice, once per dialect, matching files by path rather than by
`types: [shell]` — the deployed files have chezmoi's `dot_` prefix, no extension
and no shebang, so `identify` cannot classify them. The zsh files are deliberately
excluded (shellcheck cannot parse zsh), as is `home/dot_shell.env.tmpl`, where
shellcheck would see Go template syntax rather than shell.

## Validate (template render)

This is what CI (`.woodpecker.yml`) does — it renders every template with default
prompt values and applies as a dry run, catching template/syntax errors:

```bash
chezmoi init --promptDefaults
chezmoi apply --dry-run --verbose
bash scripts/check-rendered-shell.sh
```

The machine profile (the per-machine prompted values) is defined only in
`home/.chezmoi.toml.tmpl`; `--promptDefaults` lets CI render without restating it.

The dry run proves the templates *render*; it says nothing about whether what they
render is valid shell. `scripts/check-rendered-shell.sh` closes that gap: it renders
`~/.shell.env`, `~/.shell.aliases` and `~/.shell.rc` and parses each with `sh`,
`bash` and `zsh`. `~/.shell.env` is the reason it exists — it is sourced by every
shell invocation including login, and it is the one shell file shellcheck cannot
read in place.

## Daily use

See the **Daily use** section of `README.md` for the day-to-day
`chezmoi edit` / `diff` / `apply` / `update` workflow.

## Conventions and decisions

- Domain vocabulary: `CONTEXT.md`.
- Architecture decisions: `docs/adr/` — start at its `README.md` index. The
  conventions listed below are the summary; the ADR is the reasoning and the
  measurements behind it.
- `TODO.md` is deliberately gitignored and machine-local. This is a decided
  exception to the global "update and commit the TODO alongside the code" rule —
  edit it freely, never `git add` it, and do not propose tracking it again.
- Each shell snippet in `home/dot_zshrc` guards on its tool being present, so a
  missing tool is skipped rather than erroring.
- Shell config is layered by **when** the shell reads it, and zsh and bash share
  every layer that is not shell-specific syntax:
  `home/dot_shell.env.tmpl` (every invocation — PATH, EDITOR, LANG, GOPATH,
  DOCKER_HOST), `home/dot_shell.aliases` and `home/dot_shell.rc` (interactive,
  both shells), then `home/dot_zshenv` + `home/dot_zshrc` and
  `home/dot_bash_profile` + `home/dot_bashrc` for what only one shell can express.
  Putting an export in an rc is the bug this split exists to prevent.
- The three shared files are **POSIX sh** — no bashisms, no zsh syntax. Verified by
  `scripts/check-rendered-shell.sh`, which parses each with all three shells; run it
  rather than checking by hand. The one place
  the shells genuinely differ is word splitting: zsh does not split unquoted
  parameters, so `_dotfiles_path_prepend` turns on `SH_WORD_SPLIT` under
  `LOCAL_OPTIONS` for its loop.
- Branch on `$OSTYPE`, never `uname`. Both bash and zsh set it for free, so it costs
  nothing; `uname -s` is a fork on every shell start.
- `home/dot_bashrc` targets Linux and bash 5.x. macOS is secondary and still ships
  bash 3.2, so only the 4.0+ `shopt` names are guarded on `BASH_VERSINFO` — an
  unknown option there prints to stderr on every start.
- Set `LANG` only where the value is known to exist (macOS), never `LC_ALL`, and
  never on Linux: a minimal image has no generated locale beyond `C.utf8`, and
  naming a missing one makes bash warn on every start. Per-host overrides go in
  `~/.shell.env.local`.
- Shell completion comes from Homebrew's `share/zsh/site-functions` (already on
  `$fpath` via `brew shellenv`). Prefer that over a `<tool> completion zsh` fork at
  startup — `kubectl` alone cost 141ms.
