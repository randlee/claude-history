# Test Plan: history-search Agent

**Version**: 1.0
**Created**: 2026-02-09
**Purpose**: Systematic testing and improvement of the history-search agent prompt

## Overview

This test plan validates the history-search agent's ability to:
1. Accept fuzzy natural language queries
2. Use the claude-history CLI correctly
3. Return structured results in the Output Contract format
4. Handle errors gracefully

After each test run, we analyze the agent's tool use patterns to identify:
- Prompt clarity issues
- CLI bugs or limitations
- Missing features
- Agent confusion patterns

---

## Prerequisites

### Environment Setup

1. **CLI Path Configuration**
   - Verify `.sc/history/config.yml` exists with correct CLI path
   - Test: Run `python3 .claude/scripts/validate-claude-history-cli.py`
   - Expected: Exit code 0

2. **Agent Installation**
   - Verify `.claude/agents/history-search.md` exists
   - Verify `.claude/agents/registry.yaml` lists history-search v0.2.0
   - Verify `.claude/skills/history/SKILL.md` exists

3. **Test Data**
   - Use current repository: `/Users/randlee/Documents/github/claude-history`
   - Known test queries with expected results

---

## Test Cases

### Test Case 1: Fuzzy Text Search in Conversations

**Test ID**: TC-001
**Priority**: HIGH
**Goal**: Find conversations containing specific text ("resurrection")

#### Test Prompt
```
Find the conversation where we discussed /resurrect in this repository.
```

#### Input JSON (if using contracts)
```json
{
  "action": "search",
  "path": "/Users/randlee/Documents/github/claude-history",
  "fuzzy_query": "Find conversation about /resurrect"
}
```

#### Expected Behavior
1. Agent uses `claude-history query` with `--type user` or `--type assistant`
2. Searches conversation text (not just queue-operation entries)
3. Returns session ID `24a04e6d-4ea1-4a28-8abf-eb4b88962e3a`
4. Includes timestamp around 2026-02-08T22:00-23:00

#### Expected Output JSON
```json
{
  "success": true,
  "data": {
    "matches": [
      {
        "session_id": "24a04e6d-4ea1-4a28-8abf-eb4b88962e3a",
        "timestamp": "2026-02-08T22:48:10.993Z",
        "snippet": "### 2. `/resurrect` - Agent Resurrection",
        "confidence": 0.95,
        "match_reasons": [
          "query: 'resurrect' (exact match in text)",
          "time: February 8 (recent)"
        ]
      }
    ]
  },
  "error": null
}
```

#### Execution Steps

**a) Launch Agent**
```bash
# Task tool invocation
Task(
  subagent_type="general-purpose",  # Until history-search is registered
  description="Find resurrection conversation",
  run_in_background=true,
  model="haiku",  # Fast, cheap testing
  prompt="""
  [Full history-search.md prompt content]

  ---

  User Query: Find the conversation where we discussed /resurrect in this repository.
  Project Path: /Users/randlee/Documents/github/claude-history
  """
)
```

**b) Record Agent ID**
- Note the agent ID returned (e.g., `a34fd5f`)
- Save to tracking file: `.sc/history/test-runs.jsonl`

**c) Wait for Completion**
- Agent runs in background
- Monitor via: `tail -f /tmp/claude-*/tasks/{agent_id}.output`
- Or wait for completion notification

**d) Pass to Analyzer Agent**
- Extract agent's session JSONL
- Pass to `prompt-analyzer` agent (define below)

**e) Analyze Results**
- Review analyzer's findings
- Compare to manual analysis
- Decide on corrective actions

---

### Test Case 2: Find Agent by Type and Project

**Test ID**: TC-002
**Priority**: MEDIUM
**Goal**: Find specific agent types (e.g., Explore agents)

#### Test Prompt
```
Find the Explore agent that analyzed the beads repository.
```

#### Expected Behavior
1. Agent uses `--type queue-operation` to find agent spawns
2. Filters by `subagent_type == "Explore"`
3. Searches for "beads" in agent descriptions
4. Returns agent with session/agent IDs

---

### Test Case 3: Date Range Query

**Test ID**: TC-003
**Priority**: MEDIUM
**Goal**: Find agents from specific time period

#### Test Prompt
```
Find agents that ran last week in this repository.
```

#### Expected Behavior
1. Agent parses "last week" to date range
2. Uses `--start` and `--end` flags
3. Returns agents within that timeframe

---

### Test Case 4: Tool Usage Search

**Test ID**: TC-004
**Priority**: LOW
**Goal**: Find agents that used specific tools

#### Test Prompt
```
Which agents modified package.json files?
```

#### Expected Behavior
1. Uses `--tool Write` or `--tool-match "package.json"`
2. Returns agents that wrote to package.json

---

## Analysis Protocol

### Tool Use Analysis

For each test run, extract and categorize:

1. **Tool Calls**: List all tool names in order
2. **Failures**: Identify calls with `is_error: true`
3. **Retries**: Detect repeated similar calls
4. **Workarounds**: Patterns where agent adapts after failure

### Prompt Issues

Track when agent:
- Ignores guidance in prompt
- Can't translate fuzzy query to CLI commands
- Uses wrong approach (e.g., Read instead of Grep)
- Missing critical context

### CLI Issues

Track when:
- CLI returns errors
- CLI missing expected features
- CLI error messages are unclear
- Agent works around CLI limitations

---

## prompt-analyzer Agent Definition

**Purpose**: Analyze history-search agent execution traces and identify improvement opportunities

**Input**:
- Session JSONL for history-search agent execution
- Original test prompt

**Output**:
```json
{
  "test_case": "TC-001",
  "agent_id": "a34fd5f",
  "success": true/false,
  "tool_calls": {
    "total": 15,
    "failures": 2,
    "retries": 0
  },
  "prompt_issues": [
    {
      "severity": "high|medium|low",
      "category": "missing_guidance|contradictory|unclear",
      "issue": "Description of issue",
      "evidence": "Tool call #7: Read failed with 1.7MB file",
      "fix": "Add section: 'Always use Grep for session files'",
      "prompt_section": "Working with Session JSONL Files"
    }
  ],
  "cli_bugs": [
    {
      "severity": "high|medium|low",
      "issue": "Description",
      "evidence": "Tool call #14 error message",
      "fix": "Add filepath.Abs() in list.go",
      "file": "src/cmd/list.go"
    }
  ],
  "missing_features": [
    {
      "severity": "high|medium|low",
      "feature": "Feature name",
      "rationale": "Why needed",
      "proposal": "Implementation suggestion"
    }
  ],
  "recommendations": {
    "priority": "high|medium|low",
    "actions": [
      "1. Update prompt section X with guidance Y",
      "2. Fix CLI bug in file Z",
      "3. Add CLI feature: --search-text flag"
    ]
  }
}
```

**Agent Prompt** (to be created):
See `docs/prompt-analyzer-agent.md` (create separately)

---

## Test Execution Log

### Run 1: 2026-02-09 (Baseline)

**Test Case**: TC-001 (Find /resurrect conversation)
**Agent ID**: acaff29
**Model**: sonnet-4.5
**Duration**: 88s
**Tool Calls**: 15

**Results**:
- ✅ Found correct conversation (session 24a04e6d)
- ❌ Did not return fenced JSON Output Contract
- ❌ Used grep/sed/jq instead of CLI query commands

**Issues Identified**:
1. **Prompt Issue (HIGH)**: No guidance to use Grep for large session files
   - Evidence: Call #7 Read failed with 1.7MB file
   - Fix: Add "Working with Session JSONL Files" section

2. **CLI Bug (MEDIUM)**: Relative path `.` not resolved
   - Evidence: Call #14 error "project directory not found: .../projects/-"
   - Fix: Add filepath.Abs() before encoding (IN PROGRESS)

3. **Missing Feature (HIGH)**: No text search in CLI
   - Evidence: Agent used grep+jq to search for "resurrect" in text
   - Proposal: Add `--search-text` flag or fuzzy_query support

**Actions Taken**:
- Created worktree: fix/cli-relative-path-resolution
- Background agent fixing relative path bug (agent a34fd5f)
- Created this test plan
- Next: Create prompt-analyzer agent definition

---

## Next Steps

1. **Complete Run 1 Analysis**
   - Define prompt-analyzer agent
   - Run analyzer on acaff29 trace
   - Compare analyzer output to manual analysis
   - Debug analyzer if needed

2. **Fix Prompt (Round 1)**
   - Add "Working with Session JSONL Files" section
   - Add examples for text search in conversations
   - Clarify when to use Grep vs Read

3. **Fix CLI (Round 1)**
   - Merge relative path fix
   - Consider adding --search-text flag

4. **Run 2: Re-test TC-001**
   - Use updated prompt
   - Use fixed CLI
   - Compare results to Run 1
   - Measure improvement (fewer failures, faster execution)

5. **Iterate**
   - Run remaining test cases (TC-002 through TC-004)
   - Continue prompt/CLI improvements
   - Build regression test suite

---

## Success Metrics

### Per Test Run
- Tool call success rate (target: >90%)
- Time to completion (target: <60s for TC-001)
- Output contract compliance (target: 100%)
- Correct results (target: 100%)

### Overall Progress
- Prompt clarity score (from analyzer)
- CLI bug count (decreasing)
- Agent retry count (decreasing)
- Test pass rate (increasing)

---

## Appendix: Test Run Tracking Format

**File**: `.sc/history/test-runs.jsonl`

```json
{
  "test_id": "TC-001",
  "run_number": 1,
  "timestamp": "2026-02-09T02:00:00Z",
  "agent_id": "acaff29",
  "model": "sonnet-4.5",
  "prompt_version": "0.2.0",
  "cli_version": "0.3.0",
  "duration_ms": 88454,
  "tool_calls": 15,
  "failures": 2,
  "success": true,
  "issues_found": 3,
  "analyzer_agent_id": "xyz789"
}
```

---

**Document Version**: 1.0
**Last Updated**: 2026-02-09
**Status**: Ready for execution
