#!/usr/bin/env bash

set -euo pipefail

chart_dir="${CHART_DIR:-deploy/charts/helloworld}"
chart_file="${chart_dir}/Chart.yaml"
base_ref="${1:-origin/main}"

if [ ! -f "$chart_file" ]; then
  echo "Chart file not found: $chart_file" >&2
  exit 1
fi

if ! git cat-file -e "${base_ref}^{commit}" 2>/dev/null; then
  echo "Base ref does not exist: $base_ref" >&2
  exit 1
fi

if ! git cat-file -e "${base_ref}:${chart_file}" 2>/dev/null; then
  echo "No chart exists at ${base_ref}:${chart_file}; treating this as a new chart."
  exit 0
fi

compare_ref="$(git merge-base "$base_ref" HEAD)"

mapfile -t changed_files < <(git diff --name-only "$compare_ref" -- "$chart_dir")

if [ "${#changed_files[@]}" -eq 0 ]; then
  echo "No changes detected under ${chart_dir}."
  exit 0
fi

base_version="$(
  git show "${base_ref}:${chart_file}" | awk '$1 == "version:" { print $2; exit }'
)"
current_version="$(
  awk '$1 == "version:" { print $2; exit }' "$chart_file"
)"

if [ -z "$base_version" ] || [ -z "$current_version" ]; then
  echo "Unable to determine chart versions for comparison." >&2
  exit 1
fi

if [ "$current_version" = "$base_version" ]; then
  echo "Chart files changed, but ${chart_file} version stayed at ${current_version}." >&2
  printf 'Changed files:\n' >&2
  printf ' - %s\n' "${changed_files[@]}" >&2
  exit 1
fi

highest_version="$(printf '%s\n%s\n' "$base_version" "$current_version" | sort -V | tail -n 1)"

if [ "$highest_version" != "$current_version" ]; then
  echo "Chart version must increase. Base: ${base_version}, current: ${current_version}." >&2
  exit 1
fi

echo "Chart version increased: ${base_version} -> ${current_version}"
printf 'Changed files:\n'
printf ' - %s\n' "${changed_files[@]}"
