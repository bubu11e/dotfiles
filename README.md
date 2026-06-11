# dotfiles

Personal macOS configuration managed with [chezmoi](https://www.chezmoi.io/).

Companion to the `ansible-desktop` repository: Ansible installs software and sets
system state; this repo owns the configuration files in `$HOME`.

## What's managed

| Target | Source |
|--------|--------|
| `~/.zshrc` | `dot_zshrc.tmpl` |
| `~/.gitconfig` | `dot_gitconfig.tmpl` |
| `~/.vimrc` | `dot_vimrc` |
| `~/.config/starship.toml` | `dot_config/starship.toml` |
| `~/.config/ghostty/config` | `dot_config/ghostty/config` |
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

## Bootstrap

```bash
# Installs chezmoi, clones this repo, prompts for per-machine values, and applies.
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply github.com/bubu11e/dotfiles
```

Or, if chezmoi is already installed:

```bash
chezmoi init --apply git@github.com:bubu11e/dotfiles.git
```

On first run you are prompted for a few per-machine values (git identity, optional
GPG signing key, optional Docker socket path). They are stored locally in
`~/.config/chezmoi/chezmoi.toml` and are **never** committed here.

## Secrets

This repository is public and contains **no secrets and no personal data**:

- Personal/per-machine values are prompted at `chezmoi init` (see `.chezmoi.toml.tmpl`)
  and kept only in the local chezmoi config.
- Docker registry authentication tokens are **not** managed here — they remain in the
  `ansible-desktop` repository's Ansible Vault.

## Daily use

```bash
chezmoi edit ~/.zshrc   # edit the source of a managed file
chezmoi diff            # preview pending changes
chezmoi apply           # apply changes to $HOME
chezmoi update          # pull the latest from the remote and apply
```
