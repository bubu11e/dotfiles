# CLAUDE.md

Project instructions for working on this repository. This is **not** a deployed
dotfile — it documents how to validate changes to the dotfiles themselves.

> Note: it is listed in `.chezmoiignore` (alongside `README.md`, `CONTEXT.md`,
> and `docs/`) so `chezmoi apply` does not copy it into `$HOME`.

## What this repo is

Personal macOS configuration managed with [chezmoi](https://www.chezmoi.io/).
Source files map to targets in `$HOME`. There is no build step and no package
manager: the "code" is Go-template `.tmpl` files, a zsh rc, and YAML/TOML config.

## Validate (lint)

```bash
prek run --all-files
```

Runs the `.pre-commit-config.yaml` stack: end-of-file-fixer, trailing-whitespace,
check-added-large-files, check-merge-conflict, yamllint (`--strict -c .yamllint`),
and gitleaks. `prek` is the project's pre-commit runner.

## Validate (template render)

This is what CI (`.woodpecker.yml`) does — it renders every template with default
prompt values and applies as a dry run, catching template/syntax errors:

```bash
chezmoi init --promptDefaults
chezmoi apply --dry-run --verbose
```

The machine profile (the per-machine prompted values) is defined only in
`.chezmoi.toml.tmpl`; `--promptDefaults` lets CI render without restating it.

## Daily use

See the **Daily use** section of `README.md` for the day-to-day
`chezmoi edit` / `diff` / `apply` / `update` workflow.

## Conventions and decisions

- Domain vocabulary: `CONTEXT.md`.
- Architecture decisions: `docs/adr/`.
- Each shell snippet in `dot_zshrc.tmpl` guards on its tool being present, so a
  missing tool is skipped rather than erroring.
