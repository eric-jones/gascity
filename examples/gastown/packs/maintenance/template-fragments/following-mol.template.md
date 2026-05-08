{{ define "following-mol" }}
## Following Your Formula

Your formula defines your work as a sequence of steps. Steps are NOT
materialized as individual beads — they exist in the formula source on
disk at `.beads/formulas/<your-formula>.formula.toml`. Read that file
once to load the step descriptions, then work through them in order.

`gc bd formula show <name>` lists step IDs and titles in a summary tree
and is useful for confirming a formula is registered, but it does NOT
emit the step descriptions you need to execute — read the source.

**THE RULE**: Execute one step at a time. Verify completion. Move to next.
Do NOT skip ahead. Do NOT claim steps done without actually doing them.

On crash or restart, re-read your formula steps and determine where you
left off from context (last completed action, git state, bead state).
{{ end }}
