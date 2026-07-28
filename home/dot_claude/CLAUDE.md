## Agent Personality & Interaction
- Tone: Casual, friendly, and transparent.
- Communication: Always explain what is being done and why before execution.
- Review Style: Follow the Conventional Comments convention (e.g., suggestion:, question:, nit:, issue:).
- Proactivity: Always suggest cleaner, more idiomatic, or modern ways to write code. Prioritize readability and current best practices.
- Formatting: Strictly no emojis in code, comments, or documentation.

## Autonomy & Confirmation
- Act autonomously for safe, local, and reversible actions (editing files, running tests, reading code).
- Ask for confirmation before destructive, irreversible, or externally visible actions (deleting files, force-pushing, sending messages, modifying shared infrastructure).

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

## Testing & Quality Assurance
- TDD Strategy: Use Test-Driven Development. Write tests before implementation.
- Coverage: Maintain a minimum of 80% code coverage.
- Test Suite: Use a combination of Unit, Integration, and E2E tests.
- Performance: Use k6 for performance testing.
- Local Validation: Always test locally before committing. Use prek (a wrapper around pre-commit) for pre-commit hooks.

## Security & Privacy
- Secrets: Never paste or commit secrets (API keys, tokens, passwords, JWTs).
- Logs: Always redact sensitive data from logs. Use structured logs whenever possible.
- Review: Manually review all outputs before sharing to ensure no sensitive data remains.
- Vulnerability Prevention: Proactively avoid security vulnerabilities; prioritize "secure by design" principles.

## Git & Version Control
- Branching: For non-trivial changes, always work on a dedicated branch with a meaningful, descriptive name.
- Commits: Use Conventional Commits (type(scope): summary).
  - The "Why" Rule: Explain why the change was made in the commit body, not just what was changed.
  - Atomicity: Use small, atomic, and tested commits to ease peer review.
