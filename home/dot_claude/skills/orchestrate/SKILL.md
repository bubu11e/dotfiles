---
name: orchestrate
description: Drive a task to completion through separate implementer and reviewer subagents, looping until an objective signal passes and the review comes back clean. Use for multi-step work where the result has to be verified rather than asserted.
---

## When this applies

Work that is large enough to need more than one pass, and where "done" can be
checked by something other than an opinion: a failing test that must pass, a lint
stack that must go green, a reproduction that must stop reproducing.

Do not use it for single-edit changes. The loop costs more than the work.

## Why two agents

The implementer and the reviewer are **separate subagents**, and the reviewer does
not inherit the implementer's context.

Same-context review tends to ratify what was just written. An agent that has spent
the last twenty tool calls justifying an approach is the worst possible judge of
whether that approach was right, and it will read its own comments as evidence.
A reviewer that sees only the diff and the requirement has no such investment.

The reviewer gets: the original requirement, the diff, and the objective signal's
output. Not the implementer's reasoning, and not its summary of what it did.

## The loop

1. **Establish the objective signal first.** Before any implementation, name the
   command whose exit status decides the question — the test, the linter, the
   repro. If no such command exists, write it first; that is the work, and a task
   with no objective signal is not ready to be looped on.
2. **Implement.** One subagent, scoped to the requirement.
3. **Run the signal.** Record the actual output, not a description of it.
4. **Review.** Delegate to the `review-pr` skill rather than reviewing inline — it
   owns the diff discovery, the Conventional Comments vocabulary, and the report
   format. Do not restate those here.
5. **Gate.** Continue only if the signal passes *and* the review has no blocking
   findings. Both conditions, every iteration. A green test with an unread review
   is not a pass, and a clean review over a failing test is not either.
6. **Iterate or stop.** Feed blocking findings back to a fresh implementer pass.

## Gate categories

`review-pr` emits Conventional Comments labels. Map them onto the gate:

| Gate | Labels | Effect |
| --- | --- | --- |
| blocking | anything marked `(blocking)`, plus `issue` | loop continues |
| suggestion | `suggestion`, `todo` | reported, does not block |
| nit | `nitpick`, `thought`, `question`, `praise` | reported, does not block |

Only the blocking row gates. A reviewer that blocks on a nit is miscalibrated and
should be told so rather than obeyed.

## Iteration cap

**Three implement-review cycles.** On the fourth, stop and hand back to the user
with: what the signal says, what is still blocking, and what was tried.

The cap is not a formality. Three failed cycles means the task is
underspecified, the objective signal is measuring the wrong thing, or the approach
is wrong — and none of those is fixed by a fourth attempt. Report the state
plainly instead of grinding.

## Reporting

State the objective signal's real output and the surviving findings. Never report
a loop as converged without both the passing signal and the clean review; if the
cap was hit, say so and say which condition was still failing.
