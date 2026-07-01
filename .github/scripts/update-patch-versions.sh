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

for entry in "${COMPONENTS[@]}"; do
  IFS='|' read -r component module repo <<< "$entry"

  current_version=$(grep -E "^\\s+${module} v" go.mod | head -1 | awk '{print $2}' | sed 's/^v//') || true

  if [[ -z "$current_version" ]]; then
    echo "Warning: ${module} not found in go.mod, skipping ${component}"
    continue
  fi

  echo "Checking ${component} (${module}): current v${current_version} [mode=${UPDATE_MODE}]"

  all_tags=$(gh api "repos/${repo}/tags" --paginate --jq '.[].name' 2>/dev/null) || true

  if [[ "$UPDATE_MODE" == "latest" ]]; then
    # Pick the highest stable release tag (vX.Y.Z, no pre-release suffix)
    target_version=$(echo "$all_tags" \
      | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' \
      | sed 's/^v//' \
      | sort -V \
      | tail -1) || true
  else
    # Pick the highest patch within the same major.minor
    major_minor="${current_version%.*}"
    target_version=$(echo "$all_tags" \
      | grep -E "^v${major_minor}\.[0-9]+$" \
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
    else
      echo "  ERROR: go get failed for ${component}, skipping"
    fi
  else
    echo "  Already up to date"
  fi
done

TARGET_BRANCH="${TARGET_BRANCH:-}"
OPC_REPO="openshift-pipelines/opc"

if [[ -n "$TARGET_BRANCH" ]]; then
  current_opc=$(grep -E '^OPC_VERSION\s*:=' Makefile | sed 's/.*:= *//')

  echo ""
  echo "Checking opc version: current ${current_opc} [branch=${TARGET_BRANCH}]"

  opc_tags=$(gh api "repos/${OPC_REPO}/tags" --paginate --jq '.[].name' 2>/dev/null) || true

  if [[ "$TARGET_BRANCH" =~ ^release-v([0-9]+\.[0-9]+)\.x$ ]]; then
    branch_minor="${BASH_REMATCH[1]}"
    target_opc=$(echo "$opc_tags" \
      | grep -E "^v${branch_minor}\.[0-9]+$" \
      | sed 's/^v//' \
      | sort -V \
      | tail -1) || true
  elif [[ "$TARGET_BRANCH" == "main" ]]; then
    target_opc=$(echo "$opc_tags" \
      | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' \
      | sed 's/^v//' \
      | sort -V \
      | tail -1) || true
  fi

  if [[ -n "${target_opc:-}" && "$target_opc" != "$current_opc" ]]; then
    echo "  Update available: ${current_opc} -> ${target_opc}"
    sed -i "s/^OPC_VERSION := .*/OPC_VERSION := ${target_opc}/" Makefile
    UPDATES+=("opc: v${current_opc} -> v${target_opc}")
  else
    echo "  Already up to date"
  fi
fi

if [[ ${#UPDATES[@]} -eq 0 ]]; then
  echo ""
  echo "No updates found."
  echo "has_updates=false" >> "$GITHUB_OUTPUT"
  exit 0
fi

echo ""
echo "Running go mod tidy..."
go mod tidy

echo "Running go mod vendor..."
go mod vendor

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
