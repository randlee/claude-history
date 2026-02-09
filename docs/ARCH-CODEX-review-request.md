# Review Request for ARCH-CODEX

**Date**: 2026-02-09
**From**: Main session
**Subject**: Review prompt-analyzer agent for systematic prompt improvement

---

## Context

We're building a **meta-feedback loop** for agent development:

1. **history-search agent** - Searches Claude Code agent history using CLI tool
2. **Test execution** - Run agent with test prompts, capture tool use traces
3. **prompt-analyzer agent** - Analyzes execution traces to identify improvement opportunities
4. **Iteration** - Apply fixes, re-test, measure improvement

## The Problem We're Solving

When testing history-search agent with query: *"Find the conversation about /resurrect"*

**What happened**:
- Agent succeeded (found the answer)
- But made 15 tool calls with 2 failures
- Tool #7: Tried to Read 1.7MB file, failed, switched to Grep
- Tool #14: CLI command error, recovered with workaround

**The issue**: Agent recovered well, so it LOOKS successful. But hidden inefficiencies:
- Wasted tokens/time on failed calls
- Indicates prompt clarity gaps
- Repeated failures across test runs = systematic issues

## Our Solution: prompt-analyzer Agent

**Purpose**: Systematically analyze agent execution traces to identify:
- Prompt clarity issues (missing guidance, unclear examples)
- Root cause classification (prompt vs tool vs feature gap)
- Recovery quality (did agent adapt well?)
- Actionable recommendations with evidence

**Input**: Agent session JSONL + test case metadata
**Output**: Structured JSON with categorized issues and fixes

---

## What We Need From You

**Please review**: `.claude/agents/prompt-analyzer.md`

### Specific Questions

1. **Root Cause Classification Logic** (lines 130-180)
   - Is the decision tree clear enough?
   - Are we missing any root cause categories?
   - Is `prompt_unclear` vs `prompt_missing` distinction useful?

2. **Analysis Methodology** (lines 182-230)
   - Does the step-by-step process make sense?
   - Is there a better way to structure the failure analysis?
   - Should we add more intermediate steps?

3. **Output Contract** (lines 30-80)
   - Is the JSON structure clear and useful?
   - Are we capturing the right metadata?
   - Is `evidence` field specific enough?

4. **Examples Section** (lines 290-340)
   - Do the examples clarify the agent's job?
   - Should we add more edge cases?
   - Is the expected output realistic?

5. **Overall Prompt Quality**
   - Is the agent's job clear and focused?
   - Are instructions concise and actionable?
   - What would you do differently?

### Success Criteria

When we run prompt-analyzer on the TC-001 trace, it should:
- Identify the 2 failed tool calls
- Correctly classify root causes (prompt_missing, cli_bug)
- Provide actionable fixes with evidence
- Score agent recovery quality
- Match or exceed our manual analysis

---

## Key Design Principles

We're aiming for:
1. **Systematic** - Same analysis every time, not ad-hoc
2. **Evidence-based** - Every claim has tool call reference
3. **Actionable** - Each issue has concrete fix recommendation
4. **Prioritized** - Severity based on frequency + impact
5. **Traceable** - Can verify analyzer's reasoning

---

## Files to Review

**Primary**: `.claude/agents/prompt-analyzer.md` (the agent we need your help with)

**Context** (optional):
- `.claude/agents/history-search.md` (the agent being analyzed)
- `docs/test-plan-history-search-agent.md` (test protocol)
- `docs/2026-02-08-skill-user-stories.md` (original requirements)

---

## What Success Looks Like

After your review and our improvements:

**Run 1** (baseline with current prompt):
- Tool failures: 2
- Root causes identified: 2
- Fixes applied: 2

**Run 2** (after fixes):
- Tool failures: 0
- Execution time: -30% faster
- Output matches contract: 100%

**The goal**: Create a reproducible process for improving any agent prompt through systematic analysis.

---

## Your Expertise Needed

You're excellent at **concise, meaningful prompting**. We need your help ensuring:
- Instructions are clear without being verbose
- Examples illuminate without overwhelming
- Structure guides without constraining
- The agent knows WHEN to apply each classification rule

Please focus on **prompt quality** - we'll handle implementation details.

---

## Response Format

Whatever works best for you:
- Inline comments in the prompt file
- Structured feedback document
- Suggested rewrite of key sections
- Questions/clarifications
- Or just tell us what you'd do differently

Thank you!

---

**Files attached** (conceptually - these exist in the repo):
- `.claude/agents/prompt-analyzer.md`
- Session trace from TC-001 (if helpful for context)
