#!/bin/bash
# Script to detect empty message bubbles in HTML export
# Exit with non-zero if any empty bubbles are found

set -e

if [ $# -lt 1 ]; then
    echo "Usage: $0 <html-file>"
    exit 1
fi

HTML_FILE="$1"

if [ ! -f "$HTML_FILE" ]; then
    echo "Error: File not found: $HTML_FILE"
    exit 1
fi

# Look for empty message-content divs
# Pattern 1: <div class="message-content"></div>
# Pattern 2: <div class="message-content">\s*</div>
# Pattern 3: <div class="message-content"><!-- only whitespace or comments --></div>

echo "Analyzing $HTML_FILE for empty message bubbles..."
echo ""

# Use grep to find message-content divs and check if they're empty
EMPTY_COUNT=0
LINE_NUM=0

# Extract all message-row blocks with their UUIDs
while IFS= read -r line; do
    LINE_NUM=$((LINE_NUM + 1))

    # Check for message-row start with UUID
    if echo "$line" | grep -q 'class="message-row'; then
        UUID=$(echo "$line" | sed -n 's/.*data-uuid="\([^"]*\)".*/\1/p')

        # Read ahead to find the message-content div
        CONTENT_FOUND=false
        CONTENT_LINE=""
        TEMP_LINE_NUM=$LINE_NUM

        # Read next 20 lines looking for message-content
        for i in {1..20}; do
            TEMP_LINE_NUM=$((TEMP_LINE_NUM + 1))
            NEXT_LINE=$(sed -n "${TEMP_LINE_NUM}p" "$HTML_FILE")

            if echo "$NEXT_LINE" | grep -q 'class="message-content"'; then
                CONTENT_LINE="$NEXT_LINE"
                CONTENT_FOUND=true
                break
            fi
        done

        if [ "$CONTENT_FOUND" = true ]; then
            # Check if message-content is empty or whitespace-only
            # Pattern: <div class="message-content"></div> (on same line)
            # Pattern: <div class="message-content">    </div> (only whitespace)

            if echo "$CONTENT_LINE" | grep -q '<div class="message-content"></div>'; then
                echo "EMPTY BUBBLE FOUND:"
                echo "  UUID: $UUID"
                echo "  Line: $TEMP_LINE_NUM"
                echo "  Content: $CONTENT_LINE"
                echo ""
                EMPTY_COUNT=$((EMPTY_COUNT + 1))
            elif echo "$CONTENT_LINE" | grep -qE '<div class="message-content">\s*</div>'; then
                # Check if there's only whitespace between tags
                INNER=$(echo "$CONTENT_LINE" | sed -n 's/.*<div class="message-content">\(.*\)<\/div>.*/\1/p')
                if [ -z "$(echo "$INNER" | tr -d '[:space:]')" ]; then
                    echo "EMPTY BUBBLE FOUND (whitespace only):"
                    echo "  UUID: $UUID"
                    echo "  Line: $TEMP_LINE_NUM"
                    echo "  Content: $CONTENT_LINE"
                    echo ""
                    EMPTY_COUNT=$((EMPTY_COUNT + 1))
                fi
            fi
        fi
    fi
done < "$HTML_FILE"

echo "=================================="
echo "Total empty bubbles found: $EMPTY_COUNT"
echo "=================================="

if [ $EMPTY_COUNT -gt 0 ]; then
    exit 1
fi

exit 0
