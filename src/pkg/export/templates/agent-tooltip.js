/**
 * Agent Tooltip Module
 * Provides interactive agent statistics display with click-to-copy functionality
 */
(function() {
    'use strict';

    var tooltip = null;
    var currentTarget = null;

    // Initialize on page load
    document.addEventListener('DOMContentLoaded', function() {
        initializeAgentTooltip();
    });

    function initializeAgentTooltip() {
        var agentStats = document.querySelector('.agent-stats-interactive');
        if (!agentStats) return;

        // Create tooltip element
        tooltip = createTooltipElement();
        document.body.appendChild(tooltip);

        // Add event listeners
        agentStats.addEventListener('mouseenter', handleMouseEnter);
        agentStats.addEventListener('mouseleave', handleMouseLeave);
        agentStats.addEventListener('click', handleClick);

        // Hide tooltip when clicking outside
        document.addEventListener('click', function(e) {
            if (!agentStats.contains(e.target) && !tooltip.contains(e.target)) {
                hideTooltip();
            }
        });

        // Keep tooltip visible when hovering over it
        tooltip.addEventListener('mouseenter', function() {
            tooltip.classList.add('visible');
        });

        tooltip.addEventListener('mouseleave', function() {
            hideTooltip();
        });
    }

    function createTooltipElement() {
        var el = document.createElement('div');
        el.className = 'agent-tooltip';
        return el;
    }

    function handleMouseEnter(e) {
        currentTarget = e.target;
        showTooltip(e.target);
    }

    function handleMouseLeave(e) {
        // Don't hide immediately - wait to see if mouse moves to tooltip
        setTimeout(function() {
            if (!tooltip.matches(':hover') && !currentTarget.matches(':hover')) {
                hideTooltip();
            }
        }, 100);
    }

    function handleClick(e) {
        e.stopPropagation();
        copyToClipboard(e.target);
    }

    function showTooltip(target) {
        var agentDetailsStr = target.dataset.agentDetails || '{}';
        var agentDetails = {};

        try {
            agentDetails = JSON.parse(agentDetailsStr);
        } catch (err) {
            console.error('Failed to parse agent details:', err);
            return;
        }

        var sessionId = target.dataset.sessionId || '';

        // Build tooltip content
        var html = '<div class="agent-tooltip-header">Agent Message Counts</div>';
        html += '<div class="agent-tooltip-table">';

        // Sort by message count (descending)
        var agents = Object.entries(agentDetails).sort(function(a, b) {
            return b[1] - a[1];
        });

        if (agents.length === 0) {
            html += '<div class="agent-tooltip-empty">No agent data available</div>';
        } else {
            agents.forEach(function(entry) {
                var agentId = entry[0];
                var count = entry[1];
                html += '<div class="agent-tooltip-id">' + escapeHtml(agentId) + '</div>';
                html += '<div class="agent-tooltip-count">' + count + '</div>';
            });
        }

        html += '</div>';
        html += '<div class="agent-tooltip-footer">Click to copy to clipboard</div>';

        tooltip.innerHTML = html;

        // Position tooltip
        positionTooltip(target);

        tooltip.classList.add('visible');
    }

    function positionTooltip(target) {
        var rect = target.getBoundingClientRect();
        var tooltipHeight = 400; // max-height from CSS
        var windowHeight = window.innerHeight;
        var spaceBelow = windowHeight - rect.bottom;
        var spaceAbove = rect.top;

        // Default: show below
        var top = rect.bottom + 8;

        // If not enough space below but more space above, show above
        if (spaceBelow < tooltipHeight && spaceAbove > spaceBelow) {
            // Position above the target
            tooltip.style.bottom = (windowHeight - rect.top + 8) + 'px';
            tooltip.style.top = 'auto';
        } else {
            // Position below the target
            tooltip.style.top = top + 'px';
            tooltip.style.bottom = 'auto';
        }

        // Horizontal positioning (align with target, adjust if needed)
        var left = rect.left;
        var tooltipWidth = 350; // min-width from CSS
        var windowWidth = window.innerWidth;

        // Adjust if tooltip would overflow right edge
        if (left + tooltipWidth > windowWidth) {
            left = windowWidth - tooltipWidth - 20;
        }

        // Ensure tooltip doesn't go off left edge
        if (left < 20) {
            left = 20;
        }

        tooltip.style.left = left + 'px';
    }

    function hideTooltip() {
        if (tooltip) {
            tooltip.classList.remove('visible');
        }
    }

    function copyToClipboard(target) {
        var agentDetailsStr = target.dataset.agentDetails || '{}';
        var agentDetails = {};

        try {
            agentDetails = JSON.parse(agentDetailsStr);
        } catch (err) {
            console.error('Failed to parse agent details:', err);
            return;
        }

        var sessionId = target.dataset.sessionId || '';

        // Build clipboard text
        var text = 'session: ' + sessionId + '\n---\n';

        // Sort by message count (descending)
        var agents = Object.entries(agentDetails).sort(function(a, b) {
            return b[1] - a[1];
        });

        agents.forEach(function(entry) {
            var agentId = entry[0];
            var count = entry[1];
            text += agentId + '\t' + count + '\n';
        });

        // Copy to clipboard
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(text).then(function() {
                showCopyFeedback();
            }).catch(function(err) {
                console.error('Failed to copy:', err);
                fallbackCopy(text);
            });
        } else {
            fallbackCopy(text);
        }
    }

    function fallbackCopy(text) {
        // Fallback for older browsers
        var textarea = document.createElement('textarea');
        textarea.value = text;
        textarea.style.position = 'fixed';
        textarea.style.opacity = '0';
        document.body.appendChild(textarea);
        textarea.select();

        try {
            document.execCommand('copy');
            showCopyFeedback();
        } catch (err) {
            console.error('Fallback copy failed:', err);
        }

        document.body.removeChild(textarea);
    }

    function showCopyFeedback() {
        var feedback = document.createElement('div');
        feedback.className = 'copy-feedback';
        feedback.textContent = '✓ Copied to clipboard';
        document.body.appendChild(feedback);

        setTimeout(function() {
            feedback.style.opacity = '0';
            setTimeout(function() {
                feedback.remove();
            }, 300);
        }, 2000);
    }

    function escapeHtml(text) {
        var div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
})();
