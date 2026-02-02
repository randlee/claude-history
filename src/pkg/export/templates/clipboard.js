/**
 * Claude History Export - Clipboard Functionality
 * Provides copy-to-clipboard support for agent IDs, file paths, session IDs, and tool IDs.
 */

/**
 * Copy text to clipboard and show visual feedback.
 * Uses modern navigator.clipboard API with fallback for older browsers.
 * @param {string} text - The text to copy to clipboard
 * @param {HTMLElement} button - The button element that triggered the copy
 */
function copyToClipboard(text, button) {
    if (!text) {
        showCopyError(button, 'Nothing to copy');
        return;
    }

    // Try modern clipboard API first
    if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text)
            .then(function() {
                showCopySuccess(button);
            })
            .catch(function(err) {
                // Fall back to legacy method
                copyToClipboardLegacy(text, button);
            });
    } else {
        // Use legacy fallback for older browsers
        copyToClipboardLegacy(text, button);
    }
}

/**
 * Legacy clipboard copy using execCommand.
 * Used as fallback for browsers without navigator.clipboard support.
 * @param {string} text - The text to copy
 * @param {HTMLElement} button - The button element
 */
function copyToClipboardLegacy(text, button) {
    var textArea = document.createElement('textarea');
    textArea.value = text;

    // Make textarea invisible but keep it in DOM
    textArea.style.position = 'fixed';
    textArea.style.top = '0';
    textArea.style.left = '0';
    textArea.style.width = '2em';
    textArea.style.height = '2em';
    textArea.style.padding = '0';
    textArea.style.border = 'none';
    textArea.style.outline = 'none';
    textArea.style.boxShadow = 'none';
    textArea.style.background = 'transparent';
    textArea.style.opacity = '0';

    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    try {
        var successful = document.execCommand('copy');
        if (successful) {
            showCopySuccess(button);
        } else {
            showCopyError(button, 'Copy failed');
        }
    } catch (err) {
        showCopyError(button, 'Copy not supported');
    }

    document.body.removeChild(textArea);
}

/**
 * Show success feedback on the copy button.
 * @param {HTMLElement} button - The button element
 */
function showCopySuccess(button) {
    var originalContent = button.innerHTML;
    var originalTitle = button.title;

    // Add success class and change icon
    button.classList.add('copy-success');
    button.innerHTML = '<span class="copy-icon">&#10003;</span>';
    button.title = 'Copied!';

    // Show toast notification
    showCopyToast('Copied to clipboard');

    // Restore original state after animation
    setTimeout(function() {
        button.classList.remove('copy-success');
        button.innerHTML = originalContent;
        button.title = originalTitle;
    }, 1500);
}

/**
 * Show error feedback on the copy button.
 * @param {HTMLElement} button - The button element
 * @param {string} message - The error message
 */
function showCopyError(button, message) {
    button.classList.add('copy-error');
    showCopyToast(message, true);

    setTimeout(function() {
        button.classList.remove('copy-error');
    }, 1500);
}

/**
 * Show a toast notification.
 * @param {string} message - The message to display
 * @param {boolean} isError - Whether this is an error toast
 */
function showCopyToast(message, isError) {
    // Remove any existing toast
    var existingToast = document.querySelector('.copy-toast');
    if (existingToast) {
        existingToast.remove();
    }

    // Create new toast
    var toast = document.createElement('div');
    toast.className = 'copy-toast' + (isError ? ' copy-toast-error' : '');
    toast.textContent = message;

    document.body.appendChild(toast);

    // Trigger animation
    setTimeout(function() {
        toast.classList.add('copy-toast-visible');
    }, 10);

    // Remove toast after delay
    setTimeout(function() {
        toast.classList.remove('copy-toast-visible');
        setTimeout(function() {
            if (toast.parentNode) {
                toast.remove();
            }
        }, 300);
    }, 2000);
}

/**
 * Handle click on a copy button.
 * Extracts the text from data-copy-text attribute and copies it.
 * @param {Event} event - The click event
 */
function handleCopyClick(event) {
    var button = event.currentTarget;
    var text = button.getAttribute('data-copy-text');
    copyToClipboard(text, button);

    // Prevent event from bubbling (e.g., to tool-header toggle)
    event.stopPropagation();
}

/**
 * Initialize all copy buttons in a container.
 * @param {HTMLElement} container - The container element (defaults to document)
 */
function initCopyButtons(container) {
    container = container || document;
    var buttons = container.querySelectorAll('.copy-btn');

    buttons.forEach(function(button) {
        // Remove any existing handler to prevent duplicates
        button.removeEventListener('click', handleCopyClick);
        // Add click handler
        button.addEventListener('click', handleCopyClick);
    });
}

// Initialize copy buttons when DOM is ready
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function() {
        initCopyButtons(document);
    });
} else {
    initCopyButtons(document);
}
