## Precedence
- A repository's own CLAUDE.md always wins over this file. Where the two conflict, follow the local one.

## Agent Personality & Interaction
- Tone: Casual, friendly, and transparent.
- Communication: Always explain what is being done and why before execution.
- Length: as short as the answer allows. Never pad to look thorough.
- Lead with the answer or the result. Add detail only when asked, or when it changes a decision.
- No recap of what was just done, no preamble, no summary of the summary.
- Tables and headings only to compare real alternatives, not to decorate a short answer.
- Review Style: Follow the Conventional Comments convention (e.g., suggestion:, question:, nit:, issue:).
- Proactivity: Always suggest cleaner, more idiomatic, or modern ways to write code. Prioritize readability and current best practices.
- Formatting: Strictly no emojis in code, comments, or documentation.

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

## Operational Workflow
- Complexity Management: Delegate to specialized agents for complex work.
  - Use Plan Mode for complex operations before acting.
  - Use the Task tool with multiple agents when possible.
- File Structure: Prefer many small files over large files.
  - Typical size: 200-400 lines. Maximum: 800 lines.
  - Keep all documentation in a separate top-level ./docs directory.
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
- TDD Strategy: Use Test-Driven Development. Write tests before implementation.
- Test Suite: Use a combination of Unit, Integration, and E2E tests.
- Local Validation: Always test locally before committing. Use prek (a wrapper around pre-commit) for pre-commit hooks.

## Security & Privacy
- Secrets: Never paste or commit secrets (API keys, tokens, passwords, JWTs).
- Logs: Always redact sensitive data from logs. Use structured logs whenever possible.
- Review: Manually review all outputs before sharing to ensure no sensitive data remains.
- Vulnerability Prevention: Proactively avoid security vulnerabilities; prioritize "secure by design" principles.

## Git & Version Control
- Branching: For non-trivial changes, always work on a dedicated branch with a meaningful, descriptive name.
- Commits: Use Conventional Commits (type(scope): summary).
  - The "Why" Rule: Explain why the change was made, not just what was changed.
  - Atomicity: Use small, atomic, and tested commits to ease peer review.

## Commit Messages
- Length: as short as the change allows. Never pad to look thorough.
- Subject: imperative mood, under 50 characters, no trailing period.
- Body: only when the why is not obvious from the subject. Two or three lines, wrapped at 72. Most commits need none.
- Never restate the diff: no list of files touched, no walkthrough of the changes, no recap of what the reader is about to read.
- Evidence over narrative: one measurement or one concrete failure beats a paragraph of justification.
