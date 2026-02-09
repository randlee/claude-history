# History Skill User Stories

**Date**: 2026-02-08
**Source**: Session 24a04e6d-4ea1-4a28-8abf-eb4b88962e3a
**Context**: Discussion about extending claude-history CLI with enhanced Claude skills

## Overview

Initial discussion identified 3-4 distinct use cases for history-related skills. User's key concern: "Putting all of this information into a single skill.md seems excessive" - suggesting separate skills for each use case.

---

## Use Case 1: Passing Agent by Reference

**Status**: Already described in skill
**Purpose**: Pass entire subagent to other subagents by reference instead of duplicating content

### Performance Metrics (Real Testing)
- **40% token reduction**
- **50% time reduction (2x faster)**
- Some individual tasks: **10x faster**
- Average speedup: **2x faster**

### Implementation Challenge
Required 5-6 tries to teach primary agent proper prompting technique:
- **Problem**: When reference failed, subagent still completed task burning time/tokens
- Primary agent thought task succeeded even without reference agent
- **Success indicator**: Only when it worked first time (MUCH faster) was significance clear

### Requirements
- Minimal skill documentation (already exists)
- Clear prompting instructions for primary agent
- Error handling when reference fails (don't proceed with task)

---

## Use Case 2: Resurrecting an Agent

**Purpose**: Bring back useful agents from past sessions to answer questions or perform tasks

### Source Options
- Claude sessions in other repos
- Past Claude Code sessions from any project on this computer

### Two Operation Modes

#### Mode 1: Direct Resurrection (Easy)
```bash
/resurrect --session <session-id> --agent <agent-id>
```
- Straightforward implementation
- No search required

#### Mode 2: Fuzzy Resurrection (Complex)
```bash
/resurrect the explore agent who explored beads yesterday from ../beads-research/
```
- Requires deploying background `history-search` agent
- Agent returns session/agent ID matching fuzzy criteria
- Then proceeds with direct resurrection

### Bookmarking System
**Recommendation from previous agent**: Create bookmarking database for useful agents

**Metadata storage needed for**:
- Registry of explore agents
- Registry of expert agents (like agent-design-expert)
- Resurrection history/logs

### Requirements
- `/resurrect` skill/command
- Integration with history-search agent for fuzzy mode
- Metadata storage (`.sc/history/bookmarks.jsonl`)
- Resurrection log tracking

---

## Use Case 3: Searching for Historical Information

**Purpose**: Locate and query past conversations across all Claude sessions

### Key Capability
Claude history has access to **all Claude sessions ever executed on this computer**

### Workflow
1. Search for past conversation using fuzzy criteria
2. Locate relevant agent from that conversation
3. Resurrect agent to answer questions about that work

### Requirements
- `history-search` agent (already designed)
- Integration with resurrection workflow
- Ability to search across multiple projects/repos

---

## Use Case 4: Conversation Compacting

**Purpose**: Strategically remove unnecessary information from conversation history
**Benefits**: Reduce token use and improve compact quality
**Complexity**: HIGH - meta-programming task

### Workflow
1. User triggers custom slash command: `/compact`
2. Command launches background agent with specific instructions:
   - Leave intro prompt intact
   - Replace checklist with updated info
   - Include 2 subagent IDs for resurrection
   - Remove everything after X
3. Background agent creates new JSONL entry in `.claude` history
4. User runs `/resume` with updated history item

### Advanced Feature: Self-Reference
- Include reference to previous version of agent
- Compacted agent can ask its previous instance questions if something was missed

### Requirements
- Custom slash command infrastructure
- Background agent with compaction logic
- **Schema validation tools** to ensure properly formatted JSON output
- Resume capability with modified history
- Optional: self-reference for Q&A with previous version

---

## Decision: Separate Skills vs Combined Skill

**User's Position**: "Putting all of this information into a single skill.md seems excessive"

### Complexity Analysis

| Use Case | Complexity | Metadata Needed | Special Infrastructure |
|----------|-----------|-----------------|----------------------|
| Agent by Reference | LOW | None | Clear prompting docs |
| Resurrection | MEDIUM | Bookmarks, logs | history-search integration |
| Historical Search | MEDIUM | None | history-search agent |
| Compacting | HIGH | None | Schema validation, custom command |

### Recommendation
Create **separate, focused skills**:
1. `/history` - Search and agent-by-reference (current skill)
2. `/resurrect` - Agent resurrection with fuzzy matching
3. `/compact` - Conversation compaction (future)

**Rationale**: Each has distinct purpose, complexity, and requirements. Separation improves maintainability and user understanding.

---

## Implementation Priority

Based on user discussion:

### Phase 1: Foundation (DONE)
- ✅ `history-search` agent
- ✅ `/history` skill with agent-by-reference documentation

### Phase 2: Resurrection (NEXT)
- `/resurrect` skill
- Bookmarking infrastructure (`.sc/history/bookmarks.jsonl`)
- Fuzzy resurrection via history-search agent
- Direct resurrection mode

### Phase 3: Compaction (FUTURE)
- `/compact` custom command
- Schema validation
- Background compaction agent
- Self-reference capability

---

## Key Insights

1. **Performance is dramatic**: 40% token reduction, 2x-10x speedup when working correctly
2. **Failure mode is costly**: When reference fails, subagent still burns resources
3. **Teaching curve exists**: Required 5-6 iterations to get prompting right
4. **Bookmarking is essential**: Previous agent strongly recommended agent registry
5. **Separation of concerns**: Each use case deserves its own focused skill

---

**Document Version**: 1.0
**Last Updated**: 2026-02-09
**Status**: Requirements captured, implementation in progress
