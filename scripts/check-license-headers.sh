# Copyright (c) Hintents Authors.
# SPDX-License-Identifier: Apache-2.0

#!/bin/sh
# Check for license headers in Go and Rust files
# Exit with status 1 if any files are missing headers
set -e

EXPECTED_HEADER="Copyright 2025 Erst Users"
FAIL_FILE=$(mktemp)
echo "0" > "$FAIL_FILE"

echo "Checking for license headers in Go and Rust files..."

check_file() {
    file="$1"
    if ! head -1 "$file" | grep -q "$EXPECTED_HEADER"; then
        echo "  [FAIL] Missing license header: $file"
        count=$(cat "$FAIL_FILE")
        echo "$((count + 1))" > "$FAIL_FILE"
    else
        echo "  [OK] $file"
    fi
}

# Check Go files
echo ""
echo "Checking Go files (.go)..."
find . -type d \( -name "target" -o -name "vendor" \) -prune -o -name "*.go" -type f -print | while IFS= read -r file; do
    check_file "$file"
done

# Check Rust files
echo ""
echo "Checking Rust files (.rs)..."
find . -type d \( -name "target" -o -name "vendor" \) -prune -o -name "*.rs" -type f -print | while IFS= read -r file; do
    check_file "$file"
done

echo ""
MISSING_HEADERS=$(cat "$FAIL_FILE")
rm -f "$FAIL_FILE"

if [ "$MISSING_HEADERS" -eq 0 ]; then
    echo "[OK] All files have proper license headers"
    exit 0
else
    echo "[FAIL] Found $MISSING_HEADERS file(s) missing license headers"
    exit 1
fi
