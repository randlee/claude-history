/**
 * Claude History Export - Interactive JavaScript
 */

/**
 * Toggle visibility of a tool call body.
 * @param {HTMLElement} header - The tool header element
 */
function toggleTool(header) {
    var body = header.nextElementSibling;
    if (body && body.classList.contains('tool-body')) {
        body.classList.toggle('hidden');

        // Update toggle indicator
        var toggle = header.querySelector('.tool-toggle');
        if (toggle) {
            toggle.textContent = body.classList.contains('hidden') ? '[+]' : '[-]';
        }
    }
}

/**
 * Load subagent content via AJAX.
 * @param {HTMLElement} header - The subagent header element
 */
function loadAgent(header) {
    var container = header.nextElementSibling;
    var parent = header.parentElement;
    var agentId = parent.dataset.agentId;

    // Already loaded check
    if (container.innerHTML.trim() !== '' && !container.querySelector('.subagent-loading')) {
        // Just toggle visibility if already loaded
        container.classList.toggle('collapsed');
        return;
    }

    // Show loading state
    container.innerHTML = '<p class="subagent-loading">Loading agent content...</p>';

    // Fetch agent HTML
    fetch('agents/' + agentId + '.html')
        .then(function(response) {
            if (!response.ok) {
                throw new Error('Failed to load agent');
            }
            return response.text();
        })
        .then(function(html) {
            container.innerHTML = html;
            // Initialize any tool toggles in the loaded content
            initToolToggles(container);
        })
        .catch(function(error) {
            container.innerHTML = '<p class="subagent-error">Failed to load agent: ' + error.message + '</p>';
        });
}

/**
 * Copy code block content to clipboard.
 * @param {HTMLElement} button - The copy button element
 */
function copyCode(button) {
    var codeBlock = button.closest('.code-block');
    var codeContent = codeBlock.querySelector('.code-content code');

    if (!codeContent) {
        codeContent = codeBlock.querySelector('.code-content');
    }

    if (codeContent) {
        var text = codeContent.textContent;

        // Use modern clipboard API if available
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(text)
                .then(function() {
                    showCopySuccess(button);
                })
                .catch(function(err) {
                    fallbackCopy(text, button);
                });
        } else {
            fallbackCopy(text, button);
        }
    }
}

/**
 * Fallback copy using textarea element.
 * @param {string} text - The text to copy
 * @param {HTMLElement} button - The copy button element
 */
function fallbackCopy(text, button) {
    var textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.left = '-9999px';
    textarea.style.top = '0';
    document.body.appendChild(textarea);
    textarea.focus();
    textarea.select();

    try {
        document.execCommand('copy');
        showCopySuccess(button);
    } catch (err) {
        console.error('Failed to copy:', err);
        button.textContent = 'Failed';
        setTimeout(function() {
            button.textContent = 'Copy';
        }, 2000);
    }

    document.body.removeChild(textarea);
}

/**
 * Show copy success feedback on button.
 * @param {HTMLElement} button - The copy button element
 */
function showCopySuccess(button) {
    var originalText = button.textContent;
    button.textContent = 'Copied!';
    button.classList.add('copied');

    setTimeout(function() {
        button.textContent = originalText;
        button.classList.remove('copied');
    }, 2000);
}

/**
 * Expand all tool call bodies.
 */
function expandAll() {
    var bodies = document.querySelectorAll('.tool-body');
    bodies.forEach(function(el) {
        el.classList.remove('hidden');
    });

    // Update all toggle indicators
    var toggles = document.querySelectorAll('.tool-toggle');
    toggles.forEach(function(el) {
        el.textContent = '[-]';
    });
}

/**
 * Collapse all tool call bodies.
 */
function collapseAll() {
    var bodies = document.querySelectorAll('.tool-body');
    bodies.forEach(function(el) {
        el.classList.add('hidden');
    });

    // Update all toggle indicators
    var toggles = document.querySelectorAll('.tool-toggle');
    toggles.forEach(function(el) {
        el.textContent = '[+]';
    });
}

/**
 * Initialize tool toggle click handlers within a container.
 * @param {HTMLElement} container - The container element
 */
function initToolToggles(container) {
    var headers = container.querySelectorAll('.tool-header');
    headers.forEach(function(header) {
        // Remove any existing click handler
        header.onclick = null;
        // Add click handler
        header.onclick = function() {
            toggleTool(this);
        };
    });
}

/**
 * Initialize subagent header click handlers.
 * @param {HTMLElement} container - The container element
 */
function initSubagentHeaders(container) {
    var headers = container.querySelectorAll('.subagent-header');
    headers.forEach(function(header) {
        header.onclick = function() {
            loadAgent(this);
        };
    });
}

/**
 * Jump to a specific entry by index.
 * @param {number} index - The entry index (0-based)
 */
function jumpToEntry(index) {
    var entries = document.querySelectorAll('.entry');
    if (index >= 0 && index < entries.length) {
        entries[index].scrollIntoView({ behavior: 'smooth', block: 'start' });
    }
}

/**
 * Search for text within entries.
 * @param {string} query - The search query
 * @returns {Array} Array of matching entry elements
 */
function searchEntries(query) {
    var entries = document.querySelectorAll('.entry');
    var matches = [];
    var lowerQuery = query.toLowerCase();

    entries.forEach(function(entry) {
        var content = entry.textContent.toLowerCase();
        if (content.includes(lowerQuery)) {
            matches.push(entry);
        }
    });

    return matches;
}

/**
 * Highlight search results.
 * @param {string} query - The search query
 */
function highlightSearch(query) {
    // Clear previous highlights
    clearHighlights();

    if (!query) return;

    var matches = searchEntries(query);
    matches.forEach(function(entry) {
        entry.classList.add('search-match');
    });

    // Scroll to first match
    if (matches.length > 0) {
        matches[0].scrollIntoView({ behavior: 'smooth', block: 'center' });
    }

    return matches.length;
}

/**
 * Clear all search highlights.
 */
function clearHighlights() {
    var highlighted = document.querySelectorAll('.search-match');
    highlighted.forEach(function(el) {
        el.classList.remove('search-match');
    });
}

/**
 * Get conversation statistics.
 * @returns {Object} Statistics object
 */
function getStats() {
    var entries = document.querySelectorAll('.entry');
    var stats = {
        total: entries.length,
        user: document.querySelectorAll('.entry.user').length,
        assistant: document.querySelectorAll('.entry.assistant').length,
        system: document.querySelectorAll('.entry.system').length,
        toolCalls: document.querySelectorAll('.tool-call').length,
        subagents: document.querySelectorAll('.subagent').length
    };
    return stats;
}

/**
 * Initialize the page when DOM is ready.
 */
function init() {
    // Initialize all tool toggles
    initToolToggles(document);

    // Initialize subagent headers
    initSubagentHeaders(document);

    // Start with tool bodies collapsed
    collapseAll();
}

// Run init when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}
