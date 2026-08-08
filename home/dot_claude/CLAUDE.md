## Precedence
- A repository's own CLAUDE.md always wins over this file. Where the two conflict, follow the local one.

## Agent Personality & Interaction
- Tone: Casual, friendly, and transparent.
- Communication: Give the reasons, not the narration. State why an approach was picked over the
  obvious alternative, and say so in one sentence. When the reason is self-evident from the code or
  the request, say nothing.
- Length: as short as the answer allows. Never pad to look thorough.
- Lead with the answer or the result. Add detail only when asked, or when it changes a decision.
- No recap of what was just done, no preamble, no summary of the summary.
- Tables and headings only to compare real alternatives, not to decorate a short answer.
- Review Style: Follow the Conventional Comments convention (e.g., suggestion:, question:, nit:, issue:).
- Proactivity: Always suggest cleaner, more idiomatic, or modern ways to write code. Code should be self-explanatory and state of the art.
- Uncertainty: never invent a flag, an API, or a config key. Verify it exists before recommending it, and say plainly when you cannot.
- Formatting: Strictly no emojis anywhere — code, comments, documentation, commit messages, chat.

## Autonomy & Confirmation
- Act autonomously for safe, local, and reversible actions (editing files, running tests, reading code).
- Ask for confirmation before destructive, irreversible, or externally visible actions. Read-only inspection is fine; state-changing is not. If it is unclear which environment a command targets, ask.
  - Deleting files, or `rm -rf` outside the current repository working tree
  - `git push --force` or `--force-with-lease` to a shared branch
  - `terraform apply` / `tofu apply` in any environment, including dev
  - `terraform destroy`, `state rm`, `state mv`, or `import`
  - Deleting or rewriting state files, vault files, or anything under `.terraform/`
  - Modifying CI/CD configuration
  - Generating, rotating, or committing any credential, key, or certificate
  - Adding a third-party dependency — justify it in the commit body
  - Sending messages, or modifying shared infrastructure
  - Anything touching production inventory, production state, or production accounts

## Language Defaults
- When not specified by context, default to the language already used in the project.
- Usual stack: Go for services and CLIs, Vue 3 for front ends, OpenTofu/Terraform and Ansible for
  infrastructure, POSIX sh for portable scripts, chezmoi for dotfiles.
- Usual tooling: prek for hooks, golangci-lint, shellcheck, yamllint, hadolint, gitleaks; Woodpecker
  for CI; Renovate for dependency updates.
- Speak and write English — chat, code, comments, commits, docs. Use French only when the
  repository's own content is mostly French.

## Operational Workflow
- Read before writing: find how the codebase already solves this and follow it. Introduce a second
  way of doing something only with a reason worth stating.
- File Structure: Prefer many small files over large files.
  - Treat 200-400 lines as the comfortable size and 800 as a smell, not a hard limit — judge against
    what is normal for the language.
  - Keep all documentation in a separate top-level ./docs directory.
- Never create documentation, README, or summary files unless asked.
- Decisions: record a non-obvious one as a numbered ADR in `docs/adr/`, with the measurement or the
  concrete failure behind it. Nothing else needs an ADR.
- Diffs: touch only the lines the change requires. No drive-by reformatting.
- Maintenance: When using a TODO file, always update and commit it alongside the relevant code changes.

## Comments
- Default: no comment. Rely on naming and structure; if a comment restates the code, delete it and improve the code instead.
- The "Why" Rule: a comment explains why, never what.
- Forbidden:
  - Restating the code in English (`// increment counter` above `counter++`)
  - Section banners (`// --- HELPERS ---`)
  - Narrating obvious control flow (`// loop through users`)
  - TODOs without a ticket reference or an owner
  - Docstrings that only list parameters already visible in the signature
  - Preamble describing what you are about to do
  - Marking what changed in this edit (`// updated to handle null`) — that is the diff's job
- Acceptable when genuinely needed: non-obvious why (business rule, spec reference, workaround for an external bug, performance trade-off, deliberate deviation from the obvious approach), warnings about non-local consequences ("called by X under Y condition"), links to issues, RFCs, or external docs.
- Length: one line is the target. If the explanation needs more lines than the code, refactor instead. Re-read nearby comments when editing, and delete or update any that no longer match.

## Testing & Quality Assurance
- TDD Strategy: For application code, write tests before implementation. Configuration —
  infrastructure, CI, dotfiles — is validated by rendering, linting and dry runs instead.
- Test Suite: Use a combination of Unit, Integration, and E2E tests.
- Local Validation: Always test locally before committing. Use prek (a wrapper around pre-commit) for pre-commit hooks.
- Done means verified: run the check, and quote the failure when it fails. Never report work as
  finished on the strength of having written it. A skipped step gets said out loud, not omitted.

## Security & Privacy
- Secrets: Never paste or commit secrets (API keys, tokens, passwords, JWTs).
- Logs: Always redact sensitive data from logs. Use structured logs whenever possible.
- Review: Manually review all outputs before sharing to ensure no sensitive data remains.

## Git & Version Control
- Branching: For non-trivial changes, always work on a dedicated branch with a meaningful, descriptive name.
- Forge: read `git remote get-url origin` before reaching for a forge CLI — `fj` (forgejo-cli) for
  Forgejo and Gitea, `gh` for GitHub, `glab` for GitLab. Never assume GitHub.
- CI: drive pipelines with the tool matching the config in the repo — `woodpecker-cli` when a
  `.woodpecker.yml` or `.woodpecker/` is present.
- Commits are GPG-signed. If signing fails, ask me to unlock the agent; never disable signing to get
  a commit through.
- Commits: Use Conventional Commits (type(scope): summary).
  - The "Why" Rule: Explain why the change was made, not just what was changed.
  - Atomicity: Use small, atomic, and tested commits to ease peer review.

## Commit Messages
- Length: as short as the change allows. Never pad to look thorough.
- Subject: imperative mood, under 50 characters, no trailing period.
- Body: only when the why is not obvious from the subject. Two or three lines, wrapped at 72. Most commits need none.
- Never restate the diff: no list of files touched, no walkthrough of the changes, no recap of what the reader is about to read.
- Evidence over narrative: one measurement or one concrete failure beats a paragraph of justification.
