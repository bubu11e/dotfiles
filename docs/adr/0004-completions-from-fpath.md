# ADR-0004: Take completions from `$fpath`, not from a startup fork

Status: Accepted

## Context

The documented way to get shell completion for most CLI tools is to evaluate their
own generator at startup:

```zsh
source <(kubectl completion zsh)
```

That runs the tool, on every single shell start, to produce output that changes
only when the tool is upgraded. Measured here, `kubectl` alone cost 141ms. Several
tools doing the same is most of the difference between a shell that feels instant
and one that does not; startup went from 263ms to 126ms when they were removed.

Homebrew already installs completion functions into
`share/zsh/site-functions`, which `brew shellenv` puts on `$fpath`. zsh autoloads
from `$fpath` lazily — the function is read the first time it is needed, not at
startup.

## Decision

Prefer the `$fpath` function that Homebrew already ships. Fall back to a
`<tool> completion zsh` fork only when the tool ships no completion function at
all, and say so in a comment at the call site.

Reuse a bundled completer rather than generating a new one where the interface is
shared: `compdef nvim=vim`, `compdef gsed=sed`, `compdef pre-commit=prek`.

`compinit` itself is cached: its security audit of every `$fpath` directory runs in
full only when `~/.zcompdump` is more than 24 hours old, and `-C` reuses the cache
otherwise.

## Consequences

- Adding a tool to this setup means wiring up its completion too, from `$fpath`
  first. This is the step that is easy to forget and expensive to get wrong.
- bash gets none of this: it has no `$fpath` and cannot autoload, so completion
  functions are sourced, not looked up. `~/.bashrc` therefore defers
  `kubectl completion bash` to first use with a self-replacing stub — a shell that
  never types `kubectl` never pays.
- The remaining startup forks are the ones with no static equivalent, and each is
  commented as such: `starship init`, `mise activate`, `fzf --zsh`.
- This is the load-bearing reason dropping Homebrew is not a simple swap. Every
  completion currently comes from its `site-functions` directory, so a replacement
  has to be designed at the same time, not afterwards.
