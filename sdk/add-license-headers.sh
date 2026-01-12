#!/bin/bash

# Script to add Apache 2.0 license headers to all source files
# Copyright 2026 Brandon Guynn

HEADER='/*
 * Copyright 2026 Brandon Guynn
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

'

# Find all TypeScript files in src directory (excluding node_modules and dist)
find src -name "*.ts" -type f ! -path "*/node_modules/*" ! -path "*/dist/*" | while read file; do
    # Check if file already has Apache license header
    if ! grep -q "Apache License, Version 2.0" "$file"; then
        echo "Adding license header to: $file"
        # Create temp file with header + original content
        echo "$HEADER" | cat - "$file" > temp && mv temp "$file"
    else
        echo "Skipping (already has header): $file"
    fi
done

echo "✅ License headers added to all source files"

