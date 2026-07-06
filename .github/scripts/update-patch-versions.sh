#!/usr/bin/env bash
set -euo pipefail

# UPDATE_MODE controls version selection strategy:
#   "patch"  (default) — only bump patch within the same major.minor
#   "latest" — bump to the absolute latest stable release
UPDATE_MODE="${UPDATE_MODE:-patch}"

# Component definitions: "name|go_module|github_repo"
COMPONENTS=(
  "pac|github.com/openshift-pipelines/pipelines-as-code|tektoncd/pipelines-as-code"
  "tkn|github.com/tektoncd/cli|tektoncd/cli"
  "results|github.com/tektoncd/results|tektoncd/results"
  "manualapprovalgate|github.com/openshift-pipelines/manual-approval-gate|openshift-pipelines/manual-approval-gate"
)

UPDATES=()
GOMOD_CHANGED=false

for entry in "${COMPONENTS[@]}"; do
  IFS='|' read -r component module repo <<< "$entry"

  module_re="${module//./\\.}"
  current_version=$(grep -E "^\s+${module_re} v" go.mod | head -1 | awk '{print $2}' | sed 's/^v//') || true

  if [[ -z "$current_version" ]]; then
    echo "Warning: ${module} not found in go.mod, skipping ${component}"
    continue
  fi

  echo "Checking ${component} (${module}): current v${current_version} [mode=${UPDATE_MODE}]"

  all_tags=$(gh api "repos/${repo}/tags" --paginate --jq '.[].name' 2>/dev/null) || true

  if [[ "$UPDATE_MODE" == "latest" ]]; then
    target_version=$(echo "$all_tags" \
      | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' \
      | sed 's/^v//' \
      | sort -V \
      | tail -1) || true
  else
    major_minor="${current_version%.*}"
    major_minor_re="${major_minor//./\\.}"
    target_version=$(echo "$all_tags" \
      | grep -E "^v${major_minor_re}\.[0-9]+$" \
      | sed 's/^v//' \
      | sort -V \
      | tail -1) || true
  fi

  if [[ -z "$target_version" ]]; then
    echo "  No matching tags found, skipping"
    continue
  fi

  echo "  Target: v${target_version}"

  if [[ "$target_version" != "$current_version" ]]; then
    echo "  Update available: v${current_version} -> v${target_version}"
    if go get "${module}@v${target_version}"; then
      UPDATES+=("${component}: v${current_version} -> v${target_version}")
      GOMOD_CHANGED=true
    else
      echo "  ERROR: go get failed for ${component}, skipping"
    fi
  else
    echo "  Already up to date"
  fi
done

if [[ ${#UPDATES[@]} -eq 0 ]]; then
  echo ""
  echo "No updates found."
  echo "has_updates=false" >> "$GITHUB_OUTPUT"
  exit 0
fi

if [[ "$GOMOD_CHANGED" == "true" ]]; then
  echo ""
  echo "Running go mod tidy..."
  go mod tidy

  echo "Running go mod vendor..."
  go mod vendor
fi

echo ""
echo "Regenerating version.json..."
make generate

echo ""
echo "Updates applied:"
for update in "${UPDATES[@]}"; do
  echo "  - ${update}"
done

echo "has_updates=true" >> "$GITHUB_OUTPUT"
{
  echo "summary<<EOF"
  for update in "${UPDATES[@]}"; do
    echo "- ${update}"
  done
  echo "EOF"
} >> "$GITHUB_OUTPUT"
