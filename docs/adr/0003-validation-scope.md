# Validation is dry-run plus lint, not shell unit tests or a shell linter

Changes to this repo are validated by `prek` (the pre-commit stack) and a chezmoi
render check (`chezmoi apply --dry-run` in CI). We deliberately do not add a shell
linter or a shell test harness.

A shell linter is a poor fit: `dot_zshrc.tmpl` is a Go template, so `shellcheck`
cannot parse the `{{ }}` directives, and even on rendered output `shellcheck` does
not support zsh (its target shells are sh/bash/ksh/dash). A unit-test harness
(e.g. bats) for shell functions such as `_ssh_start_agent` is disproportionate to
a small personal dotfiles repo whose failure mode is a noisy interactive shell,
not a production outage. The render dry-run already catches the realistic break:
template and syntax errors before they reach a machine.

Revisit if the shell surface grows substantially (many functions with real
branching logic), at which point extracting that logic into a testable script
deployed as a plain file — not a template — would make a bats suite worthwhile.
