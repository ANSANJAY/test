#!/bin/bash

# Input and output file names
input_file="repos.csv"
output_file="repos_with_errors.csv"
owner="amex-eng/"

# Clear output file if it exists, or create it with headers
echo "Repository,Error" > "$output_file"

# Function to check if a repository contains a pom.xml anywhere and verify jacoco-maven-plugin
check_jacoco_in_repo() {
  local repo_name="$1"
  local full_repo_name="$owner$repo_name"

  # Fetch the list of all files in the repository
  file_list=$(gh api "repos/$full_repo_name/git/trees/main?recursive=1" --jq '.tree[].path' 2>/dev/null)

  if [[ -z "$file_list" ]]; then
    echo "Failed to fetch file list for $full_repo_name" >&2
    echo "$repo_name,Failed to fetch file list" >> "$output_file"
    return
  fi

  # Search for any pom.xml files in the repository
  pom_paths=$(echo "$file_list" | grep "pom.xml")

  if [[ -z "$pom_paths" ]]; then
    echo "No pom.xml found in $full_repo_name"
    echo "$repo_name,No pom.xml found" >> "$output_file"
    return
  fi

  # Iterate over all found pom.xml files
  for pom_path in $pom_paths; do
    # Fetch the pom.xml content using the GitHub CLI
    pom_content=$(gh api "repos/$full_repo_name/contents/$pom_path" --jq '.content' 2>/dev/null)

    # Check if we successfully fetched the content
    if [[ -z "$pom_content" ]]; then
      echo "Failed to fetch $pom_path for $full_repo_name" >&2
      echo "$repo_name,Failed to fetch $pom_path" >> "$output_file"
      continue
    fi

    # Decode the base64 content
    decoded_pom=$(echo "$pom_content" | base64 --decode)

    # Check for the jacoco-maven-plugin in the pom.xml
    if echo "$decoded_pom" | grep -q "jacoco-maven-plugin"; then
      echo "Jacoco plugin found in $full_repo_name at $pom_path"
      return
    fi
  done

  # If no pom.xml contains Jacoco plugin, add repo to output
  echo "Jacoco plugin NOT found in any pom.xml for $full_repo_name"
  echo "$repo_name,Jacoco plugin NOT found" >> "$output_file"
}

# Read the input CSV file and process each repository
while IFS=',' read -r repo_name; do
  # Check for jacoco plugin in the repository
  check_jacoco_in_repo "$repo_name"
done < "$input_file"

echo "Processing completed. Repos with errors are listed in $output_file."