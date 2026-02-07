/**
 * Claude History Export - Interactive Controls
 * Provides expand/collapse, search, keyboard shortcuts, and state persistence.
 */

(function() {
    'use strict';

    // ===========================================
    // CONFIGURATION
    // ===========================================

    var STORAGE_KEY = 'claude-history-controls-state';
    var SEARCH_HIGHLIGHT_CLASS = 'search-highlight';
    var SEARCH_MATCH_CLASS = 'search-match';
    var HIDDEN_BY_SEARCH_CLASS = 'hidden-by-search';

    // ===========================================
    // STATE MANAGEMENT
    // ===========================================

    /**
     * Load saved state from localStorage.
     * @returns {Object} The saved state or default values
     */
    function loadState() {
        try {
            var saved = localStorage.getItem(STORAGE_KEY);
            if (saved) {
                return JSON.parse(saved);
            }
        } catch (e) {
            console.warn('Failed to load controls state:', e);
        }
        return {
            collapsedIds: [],
            allCollapsed: true
        };
    }

    /**
     * Save state to localStorage.
     * @param {Object} state - The state to save
     */
    function saveState(state) {
        try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
        } catch (e) {
            console.warn('Failed to save controls state:', e);
        }
    }

    /**
     * Get current collapse state of all collapsible elements.
     * @returns {Object} State object with collapsed element IDs
     */
    function getCurrentState() {
        var collapsedIds = [];
        var collapsibles = document.querySelectorAll('.tool-body');

        collapsibles.forEach(function(el, index) {
            if (el.classList.contains('hidden')) {
                var toolCall = el.closest('.tool-call');
                var id = toolCall ? toolCall.dataset.toolId : 'tool-' + index;
                collapsedIds.push(id);
            }
        });

        return {
            collapsedIds: collapsedIds,
            allCollapsed: collapsedIds.length === collapsibles.length
        };
    }

    /**
     * Restore collapse state from saved state.
     * @param {Object} state - The saved state to restore
     */
    function restoreState(state) {
        if (!state) return;

        if (state.allCollapsed) {
            collapseAllTools();
        } else {
            // Restore individual states
            expandAllTools();
            state.collapsedIds.forEach(function(id) {
                var toolCall = document.querySelector('.tool-call[data-tool-id="' + id + '"]');
                if (toolCall) {
                    var body = toolCall.querySelector('.tool-body');
                    if (body) {
                        body.classList.add('hidden');
                    }
                }
            });
        }
    }

    // ===========================================
    // EXPAND/COLLAPSE FUNCTIONALITY
    // ===========================================

    /**
     * Expand all collapsible tool bodies.
     */
    function expandAllTools() {
        var bodies = document.querySelectorAll('.tool-body');
        bodies.forEach(function(el) {
            el.classList.remove('hidden');
        });

        // Handle <details> elements
        var detailsElements = document.querySelectorAll('details');
        detailsElements.forEach(function(el) {
            el.open = true;
        });

        // Update toggle indicators
        updateAllToggles(true);

        // Save state
        saveState({ collapsedIds: [], allCollapsed: false });
    }

    /**
     * Collapse all collapsible tool bodies.
     */
    function collapseAllTools() {
        var bodies = document.querySelectorAll('.tool-body');
        bodies.forEach(function(el) {
            el.classList.add('hidden');
        });

        // Handle <details> elements
        var detailsElements = document.querySelectorAll('details');
        detailsElements.forEach(function(el) {
            el.open = false;
        });

        // Update toggle indicators
        updateAllToggles(false);

        // Save state
        saveState(getCurrentState());
    }

    /**
     * Toggle all collapsibles between expanded and collapsed.
     */
    function toggleAllTools() {
        var bodies = document.querySelectorAll('.tool-body');
        var allHidden = true;

        bodies.forEach(function(el) {
            if (!el.classList.contains('hidden')) {
                allHidden = false;
            }
        });

        if (allHidden) {
            expandAllTools();
        } else {
            collapseAllTools();
        }
    }

    /**
     * Update all toggle indicators.
     * @param {boolean} expanded - Whether elements are expanded
     */
    function updateAllToggles(expanded) {
        var toggles = document.querySelectorAll('.tool-toggle');
        toggles.forEach(function(el) {
            el.textContent = expanded ? '[-]' : '[+]';
        });
    }

    /**
     * Toggle a single tool body.
     * @param {HTMLElement} header - The tool header element
     */
    function toggleSingleTool(header) {
        var body = header.nextElementSibling;
        if (body && body.classList.contains('tool-body')) {
            var isHidden = body.classList.toggle('hidden');

            // Update toggle indicator
            var toggle = header.querySelector('.tool-toggle');
            if (toggle) {
                toggle.textContent = isHidden ? '[+]' : '[-]';
            }

            // Save state
            saveState(getCurrentState());
        }
    }

    // ===========================================
    // SEARCH FUNCTIONALITY
    // ===========================================

    var currentSearchIndex = -1;
    var currentMatches = [];

    /**
     * Perform search and highlight matches.
     * @param {string} query - The search query
     * @returns {number} Number of matches found
     */
    function performSearch(query) {
        // Clear previous search
        clearSearch();

        if (!query || query.trim() === '') {
            updateSearchResults(0, 0);
            return 0;
        }

        var lowerQuery = query.toLowerCase().trim();
        var entries = document.querySelectorAll('.message-row');
        currentMatches = [];
        currentSearchIndex = -1;

        entries.forEach(function(entry) {
            var content = entry.querySelector('.message-content');
            if (!content) return;

            var textContent = content.textContent.toLowerCase();
            if (textContent.indexOf(lowerQuery) !== -1) {
                entry.classList.add(SEARCH_MATCH_CLASS);
                currentMatches.push(entry);

                // Highlight text matches
                highlightTextInElement(content, query);
            } else {
                entry.classList.add(HIDDEN_BY_SEARCH_CLASS);
            }
        });

        updateSearchResults(currentMatches.length, 0);

        // Navigate to first match if any
        if (currentMatches.length > 0) {
            navigateToMatch(0);
        }

        return currentMatches.length;
    }

    /**
     * Highlight text matches within an element.
     * @param {HTMLElement} element - The element to search within
     * @param {string} query - The text to highlight
     */
    function highlightTextInElement(element, query) {
        var walker = document.createTreeWalker(
            element,
            NodeFilter.SHOW_TEXT,
            null,
            false
        );

        var textNodes = [];
        var node;
        while ((node = walker.nextNode())) {
            textNodes.push(node);
        }

        textNodes.forEach(function(textNode) {
            var text = textNode.textContent;
            var lowerText = text.toLowerCase();
            var lowerQuery = query.toLowerCase();
            var index = lowerText.indexOf(lowerQuery);

            if (index !== -1) {
                var fragment = document.createDocumentFragment();
                var lastIndex = 0;

                while (index !== -1) {
                    // Add text before match
                    if (index > lastIndex) {
                        fragment.appendChild(document.createTextNode(text.substring(lastIndex, index)));
                    }

                    // Add highlighted match
                    var mark = document.createElement('mark');
                    mark.className = SEARCH_HIGHLIGHT_CLASS;
                    mark.textContent = text.substring(index, index + query.length);
                    fragment.appendChild(mark);

                    lastIndex = index + query.length;
                    index = lowerText.indexOf(lowerQuery, lastIndex);
                }

                // Add remaining text
                if (lastIndex < text.length) {
                    fragment.appendChild(document.createTextNode(text.substring(lastIndex)));
                }

                textNode.parentNode.replaceChild(fragment, textNode);
            }
        });
    }

    /**
     * Clear all search highlights and visibility states.
     */
    function clearSearch() {
        // Remove highlight marks
        var marks = document.querySelectorAll('mark.' + SEARCH_HIGHLIGHT_CLASS);
        marks.forEach(function(mark) {
            var parent = mark.parentNode;
            parent.replaceChild(document.createTextNode(mark.textContent), mark);
            parent.normalize();
        });

        // Remove match class
        var matches = document.querySelectorAll('.' + SEARCH_MATCH_CLASS);
        matches.forEach(function(el) {
            el.classList.remove(SEARCH_MATCH_CLASS);
        });

        // Remove hidden class
        var hidden = document.querySelectorAll('.' + HIDDEN_BY_SEARCH_CLASS);
        hidden.forEach(function(el) {
            el.classList.remove(HIDDEN_BY_SEARCH_CLASS);
        });

        currentMatches = [];
        currentSearchIndex = -1;
        updateSearchResults(0, 0);
    }

    /**
     * Expand any collapsed parent sections containing the element.
     * @param {HTMLElement} element - The element whose parents should be expanded
     */
    function expandParentSections(element) {
        // Walk up the DOM tree and expand any collapsed parents
        var parent = element.parentElement;
        while (parent) {
            // Expand <details> elements
            if (parent.tagName === 'DETAILS' && !parent.open) {
                parent.open = true;
            }

            // Expand hidden tool-body sections
            if (parent.classList && parent.classList.contains('tool-body')) {
                var wasHidden = parent.classList.contains('hidden') || parent.classList.contains('collapsed');
                if (wasHidden) {
                    parent.classList.remove('hidden');
                    parent.classList.remove('collapsed');

                    // Also remove collapsed from parent tool-call container
                    var toolCall = parent.closest('.tool-call');
                    if (toolCall && toolCall.classList.contains('collapsed')) {
                        toolCall.classList.remove('collapsed');
                    }

                    // Update the toggle indicator for this tool
                    var toolHeader = parent.previousElementSibling;
                    if (toolHeader && toolHeader.classList.contains('tool-header')) {
                        var toggle = toolHeader.querySelector('.tool-toggle');
                        if (toggle) {
                            toggle.textContent = '[-]';
                        }
                        // Note: Chevron rotation is handled by CSS based on .collapsed class
                    }
                }
            }

            parent = parent.parentElement;
        }
    }

    /**
     * Navigate to a specific match by index.
     * @param {number} index - The match index to navigate to
     */
    function navigateToMatch(index) {
        if (currentMatches.length === 0) return;

        // Wrap around
        if (index < 0) index = currentMatches.length - 1;
        if (index >= currentMatches.length) index = 0;

        currentSearchIndex = index;

        // Remove active class from previous
        currentMatches.forEach(function(match) {
            match.classList.remove('search-active');
        });

        // Add active class to current
        var current = currentMatches[currentSearchIndex];
        current.classList.add('search-active');

        // Auto-expand collapsed parent sections before scrolling
        expandParentSections(current);

        // Wait a tick for DOM updates to render before scrolling
        setTimeout(function() {
            smoothScrollToElement(current);
        }, 10);

        updateSearchResults(currentMatches.length, currentSearchIndex + 1);
    }

    /**
     * Navigate to the next search match.
     */
    function nextMatch() {
        navigateToMatch(currentSearchIndex + 1);
    }

    /**
     * Navigate to the previous search match.
     */
    function prevMatch() {
        navigateToMatch(currentSearchIndex - 1);
    }

    /**
     * Update the search results display.
     * @param {number} total - Total number of matches
     * @param {number} current - Current match number (1-indexed)
     */
    function updateSearchResults(total, current) {
        var resultsEl = document.querySelector('.search-results');
        if (resultsEl) {
            if (total > 0) {
                resultsEl.textContent = current + ' of ' + total;
                resultsEl.classList.add('has-results');
            } else {
                var searchBox = document.getElementById('search-box');
                if (searchBox && searchBox.value.trim() !== '') {
                    resultsEl.textContent = 'No matches';
                    resultsEl.classList.remove('has-results');
                } else {
                    resultsEl.textContent = '';
                    resultsEl.classList.remove('has-results');
                }
            }
        }
    }

    // ===========================================
    // SMOOTH SCROLL
    // ===========================================

    /**
     * Smoothly scroll to an element.
     * @param {HTMLElement} element - The element to scroll to
     */
    function smoothScrollToElement(element) {
        if (!element) return;

        // Use scrollIntoView with center alignment for better visibility
        element.scrollIntoView({
            behavior: 'smooth',
            block: 'center',
            inline: 'nearest'
        });
    }

    // ===========================================
    // KEYBOARD SHORTCUTS
    // ===========================================

    /**
     * Initialize keyboard shortcut listeners.
     */
    function initKeyboardShortcuts() {
        document.addEventListener('keydown', function(e) {
            // Ctrl+K or Cmd+K - Toggle all collapsibles
            if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
                e.preventDefault();
                toggleAllTools();
                return;
            }

            // Ctrl+F or Cmd+F - Focus search box
            if ((e.ctrlKey || e.metaKey) && e.key === 'f') {
                e.preventDefault();
                focusSearchBox();
                return;
            }

            // Escape - Close search/clear
            if (e.key === 'Escape') {
                var searchBox = document.getElementById('search-box');
                if (searchBox && document.activeElement === searchBox) {
                    searchBox.blur();
                    if (searchBox.value !== '') {
                        searchBox.value = '';
                        clearSearch();
                    }
                }
                return;
            }

            // Enter in search box - navigate to next match
            if (e.key === 'Enter') {
                var searchBox = document.getElementById('search-box');
                if (document.activeElement === searchBox) {
                    e.preventDefault();
                    if (e.shiftKey) {
                        prevMatch();
                    } else {
                        nextMatch();
                    }
                }
                return;
            }
        });
    }

    /**
     * Focus the search box.
     */
    function focusSearchBox() {
        var searchBox = document.getElementById('search-box');
        if (searchBox) {
            searchBox.focus();
            searchBox.select();
        }
    }

    // ===========================================
    // SCROLL SHADOW FOR HEADER
    // ===========================================

    /**
     * Initialize scroll listener for header shadow effect.
     */
    function initScrollShadow() {
        var header = document.querySelector('.page-header');
        if (!header) return;

        var lastScrollY = 0;
        var ticking = false;

        function updateHeaderShadow() {
            if (window.scrollY > 10) {
                header.classList.add('scrolled');
            } else {
                header.classList.remove('scrolled');
            }
            ticking = false;
        }

        window.addEventListener('scroll', function() {
            lastScrollY = window.scrollY;
            if (!ticking) {
                window.requestAnimationFrame(function() {
                    updateHeaderShadow();
                });
                ticking = true;
            }
        }, { passive: true });

        // Initial check
        updateHeaderShadow();
    }

    // ===========================================
    // INITIALIZATION
    // ===========================================

    /**
     * Initialize control panel event listeners.
     */
    function initControlPanel() {
        // Expand All button
        var expandBtn = document.getElementById('expand-all-btn');
        if (expandBtn) {
            expandBtn.addEventListener('click', function() {
                expandAllTools();
            });
        }

        // Collapse All button
        var collapseBtn = document.getElementById('collapse-all-btn');
        if (collapseBtn) {
            collapseBtn.addEventListener('click', function() {
                collapseAllTools();
            });
        }

        // Search box
        var searchBox = document.getElementById('search-box');
        if (searchBox) {
            var debounceTimer;
            searchBox.addEventListener('input', function() {
                clearTimeout(debounceTimer);
                debounceTimer = setTimeout(function() {
                    performSearch(searchBox.value);
                }, 200);
            });

            // Clear button functionality
            searchBox.addEventListener('search', function() {
                if (searchBox.value === '') {
                    clearSearch();
                }
            });
        }

        // Previous/Next buttons
        var prevBtn = document.getElementById('search-prev-btn');
        if (prevBtn) {
            prevBtn.addEventListener('click', function() {
                prevMatch();
            });
        }

        var nextBtn = document.getElementById('search-next-btn');
        if (nextBtn) {
            nextBtn.addEventListener('click', function() {
                nextMatch();
            });
        }
    }

    /**
     * Make tool headers collapsible with state tracking.
     */
    function initCollapsibleToolHeaders() {
        var headers = document.querySelectorAll('.tool-header');
        headers.forEach(function(header) {
            // Skip if already has onclick
            if (header.onclick) return;

            header.addEventListener('click', function(e) {
                // Don't toggle if clicking on a button inside the header
                if (e.target.closest('.copy-btn')) return;
                toggleSingleTool(header);
            });
        });
    }

    /**
     * Initialize all controls functionality.
     */
    function initControls() {
        // Initialize control panel buttons
        initControlPanel();

        // Initialize keyboard shortcuts
        initKeyboardShortcuts();

        // Make tool headers collapsible
        initCollapsibleToolHeaders();

        // Initialize scroll shadow effect
        initScrollShadow();

        // Restore saved state
        var savedState = loadState();
        restoreState(savedState);
    }

    // ===========================================
    // PUBLIC API
    // ===========================================

    // Expose functions globally for use by other scripts
    window.ControlsAPI = {
        expandAll: expandAllTools,
        collapseAll: collapseAllTools,
        toggleAll: toggleAllTools,
        search: performSearch,
        clearSearch: clearSearch,
        nextMatch: nextMatch,
        prevMatch: prevMatch,
        focusSearch: focusSearchBox,
        scrollTo: smoothScrollToElement,
        expandParents: expandParentSections,
        getState: getCurrentState,
        loadState: loadState,
        saveState: saveState
    };

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initControls);
    } else {
        initControls();
    }

})();
