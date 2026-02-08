#!/bin/bash
# Install git hooks
#
# This script copies the pre-commit hook from .githooks/ to .git/hooks/
# and makes it executable.

set -e

HOOKS_DIR=".githooks"
GIT_HOOKS_DIR=".git/hooks"

echo "Installing git hooks..."

if [ ! -d "$GIT_HOOKS_DIR" ]; then
    echo "Error: .git/hooks directory not found"
    echo "Make sure you're running this script from the repository root"
    exit 1
fi

cp "$HOOKS_DIR/pre-commit" "$GIT_HOOKS_DIR/pre-commit"
chmod +x "$GIT_HOOKS_DIR/pre-commit"

echo "✓ Pre-commit hook installed successfully"
echo ""
echo "The following checks will run before each commit:"
echo "  - Linting (make lint)"
echo "  - Tests (make test)"
echo ""
echo "To bypass these checks (not recommended), use:"
echo "  git commit --no-verify"
