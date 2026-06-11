---
name: review-pr
description: To audit the delta changes between the current branch and `main`, providing standardized feedback on code quality, security, and logic.
---

## Pre-flight Check

Before doing anything else, run `git diff main...HEAD`. If the output is empty (already on `main` or no commits ahead), stop and inform the user: "No changes found between this branch and `main`. Nothing to review."

## Execution Protocol

1. **Change Discovery**
   - Run `git diff main...HEAD` (three-dot diff to compare branch tip against `main` tip, not the working tree).
   - List all modified files as context for the analysis.

2. **Contextual Analysis**
   - **Logic**: Identify code smells, missing error handling, or unnecessary complexity.
   - **Security**: Scan the diff for hardcoded secrets, insecure patterns, or dangerous function calls.
   - **Consistency**: Verify changes follow conventions defined in `CLAUDE.md` (naming, commit format, file structure).

3. **Test Verification**
   - For each modified source file, check whether a corresponding test file was added or updated.
   - If a `.pre-commit-config.yaml` is present, run `pre-commit run --all-files` and report any failures.
   - If a test runner is configured (e.g., `pytest`, `go test`, `npm test`), run it and report failures. Do not proceed past this step if tests fail.

4. **Conventional Feedback**
   Generate remarks using [Conventional Comments](https://conventionalcomments.org/) syntax.
   - Labels: `praise`, `issue`, `suggestion`, `thought`, `question`, `todo`, `nitpick`.
   - Append `(blocking)` to any finding that must be resolved before merging.
   - Always include the file path and line number for each remark.
   - Write all findings to `REVIEW.md` using the template below.

## REVIEW.md Template

```markdown
# PR Review

**Branch:** <branch-name>
**Base:** main
**Date:** <date>

## Summary

<1-3 sentence overview of what the PR does and the general quality assessment.>

## Findings

### <file-path>:<line-number>

**<label>[optional (blocking)]:** <concise title>

<Detailed explanation. Include why it matters and, where applicable, a concrete suggestion.>

---
```

Repeat the `Findings` block for each remark. Group by file for readability. If there are no findings for a category, omit it.

## Output Rules

- **File Creation**: Create or overwrite `REVIEW.md` at the repo root using the template above.
- **Passive Role**: Suggest improvements; do not modify source code automatically.
