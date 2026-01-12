#!/bin/bash
# CERT Discourse - Theme Installation Helper
# This script packages the theme for easy upload to Discourse

set -e

THEME_DIR="$(dirname "$0")/theme"
OUTPUT_DIR="$(dirname "$0")/theme-package"

echo "=========================================="
echo "CERT Discourse Theme Packager"
echo "=========================================="
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Create theme package structure
echo "Packaging theme files..."

# Copy theme files
cp -r "$THEME_DIR"/* "$OUTPUT_DIR/"

# Create a tarball for easy upload
cd "$OUTPUT_DIR/.."
tar -czf cert-discourse-theme.tar.gz theme-package/

echo ""
echo "✓ Theme packaged successfully!"
echo ""
echo "Theme package location:"
echo "  $OUTPUT_DIR"
echo ""
echo "Tarball for upload:"
echo "  $(dirname "$0")/cert-discourse-theme.tar.gz"
echo ""
echo "=========================================="
echo "Installation Instructions:"
echo "=========================================="
echo ""
echo "1. Login to Discourse as admin"
echo "   https://forum.c3rt.org/admin"
echo ""
echo "2. Navigate to Customize → Themes"
echo "   https://forum.c3rt.org/admin/customize/themes"
echo ""
echo "3. Click 'Install' → 'From a Git Repository'"
echo ""
echo "4. OR manually upload theme files:"
echo "   - about.json"
echo "   - common/common.scss"
echo "   - common/header.html"
echo ""
echo "5. Set as default theme"
echo ""
echo "6. Verify theme colors match cert-web:"
echo "   - Background: #050508 (ink)"
echo "   - Mint: #00FFA3"
echo "   - Electric: #4D9FFF"
echo "   - Cyber: #9D00FF"
echo ""

