# Claude History - Issues & Improvements Tracker

This document tracks identified issues, improvements, and feature requests for future development sprints.

## Quick Summary

**Total Issues**: 6
- 🔴 **P0 Critical**: 1 (Search broken)
- 🟠 **P1 High**: 4 (URLs, newlines, expand/collapse, file paths)
- 🟡 **P2 Medium**: 1 (Links in new tab)

**Proposed Sprints**: 2
- **Phase 11** (Critical Fixes): 3 issues, ~3-7 days effort
- **Phase 12** (URL Enhancements): 3 issues, ~3-5 days effort

## Status Key
- 🆕 New - Not yet triaged
- 📋 Backlog - Triaged, not scheduled
- 🎯 Planned - Scheduled for upcoming sprint
- 🚧 In Progress - Currently being worked on
- ✅ Complete - Implemented and merged
- ❌ Wontfix - Decided not to implement

---

## Sprint Planning

### Proposed Sprint Structure

#### Sprint 1 (Phase 11): Critical Fixes
**Focus**: Fix broken core functionality
**Timeline**: 2-4 days
**Priority**: P0-P1 Critical issues

**Items**:
1. 🆕 Search functionality broken (P0) - 1-3 days
2. 🆕 Expand/collapse buttons not working (P1) - 1-3 days
3. 🆕 Newlines being duplicated (P1) - <1 day (quick win)

**Rationale**: These are core features that users expect to work. Search and expand/collapse are completely broken, making navigation difficult. Newlines issue affects readability throughout.

**Total Effort**: ~3-7 days

---

#### Sprint 2 (Phase 12): URL & Navigation Enhancements
**Focus**: Improve navigation and link handling
**Timeline**: 2-4 days
**Priority**: P1-P2 High-value enhancements

**Items**:
1. 🆕 Clickable URLs in output (P1) - <1 day (foundational)
2. 🆕 All links open in new tab (P2) - <1 day (builds on #1)
3. 🆕 File path URL generation (P1) - 1-3 days (more complex)

**Rationale**: These features work together to improve navigation. Start with basic URL clickability, add new-tab behavior, then tackle more complex file path detection. Each builds on the previous.

**Total Effort**: ~3-5 days

---

### Future Sprints
(Additional sprints will be added as more issues are identified)

---

### Backlog - Unprioritized

#### Category: HTML Export Quality
(Issues related to HTML export appearance, functionality, or usability)

##### 🆕 File Path URL Generation

**Priority**: P1: High
**Effort**: Medium: 1-3 days

**Description**:
Create clickable URLs for file paths mentioned in agent output. When an agent references a file (e.g., "Completed: gastown/schema.md"), automatically convert it to a clickable link with the full file path.

**Context**:
- Agent output frequently references files it worked on
- Users need quick access to these files
- We know the agent's working directory from session metadata
- Must verify file exists before creating link (avoid broken links)

**Proposed Solution**:
1. Parse text content for file path patterns (e.g., `path/to/file.ext`, `./relative/path`)
2. Resolve relative paths against agent's working directory
3. Check file existence with filesystem API
4. Generate `file://` URLs for existing files
5. Wrap matched paths in `<a>` tags

**Acceptance Criteria**:
- [ ] File paths in agent output are automatically linkified
- [ ] Links only created for files that exist on disk
- [ ] Relative paths resolved correctly from agent's working directory
- [ ] Common file patterns recognized (with and without `./`)
- [ ] Links open in new tab

---

##### 🆕 Clickable URLs in Output

**Priority**: P1: High
**Effort**: Small: <1 day

**Description**:
Automatically convert URLs in agent output to clickable links. This includes `http://`, `https://`, `localhost`, and `file://` URLs.

**Context**:
- Agents often output URLs (e.g., "Server running at http://localhost:3000")
- Users should be able to click these without copy/paste
- No verification needed - just make them clickable

**Proposed Solution**:
1. Use regex to detect URL patterns in text content
2. Match: `http://`, `https://`, `localhost:`, `file://`
3. Wrap in `<a href="..." target="_blank">` tags
4. Apply to all text content (user, assistant, tool results)

**Acceptance Criteria**:
- [ ] HTTP/HTTPS URLs are clickable
- [ ] localhost URLs are clickable (with or without http://)
- [ ] file:// URLs are clickable
- [ ] All links open in new tab (`target="_blank"`)
- [ ] URL detection doesn't break on URLs with query params, fragments, or ports

---

##### 🆕 All Links Open in New Tab

**Priority**: P2: Medium
**Effort**: Small: <1 day

**Description**:
Ensure all links in the HTML export open in a new tab to prevent navigating away from the conversation view.

**Context**:
- Users should not lose their place in the conversation
- External links, file links, and reference links should all open externally
- Standard UX pattern for web apps

**Proposed Solution**:
- Add `target="_blank"` and `rel="noopener noreferrer"` to all `<a>` tags
- Apply to: file paths, URLs, GitHub links, any other generated links

**Acceptance Criteria**:
- [ ] All links have `target="_blank"`
- [ ] All external links have `rel="noopener noreferrer"` for security
- [ ] Conversation view remains open when clicking links

---

##### 🆕 Newlines Being Duplicated

**Priority**: P1: High
**Effort**: Small: <1 day

**Description**:
Text content is showing doubled newlines - a single newline in the source is rendering as two line breaks in the HTML.

**Context**:
- User reported seeing double-spaced text where single-spacing was expected
- Example: "All 5 agents completed...\n\nCompleted: gastown/schema.md\n\nThe document covers..."
- Affects readability and visual density

**Proposed Solution**:
1. Investigate where newlines are being processed (markdown rendering, escapeHTML, formatUserContent)
2. Check if `<pre>` tags or CSS `white-space: pre-wrap` are doubling `\n`
3. Likely culprit: converting `\n` to `<br>\n` when CSS already preserves newlines
4. Fix: Use either CSS `white-space: pre-wrap` OR explicit `<br>` tags, not both

**Acceptance Criteria**:
- [ ] Single newlines in source render as single line breaks
- [ ] Double newlines (paragraph breaks) render correctly
- [ ] No visual doubling of spacing
- [ ] Test with user messages, assistant messages, and tool results

---

##### 🆕 Search Functionality Broken

**Priority**: P0: Critical
**Effort**: Medium: 1-3 days

**Description**:
The search feature in the HTML export does not work at all. The search box is visible but typing and searching produces no results.

**Context**:
- Search is a critical feature for navigating long conversations
- User explicitly tested and confirmed it's broken
- Search box and controls are rendered but non-functional

**Proposed Solution**:
1. Check if `controls.js` is being loaded and executed
2. Verify search event listeners are attached
3. Test search logic against conversation DOM structure
4. Common issues:
   - JS not loading due to path issues
   - Event listeners not attaching (timing issue)
   - Search selector not matching actual HTML structure
   - Case sensitivity or whitespace handling

**Acceptance Criteria**:
- [ ] Search box accepts input
- [ ] Search highlights matches in conversation
- [ ] Next/Previous buttons navigate between matches
- [ ] Match counter shows "X of Y matches"
- [ ] Search works case-insensitively
- [ ] Clear/Escape clears search results

---

##### 🆕 Expand/Collapse All Buttons Not Working

**Priority**: P1: High
**Effort**: Medium: 1-3 days

**Description**:
The "Expand All" and "Collapse All" buttons in the page header do not function. Clicking them has no effect on tool call visibility.

**Context**:
- Buttons are visible in the header but non-functional
- Users expect these to expand/collapse all tool call details blocks
- Part of the keyboard shortcuts feature (Ctrl+K)

**Proposed Solution**:
1. Check if `controls.js` is loaded and event listeners attached
2. Verify buttons have correct IDs: `#expand-all-btn`, `#collapse-all-btn`
3. Check if `<details>` elements are being targeted correctly
4. Test keyboard shortcut (Ctrl+K) separately
5. Common issues:
   - JS not executing (path or timing issue)
   - Selector mismatch between JS and HTML structure
   - Event listener not attached

**Acceptance Criteria**:
- [ ] "Expand All" button expands all tool call details
- [ ] "Collapse All" button collapses all tool call details
- [ ] Ctrl+K keyboard shortcut toggles expand/collapse
- [ ] Buttons work immediately on page load
- [ ] Visual feedback when buttons are clicked

#### Category: CLI Features
(New commands, flags, or core functionality)

#### Category: Performance
(Speed, memory usage, scalability)

#### Category: Testing
(Test coverage, test infrastructure, edge cases)

#### Category: Documentation
(README, docs, examples, help text)

#### Category: Developer Experience
(Build process, tooling, debugging)

#### Category: Security
(Input validation, XSS prevention, data handling)

#### Category: Compatibility
(Cross-platform support, version compatibility)

---

## Issue Template

When adding issues, use this format:

```markdown
### [Issue Title] - Status

**Category**: (HTML Export | CLI | Performance | Testing | etc.)
**Priority**: (P0: Critical | P1: High | P2: Medium | P3: Low)
**Effort**: (Small: <1 day | Medium: 1-3 days | Large: >3 days)

**Description**:
Clear description of the issue or improvement

**Context**:
- Why this matters
- Impact on users
- Related issues

**Proposed Solution**:
Suggested approach or alternatives

**Acceptance Criteria**:
- [ ] Specific testable outcome 1
- [ ] Specific testable outcome 2
```

---

## Notes

- Issues are added as discovered during development or user feedback
- Sprint planning happens periodically to move items from backlog to planned
- Priority and effort estimates help with sprint capacity planning
- Use GitHub issues for tracking implementation once work begins
