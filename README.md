# dotfiles

Personal macOS configuration managed with [chezmoi](https://www.chezmoi.io/).

Companion to the `ansible-desktop` repository: Ansible installs software and sets
system state; this repo owns the configuration files in `$HOME`.

## What's managed

`.chezmoiroot` points chezmoi at `home/`, so only what lives under `home/` is ever
deployed. Everything at the repo root (this README, `CLAUDE.md`, `CONTEXT.md`,
`scripts/`, CI and lint config) is invisible to `chezmoi apply` by construction.

Shell configuration is layered by **when the shell reads it**, not by topic, so zsh
and bash share everything that is not shell-specific syntax:

| Layer | Read when | Shared? |
|---|---|---|
| `~/.shell.env` | every invocation, interactive or not | POSIX, both shells |
| `~/.shell.aliases` | interactive only | POSIX, both shells |
| `~/.shell.rc` | interactive only | POSIX, both shells |
| `~/.zshenv`, `~/.zshrc` | zsh | zsh syntax only |
| `~/.bash_profile`, `~/.bashrc` | bash | bash syntax only |

An `export` belongs in `~/.shell.env`. Putting it in an rc is the bug the split
exists to prevent: `zsh -c ...`, git hooks and IDE subprocesses never read an rc.

| Target | Source (under `home/`) |
|--------|--------|
| `~/.shell.env` | `dot_shell.env.tmpl` (PATH, EDITOR, LANG, GOPATH, DOCKER_HOST) |
| `~/.shell.aliases` | `dot_shell.aliases` |
| `~/.shell.rc` | `dot_shell.rc` (ssh-agent reuse, GPG_TTY) |
| `~/.zshenv` | `dot_zshenv` (zsh wrapper: `typeset -U`, completion `$fpath`) |
| `~/.zshrc` | `dot_zshrc` (compinit, plugins, key bindings, prompt) |
| `~/.bash_profile` | `dot_bash_profile` (forwards to `~/.bashrc`) |
| `~/.bashrc` | `dot_bashrc` (history, completion, prompt) |
| `~/.gitconfig` | `dot_gitconfig.tmpl` |
| `~/.vimrc` | `dot_vimrc` |
| `~/.config/starship.toml` | `dot_config/starship.toml` |
| `~/.config/ghostty/config` | `dot_config/ghostty/config` |
| `~/.config/bat/config` | `dot_config/bat/config` |
| `~/.config/goose/config.yaml` | `dot_config/goose/private_config.yaml` |
| `~/.terraformrc` | `dot_terraformrc.tmpl` |
| `~/.ssh/config` | `private_dot_ssh/private_config` |
| `~/.gnupg/gpg.conf` | `private_dot_gnupg/private_gpg.conf.tmpl` |
| `~/.gnupg/gpg-agent.conf` | `private_dot_gnupg/private_gpg-agent.conf` |
| `~/.claude/CLAUDE.md` | `dot_claude/CLAUDE.md` (global Claude Code instructions) |
| `~/.claude/skills/` | `dot_claude/skills/` (vendored Claude Code skills) |

The config references tools (starship, fzf, mise, zsh plugins, etc.) installed by
`ansible-desktop`. Each shell snippet guards on the tool being present, so a missing
tool is silently skipped rather than erroring.

Every managed file that supports an include ends by sourcing a machine-local
counterpart that is deliberately untracked: `~/.shell.env.local`,
`~/.shell.aliases.local`, `~/.shell.rc.local`, `~/.zshenv.local`, `~/.zshrc.local`,
`~/.bashrc.local`, `~/.gitconfig.local`, `~/.ssh/config.local`.

## Bootstrap

Clone the repo and run the bootstrap script. It installs chezmoi if it is missing,
then runs `chezmoi init --apply` against this checkout — no remote URL is baked in,
which matters because the canonical remote is a self-hosted Forgejo rather than a
GitHub shorthand chezmoi understands.

```bash
git clone ssh://git@forgejo.home.mpli.fr/julien/dotfiles.git
./dotfiles/install.sh
```

Or, if chezmoi is already installed and you would rather let it do the cloning:

```bash
chezmoi init --apply ssh://git@forgejo.home.mpli.fr/julien/dotfiles.git
```

On first run you are prompted for a few per-machine values (git identity, optional
GPG signing key, optional Docker socket path). They are stored locally in
`~/.config/chezmoi/chezmoi.toml` and are **never** committed here.

## Secrets

This repository is public and contains **no secrets and no personal data**:

- Personal/per-machine values are prompted at `chezmoi init` (see
  `home/.chezmoi.toml.tmpl`) and kept only in the local chezmoi config.
- Docker registry authentication tokens are **not** managed here — they remain in the
  `ansible-desktop` repository's Ansible Vault.

## Daily use

```bash
chezmoi edit ~/.zshrc   # edit the source of a managed file
chezmoi diff            # preview pending changes
chezmoi apply           # apply changes to $HOME
chezmoi update          # pull the latest from the remote and apply
```
