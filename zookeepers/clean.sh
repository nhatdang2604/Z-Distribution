#!/bin/bash

# Define the patterns to keep
patterns=("*.gitkeep" "*.gitignore" "*.sh" "*.yml")

# Safety prompt
read -p "Are you sure you want to delete all files except for the specified patterns? (yes/no): " confirmation

if [[ "$confirmation" != "yes" ]]; then
    echo "Operation cancelled."
    exit 1
fi

# Create the exclude condition for find
exclude=""
for pattern in "${patterns[@]}"; do
    exclude+=" -not -name '$pattern'"
done

# Use find to delete files while excluding the specified patterns
eval find . -type f $exclude -exec rm -f {} +

# Optionally, remove empty directories if needed
find . -type d -empty -delete

echo "Cleanup completed successfully."
