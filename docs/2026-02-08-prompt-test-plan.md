# Prompt Development & Testing Checklist

**Date**: 2026-02-09
**Purpose**: Track development and testing of agents/skills from user stories
**Reference**: `docs/2026-02-08-skill-user-stories.md`

---

## Overview

Systematic development checklist for implementing the 4 use cases from user stories:
1. **Agent by Reference** (via history skill)
2. **Agent Resurrection** (via resurrect skill)
3. **Historical Search** (via history-search agent)
4. **Conversation Compacting** (via compact skill - future)

**Current Focus**: Phase 1 - History Search & Agent by Reference

---

## Phase 1: History-Search Agent & History Skill

### 1.1 history-search Agent Development

**Agent**: `.claude/agents/history-search.md` (v0.2.0)
**Status**: Initial version complete, needs improvements

#### Development Checklist

- [x] **D1.1** Create initial agent design (v0.1.0)
  - Completed: 2026-02-08
  - Agent-design-expert score: 82/100 (needs revision)

- [x] **D1.2** Redesign based on expert feedback (v0.2.0)
  - Completed: 2026-02-08
  - Agent-design-expert score: 92/100 (approved)
  - Changed to CLI orchestrator (not implementer)

- [x] **D1.3** Add marketplace compliance (v0.2.0)
  - Completed: 2026-02-08
  - Added Input/Output contracts
  - Added to registry.yaml
  - PreToolUse hooks for CLI validation

- [ ] **D1.4** Add Grep guidance for large files
  - Priority: HIGH
  - Issue: Agent tried Read on 1.7MB file, failed
  - Fix: Add "Working with Session JSONL Files" section
  - Evidence: TC-001 call #7

- [ ] **D1.5** Add text search examples
  - Priority: HIGH
  - Issue: No examples for searching conversation text
  - Fix: Add examples using `--type user|assistant` + text filtering
  - Evidence: TC-001 used grep+jq workaround

- [ ] **D1.6** Clarify fuzzy search capability
  - Priority: MEDIUM
  - Issue: Input Contract doesn't support fuzzy natural language queries
  - Fix: Add `fuzzy_query` field or clarify how to translate NL to CLI

#### Testing Checklist

- [x] **T1.1** Run TC-001 baseline (Find /resurrect conversation)
  - Completed: 2026-02-09 (agent acaff29)
  - Result: Success with 2 failures, good recovery
  - Issues identified: 3 (1 fixed, 2 pending)

- [x] **T1.2** Run prompt-analyzer on TC-001 trace
  - Completed: 2026-02-09 (agent a1eb5e7)
  - Result: 3 issues identified (1 cli_bug, 1 missing_feature, 1 prompt_missing)
  - Validation: Analyzer v1.1.0 works correctly

- [x] **T1.3** Apply prompt fixes (three-rule JSON + startup efficiency)
  - Updated history-search.md to v0.3.0
  - Added three-rule pattern for JSON enforcement
  - Added startup efficiency guidance
  - Commit: 76c834a

- [x] **T1.4** Run TC-001 iteration 2
  - Completed: 2026-02-09 (agent aa92e76)
  - Duration: 32s | Tool calls: 7 | Failures: 0
  - Issues: No JSON output, grep+jq workaround
  - Note: Better than baseline but still issues

- [x] **T1.5** Run prompt-analyzer on iteration 2
  - Completed: 2026-02-09 (agent a1eb5e7)
  - Identified: Output contract violation, missing --text flag docs
  - Recommended: Add three-rule pattern (applied in v0.3.0)

- [x] **T1.6** Run TC-001 iteration 3 (with proper JSON input)
  - Completed: 2026-02-09 (agent aaba92c)
  - Duration: 39s | Tool calls: 7 | Failures: 0
  - SUCCESS: JSON output fixed by three-rule pattern ✓
  - Issues: Didn't use --text flag (not documented)

- [x] **T1.7** Run prompt-analyzer on iteration 3
  - Completed: 2026-02-09 (agent ac2f33f)
  - Result: JSON output FIXED, startup efficiency FIXED
  - Remaining: --text flag not documented in prompt
  - Recommendation: Add Example 11 for --text flag

- [ ] **T1.8** Run TC-001 iteration 4 (with --text flag docs)
  - Use prompt v0.4.0 with --text flag documentation
  - Expected: Agent uses --text flag natively
  - Expected: No grep+jq workaround needed

### 1.2 History Skill Development

**Skill**: `.claude/skills/history/SKILL.md`
**Status**: Basic structure complete, needs integration testing

#### Development Checklist

- [x] **D2.1** Create skill structure
  - Completed: 2026-02-08
  - SKILL.md with YAML frontmatter
  - README.md with installation instructions
  - manifest.yaml with dependencies

- [x] **D2.2** Add agent delegation instructions
  - Completed: 2026-02-08
  - Must delegate to history-search agent
  - Pass fenced JSON input
  - Receive fenced JSON output

- [ ] **D2.3** Test skill loading in Claude Code
  - Verify `/history` command loads
  - Verify skill prompts correctly
  - Verify parameters are parsed

- [ ] **D2.4** Fix agent delegation mechanism
  - Issue: Task tool doesn't recognize custom agents
  - Options: Register agent type or use workaround
  - Document solution in SKILL.md

- [ ] **D2.5** Add agent-by-reference documentation
  - Already drafted in SKILL.md (lines 157-219)
  - Test: Pass agent ID to 10 subagents
  - Measure: Verify 40% token reduction

#### Testing Checklist

- [ ] **T2.1** Test `/history` command invocation
  - Run: `/history action=list`
  - Expected: Lists sessions for current project
  - Record: Output format and any errors

- [ ] **T2.2** Test agent delegation
  - Run: `/history action=find-agent path=. resurrection`
  - Expected: Delegates to history-search agent
  - Record: Whether delegation works or errors

- [ ] **T2.3** Test agent-by-reference pattern
  - Create explore agent for test repo
  - Pass agent ID to 3 subagents
  - Measure: Token usage vs passing full transcript
  - Target: 30-40% reduction

### 1.3 CLI Tool Improvements

**Tool**: `claude-history` CLI (v0.3.0+)

#### Development Checklist

- [x] **D3.1** Fix relative path resolution
  - PR #47: Merged to develop (commit c51e5f1)
  - Fixed: Lint + Windows CI failures
  - All CI checks passing

- [x] **D3.2** Merge PR #47 to develop
  - Completed: 2026-02-09
  - All CI checks passed
  - Ready for v0.3.1 release

- [x] **D3.3** Add --text flag for content search
  - PR #48: https://github.com/randlee/claude-history/pull/48
  - Status: CI running
  - Implementation: Case-insensitive text filtering
  - Example: `claude-history query /path --text "resurrect"`

- [ ] **D3.4** Document all flags in agent prompt
  - Ensure history-search.md covers all CLI flags
  - Add examples for each major use case
  - Update when new flags added

#### Testing Checklist

- [x] **T3.1** Test relative path fix manually
  - Completed: 2026-02-09
  - Tests: `list .`, `resolve .`, `query . --session abc`
  - Result: All pass

- [ ] **T3.2** Test relative path fix in agent context
  - Run agent with `claude-history list .`
  - Expected: No errors, correct encoded path
  - Compare to TC-001 call #14 (should succeed now)

---

## Phase 2: Agent Resurrection (Future)

### 2.1 Resurrect Skill Development

**Skill**: `.claude/skills/resurrect/` (not created yet)
**Priority**: MEDIUM (after Phase 1 complete)

#### Development Checklist

- [ ] **D4.1** Design resurrect skill architecture
  - Two modes: Direct (--session --agent) and Fuzzy (NL query)
  - Integrates with history-search agent
  - Bookmarking system design

- [ ] **D4.2** Create `.sc/history/bookmarks.jsonl` format
  - Schema design for agent bookmarks
  - Metadata: session_id, agent_id, description, tags
  - CRUD operations for bookmarks

- [ ] **D4.3** Implement fuzzy resurrection mode
  - Use history-search agent to find candidates
  - Return ranked results
  - User selects or auto-selects highest confidence

- [ ] **D4.4** Implement direct resurrection mode
  - Accept session_id and agent_id
  - Extract agent's final output
  - Format for consumption by new agent

- [ ] **D4.5** Create resurrection tracking log
  - `.sc/history/resurrection-log.jsonl`
  - Track: when, what, by whom
  - Analytics: Most resurrected agents

#### Testing Checklist

- [ ] **T4.1** Test direct resurrection
  - Resurrect agent a47c2de (agent-design-expert)
  - Verify: Complete transcript extracted
  - Verify: New agent can use the knowledge

- [ ] **T4.2** Test fuzzy resurrection
  - Query: "the explore agent who analyzed beads"
  - Expected: Returns candidate(s)
  - Expected: Can resurrect selected agent

- [ ] **T4.3** Measure performance benefits
  - Baseline: Agent without resurrection (fresh analysis)
  - With resurrection: Agent with previous analysis
  - Measure: Token reduction, time reduction
  - Target: 40% tokens, 50% time (per user stories)

---

## Phase 3: Conversation Compacting (Future)

### 3.1 Compact Skill Development

**Skill**: `.claude/skills/compact/` (not created yet)
**Priority**: LOW (most complex, after Phase 1 & 2)

#### Development Checklist

- [ ] **D5.1** Design compaction strategy
  - Keep: Intro prompt, critical context
  - Compact: Checklist → summary, remove verbose content
  - Preserve: 2 subagent IDs for resurrection
  - Add: Reference to previous version

- [ ] **D5.2** Create schema validation tools
  - Ensure output is valid JSONL
  - Validate message format
  - Check required fields

- [ ] **D5.3** Implement background compaction agent
  - Takes: Session JSONL + compaction rules
  - Returns: New compacted JSONL
  - Validates: Schema compliance

- [ ] **D5.4** Add self-reference capability
  - Compacted agent can query original version
  - Pass original session/agent ID
  - Use resurrection mechanism

#### Testing Checklist

- [ ] **T5.1** Test basic compaction
  - Input: Long session (5000+ msgs)
  - Expected: Reduced to 30-50% size
  - Verify: No loss of critical context

- [ ] **T5.2** Test resume from compact
  - Create compacted version
  - Resume with `/resume`
  - Verify: Agent has full context
  - Verify: Can ask previous version questions

---

## Phase 4: Prompt-Analyzer Validation

### 4.1 Analyzer Testing

**Agent**: `.claude/agents/prompt-analyzer.md` (v1.1.0)
**Status**: Ready for testing

#### Testing Checklist

- [ ] **T6.1** Run analyzer on TC-001 trace
  - Input: agent acaff29 JSONL
  - Expected: 3 diagnoses (2 failures + 1 pattern)
  - Validate: Root causes match manual analysis

- [ ] **T6.2** Verify deterministic classification
  - Run analyzer twice on same input
  - Expected: Identical output both times
  - Verify: No randomness or timestamps

- [ ] **T6.3** Test confidence scoring
  - Check: High confidence for clear cases
  - Check: Lower confidence for ambiguous cases
  - Validate: Scores are in 0.70-0.95 range

- [ ] **T6.4** Validate evidence structure
  - Check: All diagnoses have tool_call_ref
  - Check: All diagnoses have prompt_ref
  - Check: Error excerpts are direct quotes

- [ ] **T6.5** Test on TC-002, TC-003, TC-004 traces
  - Run analyzer on each test case
  - Compare: Consistency across different scenarios
  - Identify: Patterns in classifications

---

## Success Metrics

### Per Test Case
- **Tool call success rate**: Target >90% (baseline: 87% from TC-001)
- **Time to completion**: Target <60s (baseline: 88s from TC-001)
- **Output contract compliance**: Target 100%
- **Correct results**: Target 100%

### Prompt Improvements (Iteration 1 → 2)
- **Fewer failures**: Target 0-1 failures (baseline: 2)
- **Faster execution**: Target 30% faster
- **Better recovery**: If failures occur, immediate recovery

### Agent-by-Reference Performance
- **Token reduction**: Target 30-40% (user story: achieved 40%)
- **Time reduction**: Target 50% (user story: achieved 2x speedup)
- **Consistency**: 10 subagents using same reference

---

## Notes

### Test Run Tracking

Track each test run in `.sc/history/test-runs.jsonl`:
```json
{
  "test_id": "TC-001",
  "run_number": 1,
  "timestamp": "2026-02-09T02:00:00Z",
  "agent_id": "acaff29",
  "prompt_version": "0.2.0",
  "cli_version": "0.3.0",
  "duration_ms": 88454,
  "tool_calls": 15,
  "failures": 2,
  "success": true,
  "issues_found": 3
}
```

### Blocker Resolution

**Current Blocker**: Task tool doesn't recognize custom agents
- **Impact**: Can't delegate from skill to history-search agent
- **Workaround**: Use general-purpose agent with history-search prompt
- **Solution needed**: Register agent type or use different mechanism

---

**Document Version**: 1.0
**Last Updated**: 2026-02-09
**Status**: Active development - Phase 1 in progress
