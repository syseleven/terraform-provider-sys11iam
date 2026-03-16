#!/usr/bin/env bash
# post-generate.sh — Run after tfplugingen-framework to clean up generated code.
# Called automatically by `make tf-generate`.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "Running post-generation cleanup..."

# Fix regex patterns: the generator outputs \\u0000 in Go string literals,
# but RE2 doesn't support \u escapes. Replace with \u0000 (Go Unicode escape
# for U+0000 null byte) which RE2 handles correctly.
echo "  Fixing regex patterns..."
find "${REPO_ROOT}/internal/resource_"*/ -name '*_gen.go' -exec \
  sed -i 's/\\\\u0000/\\u0000/g' {} +

# Format generated Go files
echo "  Formatting generated code..."
goimports -w "${REPO_ROOT}/internal/resource_"*/ 2>/dev/null || true
go fmt "${REPO_ROOT}/internal/resource_"*/... 2>/dev/null || true

echo "Post-generation cleanup complete."
