# ADR-0002: Layer shell config by when the shell reads it

Status: Accepted

## Context

This setup targets zsh on macOS and bash on Linux. The naive split is by topic —
one file for git things, one for paths, one for aliases — which forces every file
to be sourced by every shell and makes it impossible to say when a given line runs.

The bug that split is designed to produce is an `export` in an interactive rc:
correct in the shell you are typing in, absent from `ssh host 'cmd'`, and absent
from anything a script spawns. It is invisible until something far away breaks.

zsh and bash disagree only on syntax, not on the shape of the problem. Duplicating
the shared 90% into two files means every future change is two edits and one
eventual divergence.

## Decision

Layer by *when* the shell reads the file, and share every layer that is not
shell-specific syntax:

| File | Read | Holds |
| --- | --- | --- |
| `~/.shell.env` | every invocation | exports: PATH, EDITOR, LANG, GOPATH, DOCKER_HOST |
| `~/.shell.aliases` | interactive only | aliases |
| `~/.shell.rc` | interactive only | interactive setup: ssh-agent, GPG_TTY |
| `~/.zshenv` + `~/.zshrc` | zsh | what only zsh can express |
| `~/.bash_profile` + `~/.bashrc` | bash | what only bash can express |

The three shared files are POSIX sh: no bashisms, no zsh syntax.

## Consequences

- "Where does this line go?" is answered by when it needs to run, not by what it is
  about. An export goes in the env layer even when it is git-related.
- The shared files must parse under `sh`, `zsh` and `bash`. This is enforced by
  `scripts/check-rendered-shell.sh` in CI and by shellcheck in the prek stack.
- One genuine incompatibility survives: zsh does not word-split unquoted
  parameters, so `_dotfiles_path_prepend` turns on `SH_WORD_SPLIT` under
  `LOCAL_OPTIONS` for its loop.
- `~/.shell.env` is sourced by `~/.bashrc` *before* its interactivity guard, so the
  non-interactive shell behind `ssh host 'cmd'` still gets a usable PATH.
- Every layer has a `.local` sibling, sourced last, for per-machine overrides that
  do not belong in a shared repository.
