---
name: commit-and-push
description: >-
  Pre-validates, commits, and pushes changes to git. Use this when the user
  asks to save, commit, or push code. Ensures atomic changes,
  Conventional Commits, and pre-commit hook compliance.
---

## Execution Flow

### 1. Inspect the Diff

- Run `git diff` (unstaged) and `git diff --cached` (staged) to understand the full set of changes.
- Do NOT run `git add` yet.

### 2. Plan Atomic Commits

- If the diff contains multiple unrelated logical changes (e.g., a bug fix mixed with a new feature), split them into separate commits.
- For each group, repeat steps 3-6 independently before moving to the next.

### 3. Stage Selectively

- Stage only the files relevant to the current logical group: `git add <file> [<file> ...]`.
- Never use `git add .` or `git add -A` -- this risks including unintended files (secrets, build artifacts, editor temp files).

### 4. Pre-Commit Validation

- If `.pre-commit-config.yaml` exists, run `pre-commit run --staged` (or `--all-files` if staged-only is not supported).
- **Hard constraint:** If pre-commit fails, stop. Do not commit. Report the exact errors to the user and wait for a fix before retrying.

### 5. Compose the Commit Message

Format: `<type>(<optional scope>): <subject>`

Valid types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `ci`, `revert`.

Rules:
- **Subject**: concise imperative summary of *what* changed (e.g., `add retry logic to auth client`).
- **Body** (required when the change is non-trivial): explain *why* the change was made -- the motivation, context, or problem it solves. Not a repeat of the subject.
- **Workarounds**: if the change includes a non-standard pattern (pinned version, bypassed check, monkey-patch), add an inline comment in the code before committing and reference it in the commit body.
- **AI Attribution**: when AI tools contribute to the changes, include an `Assisted-by` trailer in the commit message using the following format:

  ```
  Assisted-by: AGENT_NAME:MODEL_VERSION [TOOL1] [TOOL2]
  ```

  Where:
  - `AGENT_NAME` is the name of the AI tool or framework.
  - `MODEL_VERSION` is the specific model version used.
  - `[TOOL1] [TOOL2]` are optional specialized analysis tools used (e.g., coccinelle, sparse, smatch, clang-tidy).

  Basic development tools (git, gcc, make, editors) should not be listed.

  Example: `Assisted-by: Claude:claude-3-opus coccinelle sparse`

### 6. Commit

- Run `git commit -m "<message>"` (or with `--message` and a heredoc for multi-line).
- If the commit hook fails, do not retry blindly -- report the error to the user.

### 7. Push

- Run `git push`.
- If the push is rejected (remote is ahead): stop, report the rejection, and ask the user to pull/rebase first. Never force-push unless explicitly instructed.
