/**
 * Claude History Export - Deep Dive Navigation
 * Provides navigation between nested agents with breadcrumb trail and lazy loading.
 */

(function() {
    'use strict';

    // ===========================================
    // CONFIGURATION
    // ===========================================

    var NAV_STORAGE_KEY = 'claude-history-navigation-state';
    var BREADCRUMB_CONTAINER_ID = 'breadcrumbs';
    var MAIN_SESSION_ID = 'main';

    // ===========================================
    // STATE MANAGEMENT
    // ===========================================

    /**
     * Navigation state tracking.
     */
    var state = {
        currentPath: [], // Empty - no "Main Session" breadcrumb
        history: [],
        historyIndex: -1,
        loadedAgents: {},
        expandedAgents: {}
    };

    /**
     * Load navigation state from localStorage.
     * @returns {Object} The saved state or default values
     */
    function loadNavigationState() {
        try {
            var saved = localStorage.getItem(NAV_STORAGE_KEY);
            if (saved) {
                var parsed = JSON.parse(saved);
                // Merge with defaults
                state.loadedAgents = parsed.loadedAgents || {};
                state.expandedAgents = parsed.expandedAgents || {};
            }
        } catch (e) {
            console.warn('Failed to load navigation state:', e);
        }
        return state;
    }

    /**
     * Save navigation state to localStorage.
     */
    function saveNavigationState() {
        try {
            var toSave = {
                loadedAgents: state.loadedAgents,
                expandedAgents: state.expandedAgents
            };
            localStorage.setItem(NAV_STORAGE_KEY, JSON.stringify(toSave));
        } catch (e) {
            console.warn('Failed to save navigation state:', e);
        }
    }

    // ===========================================
    // BREADCRUMB MANAGEMENT
    // ===========================================

    /**
     * Update the breadcrumb trail display.
     * @param {Array} path - Array of {id, label} objects representing the navigation path
     */
    function updateBreadcrumbs(path) {
        var container = document.getElementById(BREADCRUMB_CONTAINER_ID);
        if (!container) return;

        state.currentPath = path || state.currentPath;

        // Clear existing breadcrumbs
        container.innerHTML = '';

        state.currentPath.forEach(function(item, index) {
            // Create breadcrumb link
            var link = document.createElement('a');
            link.href = '#' + item.id;
            link.className = 'breadcrumb-item';
            link.textContent = item.label;
            link.setAttribute('data-agent-id', item.id);

            // Mark current (last) item as active
            if (index === state.currentPath.length - 1) {
                link.classList.add('active');
                link.setAttribute('aria-current', 'page');
            } else {
                // Add click handler for non-active items
                link.addEventListener('click', function(e) {
                    e.preventDefault();
                    navigateToBreadcrumb(index);
                });
            }

            container.appendChild(link);

            // Add separator (except after last item)
            if (index < state.currentPath.length - 1) {
                var separator = document.createElement('span');
                separator.className = 'breadcrumb-separator';
                separator.setAttribute('aria-hidden', 'true');
                separator.innerHTML = '&#8250;'; // Right angle quote
                container.appendChild(separator);
            }
        });
    }

    /**
     * Navigate to a specific breadcrumb position.
     * @param {number} index - The breadcrumb index to navigate to
     */
    function navigateToBreadcrumb(index) {
        if (index < 0 || index >= state.currentPath.length) return;

        // Truncate path to the clicked position
        var newPath = state.currentPath.slice(0, index + 1);
        var targetId = newPath[newPath.length - 1].id;

        // Collapse agents after this point
        for (var i = index + 1; i < state.currentPath.length; i++) {
            var agentId = state.currentPath[i].id;
            if (agentId !== MAIN_SESSION_ID) {
                collapseSubagent(agentId);
            }
        }

        // Update path and scroll
        updateBreadcrumbs(newPath);
        addToHistory(newPath);
        scrollToAgent(targetId);
    }

    /**
     * Add current path to navigation history.
     * @param {Array} path - The navigation path
     */
    function addToHistory(path) {
        // Remove any forward history
        if (state.historyIndex < state.history.length - 1) {
            state.history = state.history.slice(0, state.historyIndex + 1);
        }

        // Add new path (deep copy)
        state.history.push(JSON.parse(JSON.stringify(path)));
        state.historyIndex = state.history.length - 1;
    }

    /**
     * Navigate back in history.
     * @returns {boolean} True if navigation occurred
     */
    function navigateBack() {
        if (state.historyIndex <= 0) return false;

        state.historyIndex--;
        var previousPath = state.history[state.historyIndex];
        updateBreadcrumbs(previousPath);

        var targetId = previousPath[previousPath.length - 1].id;
        scrollToAgent(targetId);

        return true;
    }

    /**
     * Navigate forward in history.
     * @returns {boolean} True if navigation occurred
     */
    function navigateForward() {
        if (state.historyIndex >= state.history.length - 1) return false;

        state.historyIndex++;
        var nextPath = state.history[state.historyIndex];
        updateBreadcrumbs(nextPath);

        var targetId = nextPath[nextPath.length - 1].id;
        scrollToAgent(targetId);

        return true;
    }

    // ===========================================
    // SUBAGENT EXPANSION/COLLAPSE
    // ===========================================

    /**
     * Expand a subagent and load its content.
     * @param {string} agentId - The agent ID to expand
     * @param {Object} options - Optional settings {scrollTo, updateBreadcrumbs}
     */
    function expandSubagent(agentId, options) {
        options = options || {};
        var overlay = document.querySelector('[data-agent-id="' + agentId + '"]');
        if (!overlay) return;

        var content = overlay.querySelector('.subagent-content');
        if (!content) return;

        // Check if already loaded
        if (state.loadedAgents[agentId] && content.innerHTML.trim() !== '' &&
            !content.querySelector('.subagent-loading')) {
            // Already loaded, just expand
            overlay.classList.add('expanded');
            content.classList.remove('collapsed');
            state.expandedAgents[agentId] = true;
            saveNavigationState();

            if (options.updateBreadcrumbs !== false) {
                addAgentToBreadcrumbs(agentId, overlay);
            }

            if (options.scrollTo !== false) {
                scrollToAgent(agentId);
            }
            return;
        }

        // Show loading state
        content.innerHTML = '<p class="subagent-loading">Loading agent content...</p>';
        content.classList.remove('collapsed');
        overlay.classList.add('expanded');

        // Fetch agent HTML (lazy load)
        fetch('agents/' + agentId + '.html')
            .then(function(response) {
                if (!response.ok) {
                    throw new Error('Failed to load agent (status: ' + response.status + ')');
                }
                return response.text();
            })
            .then(function(html) {
                content.innerHTML = html;
                state.loadedAgents[agentId] = true;
                state.expandedAgents[agentId] = true;
                saveNavigationState();

                // Initialize nested components
                initNestedComponents(content);

                if (options.updateBreadcrumbs !== false) {
                    addAgentToBreadcrumbs(agentId, overlay);
                }

                if (options.scrollTo !== false) {
                    scrollToAgent(agentId);
                }
            })
            .catch(function(error) {
                content.innerHTML = '<p class="subagent-error">Failed to load agent: ' +
                    escapeHtml(error.message) + '</p>';
                console.error('Failed to load agent ' + agentId + ':', error);
            });
    }

    /**
     * Collapse a subagent (hide content but keep it loaded).
     * @param {string} agentId - The agent ID to collapse
     */
    function collapseSubagent(agentId) {
        var overlay = document.querySelector('[data-agent-id="' + agentId + '"]');
        if (!overlay) return;

        var content = overlay.querySelector('.subagent-content');
        if (content) {
            content.classList.add('collapsed');
        }
        overlay.classList.remove('expanded');

        state.expandedAgents[agentId] = false;
        saveNavigationState();

        // Remove from breadcrumbs if present
        removeAgentFromBreadcrumbs(agentId);
    }

    /**
     * Toggle a subagent between expanded and collapsed.
     * @param {string} agentId - The agent ID to toggle
     */
    function toggleSubagent(agentId) {
        var overlay = document.querySelector('[data-agent-id="' + agentId + '"]');
        if (!overlay) return;

        if (overlay.classList.contains('expanded')) {
            collapseSubagent(agentId);
        } else {
            expandSubagent(agentId);
        }
    }

    /**
     * Add an agent to the breadcrumb path.
     * @param {string} agentId - The agent ID
     * @param {HTMLElement} overlay - The agent overlay element
     */
    function addAgentToBreadcrumbs(agentId, overlay) {
        // Check if already in path
        for (var i = 0; i < state.currentPath.length; i++) {
            if (state.currentPath[i].id === agentId) {
                return; // Already in path
            }
        }

        // Get agent label from overlay
        var title = overlay.querySelector('.subagent-title');
        var label = title ? title.textContent.replace('Subagent: ', '') : agentId.substring(0, 7);

        state.currentPath.push({ id: agentId, label: label });
        updateBreadcrumbs();
        addToHistory(state.currentPath);
    }

    /**
     * Remove an agent and all children from breadcrumbs.
     * @param {string} agentId - The agent ID to remove
     */
    function removeAgentFromBreadcrumbs(agentId) {
        var index = -1;
        for (var i = 0; i < state.currentPath.length; i++) {
            if (state.currentPath[i].id === agentId) {
                index = i;
                break;
            }
        }

        if (index > 0) {
            state.currentPath = state.currentPath.slice(0, index);
            updateBreadcrumbs();
        }
    }

    // ===========================================
    // SCROLL NAVIGATION
    // ===========================================

    /**
     * Smoothly scroll to an agent element.
     * @param {string} agentId - The agent ID to scroll to
     */
    function scrollToAgent(agentId) {
        var element;

        if (agentId === MAIN_SESSION_ID) {
            element = document.querySelector('.conversation');
        } else {
            element = document.querySelector('[data-agent-id="' + agentId + '"]');
        }

        if (!element) return;

        var headerHeight = 100; // Account for sticky header and breadcrumbs
        var elementRect = element.getBoundingClientRect();
        var absoluteElementTop = elementRect.top + window.pageYOffset;
        var targetPosition = absoluteElementTop - headerHeight;

        window.scrollTo({
            top: targetPosition,
            behavior: 'smooth'
        });

        // Add highlight effect
        element.classList.add('navigation-highlight');
        setTimeout(function() {
            element.classList.remove('navigation-highlight');
        }, 2000);
    }

    /**
     * Jump to parent agent from current nested context.
     * @param {string} currentAgentId - The current agent ID
     */
    function jumpToParent(currentAgentId) {
        // Find current agent in path
        var index = -1;
        for (var i = 0; i < state.currentPath.length; i++) {
            if (state.currentPath[i].id === currentAgentId) {
                index = i;
                break;
            }
        }

        if (index > 0) {
            navigateToBreadcrumb(index - 1);
        } else {
            // Default to main session
            scrollToAgent(MAIN_SESSION_ID);
        }
    }

    // ===========================================
    // NESTED COMPONENT INITIALIZATION
    // ===========================================

    /**
     * Initialize nested components within loaded agent content.
     * @param {HTMLElement} container - The container with new content
     */
    function initNestedComponents(container) {
        // Initialize tool overlays
        if (typeof window.initToolToggles === 'function') {
            window.initToolToggles(container);
        }

        // Initialize copy buttons
        if (typeof window.initCopyButtons === 'function') {
            window.initCopyButtons(container);
        }

        // Initialize nested subagent headers
        initNestedSubagentHeaders(container);

        // Initialize Jump to Parent buttons
        initJumpToParentButtons(container);
    }

    /**
     * Initialize subagent headers within nested content.
     * @param {HTMLElement} container - The container element
     */
    function initNestedSubagentHeaders(container) {
        var headers = container.querySelectorAll('.subagent-header');
        headers.forEach(function(header) {
            if (header.dataset.navInitialized) return;
            header.dataset.navInitialized = 'true';

            header.addEventListener('click', function(e) {
                // Check if click was on a button
                if (e.target.closest('.deep-dive-btn') || e.target.closest('.copy-btn')) {
                    return;
                }

                var overlay = header.closest('.subagent-overlay, .subagent');
                if (overlay) {
                    var agentId = overlay.dataset.agentId;
                    toggleSubagent(agentId);
                }
            });
        });

        // Initialize Deep Dive buttons
        var deepDiveBtns = container.querySelectorAll('.deep-dive-btn');
        deepDiveBtns.forEach(function(btn) {
            if (btn.dataset.navInitialized) return;
            btn.dataset.navInitialized = 'true';

            btn.addEventListener('click', function(e) {
                e.stopPropagation();
                var overlay = btn.closest('.subagent-overlay, .subagent');
                if (overlay) {
                    var agentId = overlay.dataset.agentId;
                    expandSubagent(agentId, { scrollTo: true, updateBreadcrumbs: true });
                }
            });
        });
    }

    /**
     * Initialize Jump to Parent buttons.
     * @param {HTMLElement} container - The container element
     */
    function initJumpToParentButtons(container) {
        var jumpBtns = container.querySelectorAll('.jump-to-parent-btn');
        jumpBtns.forEach(function(btn) {
            if (btn.dataset.navInitialized) return;
            btn.dataset.navInitialized = 'true';

            btn.addEventListener('click', function(e) {
                e.stopPropagation();
                var agentId = btn.dataset.agentId || btn.closest('[data-agent-id]')?.dataset.agentId;
                if (agentId) {
                    jumpToParent(agentId);
                }
            });
        });
    }

    // ===========================================
    // UTILITY FUNCTIONS
    // ===========================================

    /**
     * Escape HTML special characters.
     * @param {string} str - String to escape
     * @returns {string} Escaped string
     */
    function escapeHtml(str) {
        var div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }

    /**
     * Get the current navigation path.
     * @returns {Array} The current breadcrumb path
     */
    function getCurrentPath() {
        return state.currentPath.slice();
    }

    /**
     * Check if an agent is currently expanded.
     * @param {string} agentId - The agent ID to check
     * @returns {boolean} True if expanded
     */
    function isAgentExpanded(agentId) {
        return !!state.expandedAgents[agentId];
    }

    /**
     * Check if an agent content is loaded.
     * @param {string} agentId - The agent ID to check
     * @returns {boolean} True if loaded
     */
    function isAgentLoaded(agentId) {
        return !!state.loadedAgents[agentId];
    }

    // ===========================================
    // KEYBOARD NAVIGATION
    // ===========================================

    /**
     * Initialize keyboard shortcuts for navigation.
     */
    function initKeyboardNavigation() {
        document.addEventListener('keydown', function(e) {
            // Alt+Left - Navigate back
            if (e.altKey && e.key === 'ArrowLeft') {
                e.preventDefault();
                navigateBack();
                return;
            }

            // Alt+Right - Navigate forward
            if (e.altKey && e.key === 'ArrowRight') {
                e.preventDefault();
                navigateForward();
                return;
            }

            // Alt+Up - Jump to parent
            if (e.altKey && e.key === 'ArrowUp') {
                e.preventDefault();
                var lastItem = state.currentPath[state.currentPath.length - 1];
                if (lastItem && lastItem.id !== MAIN_SESSION_ID) {
                    jumpToParent(lastItem.id);
                }
                return;
            }

            // Escape - Go to main session
            if (e.key === 'Escape' && !e.target.matches('input, textarea')) {
                var lastItem = state.currentPath[state.currentPath.length - 1];
                if (lastItem && lastItem.id !== MAIN_SESSION_ID) {
                    navigateToBreadcrumb(0);
                }
                return;
            }
        });
    }

    // ===========================================
    // INITIALIZATION
    // ===========================================

    /**
     * Initialize the navigation system.
     */
    function initNavigation() {
        // Load saved state
        loadNavigationState();

        // Initialize breadcrumbs
        updateBreadcrumbs(state.currentPath);

        // Initialize history with current path
        addToHistory(state.currentPath);

        // Initialize all subagent headers on page
        initNestedSubagentHeaders(document);

        // Initialize keyboard navigation
        initKeyboardNavigation();

        // Restore expanded agents
        Object.keys(state.expandedAgents).forEach(function(agentId) {
            if (state.expandedAgents[agentId] && state.loadedAgents[agentId]) {
                var overlay = document.querySelector('[data-agent-id="' + agentId + '"]');
                if (overlay) {
                    overlay.classList.add('expanded');
                    var content = overlay.querySelector('.subagent-content');
                    if (content) {
                        content.classList.remove('collapsed');
                    }
                }
            }
        });
    }

    // ===========================================
    // PUBLIC API
    // ===========================================

    // Expose functions globally for use by other scripts and HTML onclick handlers
    window.NavigationAPI = {
        expandSubagent: expandSubagent,
        collapseSubagent: collapseSubagent,
        toggleSubagent: toggleSubagent,
        updateBreadcrumbs: updateBreadcrumbs,
        scrollToAgent: scrollToAgent,
        jumpToParent: jumpToParent,
        navigateBack: navigateBack,
        navigateForward: navigateForward,
        getCurrentPath: getCurrentPath,
        isAgentExpanded: isAgentExpanded,
        isAgentLoaded: isAgentLoaded,
        initNestedComponents: initNestedComponents
    };

    // Legacy function support for existing onclick handlers
    window.deepDiveAgent = function(agentId, event) {
        if (event) {
            event.stopPropagation();
        }
        expandSubagent(agentId, { scrollTo: true, updateBreadcrumbs: true });
    };

    // Initialize when DOM is ready
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', initNavigation);
    } else {
        initNavigation();
    }

})();
