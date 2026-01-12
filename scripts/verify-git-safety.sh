#!/bin/bash

# CERT Blockchain - Git Safety Verification Script
# Run this BEFORE committing to ensure no sensitive data is included

set -e

echo "================================================"
echo "  Git Safety Verification"
echo "================================================"
echo ""

ERRORS=0

# Check if .gitignore exists
echo "✓ Checking .gitignore files..."
if [ ! -f ".gitignore" ]; then
    echo "❌ ERROR: .gitignore missing in root!"
    ERRORS=$((ERRORS + 1))
else
    echo "  ✅ Root .gitignore exists"
fi

if [ ! -f "sdk/.gitignore" ]; then
    echo "❌ ERROR: sdk/.gitignore missing!"
    ERRORS=$((ERRORS + 1))
else
    echo "  ✅ SDK .gitignore exists"
fi

echo ""

# Check for sensitive files that should NOT be committed
echo "✓ Checking for sensitive files..."

SENSITIVE_FILES=(
    ".env"
    ".env.api"
    "data/certd/config/priv_validator_key.json"
    "data/certd/config/node_key.json"
    "test-node/config/priv_validator_key.json"
)

for file in "${SENSITIVE_FILES[@]}"; do
    if [ -f "$file" ]; then
        # Check if file would be committed
        if git check-ignore -q "$file" 2>/dev/null; then
            echo "  ✅ $file (ignored)"
        else
            echo "  ❌ ERROR: $file would be committed!"
            ERRORS=$((ERRORS + 1))
        fi
    fi
done

echo ""

# Check for private keys in any location
echo "✓ Scanning for private keys..."
PRIV_KEYS=$(find . -name "priv_validator_key.json" -o -name "node_key.json" 2>/dev/null | grep -v node_modules || true)
if [ -n "$PRIV_KEYS" ]; then
    echo "  Found private key files:"
    echo "$PRIV_KEYS" | while read -r key; do
        if git check-ignore -q "$key" 2>/dev/null; then
            echo "    ✅ $key (ignored)"
        else
            echo "    ❌ ERROR: $key would be committed!"
            ERRORS=$((ERRORS + 1))
        fi
    done
else
    echo "  ✅ No private key files found"
fi

echo ""

# Check for .env files
echo "✓ Scanning for .env files..."
ENV_FILES=$(find . -name ".env" -o -name ".env.local" 2>/dev/null | grep -v node_modules || true)
if [ -n "$ENV_FILES" ]; then
    echo "  Found .env files:"
    echo "$ENV_FILES" | while read -r env; do
        if git check-ignore -q "$env" 2>/dev/null; then
            echo "    ✅ $env (ignored)"
        else
            echo "    ❌ ERROR: $env would be committed!"
            ERRORS=$((ERRORS + 1))
        fi
    done
else
    echo "  ✅ No .env files found (or all ignored)"
fi

echo ""

# Check for large binary files
echo "✓ Checking for large binary files..."
LARGE_FILES=$(find . -type f -size +10M 2>/dev/null | grep -v node_modules | grep -v ".git" || true)
if [ -n "$LARGE_FILES" ]; then
    echo "  ⚠️  WARNING: Large files found (>10MB):"
    echo "$LARGE_FILES" | while read -r large; do
        SIZE=$(du -h "$large" | cut -f1)
        if git check-ignore -q "$large" 2>/dev/null; then
            echo "    ✅ $large ($SIZE) - ignored"
        else
            echo "    ⚠️  $large ($SIZE) - will be committed"
        fi
    done
else
    echo "  ✅ No large files found"
fi

echo ""

# Check for node_modules
echo "✓ Checking for node_modules..."
NODE_MODULES=$(find . -type d -name "node_modules" 2>/dev/null | head -5 || true)
if [ -n "$NODE_MODULES" ]; then
    echo "  Found node_modules directories:"
    echo "$NODE_MODULES" | while read -r nm; do
        if git check-ignore -q "$nm" 2>/dev/null; then
            echo "    ✅ $nm (ignored)"
        else
            echo "    ❌ ERROR: $nm would be committed!"
            ERRORS=$((ERRORS + 1))
        fi
    done
else
    echo "  ✅ No node_modules found"
fi

echo ""

# Check for build artifacts
echo "✓ Checking for build artifacts..."
BUILD_ARTIFACTS=(
    "certd"
    "certd.exe"
    "cert-api"
    "cert-api-linux"
    "data/"
    "build/"
)

for artifact in "${BUILD_ARTIFACTS[@]}"; do
    if [ -e "$artifact" ]; then
        if git check-ignore -q "$artifact" 2>/dev/null; then
            echo "  ✅ $artifact (ignored)"
        else
            echo "  ⚠️  WARNING: $artifact would be committed"
        fi
    fi
done

echo ""
echo "================================================"
echo "  Summary"
echo "================================================"

if [ $ERRORS -eq 0 ]; then
    echo "✅ SAFE TO COMMIT - No sensitive data detected"
    echo ""
    echo "Next steps:"
    echo "  git add ."
    echo "  git commit -m 'Initial commit'"
    echo "  git push"
    exit 0
else
    echo "❌ NOT SAFE TO COMMIT - $ERRORS error(s) found"
    echo ""
    echo "Fix the errors above before committing!"
    echo ""
    echo "Common fixes:"
    echo "  1. Ensure .gitignore files are in place"
    echo "  2. Remove sensitive files from staging: git rm --cached <file>"
    echo "  3. Add patterns to .gitignore"
    exit 1
fi

